package allrounder

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func liveResultOf(wcaID, name string, ranking, best *int, attempts int) liveResult {
	result := liveResult{Ranking: ranking, Best: best}
	result.Person.WCAID = wcaID
	result.Person.Name = name
	for range attempts {
		result.Attempts = append(result.Attempts, struct {
			Result int `json:"result"`
		}{Result: 700})
	}

	return result
}

func intPtr(value int) *int { return &value }

func TestRoundResultsFromLive_MapsRankingToPosition(t *testing.T) {
	round := roundResultsFromLive(
		roundIdentity{LiveRoundID: 1, EventID: "333", RoundNumber: 2},
		liveRound{Results: []liveResult{
			liveResultOf("2020BBBB01", "Bence", intPtr(2), intPtr(800), 5),
			liveResultOf("2020AAAA01", "Anna", intPtr(1), intPtr(700), 5),
		}},
	)

	assert.Equal(t, "333", round.EventID)
	assert.Equal(t, 2, round.Rank, "a WCA Live körsorszáma közvetlenül a kör rangja")

	require.Len(t, round.Placements, 2)
	assert.Equal(t, "2020AAAA01", round.Placements[0].WCAID, "helyezés szerint rendezve")
	assert.Equal(t, 1, round.Placements[0].Pos)
	assert.Equal(t, 700, round.Placements[0].Best)
}

func TestRoundResultsFromLive_SkipsCompetitorsWithoutUsableResult(t *testing.T) {
	round := roundResultsFromLive(
		roundIdentity{LiveRoundID: 1, EventID: "333", RoundNumber: 1},
		liveRound{Results: []liveResult{
			liveResultOf("2020AAAA01", "Anna", intPtr(1), intPtr(700), 5),
			liveResultOf("2020BBBB01", "Bence", nil, intPtr(0), 0),
			liveResultOf("2020CCCC01", "Cili", intPtr(2), intPtr(0), 0),
			liveResultOf("", "Dani", intPtr(3), intPtr(900), 5),
		}},
	)

	require.Len(t, round.Placements, 1)
	assert.Equal(t, "2020AAAA01", round.Placements[0].WCAID)
}

func TestRoundResultsFromLive_KeepsDNFAsInvalid(t *testing.T) {
	round := roundResultsFromLive(
		roundIdentity{LiveRoundID: 1, EventID: "333", RoundNumber: 1},
		liveRound{Results: []liveResult{
			liveResultOf("2020AAAA01", "Anna", intPtr(1), intPtr(-1), 5),
		}},
	)

	require.Len(t, round.Placements, 1)
	assert.Equal(t, -1, round.Placements[0].Best)
	assert.Equal(t, 0, pointsForPlacement(round.Placements[0], testPointsTable))
}

func TestCurrentLeg_PicksTheCompetitionRunningToday(t *testing.T) {
	series := ResolveSeries(2026)
	first := series.Competitions[0].CompetitionWCAID
	second := series.Competitions[1].CompetitionWCAID

	day := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
	states := map[string]SyncState{
		first:  {CompetitionWCAID: first, ResultCount: 280, StartDate: &day, EndDate: &day},
		second: {CompetitionWCAID: second, ResultCount: 0, StartDate: &day, EndDate: &day},
	}

	competition, _, found := currentLeg(series, states, day.Add(11*time.Hour))

	require.True(t, found)
	assert.Equal(t, second, competition.CompetitionWCAID)
}

func TestCurrentLeg_IgnoresCompetitionsOutsideTheirDay(t *testing.T) {
	series := ResolveSeries(2026)
	id := series.Competitions[0].CompetitionWCAID

	day := time.Date(2026, 3, 14, 0, 0, 0, 0, time.UTC)
	states := map[string]SyncState{
		id: {CompetitionWCAID: id, ResultCount: 0, StartDate: &day, EndDate: &day},
	}

	_, _, tooEarly := currentLeg(series, states, day.Add(-72*time.Hour))
	assert.False(t, tooEarly, "a verseny előtt jóval nem nyitunk WCA Live kapcsolatot")

	_, _, tooLate := currentLeg(series, states, day.Add(120*time.Hour))
	assert.False(t, tooLate, "napokkal később már nem")
}

func TestCurrentLeg_IgnoresCompetitionsWithoutKnownDate(t *testing.T) {
	series := ResolveSeries(2026)
	id := series.Competitions[0].CompetitionWCAID

	states := map[string]SyncState{
		id: {CompetitionWCAID: id, ResultCount: 0},
	}

	_, _, found := currentLeg(series, states, time.Now())
	assert.False(t, found)
}

func TestComputeStandingsByRound_MergesLiveLegWithOfficialOnes(t *testing.T) {
	official := leg("C1",
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 2, Best: 800},
	)

	live := CompetitionRounds{
		CompetitionWCAID: "C2",
		Rounds: []RoundResults{roundResultsFromLive(
			roundIdentity{LiveRoundID: 9, EventID: "333", RoundNumber: 1},
			liveRound{Results: []liveResult{
				liveResultOf("2020AAAA01", "Anna", intPtr(1), intPtr(700), 5),
			}},
		)},
	}

	standings := ComputeStandingsByRound([]CompetitionRounds{official, live}, testPointsTable, 10)

	require.Len(t, standings, 1)
	assert.Equal(t, 16, standings[0].TotalPoints, "6 pont a lezajlott + 10 az élő állomásról")
	require.Len(t, standings[0].Results, 2)
	assert.Equal(t, "C2", standings[0].Results[1].CompetitionWCAID)
}
