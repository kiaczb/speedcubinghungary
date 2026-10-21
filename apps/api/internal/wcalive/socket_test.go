package wcalive

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSubscriptionDoc = `subscription RoundUpdated($id: ID!) { roundUpdated(id: $id) { id } }`

type fakeLiveSocket struct {
	server *httptest.Server

	mu           sync.Mutex
	docQueries   []string
	unsubscribed []string
	connections  int
	dropAfterDoc bool
	broadcast    chan string
}

func newFakeLiveSocket(t *testing.T) *fakeLiveSocket {
	t.Helper()

	f := &fakeLiveSocket{broadcast: make(chan string, 4)}

	f.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()

		f.mu.Lock()
		f.connections++
		drop := f.dropAfterDoc
		f.mu.Unlock()

		ctx := r.Context()

		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				return
			}

			var msg phoenixMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				return
			}

			switch msg.Event {
			case eventJoin:
				f.reply(ctx, conn, msg, map[string]any{})

			case eventHeartbeat:
				f.reply(ctx, conn, msg, map[string]any{})

			case eventUnsubscribe:
				var payload struct {
					SubscriptionID string `json:"subscriptionId"`
				}
				_ = json.Unmarshal(msg.Payload, &payload)

				f.mu.Lock()
				f.unsubscribed = append(f.unsubscribed, payload.SubscriptionID)
				f.mu.Unlock()

				f.reply(ctx, conn, msg, map[string]any{"subscriptionId": payload.SubscriptionID})

			case eventDoc:
				var payload struct {
					Query     string         `json:"query"`
					Variables map[string]any `json:"variables"`
				}
				_ = json.Unmarshal(msg.Payload, &payload)

				f.mu.Lock()
				f.docQueries = append(f.docQueries, payload.Query)
				f.mu.Unlock()

				subscriptionID := "__absinthe__:doc:global:" + payload.Variables["id"].(string)
				f.reply(ctx, conn, msg, map[string]any{"subscriptionId": subscriptionID})

				if drop {
					conn.CloseNow()
					return
				}

				go func() {
					select {
					case f.broadcast <- subscriptionID:
					case <-ctx.Done():
					}
				}()

			default:
			}
		}
	}))

	t.Cleanup(f.server.Close)

	return f
}

func (f *fakeLiveSocket) reply(ctx context.Context, conn *websocket.Conn, msg phoenixMessage, response map[string]any) {
	ref := ""
	if msg.Ref != nil {
		ref = *msg.Ref
	}
	frame, _ := json.Marshal([]any{"1", ref, msg.Topic, eventReply, map[string]any{
		"status":   "ok",
		"response": response,
	}})
	_ = conn.Write(ctx, websocket.MessageText, frame)
}

func (f *fakeLiveSocket) push(ctx context.Context, conn *websocket.Conn, topic string, data any) error {
	frame, err := json.Marshal([]any{nil, nil, topic, eventSubscriptionData, map[string]any{
		"subscriptionId": topic,
		"result":         map[string]any{"data": data},
	}})
	if err != nil {
		return err
	}
	return conn.Write(ctx, websocket.MessageText, frame)
}

func (f *fakeLiveSocket) baseURL() string {
	return f.server.URL
}

func (f *fakeLiveSocket) stats() (connections int, queries []string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.connections, append([]string(nil), f.docQueries...)
}

func (f *fakeLiveSocket) unsubscribes() []string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return append([]string(nil), f.unsubscribed...)
}

func TestSocketURL(t *testing.T) {
	assert.Equal(t,
		"wss://live.worldcubeassociation.org/socket/websocket?vsn=2.0.0",
		socketURL("https://live.worldcubeassociation.org"))
	assert.Equal(t,
		"ws://localhost:4000/socket/websocket?vsn=2.0.0",
		socketURL("http://localhost:4000/"))
	assert.True(t, strings.HasPrefix(socketURL(""), "wss://live.worldcubeassociation.org"))
}

func TestPhoenixMessageUnmarshal(t *testing.T) {
	var msg phoenixMessage
	require.NoError(t, json.Unmarshal(
		[]byte(`[null,null,"__absinthe__:doc:global:X","subscription:data",{"result":{"data":{"roundUpdated":{"id":"1"}}}}]`),
		&msg))

	assert.Nil(t, msg.JoinRef)
	assert.Nil(t, msg.Ref)
	assert.Equal(t, "__absinthe__:doc:global:X", msg.Topic)
	assert.Equal(t, eventSubscriptionData, msg.Event)
	assert.Contains(t, string(msg.Payload), "roundUpdated")

	assert.Error(t, json.Unmarshal([]byte(`["1","2","topic"]`), &phoenixMessage{}))
}

func TestSocketClient_DeliversSubscriptionData(t *testing.T) {
	fake := newFakeLiveSocket(t)

	var (
		mu       sync.Mutex
		liveConn *websocket.Conn
	)
	fake.server.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()

		mu.Lock()
		liveConn = conn
		mu.Unlock()

		ctx := r.Context()
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				return
			}
			var msg phoenixMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				return
			}

			switch msg.Event {
			case eventJoin, eventHeartbeat:
				fake.reply(ctx, conn, msg, map[string]any{})
			case eventDoc:
				var payload struct {
					Variables map[string]any `json:"variables"`
				}
				_ = json.Unmarshal(msg.Payload, &payload)
				topic := "__absinthe__:doc:global:" + payload.Variables["id"].(string)
				fake.reply(ctx, conn, msg, map[string]any{"subscriptionId": topic})

				_ = fake.push(ctx, conn, topic, map[string]any{
					"roundUpdated": map[string]any{"id": "147464"},
				})
			}
		}
	})

	client := NewSocketClient(fake.baseURL())

	connected := make(chan struct{}, 1)
	client.SetOnConnect(func() {
		select {
		case connected <- struct{}{}:
		default:
		}
	})

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client.Subscribe("round:147464", testSubscriptionDoc, map[string]any{"id": "147464"})
	go client.Run(ctx)

	select {
	case <-connected:
	case <-ctx.Done():
		t.Fatal("onConnect never fired")
	}

	select {
	case event := <-client.Events():
		assert.Equal(t, "round:147464", event.Key, "az esemény a hívó kulcsával érkezik")
		assert.JSONEq(t, `{"roundUpdated":{"id":"147464"}}`, string(event.Data))
	case <-ctx.Done():
		t.Fatal("no subscription event received")
	}

	mu.Lock()
	require.NotNil(t, liveConn)
	mu.Unlock()
}

func TestSocketClient_SubscribeIsIdempotent(t *testing.T) {
	fake := newFakeLiveSocket(t)
	client := NewSocketClient(fake.baseURL())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client.Subscribe("round:1", testSubscriptionDoc, map[string]any{"id": "1"})
	client.Subscribe("round:1", testSubscriptionDoc, map[string]any{"id": "1"})
	client.Subscribe("round:1", testSubscriptionDoc, map[string]any{"id": "1"})

	assert.Len(t, client.Subscriptions(), 1)

	go client.Run(ctx)

	require.Eventually(t, func() bool {
		_, queries := fake.stats()
		return len(queries) >= 1
	}, 5*time.Second, 20*time.Millisecond)

	time.Sleep(200 * time.Millisecond)

	_, queries := fake.stats()
	assert.Len(t, queries, 1, "egy kulcshoz egy feliratkozás tartozik")
	assert.Contains(t, queries[0], "roundUpdated")
}

func TestSocketClient_ReconnectsAndResubscribes(t *testing.T) {
	fake := newFakeLiveSocket(t)

	fake.mu.Lock()
	fake.dropAfterDoc = true
	fake.mu.Unlock()

	client := NewSocketClient(fake.baseURL())

	var connects int
	var mu sync.Mutex
	client.SetOnConnect(func() {
		mu.Lock()
		connects++
		mu.Unlock()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	client.Subscribe("round:42", testSubscriptionDoc, map[string]any{"id": "42"})
	go client.Run(ctx)

	require.Eventually(t, func() bool {
		connections, queries := fake.stats()
		return connections >= 2 && len(queries) >= 2
	}, 15*time.Second, 100*time.Millisecond, "a kliensnek újra kell csatlakoznia és újra feliratkoznia")

	mu.Lock()
	defer mu.Unlock()
	assert.GreaterOrEqual(t, connects, 2,
		"minden újracsatlakozás után újra le kell kérdezni a teljes állapotot")
}

func TestSocketClient_DoesNotConnectWithoutSubscriptions(t *testing.T) {
	fake := newFakeLiveSocket(t)
	client := NewSocketClient(fake.baseURL())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go client.Run(ctx)

	time.Sleep(300 * time.Millisecond)
	connections, _ := fake.stats()
	assert.Zero(t, connections, "üres feliratkozás-listával nem csatlakozunk")

	client.Subscribe("round:7", testSubscriptionDoc, map[string]any{"id": "7"})

	require.Eventually(t, func() bool {
		connections, queries := fake.stats()
		return connections == 1 && len(queries) == 1
	}, 5*time.Second, 20*time.Millisecond, "feliratkozásra azonnal csatlakozunk")
}

func TestSocketClient_Unsubscribe(t *testing.T) {
	client := NewSocketClient("http://example.invalid")

	client.Subscribe("round:1", testSubscriptionDoc, map[string]any{"id": "1"})
	client.Subscribe("round:2", testSubscriptionDoc, map[string]any{"id": "2"})
	require.Len(t, client.Subscriptions(), 2)

	client.Unsubscribe("round:1")
	assert.Equal(t, []string{"round:2"}, client.Subscriptions())
}

func TestSocketClient_UnsubscribeTellsWCALive(t *testing.T) {
	fake := newFakeLiveSocket(t)
	client := NewSocketClient(fake.baseURL())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client.Subscribe("round:7", testSubscriptionDoc, map[string]any{"id": "7"})
	client.Subscribe("round:8", testSubscriptionDoc, map[string]any{"id": "8"})
	go client.Run(ctx)

	require.Eventually(t, func() bool {
		_, queries := fake.stats()
		return len(queries) == 2
	}, 5*time.Second, 20*time.Millisecond, "mindkét feliratkozás kimegy")

	client.Unsubscribe("round:7")

	require.Eventually(t, func() bool {
		return len(fake.unsubscribes()) == 1
	}, 5*time.Second, 20*time.Millisecond, "a lemondásnak el kell jutnia a WCA Live-ig")

	assert.Equal(t, []string{"__absinthe__:doc:global:7"}, fake.unsubscribes(),
		"pontosan az elengedett kör feliratkozását mondjuk le, a másikat nem")

	_, queries := fake.stats()
	assert.Len(t, queries, 2, "a lemondás nem indít újra feliratkozást")
}

func TestSocketClient_DropsEventsAfterUnsubscribe(t *testing.T) {
	const topic = "__absinthe__:doc:global:7"

	subscribed := make(chan struct{}, 1)
	unsubscribed := make(chan struct{}, 1)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Accept(w, r, nil)
		if err != nil {
			return
		}
		defer conn.CloseNow()

		ctx := r.Context()
		fake := &fakeLiveSocket{}

		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				return
			}
			var msg phoenixMessage
			if err := json.Unmarshal(data, &msg); err != nil {
				return
			}

			switch msg.Event {
			case eventJoin, eventHeartbeat:
				fake.reply(ctx, conn, msg, map[string]any{})

			case eventDoc:
				fake.reply(ctx, conn, msg, map[string]any{"subscriptionId": topic})
				signal(subscribed)

			case eventUnsubscribe:
				fake.reply(ctx, conn, msg, map[string]any{"subscriptionId": topic})

				_ = fake.push(ctx, conn, topic, map[string]any{
					"roundUpdated": map[string]any{"id": "7"},
				})

				signal(unsubscribed)
			}
		}
	}))
	t.Cleanup(server.Close)

	client := NewSocketClient(server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client.Subscribe("round:7", testSubscriptionDoc, map[string]any{"id": "7"})
	go client.Run(ctx)

	waitForSignal(t, subscribed, "a feliratkozás nem ért ki")

	client.Unsubscribe("round:7")
	waitForSignal(t, unsubscribed, "a lemondás nem ért ki")

	select {
	case event := <-client.Events():
		t.Fatalf("a lemondott feliratkozás eseménye nem juthat ki: %+v", event)
	case <-time.After(300 * time.Millisecond):
	}
}

func signal(ch chan struct{}) {
	select {
	case ch <- struct{}{}:
	default:
	}
}

func waitForSignal(t *testing.T, ch <-chan struct{}, message string) {
	t.Helper()

	select {
	case <-ch:
	case <-time.After(5 * time.Second):
		t.Fatal(message)
	}
}
