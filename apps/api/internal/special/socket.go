package special

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

// A WCA Live subscription NEM graphql-ws: Phoenix csatornákon fut
// (absinthe_phoenix 2.0.3). A drótformátum a Phoenix v2 szerializáló tömbje:
//
//	[join_ref, ref, topic, event, payload]
//
// A menet: phx_join a "__absinthe__:control" topicra, majd "doc" push a
// lekérdezéssel, amire a szerver {"subscriptionId": "__absinthe__:doc:..."}
// választ ad. Az adat ezután azon a topicon, "subscription:data" eseményként
// érkezik (absinthe_phoenix Endpoint.publish_subscription).
const (
	absintheControlTopic  = "__absinthe__:control"
	phoenixTopic          = "phoenix"
	eventJoin             = "phx_join"
	eventDoc              = "doc"
	eventReply            = "phx_reply"
	eventHeartbeat        = "heartbeat"
	eventSubscriptionData = "subscription:data"

	// A Phoenix websocket transport alapértelmezett timeoutja 60s, ezért
	// ennél sűrűbben kell életjelet küldeni.
	heartbeatInterval = 25 * time.Second

	reconnectMinDelay = 2 * time.Second
	reconnectMaxDelay = 60 * time.Second

	// Ennél hosszabb kapcsolatot egészségesnek tekintünk, és a következő
	// szakadás után újra a legrövidebb várakozással próbálkozunk.
	healthySessionDuration = time.Minute

	// Amikor nincs mit figyelni, ilyen sűrűn nézzük meg, hogy jött-e feliratkozás.
	idleCheckInterval = 30 * time.Second
)

// SubscriptionEvent egy megérkezett subscription payload. A Key az a kulcs,
// amivel a hívó feliratkozott — így nem kell az Absinthe topicokkal bajlódnia.
type SubscriptionEvent struct {
	Key  string
	Data json.RawMessage // a GraphQL "data" objektum
}

type subscriptionSpec struct {
	query     string
	variables map[string]any
}

// SocketClient EGYETLEN WebSocket kapcsolatot tart fenn a WCA Live felé, és
// azon multiplexeli az összes feliratkozást (a Phoenix csatornák eleve erre
// valók). Automatikusan újracsatlakozik, és minden kapcsolódás után magától
// újraküldi a nyilvántartott feliratkozásokat.
type SocketClient struct {
	url    string
	events chan SubscriptionEvent

	// onConnect minden (újra)csatlakozás UTÁN fut le, a feliratkozások
	// újraküldését követően — itt kell a hívónak teljes állapotot lekérdeznie,
	// hogy a szünet alatt kimaradt eseményeket behozza.
	onConnect func()

	mu      sync.Mutex
	specs   map[string]subscriptionSpec // key -> feliratkozás
	pending chan struct{}               // jelzés: változott a specs

	// A következő mezők csak az aktív session goroutine-ból íródnak.
	conn     *websocket.Conn
	writeMu  sync.Mutex
	ref      int
	topicKey map[string]string // absinthe topic -> key
	sentKeys map[string]bool
}

func NewSocketClient(baseURL string) *SocketClient {
	return &SocketClient{
		url:     socketURL(baseURL),
		events:  make(chan SubscriptionEvent, 64),
		specs:   make(map[string]subscriptionSpec),
		pending: make(chan struct{}, 1),
	}
}

// socketURL a HTTP alap-URL-ből a Phoenix websocket végpontot képzi.
func socketURL(baseURL string) string {
	if baseURL == "" {
		baseURL = DefaultWCALiveURL
	}
	url := strings.TrimSuffix(baseURL, "/")
	url = strings.Replace(url, "https://", "wss://", 1)
	url = strings.Replace(url, "http://", "ws://", 1)
	return url + "/socket/websocket?vsn=2.0.0"
}

func (s *SocketClient) Events() <-chan SubscriptionEvent { return s.events }

func (s *SocketClient) SetOnConnect(fn func()) { s.onConnect = fn }

// Subscribe nyilvántartásba vesz egy feliratkozást. Bármikor hívható: ha épp
// van élő kapcsolat, azonnal kimegy, ha nincs, a következő csatlakozáskor.
// Ugyanazzal a kulccsal kétszer hívva nem keletkezik két feliratkozás.
func (s *SocketClient) Subscribe(key, query string, variables map[string]any) {
	s.mu.Lock()
	if _, exists := s.specs[key]; exists {
		s.mu.Unlock()
		return
	}
	s.specs[key] = subscriptionSpec{query: query, variables: variables}
	s.mu.Unlock()

	s.notify()
}

// Unsubscribe kivesz egy feliratkozást a nyilvántartásból. A meglévő
// kapcsolaton nem mondjuk le külön: a kör lezárása után a WCA Live úgysem küld
// többet, a következő újracsatlakozáskor pedig már nem megy ki újra.
func (s *SocketClient) Unsubscribe(key string) {
	s.mu.Lock()
	delete(s.specs, key)
	s.mu.Unlock()
}

func (s *SocketClient) Subscriptions() []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	keys := make([]string, 0, len(s.specs))
	for key := range s.specs {
		keys = append(keys, key)
	}
	return keys
}

func (s *SocketClient) notify() {
	select {
	case s.pending <- struct{}{}:
	default: // már van függő jelzés, elég egy
	}
}

// Run a kapcsolatot tartó ciklus: blokkol, amíg a context él. Kapcsolatvesztés
// után exponenciálisan növekvő várakozással próbálkozik újra.
func (s *SocketClient) Run(ctx context.Context) {
	delay := reconnectMinDelay

	for ctx.Err() == nil {
		// Ha nincs figyelt kör, nem tartunk nyitva kapcsolatot a WCA Live felé
		// — a szolgáltatást önkéntesek üzemeltetik, ne terheljük fölöslegesen.
		if len(s.Subscriptions()) == 0 {
			select {
			case <-ctx.Done():
				return
			case <-s.pending:
			case <-time.After(idleCheckInterval):
			}
			continue
		}

		startedAt := time.Now()
		err := s.session(ctx)

		if ctx.Err() != nil {
			return
		}

		// Egy hosszan élt kapcsolat után a szakadás valószínűleg egyszeri —
		// ilyenkor ne örököljük a korábbi hosszú várakozást.
		if time.Since(startedAt) >= healthySessionDuration {
			delay = reconnectMinDelay
		}

		if err != nil {
			log.Printf("[special] socket disconnected: %v (reconnecting in %s)", err, delay)
		} else {
			log.Printf("[special] socket closed (reconnecting in %s)", delay)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(delay):
		}

		delay *= 2
		if delay > reconnectMaxDelay {
			delay = reconnectMaxDelay
		}
	}
}

// session egy teljes kapcsolat élettartama: csatlakozás, csatorna-join,
// feliratkozások kiküldése, majd olvasás a kapcsolat végéig.
func (s *SocketClient) session(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dialCtx, dialCancel := context.WithTimeout(ctx, 15*time.Second)
	conn, _, err := websocket.Dial(dialCtx, s.url, &websocket.DialOptions{
		// Origin fejlécet szándékosan nem küldünk: a Phoenix check_origin
		// az Origin nélküli kéréseket átengedi, a WCA Live domainjét
		// hamisítani viszont nem akarjuk.
		HTTPHeader: http.Header{},
	})
	dialCancel()
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	// A kör-pillanatképek nagyok lehetnek (több száz eredmény).
	conn.SetReadLimit(8 << 20)
	defer conn.CloseNow()

	s.conn = conn
	s.ref = 0
	s.topicKey = make(map[string]string)
	s.sentKeys = make(map[string]bool)

	log.Printf("[special] socket connected to %s", s.url)

	frames := make(chan phoenixMessage, 32)
	readErr := make(chan error, 1)
	go func() {
		defer close(frames)
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				readErr <- err
				return
			}
			var msg phoenixMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				log.Printf("[special] dropping unparseable frame: %v", err)
				continue
			}
			select {
			case frames <- msg:
			case <-ctx.Done():
				return
			}
		}
	}()

	if err := s.send(ctx, absintheControlTopic, eventJoin, map[string]any{}); err != nil {
		return fmt.Errorf("join: %w", err)
	}

	heartbeat := time.NewTicker(heartbeatInterval)
	defer heartbeat.Stop()

	joined := false

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case err := <-readErr:
			return err

		case <-heartbeat.C:
			if err := s.send(ctx, phoenixTopic, eventHeartbeat, map[string]any{}); err != nil {
				return fmt.Errorf("heartbeat: %w", err)
			}

		case <-s.pending:
			if joined {
				s.syncSubscriptions(ctx)
			}

		case msg, ok := <-frames:
			if !ok {
				return errors.New("read loop ended")
			}

			switch {
			case msg.Event == eventReply && msg.Topic == absintheControlTopic:
				subscriptionID, ok := s.handleReply(msg)
				if !ok {
					continue
				}
				if subscriptionID == "" && !joined {
					// Ez a phx_join nyugtája: innentől mehetnek a doc-ok.
					joined = true
					log.Printf("[special] joined %s", absintheControlTopic)
					s.syncSubscriptions(ctx)
					if s.onConnect != nil {
						go s.onConnect()
					}
				}

			case msg.Event == eventSubscriptionData:
				s.handleData(ctx, msg)
			}
		}
	}
}

// syncSubscriptions kiküldi azokat a feliratkozásokat, amik ezen a kapcsolaton
// még nem mentek ki.
func (s *SocketClient) syncSubscriptions(ctx context.Context) {
	s.mu.Lock()
	specs := make(map[string]subscriptionSpec, len(s.specs))
	for key, spec := range s.specs {
		specs[key] = spec
	}
	s.mu.Unlock()

	for key, spec := range specs {
		if s.sentKeys[key] {
			continue
		}

		ref, err := s.sendWithRef(ctx, absintheControlTopic, eventDoc, map[string]any{
			"query":     spec.query,
			"variables": spec.variables,
		})
		if err != nil {
			log.Printf("[special] failed to send subscription %s: %v", key, err)
			return
		}

		s.sentKeys[key] = true
		// A választ ref alapján párosítjuk a kulcshoz.
		s.topicKey["ref:"+ref] = key
	}
}

// handleReply feldolgoz egy phx_reply-t. A visszaadott string a subscriptionId,
// ha a válasz feliratkozásé volt.
func (s *SocketClient) handleReply(msg phoenixMessage) (string, bool) {
	var reply struct {
		Status   string `json:"status"`
		Response struct {
			SubscriptionID string `json:"subscriptionId"`
		} `json:"response"`
	}
	if err := json.Unmarshal(msg.Payload, &reply); err != nil {
		return "", false
	}

	if reply.Status != "ok" {
		log.Printf("[special] control reply not ok: %s", string(msg.Payload))
		return "", false
	}

	subscriptionID := reply.Response.SubscriptionID
	if subscriptionID == "" {
		return "", true
	}

	if msg.Ref != nil {
		if key, ok := s.topicKey["ref:"+*msg.Ref]; ok {
			delete(s.topicKey, "ref:"+*msg.Ref)
			s.topicKey[subscriptionID] = key
			log.Printf("[special] subscribed %s", key)
		}
	}

	return subscriptionID, true
}

func (s *SocketClient) handleData(ctx context.Context, msg phoenixMessage) {
	key, ok := s.topicKey[msg.Topic]
	if !ok {
		// Ismeretlen topic: nem a mi feliratkozásunk (vagy már leiratkoztunk).
		return
	}

	var payload struct {
		Result struct {
			Data json.RawMessage `json:"data"`
		} `json:"result"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("[special] dropping malformed subscription payload for %s: %v", key, err)
		return
	}

	select {
	case s.events <- SubscriptionEvent{Key: key, Data: payload.Result.Data}:
	case <-ctx.Done():
	}
}

func (s *SocketClient) send(ctx context.Context, topic, event string, payload map[string]any) error {
	_, err := s.sendWithRef(ctx, topic, event, payload)
	return err
}

func (s *SocketClient) sendWithRef(ctx context.Context, topic, event string, payload map[string]any) (string, error) {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()

	s.ref++
	ref := strconv.Itoa(s.ref)

	// join_ref-nek végig "1"-et használunk: egyetlen csatornát tartunk nyitva.
	frame, err := json.Marshal([]any{"1", ref, topic, event, payload})
	if err != nil {
		return "", err
	}

	writeCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := s.conn.Write(writeCtx, websocket.MessageText, frame); err != nil {
		return "", err
	}

	return ref, nil
}

// phoenixMessage a Phoenix v2 tömbös üzenetformátuma.
type phoenixMessage struct {
	JoinRef *string
	Ref     *string
	Topic   string
	Event   string
	Payload json.RawMessage
}

func (m *phoenixMessage) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if len(raw) != 5 {
		return fmt.Errorf("expected 5 phoenix message fields, got %d", len(raw))
	}

	if err := json.Unmarshal(raw[0], &m.JoinRef); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[1], &m.Ref); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[2], &m.Topic); err != nil {
		return err
	}
	if err := json.Unmarshal(raw[3], &m.Event); err != nil {
		return err
	}
	m.Payload = raw[4]

	return nil
}
