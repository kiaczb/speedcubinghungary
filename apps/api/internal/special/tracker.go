package special

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	refreshInterval = 2 * time.Minute

	cacheTTL = 15 * time.Second

	idleRetention = 5 * time.Minute

	maxTrackedCompetitions = 20

	roundFetchConcurrency = 4
)

var ErrTooManyCompetitions = errors.New("too many competitions are being watched at once")

type Tracker struct {
	client      *LiveClient
	broadcaster *Broadcaster

	// session a WCA Live kapcsolat életciklusa (nézőszámlálás, socket
	// indítás/leállítás, feliratkozás-szinkron) — lásd internal/wcalive/session.go.
	// Ez a réteg közös az allrounder csomaggal; itt CSAK az marad, ami a
	// Special sajátja: több verseny követése egyszerre, és a pillanatkép alakja.
	session *Session

	mu           sync.Mutex
	competitions map[string]*trackedCompetition
}

type trackedCompetition struct {
	competition Competition

	loadMu sync.Mutex

	rounds       map[int]trackedRound
	loadedAt     time.Time
	viewers      int
	trackedSince time.Time
	lastSeen     time.Time
}

type trackedRound struct {
	identity  RoundIdentity
	timeLimit *liveTimeLimit
	results   []ResultSnapshot
}

func NewTracker(client *LiveClient, socket *SocketClient, broadcaster *Broadcaster) *Tracker {
	tracker := &Tracker{
		client:       client,
		broadcaster:  broadcaster,
		session:      NewSession("special", socket, refreshInterval),
		competitions: make(map[string]*trackedCompetition),
	}

	tracker.session.Watch(tracker)

	return tracker
}

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

func (t *Tracker) AddViewer(competition Competition) (func(), error) {
	entry, err := t.entry(competition)
	if err != nil {
		return nil, err
	}

	// A versenyenkénti nézőszám a MI könyvelésünk (ebből derül ki, melyik
	// verseny köreire kell feliratkozni, és hogy a pillanatkép magától
	// frissül-e); a socket életciklusát a session számolja.
	t.mu.Lock()
	entry.viewers++
	if entry.viewers == 1 {
		entry.trackedSince = time.Now()
	}
	t.mu.Unlock()

	releaseSession, startedSession := t.session.AddViewer()

	if !startedSession {
		// A session már futott, tehát EZT a versenyt külön kell betölteni és
		// feliratkoztatni; az induló session ezt magától megtenné.
		go t.ensureLive(t.session.Context(), entry)
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
			t.mu.Unlock()

			releaseSession()

			// Ha a session él tovább (más versenyt még néznek), ennek a
			// versenynek a feliratkozásait el kell engedni.
			if t.session.Running() {
				t.session.SyncSubscriptions()
			}
		})
	}, nil
}

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

	live := t.session.Running() && entry.viewers > 0 && entry.loadedAt.After(entry.trackedSince)
	if !live && time.Since(entry.loadedAt) > cacheTTL {
		return nil, false
	}

	return collect(entry), true
}

func collect(entry *trackedCompetition) []ResultSnapshot {
	results := make([]ResultSnapshot, 0)
	for _, round := range entry.rounds {
		results = append(results, round.results...)
	}
	return results
}

func (t *Tracker) load(ctx context.Context, entry *trackedCompetition, force bool) ([]ResultSnapshot, error) {
	entry.loadMu.Lock()
	defer entry.loadMu.Unlock()

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
	failed := make(map[int]bool)
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
				log.Printf("[special] failed to query round %d: %v", tracked.identity.LiveRoundID, err)

				mu.Lock()
				failed[tracked.identity.LiveRoundID] = true
				mu.Unlock()
				return
			}
			tracked.results = snapshotResults(tracked.identity, *round)

			mu.Lock()
			rounds[tracked.identity.LiveRoundID] = tracked
			mu.Unlock()
		}(tracked)
	}

	wg.Wait()

	// Ha EGYETLEN kör sem jött le, az nem részleges hiba, hanem elérhetetlen
	// upstream. Ilyenkor ne írjuk felül a meglévő állapotot, és a loadedAt-ot se
	// frissítsük — így a következő kérés azonnal újrapróbálja, ahelyett hogy egy
	// üres pillanatképet szolgálnánk ki a cache TTL végéig.
	if len(failed) > 0 && len(rounds) == 0 {
		return nil, fmt.Errorf("every round query failed for %s", entry.competition.CompetitionWCAID)
	}

	t.mu.Lock()
	// ⚠️ Egy kör lekérési hibája NEM törölheti a kör addigi eredményeit. A map
	// teljes cseréjével egy pillanatnyi WCA Live hiba eltüntetné a kört a
	// riportból, és a kumulatív keret (A1a3) hibás „maradék időt” mutatna —
	// ezért a hibás körök az ELŐZŐ pillanatképükkel élnek tovább.
	for roundID := range failed {
		if previous, known := entry.rounds[roundID]; known {
			rounds[roundID] = previous
		}
	}

	entry.rounds = rounds
	entry.loadedAt = time.Now()
	entry.lastSeen = entry.loadedAt
	results := collect(entry)
	t.mu.Unlock()

	if t.session.Running() {
		t.session.SyncSubscriptions()
	}

	return results, nil
}

func (t *Tracker) ensureLive(ctx context.Context, entry *trackedCompetition) {
	if ctx == nil {
		return
	}

	if _, ok := t.cachedResults(entry); ok {
		t.session.SyncSubscriptions()
		return
	}

	if _, err := t.load(ctx, entry, false); err != nil {
		log.Printf("[special] failed to load competition %s: %v",
			entry.competition.CompetitionWCAID, err)
	}
}

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

// Reload a wcalive.Watcher tagja: a session indulásakor, a socket
// (újra)kapcsolódásakor és a frissítési időzítőre hívódik. Minden ÉPPEN NÉZETT
// verseny minden körét újratölti — ez az, ami pótolja a kapcsolat nélkül
// eltelt időt, és ami észreveszi, hogy új kör nyílt.
func (t *Tracker) Reload(ctx context.Context) {
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

// HandleEvent a wcalive.Watcher tagja. Minden roundUpdated esemény TELJES kör-
// pillanatképet hoz (nem különbséget), ezért a duplikált, késve vagy egyáltalán
// meg nem érkező esemény magától gyógyul — ne építs rá eseménynaplót.
func (t *Tracker) HandleEvent(event SubscriptionEvent) {
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

	round := *payload.RoundUpdated
	round.TimeLimit = tracked.timeLimit
	tracked.results = snapshotResults(tracked.identity, round)

	t.mu.Lock()
	_, still := owner.rounds[roundID]
	if still {
		owner.rounds[roundID] = tracked
	}
	t.mu.Unlock()

	if still {
		t.broadcaster.Notify(owner.competition.CompetitionWCAID)
	}
}

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

// WantedSubscriptions a wcalive.Watcher tagja: a MOST nézett versenyek köreire
// akarunk feliratkozni. Nézők nélküli verseny köre nincs benne, így a session
// magától elengedi.
func (t *Tracker) WantedSubscriptions() map[string]Subscription {
	t.mu.Lock()
	defer t.mu.Unlock()

	wanted := make(map[string]Subscription)

	for _, entry := range t.competitions {
		if entry.viewers == 0 {
			continue
		}
		for roundID := range entry.rounds {
			wanted[subscriptionKey(roundID)] = Subscription{
				Query:     RoundSubscription,
				Variables: map[string]any{"id": strconv.Itoa(roundID)},
			}
		}
	}

	return wanted
}
