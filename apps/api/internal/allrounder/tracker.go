package allrounder

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
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
	socket      *SocketClient
	broadcaster *Broadcaster

	mu      sync.Mutex
	tracked *trackedCompetition

	viewers        int
	sessionRunning bool
	stopSession    context.CancelFunc
	sessionCtx     context.Context
	sessionDone    chan struct{}
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
		socket:      socket,
		broadcaster: broadcaster,
	}

	socket.SetOnConnect(func() {
		log.Printf("[allrounder] socket (re)connected, resyncing rounds")
		tracker.refresh(tracker.currentSessionContext())
	})

	return tracker
}

func (t *LiveTracker) currentSessionContext() context.Context {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sessionCtx == nil {
		return context.Background()
	}
	return t.sessionCtx
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
	startSession := t.viewers == 1
	if startSession {
		entry.trackedSince = time.Now()
		t.startSessionLocked()
	}
	sessionCtx := t.sessionCtx
	t.mu.Unlock()

	if !startSession {
		go t.ensureLoaded(sessionCtx, entry)
	}

	var once sync.Once
	return func() {
		once.Do(func() {
			t.mu.Lock()
			t.viewers--
			last := t.viewers == 0
			stop := t.stopSession
			if last {
				t.sessionRunning = false
				t.stopSession = nil
				t.sessionCtx = nil
				if t.tracked != nil {
					t.tracked.trackedSince = time.Time{}
				}
			}
			t.mu.Unlock()

			if last && stop != nil {
				stop()
			}
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

	live := t.sessionRunning && t.viewers > 0 && entry.loadedAt.After(entry.trackedSince)
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
				return
			}

			mu.Lock()
			rounds[identity.LiveRoundID] = roundResultsFromLive(identity, *round)
			mu.Unlock()
		}(identity)
	}

	wg.Wait()

	t.mu.Lock()
	entry.identities = byID
	entry.rounds = rounds
	entry.loadedAt = time.Now()
	collected := collectRounds(entry)
	sessionRunning := t.sessionRunning
	t.mu.Unlock()

	if sessionRunning {
		t.syncSubscriptions()
	}

	return collected, nil
}

func (t *LiveTracker) ensureLoaded(ctx context.Context, entry *trackedCompetition) {
	if ctx == nil {
		return
	}

	if _, fresh := t.cachedRounds(entry); fresh {
		t.syncSubscriptions()
		return
	}

	if _, err := t.load(ctx, entry); err != nil {
		log.Printf("[allrounder] failed to load competition %s: %v", entry.competitionWCAID, err)
	}
}

func (t *LiveTracker) startSessionLocked() {
	previous := t.sessionDone

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	t.stopSession = cancel
	t.sessionCtx = ctx
	t.sessionDone = done
	t.sessionRunning = true

	go func() {
		defer close(done)
		if previous != nil {
			<-previous
		}
		t.runSession(ctx)
	}()
}

func (t *LiveTracker) runSession(ctx context.Context) {
	log.Printf("[allrounder] live session starting")

	socketDone := make(chan struct{})
	go func() {
		defer close(socketDone)
		t.socket.Run(ctx)
	}()

	defer func() {
		<-socketDone
		t.clearSubscriptions()
		log.Printf("[allrounder] live session stopped")
	}()

	t.refresh(ctx)

	ticker := time.NewTicker(liveRefreshInterval)
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

func (t *LiveTracker) refresh(ctx context.Context) {
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

func (t *LiveTracker) handleEvent(event SubscriptionEvent) {
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

func (t *LiveTracker) syncSubscriptions() {
	t.mu.Lock()
	wanted := make(map[string]int)
	if t.tracked != nil && t.viewers > 0 {
		for roundID := range t.tracked.identities {
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
		t.socket.Subscribe(key, RoundRankingSubscription, map[string]any{
			"id": strconv.Itoa(roundID),
		})
	}
}

func (t *LiveTracker) clearSubscriptions() {
	for _, key := range t.socket.Subscriptions() {
		t.socket.Unsubscribe(key)
	}
}

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
