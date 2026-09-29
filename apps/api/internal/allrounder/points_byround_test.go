package allrounder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func leg(competitionWCAID string, rows ...Result) CompetitionRounds {
	return CompetitionRounds{
		CompetitionWCAID: competitionWCAID,
		Rounds:           RoundsFromResults(results(rows...)),
	}
}

func TestComputeStandingsByRound_AwardsPointsByPosition(t *testing.T) {
	standings := ComputeStandingsByRound([]CompetitionRounds{leg("C1",
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
		Result{WCAID: "2020BBBB01", Name: "Bence", EventID: "333", RoundTypeID: "f", Pos: 2, Best: 800},
		Result{WCAID: "2020CCCC01", Name: "Cili", EventID: "333", RoundTypeID: "f", Pos: 3, Best: 900},
		Result{WCAID: "2020DDDD01", Name: "Dani", EventID: "333", RoundTypeID: "f", Pos: 4, Best: 1000},
	)}, testPointsTable, 10)

	require.Len(t, standings, 3)
	assert.Equal(t, "2020AAAA01", standings[0].WCAID)
	assert.Equal(t, 10, standings[0].TotalPoints)
	assert.Equal(t, 6, standings[1].TotalPoints)
	assert.Equal(t, 3, standings[2].TotalPoints)
}

func TestComputeStandingsByRound_LaterRoundWins(t *testing.T) {
	standings := ComputeStandingsByRound([]CompetitionRounds{leg("C1",
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "1", Pos: 3, Best: 900},
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
	)}, testPointsTable, 10)

	require.Len(t, standings, 1)
	assert.Equal(t, 10, standings[0].TotalPoints, "a döntőbeli 1. helyért járó pont")
	require.Len(t, standings[0].Results, 1, "versenyszámonként csak egy eredmény számít")
	assert.Equal(t, 1, standings[0].Results[0].Ranking)
}

func TestComputeStandingsByRound_DNFInFinalFallsBackToEarlierRound(t *testing.T) {
	standings := ComputeStandingsByRound([]CompetitionRounds{leg("C1",
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "1", Pos: 2, Best: 900},
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 3, Best: -1},
	)}, testPointsTable, 10)

	require.Len(t, standings, 1)
	assert.Equal(t, 6, standings[0].TotalPoints, "a selejtezőbeli 2. helyért járó pont")
	assert.Equal(t, 2, standings[0].Results[0].Ranking)
}

func TestComputeStandingsByRound_DNFDoesNotShiftPointsUp(t *testing.T) {
	standings := ComputeStandingsByRound([]CompetitionRounds{leg("C1",
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "1", Pos: 1, Best: 700},
		Result{WCAID: "2020BBBB01", Name: "Bence", EventID: "333", RoundTypeID: "1", Pos: 2, Best: 800},
		Result{WCAID: "2020CCCC01", Name: "Cili", EventID: "333", RoundTypeID: "1", Pos: 3, Best: 900},
		Result{WCAID: "2020DDDD01", Name: "Dani", EventID: "333", RoundTypeID: "1", Pos: 4, Best: 1000},
		Result{WCAID: "2020EEEE01", Name: "Eszter", EventID: "333", RoundTypeID: "1", Pos: 5, Best: -1},
		Result{WCAID: "2020FFFF01", Name: "Feri", EventID: "333", RoundTypeID: "1", Pos: 5, Best: -1},
	)}, testPointsTable, 10)

	require.Len(t, standings, 3, "csak az első három helyezett kap pontot")
	for _, standing := range standings {
		assert.NotEqual(t, "2020DDDD01", standing.WCAID,
			"a 4. hely a 3 fős pontozott mezőnyön kívül van, DNF-ek ellenére sem kap pontot")
	}
	assert.Equal(t, 3, standings[2].TotalPoints, "a 3. hely pontja változatlan")
}

func TestComputeStandingsByRound_IgnoresResultsWithoutTime(t *testing.T) {
	standings := ComputeStandingsByRound([]CompetitionRounds{{
		CompetitionWCAID: "C1",
		Rounds: []RoundResults{{
			EventID: "333", Rank: 1,
			Placements: []Placement{
				{WCAID: "2020AAAA01", Name: "Anna", Pos: 1, Best: 700},
				{WCAID: "2020BBBB01", Name: "Bence", Pos: 2, Best: 0},
			},
		}},
	}}, testPointsTable, 10)

	require.Len(t, standings, 1)
	assert.Equal(t, "2020AAAA01", standings[0].WCAID)
}

func TestComputeStandingsByRound_ScoresEachCompetitionSeparately(t *testing.T) {
	standings := ComputeStandingsByRound([]CompetitionRounds{
		leg("C1", Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700}),
		leg("C2", Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 2, Best: 800}),
	}, testPointsTable, 10)

	require.Len(t, standings, 1)
	assert.Equal(t, 16, standings[0].TotalPoints, "ugyanaz a versenyszám két állomáson külön pontot ér")
	assert.Len(t, standings[0].Results, 2)
}

func TestComputeStandingsByRound_KeepsOnlyBestCountingResults(t *testing.T) {
	standings := ComputeStandingsByRound([]CompetitionRounds{leg("C1",
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "222", RoundTypeID: "f", Pos: 2, Best: 300},
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "444", RoundTypeID: "f", Pos: 3, Best: 2000},
	)}, testPointsTable, 2)

	require.Len(t, standings, 1)
	assert.Equal(t, 16, standings[0].TotalPoints, "csak a két legjobb eredmény számít")
	assert.Len(t, standings[0].Results, 2)
}

func TestComputeStandingsByRound_IsIndependentOfInputOrder(t *testing.T) {
	forward := ComputeStandingsByRound([]CompetitionRounds{leg("C1",
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "1", Pos: 2, Best: 900},
		Result{WCAID: "2020BBBB01", Name: "Bence", EventID: "333", RoundTypeID: "1", Pos: 1, Best: 800},
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
		Result{WCAID: "2020BBBB01", Name: "Bence", EventID: "333", RoundTypeID: "f", Pos: 2, Best: 750},
	)}, testPointsTable, 10)

	shuffled := ComputeStandingsByRound([]CompetitionRounds{leg("C1",
		Result{WCAID: "2020BBBB01", Name: "Bence", EventID: "333", RoundTypeID: "f", Pos: 2, Best: 750},
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "1", Pos: 2, Best: 900},
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
		Result{WCAID: "2020BBBB01", Name: "Bence", EventID: "333", RoundTypeID: "1", Pos: 1, Best: 800},
	)}, testPointsTable, 10)

	assert.Equal(t, forward, shuffled)
}
