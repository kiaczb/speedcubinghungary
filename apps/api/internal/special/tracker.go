package special

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
)

const (
	// Amíg valaki nézi az oldalt, a roundUpdated subscription tartja frissen az
	// eredményeket. Ez csak annak az üteme, hogy megnézzük: nyílt-e új kör.
	// A körök napközben nyílnak, nem másodpercekre pontosan kell látni őket.
	refreshInterval = 2 * time.Minute

	// Ha épp senki nincs az adott verseny oldalán, nincs subscription sem —
	// ilyenkor ennyi ideig tekintjük frissnek a legutóbbi lekérdezést.
	cacheTTL = 15 * time.Second

	// Nézők nélkül maradt verseny pillanatképét ennyi ideig tartjuk még meg:
	// az oldalak közti ugrálás így nem indít újra mindent a WCA Live felé.
	idleRetention = 5 * time.Minute

	// Egyszerre ennyi versenyt tartunk nyilván. A versenyt a NÉZŐ választja,
	// tehát elvileg bárki bármit megnyithat — a WCA Live-ot önkéntesek
	// üzemeltetik, ezért felső korlátot adunk a párhuzamos figyelésnek.
	maxTrackedCompetitions = 20

	// Egyszerre ennyi kört kérdezünk le. A köröket külön-külön kell lekérni:
	// eredményekkel együtt a competition lekérdezés túllépi a WCA Live 5000-es
	// komplexitási korlátját.
	roundFetchConcurrency = 4
)

// ErrTooManyCompetitions akkor jön, ha minden figyelőhelyet néző foglal.
var ErrTooManyCompetitions = errors.New("too many competitions are being watched at once")

// Tracker tartja a nézett versenyek eredményeit — KIZÁRÓLAG a memóriában.
//
// A hivatalos eredmény a WCA Live-é; nálunk csak másolat, ezért nem tároljuk.
// Két forrásból tartjuk naprakészen:
//
//   - lekérdezés (GetRound): a körök teljes pillanatképe, induláskor és
//     időnként újra,
//   - roundUpdated subscription: élő frissítés, amíg legalább egy nézet nyitva
//     van (AddViewer). Nézők nélkül nem tartunk kapcsolatot a WCA Live felé.
//
// Minden esemény TELJES kör-pillanatkép, nem delta — ezért a duplikált,
// sorrenden kívüli és kimaradt események maguktól helyreállnak, és nincs
// szükség eseménynaplóra vagy sorszámozásra.
//
// ⚠️ TÖBB versenyt kezel egyszerre: a versenyt a néző választja ki, nincs
// kitüntetett "figyelt verseny". Az összes feliratkozás EGYETLEN websocketen
// multiplexálódik (a Phoenix csatornák eleve erre valók), mert a WCA Live
// kör-azonosítói globálisan egyediek — így a kulcs (round:<id>) versenyek
// között sem ütközhet.
type Tracker struct {
	client      *LiveClient
	socket      *SocketClient
	broadcaster *Broadcaster

	mu           sync.Mutex
	competitions map[string]*trackedCompetition

	// A munkamenet (websocket) az ÖSSZES nézetre közös: az elsőnél indul, az
	// utolsó bezárásakor áll le.
	viewers        int
	sessionRunning bool
	stopSession    context.CancelFunc
	sessionCtx     context.Context
	// sessionDone az ELŐZŐ munkamenet lezárását jelzi. Két munkamenet nem
	// futhat átfedésben: a SocketClient a kapcsolatonkénti állapotát
	// (elküldött feliratkozások) session-önként állítja vissza.
	sessionDone chan struct{}
}

type trackedCompetition struct {
	competition Competition

	// loadMu: versenyenként egyszerre egy teljes lekérdezés fusson, hogy az
	// egyszerre érkező kérések ne szorozzák fel a WCA Live felé menő forgalmat.
	loadMu sync.Mutex

	// A további mezőket a Tracker.mu védi.
	rounds   map[int]trackedRound
	loadedAt time.Time
	viewers  int
	// trackedSince az élő figyelés kezdete (0 → 1 néző). Csak az ez UTÁN
	// betöltött pillanatképet tekintjük magától frissülőnek — különben egy
	// régi, gyorsítótárazott állapotot néznénk élőnek.
	trackedSince time.Time
	// lastSeen az utolsó hozzáférés; a nézők nélkül maradt versenyek ez
	// alapján esnek ki a memóriából.
	lastSeen time.Time
}

// trackedRound egy figyelt kör. A timeLimit-et azért tartjuk el, mert a
// roundUpdated payload csak az eredményeket küldi — a limit viszont a kör
// futása alatt nem változik, így nem kell miatta újra lekérdezni.
type trackedRound struct {
	identity  RoundIdentity
	timeLimit *liveTimeLimit
	results   []ResultSnapshot
}

func NewTracker(client *LiveClient, socket *SocketClient, broadcaster *Broadcaster) *Tracker {
	tracker := &Tracker{
		client:       client,
		socket:       socket,
		broadcaster:  broadcaster,
		competitions: make(map[string]*trackedCompetition),
	}

	// EGYSZER állítjuk be, nem munkamenetenként: a SocketClient a saját
	// goroutine-jából olvassa, és az ismételt írás versenyhelyzet lenne.
	// Minden (újra)csatlakozás után teljes állapotot kérdezünk, hogy a
	// kapcsolat szünetében kimaradt bevitelek is bekerüljenek.
	socket.SetOnConnect(func() {
		log.Printf("[special] socket (re)connected, resyncing rounds")
		tracker.refresh(tracker.currentSessionContext())
	})

	return tracker
}

// currentSessionContext a futó munkamenet contextje, hogy az újracsatlakozás
// utáni lekérdezés is leálljon, amikor az utolsó néző elmegy.
func (t *Tracker) currentSessionContext() context.Context {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sessionCtx == nil {
		return context.Background()
	}
	return t.sessionCtx
}

// Results a verseny aktuális eredményei. Ha a versenyt élőben figyeljük, a
// memóriában lévő pillanatképet adja vissza (azt a subscription tartja
// frissen); egyébként szükség szerint újrakérdezi a WCA Live-ot.
func (t *Tracker) Results(ctx context.Context, competition Competition) ([]ResultSnapshot, error) {
	entry, err := t.entry(competition)
	if err != nil {
		return nil, err
	}

	if results, ok := t.cachedResults(entry); ok {
		return results, nil
	}

	return t.load(ctx, entry, false)
}

// AddViewer jelzi, hogy megnyílt egy nézet (SSE kapcsolat) az adott versenyre.
// Az első nézőnél elindul az élő subscription, az utolsó bezárásakor leáll — a
// WCA Live felé csak addig tartunk kapcsolatot, amíg tényleg nézi valaki. A
// visszaadott függvény zárja a nézetet, és többször hívva is biztonságos.
func (t *Tracker) AddViewer(competition Competition) (func(), error) {
	entry, err := t.entry(competition)
	if err != nil {
		return nil, err
	}

	t.mu.Lock()
	entry.viewers++
	if entry.viewers == 1 {
		entry.trackedSince = time.Now()
	}

	t.viewers++
	startSession := t.viewers == 1
	if startSession {
		t.startSessionLocked()
	}
	sessionCtx := t.sessionCtx
	t.mu.Unlock()

	// Induló munkamenetnél a runSession maga tölt be mindent; egyébként ennek
	// a versenynek a feliratkozásait most kell felvenni.
	if !startSession {
		go t.ensureLive(sessionCtx, entry)
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			t.mu.Lock()
			entry.viewers--
			entry.lastSeen = time.Now()
			if entry.viewers == 0 {
				entry.trackedSince = time.Time{}
			}

			t.viewers--
			last := t.viewers == 0
			stop := t.stopSession
			if last {
				t.sessionRunning = false
				t.stopSession = nil
				t.sessionCtx = nil
			}
			t.mu.Unlock()

			if last && stop != nil {
				stop()
				return
			}

			// Más versenyt még néznek: csak ennek a feliratkozásait engedjük el.
			t.syncSubscriptions()
		})
	}, nil
}

// startSessionLocked elindítja az élő munkamenetet. A hívó tartja a mu-t.
func (t *Tracker) startSessionLocked() {
	previous := t.sessionDone

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	t.stopSession = cancel
	t.sessionCtx = ctx
	t.sessionDone = done
	t.sessionRunning = true

	go func() {
		defer close(done)
		// Az előző munkamenet teljes leállását megvárjuk: a SocketClient
		// állapotán két párhuzamos kapcsolat összeakadna.
		if previous != nil {
			<-previous
		}
		t.runSession(ctx)
	}()
}

// entry a verseny nyilvántartott állapota, szükség szerint létrehozva.
func (t *Tracker) entry(competition Competition) (*trackedCompetition, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if entry, known := t.competitions[competition.CompetitionWCAID]; known {
		entry.lastSeen = time.Now()
		return entry, nil
	}

	t.evictLocked()

	if len(t.competitions) >= maxTrackedCompetitions {
		return nil, ErrTooManyCompetitions
	}

	entry := &trackedCompetition{competition: competition, lastSeen: time.Now()}
	t.competitions[competition.CompetitionWCAID] = entry

	return entry, nil
}

// evictLocked kiejti a nézők nélkül maradt, elévült versenyeket. Ha így sem fér
// be új, a legrégebben látott néző nélkülit dobja el. A hívó tartja a mu-t.
func (t *Tracker) evictLocked() {
	for id, entry := range t.competitions {
		if entry.viewers == 0 && time.Since(entry.lastSeen) > idleRetention {
			delete(t.competitions, id)
		}
	}

	for len(t.competitions) >= maxTrackedCompetitions {
		oldestID := ""
		var oldest time.Time

		for id, entry := range t.competitions {
			if entry.viewers > 0 {
				continue
			}
			if oldestID == "" || entry.lastSeen.Before(oldest) {
				oldestID, oldest = id, entry.lastSeen
			}
		}

		// Minden helyet néző foglal: a hívó kapja meg a hibát.
		if oldestID == "" {
			return
		}
		delete(t.competitions, oldestID)
	}
}

func (t *Tracker) cachedResults(entry *trackedCompetition) ([]ResultSnapshot, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if entry.rounds == nil {
		return nil, false
	}

	// Élő figyelés mellett a pillanatképek maguktól frissülnek, ezért nem
	// járnak le — de csak az a pillanatkép él magától, amelyik a figyelés
	// kezdete UTÁN készült. Élő figyelés nélkül rövid ideig tekintjük frissnek
	// a legutóbbi lekérdezést.
	live := t.sessionRunning && entry.viewers > 0 && entry.loadedAt.After(entry.trackedSince)
	if !live && time.Since(entry.loadedAt) > cacheTTL {
		return nil, false
	}

	return collect(entry), true
}

// collect a körök eredményeit fűzi egybe. A hívó tartja a mu-t.
func collect(entry *trackedCompetition) []ResultSnapshot {
	results := make([]ResultSnapshot, 0)
	for _, round := range entry.rounds {
		results = append(results, round.results...)
	}
	return results
}

// load lekérdezi a verseny minden megnyitott körét. force esetén akkor is,
// ha közben más már betöltötte (időzített frissítés, újracsatlakozás).
func (t *Tracker) load(ctx context.Context, entry *trackedCompetition, force bool) ([]ResultSnapshot, error) {
	entry.loadMu.Lock()
	defer entry.loadMu.Unlock()

	// Amíg a zárra vártunk, egy másik kérés már betölthette.
	if !force {
		if results, ok := t.cachedResults(entry); ok {
			return results, nil
		}
	}

	liveComp, err := t.client.GetCompetition(ctx, entry.competition.LiveCompetitionID)
	if err != nil {
		return nil, err
	}

	rounds := make(map[int]trackedRound)
	var mu sync.Mutex
	var wg sync.WaitGroup
	slots := make(chan struct{}, roundFetchConcurrency)

	for _, tracked := range openRounds(liveComp) {
		wg.Add(1)
		go func(tracked trackedRound) {
			defer wg.Done()

			slots <- struct{}{}
			defer func() { <-slots }()

			round, err := t.client.GetRound(ctx, tracked.identity.LiveRoundID)
			if err != nil {
				// Egy akadozó kör ne ürítse ki az egész oldalt verseny közben:
				// a többit megtartjuk, ez a következő frissítéskor bejön.
				log.Printf("[special] failed to query round %d: %v", tracked.identity.LiveRoundID, err)
				return
			}
			tracked.results = snapshotResults(tracked.identity, *round)

			mu.Lock()
			rounds[tracked.identity.LiveRoundID] = tracked
			mu.Unlock()
		}(tracked)
	}

	wg.Wait()

	t.mu.Lock()
	entry.rounds = rounds
	entry.loadedAt = time.Now()
	entry.lastSeen = entry.loadedAt
	results := collect(entry)
	sessionRunning := t.sessionRunning
	t.mu.Unlock()

	if sessionRunning {
		t.syncSubscriptions()
	}

	return results, nil
}

// ensureLive gondoskodik róla, hogy a most megnyílt nézet versenye friss
// legyen, és a körei feliratkozva legyenek.
func (t *Tracker) ensureLive(ctx context.Context, entry *trackedCompetition) {
	if ctx == nil {
		return
	}

	if _, ok := t.cachedResults(entry); ok {
		t.syncSubscriptions()
		return
	}

	if _, err := t.load(ctx, entry, false); err != nil {
		log.Printf("[special] failed to load competition %s: %v",
			entry.competition.CompetitionWCAID, err)
	}
}

// openRounds a verseny MINDEN megnyitott körét adja — a már lezártakat is.
//
// ⚠️ Szándékosan tágabb, mint a WCA Live kliensének feltétele
// (`open && (!finished || active)`). A Special pontszáma az ÖSSZES versenyszám
// összege, az oldalt viszont tipikusan verseny közben nyitják meg — a szűkebb
// feltétellel a addigra lezárult körök (a reggeli 2x2) sosem kerülnének be.
func openRounds(competition *liveCompetition) []trackedRound {
	rounds := make([]trackedRound, 0)

	for _, competitionEvent := range competition.CompetitionEvents {
		for _, round := range competitionEvent.Rounds {
			if !round.Open {
				continue
			}
			roundID, err := strconv.Atoi(round.ID)
			if err != nil {
				continue
			}

			rounds = append(rounds, trackedRound{
				identity: RoundIdentity{
					LiveRoundID: roundID,
					EventID:     competitionEvent.Event.ID,
					RoundNumber: round.Number,
				},
				timeLimit: round.TimeLimit,
			})
		}
	}

	return rounds
}

// runSession az élő frissítés élettartama: socket + feliratkozások, amíg van
// legalább egy nyitott nézet.
func (t *Tracker) runSession(ctx context.Context) {
	log.Printf("[special] live session starting")

	socketDone := make(chan struct{})
	go func() {
		defer close(socketDone)
		t.socket.Run(ctx)
	}()

	defer func() {
		// Csak akkor tekintjük lezártnak a munkamenetet, ha a kapcsolat is
		// tényleg leállt — különben a következő munkamenet ráindulna.
		<-socketDone
		t.clearSubscriptions()
		log.Printf("[special] live session stopped")
	}()

	t.refresh(ctx)

	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			t.refresh(ctx)

		case event := <-t.socket.Events():
			t.handleEvent(event)
		}
	}
}

// refresh újrakérdezi a NÉZETT versenyek köreit, majd egyezteti a
// feliratkozásokat. Egyben ez ejti ki a nézők nélkül maradt versenyeket.
func (t *Tracker) refresh(ctx context.Context) {
	t.mu.Lock()
	t.evictLocked()
	watched := make([]*trackedCompetition, 0, len(t.competitions))
	for _, entry := range t.competitions {
		if entry.viewers > 0 {
			watched = append(watched, entry)
		}
	}
	t.mu.Unlock()

	for _, entry := range watched {
		if _, err := t.load(ctx, entry, true); err != nil {
			log.Printf("[special] failed to load competition %s: %v",
				entry.competition.CompetitionWCAID, err)
		}
	}
}

func (t *Tracker) handleEvent(event SubscriptionEvent) {
	roundID, err := roundIDFromKey(event.Key)
	if err != nil {
		log.Printf("[special] ignoring event with unexpected key %q", event.Key)
		return
	}

	owner, tracked, watched := t.roundOwner(roundID)
	if !watched {
		return
	}

	var payload struct {
		RoundUpdated *liveRound `json:"roundUpdated"`
	}
	if err := json.Unmarshal(event.Data, &payload); err != nil || payload.RoundUpdated == nil {
		log.Printf("[special] failed to decode roundUpdated payload for round %d: %v", roundID, err)
		return
	}

	// A payload csak az eredményeket hozza; a limitet a lekérdezéskor eltett
	// értékkel egészítjük ki, így nem kell újra lekérdezni a kört.
	round := *payload.RoundUpdated
	round.TimeLimit = tracked.timeLimit
	tracked.results = snapshotResults(tracked.identity, round)

	t.mu.Lock()
	// Közben kieshetett a verseny — csak akkor írunk, ha a kör még a miénk.
	_, still := owner.rounds[roundID]
	if still {
		owner.rounds[roundID] = tracked
	}
	t.mu.Unlock()

	if still {
		t.broadcaster.Notify(owner.competition.CompetitionWCAID)
	}
}

// roundOwner megkeresi, melyik nézett versenyhez tartozik a kör. A WCA Live
// kör-azonosítói globálisan egyediek, ezért legfeljebb egy találat lehet.
func (t *Tracker) roundOwner(roundID int) (*trackedCompetition, trackedRound, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, entry := range t.competitions {
		if round, ok := entry.rounds[roundID]; ok {
			return entry, round, true
		}
	}

	return nil, trackedRound{}, false
}

// syncSubscriptions a feliratkozásokat a NÉZETT versenyek köreihez igazítja.
// Amit már senki nem néz, arról leiratkozunk.
func (t *Tracker) syncSubscriptions() {
	t.mu.Lock()
	wanted := make(map[string]int)
	for _, entry := range t.competitions {
		if entry.viewers == 0 {
			continue
		}
		for roundID := range entry.rounds {
			wanted[subscriptionKey(roundID)] = roundID
		}
	}
	t.mu.Unlock()

	for _, key := range t.socket.Subscriptions() {
		if _, keep := wanted[key]; !keep {
			t.socket.Unsubscribe(key)
		}
	}

	for key, roundID := range wanted {
		t.socket.Subscribe(key, RoundSubscription, map[string]any{
			"id": strconv.Itoa(roundID),
		})
	}
}

func (t *Tracker) clearSubscriptions() {
	for _, key := range t.socket.Subscriptions() {
		t.socket.Unsubscribe(key)
	}
}

// subscriptionKey / roundIDFromKey a socket kulcs és a kör-id közti fordítás.
const subscriptionKeyPrefix = "round:"

func subscriptionKey(roundID int) string {
	return subscriptionKeyPrefix + strconv.Itoa(roundID)
}

func roundIDFromKey(key string) (int, error) {
	trimmed, found := strings.CutPrefix(key, subscriptionKeyPrefix)
	if !found {
		return 0, fmt.Errorf("subscription key %q has no %q prefix", key, subscriptionKeyPrefix)
	}
	return strconv.Atoi(trimmed)
}
