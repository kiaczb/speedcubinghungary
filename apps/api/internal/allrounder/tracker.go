package allrounder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"speedcubinghungary/internal/wcalive"
)

const (
	liveCacheTTL         = 15 * time.Second
	liveRefreshInterval  = 2 * time.Minute
	liveRoundConcurrency = 4
)

type LiveTracker struct {
	client      *LiveClient
	broadcaster *wcalive.Broadcaster

	session *wcalive.Session

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

func NewLiveTracker(client *LiveClient, socket *wcalive.SocketClient, broadcaster *wcalive.Broadcaster) *LiveTracker {
	tracker := &LiveTracker{
		client:      client,
		broadcaster: broadcaster,
		session:     wcalive.NewSession("allrounder", socket, liveRefreshInterval),
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

	t.mu.Lock()
	t.viewers++
	if t.viewers == 1 {
		entry.trackedSince = time.Now()
	}
	t.mu.Unlock()

	releaseSession, startedSession := t.session.AddViewer()

	if !startedSession {
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

	if len(failed) > 0 && len(rounds) == 0 {
		return nil, fmt.Errorf("every round query failed for %s", entry.competitionWCAID)
	}

	t.mu.Lock()
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

func (t *LiveTracker) HandleEvent(event wcalive.SubscriptionEvent) {
	roundID, err := wcalive.RoundIDFromKey(event.Key)
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

func (t *LiveTracker) WantedSubscriptions() map[string]wcalive.Subscription {
	t.mu.Lock()
	defer t.mu.Unlock()

	wanted := make(map[string]wcalive.Subscription)
	if t.tracked == nil || t.viewers == 0 {
		return wanted
	}

	for roundID := range t.tracked.identities {
		wanted[wcalive.RoundSubscriptionKey(roundID)] = wcalive.Subscription{
			Query:     RoundRankingSubscription,
			Variables: map[string]any{"id": strconv.Itoa(roundID)},
		}
	}

	return wanted
}
