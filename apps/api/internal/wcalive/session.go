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

// A WCA Live kapcsolat ÉLETCIKLUSA — az a rész, ami a `special` és az
// `allrounder` csomagban korábban külön-külön, majdnem szó szerint egyformán
// élt (~250 sor). Ami a két feature-ben tényleg különbözik, az NEM itt van: a
// GraphQL lekérdezés (a complexity limit miatt kötelezően külön), a
// pillanatkép alakja, és hogy egy vagy több versenyt követünk-e.
//
// Amit ez a réteg garantál, és amit könnyű elrontani, ha valaki újra
// lemásolja:
//
//   - A websocket kapcsolat CSAK addig él, amíg van nézője. Nincs háttérben
//     futó figyelő: az első néző indítja a session-t, az utolsó leállítja.
//   - A session-ök NEM fedhetnek át. A SocketClient kapcsolatonként nullázza a
//     belső állapotát (sentKeys, topicKey), ezért az új session megvárja az
//     előző lefutását (sessionDone), a futó session pedig megvárja, amíg a
//     socket.Run visszatér.
//   - A SetOnConnect EGYSZER kerül beállításra (a Watch-ban), nem session-önként
//     — különben a socket saját goroutine-jával versenyezne.
//   - A Session SOHA nem hívja a Watcher-t a saját mutexe alatt. Enélkül a
//     feature (ami a saját zárja alatt kérdezi meg, fut-e a session)
//     zárolás-sorrendi holtpontba futna.

// Subscription egy feliratkozás GraphQL dokumentuma és változói.
type Subscription struct {
	Query     string
	Variables map[string]any
}

// Watcher az a réteg, amit a feature ad hozzá: mit töltsünk újra, mire
// iratkozzunk fel, és mi történjék egy beérkező eseménnyel.
type Watcher interface {
	// Reload újratölti a jelenleg figyelt adatot. A session indulásakor, a
	// socket (újra)kapcsolódásakor és a frissítési időzítőre hívjuk — az
	// (újra)kapcsolódás miatt ez az, ami pótolja a kapcsolat nélkül eltelt időt.
	Reload(ctx context.Context)

	// WantedSubscriptions a MOST kívánt feliratkozások kulcs szerint. A
	// SyncSubscriptions ehhez igazítja a socketet: ami kiesett, azt lemondja,
	// ami új, arra feliratkozik.
	WantedSubscriptions() map[string]Subscription

	// HandleEvent egy beérkezett subscription payloadot dolgoz fel.
	HandleEvent(event SubscriptionEvent)
}

type Session struct {
	// name csak a naplóba kerül: melyik feature session-je ez.
	name     string
	socket   *SocketClient
	interval time.Duration

	watcher Watcher

	// running atomi, mert a feature-ök a SAJÁT zárjuk alatt kérdezik meg
	// (a gyorsítótár frissessége függ tőle) — így nincs zárolás-sorrendi kérdés.
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

// Watch beköti a feature-t. Egyszer kell hívni, a használatba vétel előtt
// (tipikusan a tracker konstruktorában).
func (s *Session) Watch(watcher Watcher) {
	s.watcher = watcher

	s.socket.SetOnConnect(func() {
		log.Printf("[%s] socket (re)connected, resyncing", s.name)
		watcher.Reload(s.Context())
	})
}

// Running igaz, amíg fut a session. Gyorsítótár-frissesség eldöntéséhez való:
// egy session nélkül betöltött pillanatkép nem frissül magától.
func (s *Session) Running() bool { return s.running.Load() }

// Context a futó session context-je, vagy context.Background(), ha nincs futó
// session. A háttérben indított betöltések ezt használják, hogy a session
// leállásakor ők is elengedjenek.
func (s *Session) Context() context.Context {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ctx == nil {
		return context.Background()
	}
	return s.ctx
}

// AddViewer egy nézőt jelez be. Az első néző indítja a WCA Live kapcsolatot, az
// utolsó leállítja. A visszaadott függvény többszöri hívása ártalmatlan.
//
// A `started` azt mondja meg, hogy EZ a hívás indította-e a session-t. Aki
// indította, annak nem kell külön betöltenie semmit: a session első dolga egy
// Reload. Aki viszont egy már futó session-höz csatlakozott, annak magának kell
// betöltetnie, amit néz. Ezt a döntést a számláló mellett, ugyanazon zár alatt
// kell meghozni — egy előzetes „fut-e már?” kérdés két egyszerre érkező nézőnél
// oda vezetne, hogy egyikük sem tölt be semmit.
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

		// A session-ök nem fedhetnek át: várjuk meg az előző lefutását.
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

// SyncSubscriptions a socket feliratkozásait a Watcher által kívánt halmazhoz
// igazítja. A feature akkor hívja, amikor a figyelt körök halmaza változhatott
// (betöltés után, illetve néző érkezésekor/távozásakor).
func (s *Session) SyncSubscriptions() {
	// Szándékosan a mutexen KÍVÜL: a Watcher a saját zárját veszi fel.
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

// Subscriptions a socket JELENLEG élő feliratkozásainak kulcsai. Nem azonos a
// Watcher kívánságlistájával: ez a tényleges állapot, az a szándék.
func (s *Session) Subscriptions() []string {
	return s.socket.Subscriptions()
}

func (s *Session) clearSubscriptions() {
	for _, key := range s.socket.Subscriptions() {
		s.socket.Unsubscribe(key)
	}
}

// A kör-feliratkozások kulcsa. A WCA Live kör-azonosítók globálisan egyediek,
// ezért ez a kulcs versenyek között sem ütközhet — emiatt tud egyetlen socket
// több versenyt is kiszolgálni.
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
