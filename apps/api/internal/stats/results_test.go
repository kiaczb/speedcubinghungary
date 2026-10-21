package stats

import (
	"testing"
	"time"

	"speedcubinghungary/internal/competition"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeResults(t *testing.T) {
	results := []competition.Result{
		{WCAID: "2010AAAA01", Name: "Anna", CountryISO2: "HU", EventID: "333", RoundTypeID: "1", Pos: 1, Best: 820, Average: 950, Attempts: []int{820, 950, 990, 1000, -1}},
		{WCAID: "2010AAAA01", Name: "Anna", CountryISO2: "HU", EventID: "333", RoundTypeID: "f", Pos: 2, Best: 990, Average: 1050, Attempts: []int{990, 1050, 1100, 1080, 1010}},
		{WCAID: "2012BBBB01", Name: "Béla", CountryISO2: "SK", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 701, Average: 899, Attempts: []int{701, 899, 920, 850, 930}, SingleRecord: "NR"},
		{WCAID: "2014CCCC01", Name: "Cili", CountryISO2: "HU", EventID: "333", RoundTypeID: "f", Pos: 3, Best: 1100, Average: 1200, Attempts: []int{1100, 1200, 1250, 1190, 1210}},
		{WCAID: "2010AAAA01", Name: "Anna", CountryISO2: "HU", EventID: "333fm", RoundTypeID: "f", Pos: 1, Best: 24, Average: 2633, Attempts: []int{24, 27, 28}, AverageRecord: "NR"},
	}

	stats := computeResults(testInfo(), results)

	require.Len(t, stats.Winners, 2)
	assert.Equal(t, EventWinner{EventID: "333", Person: Person{Name: "Béla", WCAID: "2012BBBB01", CountryISO2: "SK"}, Best: "7.01", Average: "8.99"}, stats.Winners[0],
		"a győztes a döntő első helyezettje, nem az első fordulóé")
	assert.Equal(t, "333fm", stats.Winners[1].EventID)

	require.Len(t, stats.Records, 2)
	assert.Equal(t, RecordEntry{Person: Person{Name: "Béla", WCAID: "2012BBBB01", CountryISO2: "SK"}, EventID: "333", Type: "single", Record: "NR", Display: "7.01"}, stats.Records[0])
	assert.Equal(t, "26.33", stats.Records[1].Display)

	require.Len(t, stats.Podiums, 3)
	assert.Equal(t, MedalHolder{Person: Person{Name: "Anna", WCAID: "2010AAAA01", CountryISO2: "HU"}, Gold: 1, Silver: 1}, stats.Podiums[0])
	assert.Equal(t, []CountryMedals{{CountryISO2: "HU", Gold: 1, Silver: 1, Bronze: 1}, {CountryISO2: "SK", Gold: 1}}, stats.CountryPodiums)

	assert.Equal(t, 23, stats.Solves)
	assert.Equal(t, 1, stats.DNFs)
	assert.Equal(t, int64(820+950+990+1000+990+1050+1100+1080+1010+701+899+920+850+930+1100+1200+1250+1190+1210), stats.SolvingCentiseconds,
		"az FMC lépésszám nem idő, kimarad az összegből")
}

func TestComputeLive(t *testing.T) {
	ranking := 1
	live := &LiveCompetition{Rounds: []LiveRound{
		{EventID: "333", Number: 1, Open: true, Finished: true, Results: []LiveResult{
			{Person: Person{Name: "Anna", WCAID: "2010AAAA01"}, Ranking: &ranking, Best: 820, Average: 950, SingleRecordTag: "PR", AverageRecordTag: "NR", Attempts: []int{820, 950, 990, 1000, -1}},
			{Person: Person{Name: "Béla", WCAID: "2012BBBB01"}, Best: 900, Average: 1000, AverageRecordTag: "PR", Attempts: []int{900, 1000, 1100, 1000, 950}},
			{Person: Person{Name: "Nem indult"}, Attempts: []int{0, 0, 0, 0, 0}},
		}},
		{EventID: "333", Number: 2, Open: true},
		{EventID: "222", Number: 1},
	}}

	stats := computeLive(testInfo(), live, time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC))

	assert.Equal(t, 3, stats.RoundsTotal)
	assert.Equal(t, 2, stats.RoundsOpen)
	assert.Equal(t, 1, stats.RoundsFinished)
	assert.Equal(t, 2, stats.ResultsEntered, "akinek még nincs kísérlete, nem számít")
	assert.Equal(t, 10, stats.Attempts)
	assert.Equal(t, 1, stats.DNFs)
	assert.Equal(t, 3, stats.PRs, "a rekord is PR")
	require.Len(t, stats.Records, 1)
	assert.Equal(t, RecordEntry{Person: Person{Name: "Anna", WCAID: "2010AAAA01"}, EventID: "333", Type: "average", Record: "NR", Display: "9.50"}, stats.Records[0])
	assert.Equal(t, PersonValue{Person: Person{Name: "Anna", WCAID: "2010AAAA01"}, Value: 2}, stats.PRLeaders[0])

	require.Len(t, stats.EventBests, 1)
	assert.Equal(t, "8.20", stats.EventBests[0].Single.Display)
	assert.Equal(t, "Anna", stats.EventBests[0].Average.Person.Name)
}

func TestPhaseFor(t *testing.T) {
	info := testInfo()
	budapest, err := time.LoadLocation("Europe/Budapest")
	require.NoError(t, err)

	assert.Equal(t, PhaseBefore, phaseFor(info, time.Date(2026, 10, 22, 23, 59, 0, 0, budapest)))
	assert.Equal(t, PhaseDuring, phaseFor(info, time.Date(2026, 10, 23, 0, 30, 0, 0, budapest)),
		"budapesti éjfél után már a verseny napja van, akkor is, ha UTC-ben még előző nap")
	assert.Equal(t, PhaseDuring, phaseFor(info, time.Date(2026, 10, 28, 12, 0, 0, 0, budapest)),
		"amíg nincs feltöltve az eredmény, a WCA Live a forrás")

	info.ResultsPostedAt = "2026-10-27T10:00:00.000Z"
	assert.Equal(t, PhaseAfter, phaseFor(info, time.Date(2026, 10, 28, 12, 0, 0, 0, budapest)))
}
