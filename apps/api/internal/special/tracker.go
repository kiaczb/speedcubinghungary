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
	refreshInterval = 2 * time.Minute

	cacheTTL = 15 * time.Second

	idleRetention = 5 * time.Minute

	maxTrackedCompetitions = 20

	roundFetchConcurrency = 4
)

var ErrTooManyCompetitions = errors.New("too many competitions are being watched at once")

type Tracker struct {
	client      *LiveClient
	socket      *SocketClient
	broadcaster *Broadcaster

	mu           sync.Mutex
	competitions map[string]*trackedCompetition

	viewers        int
	sessionRunning bool
	stopSession    context.CancelFunc
	sessionCtx     context.Context
	sessionDone    chan struct{}
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
		socket:       socket,
		broadcaster:  broadcaster,
		competitions: make(map[string]*trackedCompetition),
	}

	socket.SetOnConnect(func() {
		log.Printf("[special] socket (re)connected, resyncing rounds")
		tracker.refresh(tracker.currentSessionContext())
	})

	return tracker
}

func (t *Tracker) currentSessionContext() context.Context {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.sessionCtx == nil {
		return context.Background()
	}
	return t.sessionCtx
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

			t.syncSubscriptions()
		})
	}, nil
}

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
		if previous != nil {
			<-previous
		}
		t.runSession(ctx)
	}()
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

	live := t.sessionRunning && entry.viewers > 0 && entry.loadedAt.After(entry.trackedSince)
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

func (t *Tracker) runSession(ctx context.Context) {
	log.Printf("[special] live session starting")

	socketDone := make(chan struct{})
	go func() {
		defer close(socketDone)
		t.socket.Run(ctx)
	}()

	defer func() {
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
