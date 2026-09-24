package allrounder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"
)

const (
	liveCacheTTL         = 15 * time.Second
	liveRefreshInterval  = 2 * time.Minute
	liveRoundConcurrency = 4
)

type LiveTracker struct {
	client      *LiveClient
	broadcaster *Broadcaster

	// session a WCA Live kapcsolat életciklusa — közös a special csomaggal,
	// lásd internal/wcalive/session.go. Itt CSAK az marad, ami az All-Rounder
	// sajátja: egyszerre EGY állomást követünk (a sorozat dönti el, melyiket),
	// ezért nincs szükség a special többversenyes gépezetére.
	session *Session

	mu      sync.Mutex
	tracked *trackedCompetition
	viewers int
}

type trackedCompetition struct {
	competitionWCAID  string
	liveCompetitionID int
	notifyKey         string

	loadMu sync.Mutex

	identities   map[int]roundIdentity
	rounds       map[int]RoundResults
	loadedAt     time.Time
	trackedSince time.Time
}

func NewLiveTracker(client *LiveClient, socket *SocketClient, broadcaster *Broadcaster) *LiveTracker {
	tracker := &LiveTracker{
		client:      client,
		broadcaster: broadcaster,
		session:     NewSession("allrounder", socket, liveRefreshInterval),
	}

	tracker.session.Watch(tracker)

	return tracker
}

func (t *LiveTracker) Rounds(ctx context.Context, competitionWCAID string, liveCompetitionID, year int) ([]RoundResults, error) {
	entry := t.entry(competitionWCAID, liveCompetitionID, year)

	if rounds, fresh := t.cachedRounds(entry); fresh {
		return rounds, nil
	}

	return t.load(ctx, entry)
}

func (t *LiveTracker) AddViewer(competitionWCAID string, liveCompetitionID, year int) func() {
	entry := t.entry(competitionWCAID, liveCompetitionID, year)

	// ⚠️ A nézőszámot a session INDÍTÁSA ELŐTT kell növelni: az induló session
	// első dolga egy Reload, az pedig néző nélkül nem csinál semmit. Fordított
	// sorrendben a nyitó betöltés kimaradhatna, és az állomás csak a következő
	// (2 perces) frissítéskor jelenne meg.
	t.mu.Lock()
	t.viewers++
	if t.viewers == 1 {
		entry.trackedSince = time.Now()
	}
	t.mu.Unlock()

	releaseSession, startedSession := t.session.AddViewer()

	if !startedSession {
		// A session már futott: a saját állomásunkat külön kell betöltetni.
		go t.ensureLoaded(t.session.Context(), entry)
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			t.mu.Lock()
			t.viewers--
			if t.viewers == 0 && t.tracked != nil {
				t.tracked.trackedSince = time.Time{}
			}
			t.mu.Unlock()

			releaseSession()
		})
	}
}

func (t *LiveTracker) entry(competitionWCAID string, liveCompetitionID, year int) *trackedCompetition {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.tracked != nil && t.tracked.competitionWCAID == competitionWCAID {
		return t.tracked
	}

	t.tracked = &trackedCompetition{
		competitionWCAID:  competitionWCAID,
		liveCompetitionID: liveCompetitionID,
		notifyKey:         strconv.Itoa(year),
		identities:        make(map[int]roundIdentity),
		rounds:            make(map[int]RoundResults),
	}

	return t.tracked
}

func (t *LiveTracker) cachedRounds(entry *trackedCompetition) ([]RoundResults, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if entry.loadedAt.IsZero() {
		return nil, false
	}

	live := t.session.Running() && t.viewers > 0 && entry.loadedAt.After(entry.trackedSince)
	if !live && time.Since(entry.loadedAt) > liveCacheTTL {
		return nil, false
	}

	return collectRounds(entry), true
}

func collectRounds(entry *trackedCompetition) []RoundResults {
	rounds := make([]RoundResults, 0, len(entry.rounds))
	for _, round := range entry.rounds {
		rounds = append(rounds, round)
	}
	sortRounds(rounds)

	return rounds
}

func (t *LiveTracker) load(ctx context.Context, entry *trackedCompetition) ([]RoundResults, error) {
	entry.loadMu.Lock()
	defer entry.loadMu.Unlock()

	if rounds, fresh := t.cachedRounds(entry); fresh {
		return rounds, nil
	}

	identities, err := t.client.OpenRounds(ctx, entry.liveCompetitionID)
	if err != nil {
		return nil, err
	}

	var (
		mu     sync.Mutex
		wg     sync.WaitGroup
		slots  = make(chan struct{}, liveRoundConcurrency)
		rounds = make(map[int]RoundResults, len(identities))
		byID   = make(map[int]roundIdentity, len(identities))
		failed = make(map[int]bool)
	)

	for _, identity := range identities {
		byID[identity.LiveRoundID] = identity

		wg.Add(1)
		go func(identity roundIdentity) {
			defer wg.Done()

			slots <- struct{}{}
			defer func() { <-slots }()

			round, err := t.client.GetRound(ctx, identity.LiveRoundID)
			if err != nil {
				log.Printf("[allrounder] failed to query round %d: %v", identity.LiveRoundID, err)

				mu.Lock()
				failed[identity.LiveRoundID] = true
				mu.Unlock()
				return
			}

			mu.Lock()
			rounds[identity.LiveRoundID] = roundResultsFromLive(identity, *round)
			mu.Unlock()
		}(identity)
	}

	wg.Wait()

	// Ha egyetlen kör sem jött le, az elérhetetlen upstream, nem részleges hiba:
	// ne írjuk felül a meglévő állapotot, és a loadedAt-ot se frissítsük.
	if len(failed) > 0 && len(rounds) == 0 {
		return nil, fmt.Errorf("every round query failed for %s", entry.competitionWCAID)
	}

	t.mu.Lock()
	// ⚠️ Egy kör lekérési hibája NEM törölheti a kör addigi eredményeit: a map
	// teljes cseréjével az élő állomás pontjai látványosan lecsökkennének, majd
	// a következő frissítéskor visszajönnének — a versenyzőnek ez elvesztett
	// pontnak látszik. A hibás körök ezért az ELŐZŐ pillanatképükkel élnek tovább.
	for roundID := range failed {
		if previous, known := entry.rounds[roundID]; known {
			rounds[roundID] = previous
		}
	}

	entry.identities = byID
	entry.rounds = rounds
	entry.loadedAt = time.Now()
	collected := collectRounds(entry)
	t.mu.Unlock()

	if t.session.Running() {
		t.session.SyncSubscriptions()
	}

	return collected, nil
}

func (t *LiveTracker) ensureLoaded(ctx context.Context, entry *trackedCompetition) {
	if ctx == nil {
		return
	}

	if _, fresh := t.cachedRounds(entry); fresh {
		t.session.SyncSubscriptions()
		return
	}

	if _, err := t.load(ctx, entry); err != nil {
		log.Printf("[allrounder] failed to load competition %s: %v", entry.competitionWCAID, err)
	}
}

// Reload a wcalive.Watcher tagja: a session indulásakor, a socket
// (újra)kapcsolódásakor és a frissítési időzítőre hívódik.
func (t *LiveTracker) Reload(ctx context.Context) {
	t.mu.Lock()
	entry := t.tracked
	watched := t.viewers > 0
	t.mu.Unlock()

	if entry == nil || !watched {
		return
	}

	if _, err := t.load(ctx, entry); err != nil {
		log.Printf("[allrounder] failed to refresh competition %s: %v", entry.competitionWCAID, err)
	}
}

// HandleEvent a wcalive.Watcher tagja. Minden roundUpdated esemény TELJES
// kör-pillanatképet hoz, ezért a kimaradt vagy duplikált esemény magától gyógyul.
func (t *LiveTracker) HandleEvent(event SubscriptionEvent) {
	roundID, err := roundIDFromKey(event.Key)
	if err != nil {
		log.Printf("[allrounder] ignoring event with unexpected key %q", event.Key)
		return
	}

	t.mu.Lock()
	entry := t.tracked
	var identity roundIdentity
	watched := false
	if entry != nil {
		identity, watched = entry.identities[roundID]
	}
	t.mu.Unlock()

	if !watched {
		return
	}

	var payload struct {
		RoundUpdated *liveRound `json:"roundUpdated"`
	}
	if err := json.Unmarshal(event.Data, &payload); err != nil || payload.RoundUpdated == nil {
		log.Printf("[allrounder] failed to decode roundUpdated payload for round %d: %v", roundID, err)
		return
	}

	round := roundResultsFromLive(identity, *payload.RoundUpdated)

	t.mu.Lock()
	still := t.tracked == entry
	if still {
		entry.rounds[roundID] = round
	}
	notifyKey := entry.notifyKey
	t.mu.Unlock()

	if still {
		t.broadcaster.Notify(notifyKey)
	}
}

// WantedSubscriptions a wcalive.Watcher tagja: az éppen figyelt állomás nyitott
// köreire akarunk feliratkozni. Néző nélkül üres, így a session elengedi őket.
func (t *LiveTracker) WantedSubscriptions() map[string]Subscription {
	t.mu.Lock()
	defer t.mu.Unlock()

	wanted := make(map[string]Subscription)
	if t.tracked == nil || t.viewers == 0 {
		return wanted
	}

	for roundID := range t.tracked.identities {
		wanted[subscriptionKey(roundID)] = Subscription{
			Query:     RoundRankingSubscription,
			Variables: map[string]any{"id": strconv.Itoa(roundID)},
		}
	}

	return wanted
}
