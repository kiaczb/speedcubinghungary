package special

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmoke_AgainstRealWCALive(t *testing.T) {
	competitionWCAID := os.Getenv("WCA_LIVE_SMOKE")
	if competitionWCAID == "" {
		t.Skip("set WCA_LIVE_SMOKE=<competitionWcaId> to run against the real WCA Live")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := NewLiveClient(DefaultWCALiveURL)

	liveID, err := client.ResolveCompetitionID(ctx, competitionWCAID)
	require.NoError(t, err, "a /link/ átirányításból kell jönnie a numerikus id-nak")
	t.Logf("competition %s -> live id %d", competitionWCAID, liveID)

	competition, err := client.GetCompetition(ctx, liveID)
	require.NoError(t, err)
	require.NotEmpty(t, competition.CompetitionEvents, "a versenynek vannak eseményei")

	var (
		roundID   int
		eventID   string
		roundNum  int
		foundOpen bool
	)
	for _, competitionEvent := range competition.CompetitionEvents {
		for _, round := range competitionEvent.Rounds {
			if !round.Open {
				continue
			}
			id, err := strconv.Atoi(round.ID)
			require.NoError(t, err)
			roundID, eventID, roundNum, foundOpen = id, competitionEvent.Event.ID, round.Number, true
			break
		}
		if foundOpen {
			break
		}
	}
	require.True(t, foundOpen, "kell legyen legalább egy megnyitott kör")

	round, err := client.GetRound(ctx, roundID)
	require.NoError(t, err)
	t.Logf("round %s: %d results", RoundWCIFID(eventID, roundNum), len(round.Results))

	snapshots := snapshotResults(RoundIdentity{
		LiveRoundID: roundID,
		EventID:     eventID,
		RoundNumber: roundNum,
	}, *round)
	t.Logf("%d results in the snapshot", len(snapshots))

	for _, snapshot := range snapshots {
		assert.NotEmpty(t, snapshot.Attempts, "üres eredmény nem kerül a pillanatképbe")
	}

	socket := NewSocketClient(DefaultWCALiveURL)
	connected := make(chan struct{}, 1)
	socket.SetOnConnect(func() {
		select {
		case connected <- struct{}{}:
		default:
		}
	})

	socket.Subscribe(subscriptionKey(roundID), RoundSubscription, map[string]any{
		"id": strconv.Itoa(roundID),
	})
	go socket.Run(ctx)

	select {
	case <-connected:
		t.Logf("subscribed to roundUpdated(id: %d) on the real WCA Live socket", roundID)
	case <-ctx.Done():
		t.Fatal("could not subscribe to the real WCA Live socket")
	}
}

func TestSmokeDatabase_EndToEnd(t *testing.T) {
	databaseURL := os.Getenv("SMOKE_DATABASE_URL")
	competitionWCAID := os.Getenv("WCA_LIVE_SMOKE")
	if databaseURL == "" || competitionWCAID == "" {
		t.Skip("set SMOKE_DATABASE_URL and WCA_LIVE_SMOKE to run the end-to-end smoke test")
	}

	db, err := sql.Open("pgx", databaseURL)
	require.NoError(t, err)
	defer db.Close()
	require.NoError(t, db.Ping())

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	repo := NewPostgresRepository(db)
	client := NewLiveClient(DefaultWCALiveURL)
	broadcaster := NewBroadcaster()
	tracker := NewTracker(client, NewSocketClient(DefaultWCALiveURL), broadcaster)
	service := NewService(repo, client, tracker, broadcaster)

	userID, _ := strconv.Atoi(os.Getenv("SMOKE_USER_ID"))
	require.Positive(t, userID, "set SMOKE_USER_ID to an existing users.id")

	tracked, err := service.ResolveCompetition(ctx, competitionWCAID)
	require.NoError(t, err)
	t.Logf("watching %s (live id %d)", tracked.CompetitionWCAID, tracked.LiveCompetitionID)

	report, err := service.GetReport(ctx, competitionWCAID, nil)
	require.NoError(t, err)
	require.NotEmpty(t, report.Competitors, "a versenyzői nézet nem lehet üres")

	t.Logf("%d competitors, %d DNF attempts",
		len(report.Competitors), countDNFAttempts(report.Competitors))

	var (
		liveResultID  int
		attemptNumber int
	)
	for _, competitor := range report.Competitors {
		for _, round := range competitor.Rounds {
			for _, attempt := range round.Attempts {
				assert.NotEqual(t, 0, attempt.AttemptNumber)

				if attempt.Result == AttemptDNF && attempt.ElapsedCs != nil {
					t.Fatalf("a WCA Live nem ad tényleges időt, mégis ki van töltve: %v", attempt)
				}

				if attempt.Result == AttemptDNF && liveResultID == 0 {
					liveResultID = round.LiveResultID
					attemptNumber = attempt.AttemptNumber
				}
			}
		}
	}
	require.Positive(t, liveResultID, "kell legyen legalább egy DNF kísérlet")

	const elapsed = 3742
	value := elapsed
	require.NoError(t, service.RecordElapsed(ctx, competitionWCAID, liveResultID, attemptNumber, &value, "smoke test", userID))

	after, err := service.GetReport(ctx, competitionWCAID, nil)
	require.NoError(t, err)
	require.NotNil(t, findAttempt(after, liveResultID, attemptNumber).ElapsedCs,
		"a rögzített időnek meg kell jelennie a nézetben")
	assert.Equal(t, elapsed, *findAttempt(after, liveResultID, attemptNumber).ElapsedCs)

	require.NoError(t, service.RecordElapsed(ctx, competitionWCAID, liveResultID, attemptNumber, &value, "smoke test", userID))

	recorded, err := repo.ListRecordedTimes(tracked.CompetitionWCAID)
	require.NoError(t, err)
	assert.Len(t, recorded, 1, "az ismételt rögzítés nem duplikálhat")

	require.NoError(t, service.RecordElapsed(ctx, competitionWCAID, liveResultID, attemptNumber, nil, "", userID))

	recorded, err = repo.ListRecordedTimes(tracked.CompetitionWCAID)
	require.NoError(t, err)
	assert.Empty(t, recorded, "a null idő törli a bejegyzést")

	_, err = db.Exec(`DELETE FROM special_competitions WHERE competition_wca_id = $1`,
		tracked.CompetitionWCAID)
	require.NoError(t, err)
}

func findAttempt(report *Report, liveResultID, attemptNumber int) CompetitorAttempt {
	for _, competitor := range report.Competitors {
		for _, round := range competitor.Rounds {
			if round.LiveResultID != liveResultID {
				continue
			}
			for _, attempt := range round.Attempts {
				if attempt.AttemptNumber == attemptNumber {
					return attempt
				}
			}
		}
	}
	return CompetitorAttempt{}
}

func countDNFAttempts(competitors []Competitor) int {
	count := 0
	for _, competitor := range competitors {
		for _, round := range competitor.Rounds {
			for _, attempt := range round.Attempts {
				if attempt.Result == AttemptDNF {
					count++
				}
			}
		}
	}
	return count
}

func TestSmokeTracker_LoadsWholeCompetition(t *testing.T) {
	competitionWCAID := os.Getenv("WCA_LIVE_SMOKE")
	if competitionWCAID == "" {
		t.Skip("set WCA_LIVE_SMOKE=<competitionWcaId> to run against the real WCA Live")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := NewLiveClient(DefaultWCALiveURL)
	liveID, err := client.ResolveCompetitionID(ctx, competitionWCAID)
	require.NoError(t, err)

	tracker := NewTracker(client, NewSocketClient(DefaultWCALiveURL), NewBroadcaster())

	started := time.Now()
	results, err := tracker.Results(ctx, Competition{
		CompetitionWCAID:  competitionWCAID,
		LiveCompetitionID: liveID,
	})
	require.NoError(t, err)
	require.NotEmpty(t, results, "a versenynek vannak eredményei")

	rounds := make(map[string]int)
	for _, result := range results {
		rounds[result.RoundWCIFID]++
	}
	t.Logf("%d results across %d rounds in %s", len(results), len(rounds), time.Since(started))

	require.Greater(t, len(rounds), 1,
		"a Special minden versenyszáma kell, nem csak egy kör")

	cachedAt := time.Now()
	cached, err := tracker.Results(ctx, Competition{
		CompetitionWCAID:  competitionWCAID,
		LiveCompetitionID: liveID,
	})
	require.NoError(t, err)
	assert.Len(t, cached, len(results))
	assert.Less(t, time.Since(cachedAt), 50*time.Millisecond, "a gyorsítótárnak azonnalinak kell lennie")
}

func TestSmokeCompetitions_AgainstRealWCALive(t *testing.T) {
	competitionWCAID := os.Getenv("WCA_LIVE_SMOKE")
	if competitionWCAID == "" {
		t.Skip("set WCA_LIVE_SMOKE=<competitionWcaId> to run against the real WCA Live")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := NewLiveClient(DefaultWCALiveURL)
	service := NewService(nil, client, nil, NewBroadcaster())

	found, err := service.Competitions(ctx, "Hungarian")
	require.NoError(t, err)
	require.NotEmpty(t, found, "a névre keresésnek adnia kell találatot")

	for _, competition := range found {
		assert.NotEmpty(t, competition.CompetitionWCAID, "a wcaId nélkül nem megnyitható a verseny")
		assert.Positive(t, competition.LiveCompetitionID, "a numerikus id kell a lekérdezéshez")
	}
	t.Logf("%d competitions matched %q, first: %s", len(found), "Hungarian", found[0].Name)

	local, err := service.Competitions(ctx, "")
	require.NoError(t, err)
	for _, competition := range local {
		assert.Equal(t, defaultCountryISO2, competition.CountryISO2,
			"az alaplistába csak magyar verseny kerülhet")
	}
	t.Logf("%d Hungarian competitions in the picker", len(local))

	competition, err := service.ResolveCompetition(ctx, competitionWCAID)
	require.NoError(t, err)
	assert.Equal(t, competitionWCAID, competition.CompetitionWCAID)
	assert.Positive(t, competition.LiveCompetitionID)
	assert.NotEmpty(t, competition.Name)
}
