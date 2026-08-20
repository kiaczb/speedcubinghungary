package special

import (
	"context"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	waitTimeout = 2 * time.Second
	waitTick    = 10 * time.Millisecond
	// Annyi ideig figyeljük, hogy NEM történik meg valami. Rövid, mert a
	// munkamenet leállása azonnali lenne, nem időzített.
	neverWindow = 150 * time.Millisecond
)

func newTestTracker(liveURL string) *Tracker {
	return NewTracker(NewLiveClient(liveURL), NewSocketClient(liveURL), NewBroadcaster())
}

// A Special pontszáma minden versenyszám összege, az oldalt viszont tipikusan
// verseny KÖZBEN nyitják meg. A már lezárult versenyszámoknak (a reggeli 2x2)
// ezért ugyanúgy be kell kerülniük, nem csak a megnyitás után nyíló köröknek.
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

// A második lekérdezés a memóriából jön: a WCA Live-ot nem szólítjuk meg
// minden egyes kérésnél.
func TestTracker_CachesBetweenRequests(t *testing.T) {
	server := fakeWCALive(t, fakeRound{
		id: "1", eventID: "333", number: 1,
		results: []liveResult{result("11", 7, "Teszt Elek", attempts(900))},
	})

	tracker := newTestTracker(server.URL)

	first, err := tracker.Results(context.Background(), testCompetition())
	require.NoError(t, err)

	// A hamis szervert leállítjuk: ha most is lekérdezne, hibát kapnánk.
	server.Close()

	second, err := tracker.Results(context.Background(), testCompetition())
	require.NoError(t, err)
	assert.Equal(t, first, second)
}

// Másik verseny adatát nem adjuk vissza a korábbi gyorsítótárból: a versenyt a
// néző választja, egyszerre több is nyitva lehet.
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

// A feliratkozás nézőhöz kötött: az első nézőnél indul, az utolsó után leáll.
// Ez váltja ki a korábbi, folyamatosan futó háttérfigyelőt.
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

// Több néző: a munkamenet az UTOLSÓ távozásáig él.
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

// A nézet lezárása többször is hívható (a defer és egy korai visszatérés is
// meghívhatja) — a számláló nem csúszhat el.
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

	// Előbb megvárjuk a feliratkozást, különben a "még nincs" állapotot
	// néznénk "már leállt"-nak.
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

// Csak arra a versenyre iratkozunk fel, amit tényleg néz valaki — a
// memóriában tartott, de nézetlen verseny köreiért nem tartunk kapcsolatot.
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

// A versenyt bárki megnyithatja, ezért felső korlát van a párhuzamos
// figyelésen: a nézetlen versenyek kiesnek, a nézetteket viszont nem dobjuk el.
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

	// Amint egy verseny nézők nélkül marad, a helyére bejöhet a következő.
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
