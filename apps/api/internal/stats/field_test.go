package stats

import (
	"testing"

	"speedcubinghungary/internal/competition"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testInfo() CompetitionInfo {
	return CompetitionInfo{
		ID:          "Teszt2026",
		StartDate:   "2026-10-23",
		EndDate:     "2026-10-25",
		CountryISO2: "HU",
		EventIDs:    []string{"333", "222", "333fm"},
	}
}

func testRegistrations() competition.Registrations {
	return competition.Registrations{
		EventIDs: []string{"333", "222", "333fm"},
		Registrants: []competition.Registrant{
			{
				Name: "Régi Róbert", WCAID: "2006REGI01", CountryISO2: "HU", Gender: "m",
				EventIDs: []string{"333", "222", "333fm"},
				PersonalBests: []competition.PersonalBest{
					{EventID: "333", Type: "single", Best: 642, WorldRanking: 900, NationalRanking: 3},
					{EventID: "333fm", Type: "single", Best: 22, WorldRanking: 150, NationalRanking: 1},
				},
			},
			{
				Name: "Külföldi Kata", WCAID: "2016KULF01", CountryISO2: "SK", Gender: "f",
				EventIDs: []string{"333", "222"},
				PersonalBests: []competition.PersonalBest{
					{EventID: "222", Type: "average", Best: 189, WorldRanking: 40, ContinentalRanking: 1, NationalRanking: 1},
				},
			},
			{
				Name: "Új Ubul", WCAID: "", CountryISO2: "HU", Gender: "m",
				EventIDs: []string{"333"},
			},
			{
				Name: "Tízéves Tamás", WCAID: "2016TIZE01", CountryISO2: "HU", Gender: "m",
				EventIDs: []string{"333", "ismeretlen"},
			},
		},
	}
}

func TestComputeField_CountsTheField(t *testing.T) {
	field := computeField(testInfo(), PhaseBefore, testRegistrations(), nil)

	assert.Equal(t, 4, field.Competitors)
	assert.Equal(t, 1, field.Newcomers)
	assert.Equal(t, 1, field.Foreigners)
	assert.Equal(t, []Count{{Key: "HU", Count: 3}, {Key: "SK", Count: 1}}, field.Countries)
	assert.Equal(t, []Count{{Key: "m", Count: 3}, {Key: "f", Count: 1}}, field.Genders)
	assert.Equal(t, []Count{{Key: "333", Count: 4}, {Key: "222", Count: 2}, {Key: "333fm", Count: 1}}, field.Events,
		"a verseny sorrendjében, a versenyen nem szereplő számot kihagyva")
	assert.Equal(t, 1.8, field.AverageEvents)
	assert.Equal(t, []Person{{Name: "Régi Róbert", WCAID: "2006REGI01", CountryISO2: "HU"}}, field.AllEvents)
}

func TestComputeField_RankingsRecordsAndVeterans(t *testing.T) {
	field := computeField(testInfo(), PhaseBefore, testRegistrations(), nil)

	require.Len(t, field.TopRanked, 2)
	assert.Equal(t, "Külföldi Kata", field.TopRanked[0].Person.Name)
	assert.Equal(t, "CR", field.TopRanked[0].Record)
	assert.Equal(t, "1.89", field.TopRanked[0].Display)
	assert.Equal(t, 150, field.TopRanked[1].WorldRanking, "személyenként a legjobb világranglista-helyezés számít")

	require.Len(t, field.RecordHolders, 2)
	assert.Equal(t, "222", field.RecordHolders[0].EventID, "a verseny versenyszám-sorrendjében")
	assert.Equal(t, "333fm", field.RecordHolders[1].EventID)
	assert.Equal(t, "22", field.RecordHolders[1].Display)

	require.Len(t, field.Veterans, 3)
	assert.Equal(t, 2006, field.Veterans[0].Value)

	require.Len(t, field.Anniversaries, 2)
	assert.Equal(t, 20, field.Anniversaries[0].Years)
	assert.Equal(t, "Régi Róbert", field.Anniversaries[0].People[0].Name)
	assert.Equal(t, 10, field.Anniversaries[1].Years)
	assert.Len(t, field.Anniversaries[1].People, 2)
}

func TestComputeField_MilestonesDependOnWhetherResultsAreIn(t *testing.T) {
	summaries := map[string]competition.PersonSummary{
		"2006REGI01": {WCAID: "2006REGI01", CompetitionCount: 99, TotalSolves: 4000, Medals: competition.Medals{Gold: 3, Total: 3}},
		"2016KULF01": {WCAID: "2016KULF01", CompetitionCount: 25, TotalSolves: 900, Records: competition.RecordCounts{Continental: 1, National: 4, Total: 5}},
	}

	before := computeField(testInfo(), PhaseBefore, testRegistrations(), summaries)
	require.Len(t, before.Milestones, 1)
	assert.Equal(t, PersonValue{Person: Person{Name: "Régi Róbert", WCAID: "2006REGI01", CountryISO2: "HU"}, Value: 100}, before.Milestones[0],
		"az eredmények feltöltése előtt ez a következő, 100. verseny")
	assert.Equal(t, 4900, before.CareerSolves)
	assert.Equal(t, "Régi Róbert", before.MostCompetitions[0].Person.Name)
	require.Len(t, before.MostMedals, 1)
	require.Len(t, before.MostRecords, 1)
	assert.Equal(t, 1, before.MostRecords[0].Continental)

	after := computeField(testInfo(), PhaseAfter, testRegistrations(), summaries)
	require.Len(t, after.Milestones, 1)
	assert.Equal(t, 25, after.Milestones[0].Value, "feltöltés után a WCA már beleszámolja ezt a versenyt")
}

func TestIsMilestone(t *testing.T) {
	for _, n := range []int{10, 25, 50, 75, 100, 125, 200} {
		assert.True(t, isMilestone(n), n)
	}
	for _, n := range []int{1, 5, 11, 30, 60, 99} {
		assert.False(t, isMilestone(n), n)
	}
}
