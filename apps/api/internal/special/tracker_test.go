package special

import (
	"context"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	waitTimeout = 2 * time.Second
	waitTick    = 10 * time.Millisecond
	neverWindow = 150 * time.Millisecond
)

func newTestTracker(liveURL string) *Tracker {
	return NewTracker(NewLiveClient(liveURL), NewSocketClient(liveURL), NewBroadcaster())
}

func TestTracker_LoadsAlreadyFinishedEvents(t *testing.T) {
	server := fakeWCALive(t,
		fakeRound{id: "1", eventID: "222", number: 1, results: []liveResult{
			result("11", 7, "Korán Kelő", attempts(300, 350, 320)),
		}},
		fakeRound{id: "2", eventID: "333", number: 1, results: []liveResult{
			result("12", 7, "Korán Kelő", attempts(900, AttemptDNF, 850)),
		}},
	)

	tracker := newTestTracker(server.URL)

	results, err := tracker.Results(context.Background(), testCompetition())
	require.NoError(t, err)

	rounds := make([]string, 0, len(results))
	for _, result := range results {
		rounds = append(rounds, result.RoundWCIFID)
	}

	assert.ElementsMatch(t, []string{"222-r1", "333-r1"}, rounds,
		"a figyelés megkezdése előtt lezárult versenyszám is kell")
}

func TestTracker_CachesBetweenRequests(t *testing.T) {
	server := fakeWCALive(t, fakeRound{
		id: "1", eventID: "333", number: 1,
		results: []liveResult{result("11", 7, "Teszt Elek", attempts(900))},
	})

	tracker := newTestTracker(server.URL)

	first, err := tracker.Results(context.Background(), testCompetition())
	require.NoError(t, err)

	server.Close()

	second, err := tracker.Results(context.Background(), testCompetition())
	require.NoError(t, err)
	assert.Equal(t, first, second)
}

func TestTracker_CacheIsScopedToCompetition(t *testing.T) {
	server := fakeWCALive(t, fakeRound{
		id: "1", eventID: "333", number: 1,
		results: []liveResult{result("11", 7, "Teszt Elek", attempts(900))},
	})

	tracker := newTestTracker(server.URL)

	_, err := tracker.Results(context.Background(), testCompetition())
	require.NoError(t, err)

	other := testCompetition()
	other.CompetitionWCAID = "MasikVerseny2026"
	other.LiveCompetitionID = 99999

	server.Close()

	_, err = tracker.Results(context.Background(), other)
	assert.Error(t, err, "másik versenyhez újra kell kérdezni")
}

func TestTracker_SubscribesOnlyWhileViewersArePresent(t *testing.T) {
	server := fakeWCALive(t, fakeRound{
		id: "147464", eventID: "333", number: 1,
		results: []liveResult{result("11", 7, "Teszt Elek", attempts(900))},
	})

	tracker := newTestTracker(server.URL)

	assert.Empty(t, tracker.socket.Subscriptions(), "néző nélkül nincs feliratkozás")

	closeViewer, err := tracker.AddViewer(testCompetition())
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		return len(tracker.socket.Subscriptions()) == 1
	}, waitTimeout, waitTick, "a néző megjelenésekor fel kell iratkozni")

	closeViewer()
	assert.Eventually(t, func() bool {
		return len(tracker.socket.Subscriptions()) == 0
	}, waitTimeout, waitTick, "az utolsó néző távozásakor le kell állni")
}

func TestTracker_SessionOutlivesFirstViewer(t *testing.T) {
	server := fakeWCALive(t, fakeRound{
		id: "147464", eventID: "333", number: 1,
		results: []liveResult{result("11", 7, "Teszt Elek", attempts(900))},
	})

	tracker := newTestTracker(server.URL)

	first, err := tracker.AddViewer(testCompetition())
	require.NoError(t, err)
	second, err := tracker.AddViewer(testCompetition())
	require.NoError(t, err)

	assert.Eventually(t, func() bool {
		return len(tracker.socket.Subscriptions()) == 1
	}, waitTimeout, waitTick)

	first()
	assert.Never(t, func() bool {
		return len(tracker.socket.Subscriptions()) == 0
	}, neverWindow, waitTick, "amíg van néző, marad a feliratkozás")

	second()
	assert.Eventually(t, func() bool {
		return len(tracker.socket.Subscriptions()) == 0
	}, waitTimeout, waitTick)
}

func TestTracker_CloseViewerIsIdempotent(t *testing.T) {
	server := fakeWCALive(t, fakeRound{
		id: "147464", eventID: "333", number: 1,
		results: []liveResult{result("11", 7, "Teszt Elek", attempts(900))},
	})

	tracker := newTestTracker(server.URL)

	first, err := tracker.AddViewer(testCompetition())
	require.NoError(t, err)
	second, err := tracker.AddViewer(testCompetition())
	require.NoError(t, err)

	require.Eventually(t, func() bool {
		return len(tracker.socket.Subscriptions()) == 1
	}, waitTimeout, waitTick)

	first()
	first()

	assert.Never(t, func() bool {
		return len(tracker.socket.Subscriptions()) == 0
	}, neverWindow, waitTick, "a kétszer lezárt nézet nem vihet nullába")

	second()
	assert.Eventually(t, func() bool {
		return len(tracker.socket.Subscriptions()) == 0
	}, waitTimeout, waitTick)
}

func TestTracker_SubscribesOnlyToWatchedCompetitions(t *testing.T) {
	tracker := newTestTracker("http://example.invalid")

	watched, err := tracker.entry(testCompetition())
	require.NoError(t, err)
	watched.rounds = map[int]trackedRound{147464: {}}
	watched.viewers = 1

	idle, err := tracker.entry(Competition{CompetitionWCAID: "MasikVerseny2026", LiveCompetitionID: 2})
	require.NoError(t, err)
	idle.rounds = map[int]trackedRound{999: {}}

	tracker.syncSubscriptions()

	assert.Equal(t, []string{subscriptionKey(147464)}, tracker.socket.Subscriptions())
}

func TestTracker_CapsConcurrentlyTrackedCompetitions(t *testing.T) {
	tracker := newTestTracker("http://example.invalid")

	for i := 0; i < maxTrackedCompetitions; i++ {
		entry, err := tracker.entry(Competition{
			CompetitionWCAID:  "Verseny" + strconv.Itoa(i),
			LiveCompetitionID: i + 1,
		})
		require.NoError(t, err)
		entry.viewers = 1
	}

	_, err := tracker.entry(testCompetition())
	assert.ErrorIs(t, err, ErrTooManyCompetitions, "minden helyet néző foglal")

	tracker.competitions["Verseny0"].viewers = 0

	_, err = tracker.entry(testCompetition())
	assert.NoError(t, err)
	assert.NotContains(t, tracker.competitions, "Verseny0")
}

func TestOpenRounds_SkipsUnopenedRounds(t *testing.T) {
	competition := &liveCompetition{
		CompetitionEvents: []liveCompetitionEvnt{{
			Event: liveEvent{ID: "333"},
			Rounds: []liveRound{
				{ID: "1", Number: 1, Open: true, Finished: true},
				{ID: "2", Number: 2, Open: false},
			},
		}},
	}

	rounds := openRounds(competition)

	require.Len(t, rounds, 1, "a meg nem nyitott kört nem figyeljük")
	assert.Equal(t, 1, rounds[0].identity.LiveRoundID)
	assert.Equal(t, "333", rounds[0].identity.EventID)
}

func TestSubscriptionKeyRoundTrip(t *testing.T) {
	roundID, err := roundIDFromKey(subscriptionKey(147464))

	require.NoError(t, err)
	assert.Equal(t, 147464, roundID)

	_, err = roundIDFromKey("nem-ilyen-kulcs")
	assert.Error(t, err)
}

// ⚠️ Egy kör lekérési hibája nem törölheti a kör addigi eredményeit.
//
// A load korábban a TELJES rounds map-et lecserélte, így egy pillanatnyi WCA
// Live hiba eltüntette a kört a riportból — a versenyző kumulatív kerete
// (A1a3) pedig hibás „maradék időt” mutatott, amíg a következő frissítés
// vissza nem hozta.
func TestTracker_KeepsRoundWhenItsQueryFails(t *testing.T) {
	var failing atomic.Bool

	server := fakeWCALive(t,
		fakeRound{id: "1", eventID: "222", number: 1, results: []liveResult{
			result("11", 7, "Teszt Elek", attempts(300, 350, 320)),
		}},
		fakeRound{id: "2", eventID: "333", number: 1, fail: &failing, results: []liveResult{
			result("12", 7, "Teszt Elek", attempts(900, 850, 880)),
		}},
	)

	tracker := newTestTracker(server.URL)

	first, err := tracker.Results(context.Background(), testCompetition())
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"222-r1", "333-r1"}, roundWCIFIDs(first))

	failing.Store(true)

	entry, err := tracker.entry(testCompetition())
	require.NoError(t, err)

	results, err := tracker.load(context.Background(), entry, true)
	require.NoError(t, err, "a részleges hiba nem teszi tönkre a betöltést")

	assert.ElementsMatch(t, []string{"222-r1", "333-r1"}, roundWCIFIDs(results),
		"a hibás kör az előző pillanatképével él tovább")
}

// Ha EGYETLEN kör sem jött le, az nem részleges hiba: ilyenkor hibát adunk, és
// a korábbi állapotot érintetlenül hagyjuk, hogy a következő kérés azonnal
// újrapróbálja ahelyett, hogy üres riportot szolgálnánk ki a cache TTL végéig.
func TestTracker_TotalRoundFailureKeepsPreviousState(t *testing.T) {
	var failing atomic.Bool

	server := fakeWCALive(t, fakeRound{
		id: "1", eventID: "333", number: 1, fail: &failing,
		results: []liveResult{result("11", 7, "Teszt Elek", attempts(900))},
	})

	tracker := newTestTracker(server.URL)

	first, err := tracker.Results(context.Background(), testCompetition())
	require.NoError(t, err)
	require.Len(t, first, 1)

	failing.Store(true)

	entry, err := tracker.entry(testCompetition())
	require.NoError(t, err)

	loadedAt := entry.loadedAt

	_, err = tracker.load(context.Background(), entry, true)
	require.Error(t, err, "teljes kiesésnél hibát kell jelezni")

	assert.Len(t, entry.rounds, 1, "a korábbi pillanatkép megmarad")
	assert.Equal(t, loadedAt, entry.loadedAt,
		"a loadedAt sem frissülhet, különben a cache TTL elnyelné az újrapróbálkozást")
}

func roundWCIFIDs(results []ResultSnapshot) []string {
	seen := make(map[string]bool, len(results))
	ids := make([]string, 0, len(results))

	for _, result := range results {
		if seen[result.RoundWCIFID] {
			continue
		}
		seen[result.RoundWCIFID] = true
		ids = append(ids, result.RoundWCIFID)
	}

	return ids
}
