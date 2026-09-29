package wcalive

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/coder/websocket"
)

const (
	absintheControlTopic  = "__absinthe__:control"
	phoenixTopic          = "phoenix"
	eventJoin             = "phx_join"
	eventDoc              = "doc"
	eventUnsubscribe      = "unsubscribe"
	eventReply            = "phx_reply"
	eventHeartbeat        = "heartbeat"
	eventSubscriptionData = "subscription:data"

	heartbeatInterval = 25 * time.Second

	reconnectMinDelay = 2 * time.Second
	reconnectMaxDelay = 60 * time.Second

	healthySessionDuration = time.Minute

	idleCheckInterval = 30 * time.Second
)

type SubscriptionEvent struct {
	Key  string
	Data json.RawMessage
}

type subscriptionSpec struct {
	query     string
	variables map[string]any
}

type SocketClient struct {
	url    string
	events chan SubscriptionEvent

	onConnect func()

	mu      sync.Mutex
	specs   map[string]subscriptionSpec
	pending chan struct{}

	// Az alábbi mezőket KIZÁRÓLAG a session goroutine írja és olvassa
	// (syncSubscriptions, handleReply, handleData) — ezért nincs rajtuk zár.
	conn    *websocket.Conn
	writeMu sync.Mutex
	ref     int
	// topicKey: Absinthe subscriptionId (és átmenetileg "ref:<ref>") → kulcs.
	// keyTopic ennek a megfordítása — a LEMONDÁSHOZ kell, mert az upstream a
	// subscriptionId-t várja, mi viszont kulcs szerint mondunk le.
	topicKey map[string]string
	keyTopic map[string]string
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

func socketURL(baseURL string) string {
	if baseURL == "" {
		baseURL = DefaultURL
	}
	url := strings.TrimSuffix(baseURL, "/")
	url = strings.Replace(url, "https://", "wss://", 1)
	url = strings.Replace(url, "http://", "ws://", 1)
	return url + "/socket/websocket?vsn=2.0.0"
}

func (s *SocketClient) Events() <-chan SubscriptionEvent { return s.events }

func (s *SocketClient) SetOnConnect(fn func()) { s.onConnect = fn }

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

// Unsubscribe elengedi a feliratkozást. A kulcs helyi törlése önmagában NEM
// elég: a WCA Live-nak is meg kell mondani, különben a kapcsolat végéig tovább
// küldi az adott kör roundUpdated eseményeit, amit aztán feleslegesen dolgozunk
// fel (a `special`-ben akár egy olyan versenyét, amit már senki nem néz). A
// tényleges lemondó keretet a session goroutine küldi ki — ezért csak jelzünk.
func (s *SocketClient) Unsubscribe(key string) {
	s.mu.Lock()
	_, existed := s.specs[key]
	delete(s.specs, key)
	s.mu.Unlock()

	if existed {
		s.notify()
	}
}

func (s *SocketClient) wants(key string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.specs[key]
	return ok
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
	default:
	}
}

func (s *SocketClient) Run(ctx context.Context) {
	delay := reconnectMinDelay

	for ctx.Err() == nil {
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

		if time.Since(startedAt) >= healthySessionDuration {
			delay = reconnectMinDelay
		}

		if err != nil {
			log.Printf("[wcalive] socket disconnected: %v (reconnecting in %s)", err, delay)
		} else {
			log.Printf("[wcalive] socket closed (reconnecting in %s)", delay)
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

func (s *SocketClient) session(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dialCtx, dialCancel := context.WithTimeout(ctx, 15*time.Second)
	// Origin fejléc nélkül: a Phoenix check_origin csak így enged be.
	conn, _, err := websocket.Dial(dialCtx, s.url, nil)
	dialCancel()
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	conn.SetReadLimit(8 << 20)
	defer conn.CloseNow()

	s.conn = conn
	s.ref = 0
	s.topicKey = make(map[string]string)
	s.keyTopic = make(map[string]string)
	s.sentKeys = make(map[string]bool)

	log.Printf("[wcalive] socket connected to %s", s.url)

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
				log.Printf("[wcalive] dropping unparseable frame: %v", err)
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
					joined = true
					log.Printf("[wcalive] joined %s", absintheControlTopic)
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

func (s *SocketClient) syncSubscriptions(ctx context.Context) {
	s.mu.Lock()
	specs := make(map[string]subscriptionSpec, len(s.specs))
	for key, spec := range s.specs {
		specs[key] = spec
	}
	s.mu.Unlock()

	// Előbb a kiesettek: amit már nem akarunk, azt az upstream felé is le kell
	// mondani, különben a kapcsolat végéig tovább érkeznek az eseményei.
	for key := range s.sentKeys {
		if _, keep := specs[key]; !keep {
			s.dropSubscription(ctx, key)
		}
	}

	for key, spec := range specs {
		if s.sentKeys[key] {
			continue
		}

		ref, err := s.sendWithRef(ctx, absintheControlTopic, eventDoc, map[string]any{
			"query":     spec.query,
			"variables": spec.variables,
		})
		if err != nil {
			log.Printf("[wcalive] failed to send subscription %s: %v", key, err)
			return
		}

		s.sentKeys[key] = true
		s.topicKey["ref:"+ref] = key
	}
}

// dropSubscription lemondja a feliratkozást az Absinthe control csatornáján, és
// elengedi a hozzá tartozó helyi állapotot.
func (s *SocketClient) dropSubscription(ctx context.Context, key string) {
	subscriptionID, known := s.keyTopic[key]
	if !known {
		// A feliratkozás visszaigazolása — és vele a subscriptionId — még nem
		// érkezett meg, lemondani tehát nincs mit. A kulcsot szándékosan
		// BENNHAGYJUK a sentKeys-ben: a handleReply látja majd, hogy közben
		// elengedtük, és onnan indul újra ez az ág.
		return
	}

	delete(s.sentKeys, key)
	delete(s.keyTopic, key)
	delete(s.topicKey, subscriptionID)

	if err := s.send(ctx, absintheControlTopic, eventUnsubscribe, map[string]any{
		"subscriptionId": subscriptionID,
	}); err != nil {
		log.Printf("[wcalive] failed to unsubscribe %s: %v", key, err)
		return
	}

	log.Printf("[wcalive] unsubscribed %s", key)
}

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
		log.Printf("[wcalive] control reply not ok: %s", string(msg.Payload))
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
			s.keyTopic[key] = subscriptionID
			log.Printf("[wcalive] subscribed %s", key)

			// A kulcsot még a visszaigazolás előtt is elengedhették. Csak most
			// lett subscriptionId-nk, amivel le lehet mondani — kérjünk hát egy
			// újabb szinkront.
			if !s.wants(key) {
				s.notify()
			}
		}
	}

	return subscriptionID, true
}

func (s *SocketClient) handleData(ctx context.Context, msg phoenixMessage) {
	key, ok := s.topicKey[msg.Topic]
	if !ok {
		return
	}

	var payload struct {
		Result struct {
			Data json.RawMessage `json:"data"`
		} `json:"result"`
	}
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		log.Printf("[wcalive] dropping malformed subscription payload for %s: %v", key, err)
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
