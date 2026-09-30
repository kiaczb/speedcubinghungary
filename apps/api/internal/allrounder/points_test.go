package allrounder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testPointsTable = []int{10, 6, 3}

func results(rows ...Result) []Result {
	for index := range rows {
		rows[index].Ordinal = index
	}
	return rows
}

func TestComputeStandingsLegacy_AwardsPointsByPosition(t *testing.T) {
	standings := ComputeStandingsLegacy([]competitionResults{{
		CompetitionWCAID: "C1",
		Results: results(
			Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
			Result{WCAID: "2020BBBB01", Name: "Bence", EventID: "333", RoundTypeID: "f", Pos: 2, Best: 800},
			Result{WCAID: "2020CCCC01", Name: "Cili", EventID: "333", RoundTypeID: "f", Pos: 3, Best: 900},
			Result{WCAID: "2020DDDD01", Name: "Dani", EventID: "333", RoundTypeID: "f", Pos: 4, Best: 1000},
		),
	}}, testPointsTable, 10)

	require.Len(t, standings, 3)
	assert.Equal(t, "2020AAAA01", standings[0].WCAID)
	assert.Equal(t, 10, standings[0].TotalPoints)
	assert.Equal(t, 6, standings[1].TotalPoints)
	assert.Equal(t, 3, standings[2].TotalPoints)
}

func TestComputeStandingsLegacy_LaterRoundWins(t *testing.T) {
	standings := ComputeStandingsLegacy([]competitionResults{{
		CompetitionWCAID: "C1",
		Results: results(
			Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "1", Pos: 3, Best: 900},
			Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
		),
	}}, testPointsTable, 10)

	require.Len(t, standings, 1)
	assert.Equal(t, 10, standings[0].TotalPoints, "a döntőbeli 1. helyért járó pont")
	require.Len(t, standings[0].Results, 1, "versenyszámonként csak egy eredmény számít")
	assert.Equal(t, 1, standings[0].Results[0].Ranking)
}

func TestComputeStandingsLegacy_DNFShiftsPointsDown(t *testing.T) {
	standings := ComputeStandingsLegacy([]competitionResults{{
		CompetitionWCAID: "C1",
		Results: results(
			Result{WCAID: "2020DDDD01", Name: "Dani", EventID: "333", RoundTypeID: "1", Pos: 4, Best: 1000},
			Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
			Result{WCAID: "2020XXXX01", Name: "Xavér", EventID: "333", RoundTypeID: "f", Pos: 2, Best: -1},
		),
	}}, testPointsTable, 10)

	byID := make(map[string]PersonStandings, len(standings))
	for _, person := range standings {
		byID[person.WCAID] = person
	}

	assert.Equal(t, 10, byID["2020AAAA01"].TotalPoints)
	assert.NotContains(t, byID, "2020XXXX01", "a DNF nem ér pontot")
	assert.Equal(t, 3, byID["2020DDDD01"].TotalPoints,
		"a 4. hely az egy érvénytelen eredmény miatt a 3. helyért járó pontot kapja")
}

func TestComputeStandingsLegacy_KeepsOnlyBestCountingResults(t *testing.T) {
	standings := ComputeStandingsLegacy([]competitionResults{{
		CompetitionWCAID: "C1",
		Results: results(
			Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "222", RoundTypeID: "f", Pos: 3, Best: 300},
			Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
			Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "444", RoundTypeID: "f", Pos: 2, Best: 3000},
		),
	}}, testPointsTable, 2)

	require.Len(t, standings, 1)
	assert.Equal(t, 16, standings[0].TotalPoints, "csak a két legjobb eredmény számít")
	require.Len(t, standings[0].Results, 2)
}

func TestComputeStandingsLegacy_ScoresEachCompetitionSeparately(t *testing.T) {
	standings := ComputeStandingsLegacy([]competitionResults{
		{CompetitionWCAID: "C1", Results: results(
			Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
		)},
		{CompetitionWCAID: "C2", Results: results(
			Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 2, Best: 750},
		)},
	}, testPointsTable, 10)

	require.Len(t, standings, 1)
	assert.Equal(t, 16, standings[0].TotalPoints)
	require.Len(t, standings[0].Results, 2)
	assert.Equal(t, "C1", standings[0].Results[0].CompetitionWCAID)
	assert.Equal(t, "C2", standings[0].Results[1].CompetitionWCAID)
}

func TestComputeStandingsLegacy_SkipsResultsWithoutWCAID(t *testing.T) {
	standings := ComputeStandingsLegacy([]competitionResults{{
		CompetitionWCAID: "C1",
		Results: results(
			Result{WCAID: "", Name: "Újonc", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
			Result{WCAID: "", Name: "Másik újonc", EventID: "333", RoundTypeID: "f", Pos: 2, Best: 800},
		),
	}}, testPointsTable, 10)

	assert.Empty(t, standings)
}

func TestComputeStandingsLegacy_TiesAreDeterministic(t *testing.T) {
	standings := ComputeStandingsLegacy([]competitionResults{{
		CompetitionWCAID: "C1",
		Results: results(
			Result{WCAID: "2020ZZZZ01", Name: "Zsolt", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
			Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "222", RoundTypeID: "f", Pos: 1, Best: 300},
		),
	}}, testPointsTable, 10)

	require.Len(t, standings, 2)
	assert.Equal(t, "Anna", standings[0].Name)
	assert.Equal(t, "Zsolt", standings[1].Name)
}
