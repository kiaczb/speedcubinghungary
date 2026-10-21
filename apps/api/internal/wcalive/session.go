package wcalive

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Subscription struct {
	Query     string
	Variables map[string]any
}

type Watcher interface {
	Reload(ctx context.Context)

	WantedSubscriptions() map[string]Subscription

	HandleEvent(event SubscriptionEvent)
}

type Session struct {
	name     string
	socket   *SocketClient
	interval time.Duration

	watcher Watcher

	running atomic.Bool

	mu      sync.Mutex
	viewers int
	stop    context.CancelFunc
	ctx     context.Context
	done    chan struct{}
}

func NewSession(name string, socket *SocketClient, refreshInterval time.Duration) *Session {
	return &Session{name: name, socket: socket, interval: refreshInterval}
}

func (s *Session) Watch(watcher Watcher) {
	s.watcher = watcher

	s.socket.SetOnConnect(func() {
		log.Printf("[%s] socket (re)connected, resyncing", s.name)
		watcher.Reload(s.Context())
	})
}

func (s *Session) Running() bool { return s.running.Load() }

func (s *Session) Context() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

func (s *Session) AddViewer() (release func(), started bool) {
	s.mu.Lock()
	s.viewers++
	started = s.viewers == 1
	if started {
		s.startLocked()
	}
	s.mu.Unlock()

	var once sync.Once
	return func() {
		once.Do(func() {
			s.mu.Lock()
			s.viewers--
			last := s.viewers == 0
			stop := s.stop
			if last {
				s.running.Store(false)
				s.stop = nil
				s.ctx = nil
			}
			s.mu.Unlock()

			if last && stop != nil {
				stop()
			}
		})
	}, started
}

func (s *Session) startLocked() {
	previous := s.done

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	s.stop = cancel
	s.ctx = ctx
	s.done = done
	s.running.Store(true)

	go func() {
		defer close(done)

		if previous != nil {
			<-previous
		}
		s.run(ctx)
	}()
}

func (s *Session) run(ctx context.Context) {
	log.Printf("[%s] live session starting", s.name)

	socketDone := make(chan struct{})
	go func() {
		defer close(socketDone)
		s.socket.Run(ctx)
	}()

	defer func() {
		<-socketDone
		s.clearSubscriptions()
		log.Printf("[%s] live session stopped", s.name)
	}()

	s.watcher.Reload(ctx)

	ticker := time.NewTicker(s.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			s.watcher.Reload(ctx)

		case event := <-s.socket.Events():
			s.watcher.HandleEvent(event)
		}
	}
}

func (s *Session) SyncSubscriptions() {
	wanted := s.watcher.WantedSubscriptions()

	for _, key := range s.socket.Subscriptions() {
		if _, keep := wanted[key]; !keep {
			s.socket.Unsubscribe(key)
		}
	}

	for key, subscription := range wanted {
		s.socket.Subscribe(key, subscription.Query, subscription.Variables)
	}
}

func (s *Session) Subscriptions() []string {
	return s.socket.Subscriptions()
}

func (s *Session) clearSubscriptions() {
	for _, key := range s.socket.Subscriptions() {
		s.socket.Unsubscribe(key)
	}
}

const roundKeyPrefix = "round:"

func RoundSubscriptionKey(liveRoundID int) string {
	return roundKeyPrefix + strconv.Itoa(liveRoundID)
}

func RoundIDFromKey(key string) (int, error) {
	trimmed, found := strings.CutPrefix(key, roundKeyPrefix)
	if !found {
		return 0, fmt.Errorf("subscription key %q has no %q prefix", key, roundKeyPrefix)
	}
	return strconv.Atoi(trimmed)
}
