package certificate

import (
	"testing"

	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/stats"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPodiumFromResults(t *testing.T) {
	results := []competition.Result{
		{WCAID: "2010AAAA01", Name: "Anna", CountryISO2: "HU", EventID: "333", RoundTypeID: "1", Pos: 1, Best: 820, Average: 950, Attempts: []int{820, 950, 990, 1000, -1}},
		{WCAID: "2010AAAA01", Name: "Anna", CountryISO2: "HU", EventID: "333", RoundTypeID: "f", Pos: 2, Best: 990, Average: 1050, Attempts: []int{990, 1050, 1100, 1080, 1010}},
		{WCAID: "2012BBBB01", Name: "Béla", CountryISO2: "SK", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 701, Average: 899, Attempts: []int{701, 899, 920, 850, 930}},
		{WCAID: "2014CCCC01", Name: "Cili", CountryISO2: "HU", EventID: "333", RoundTypeID: "f", Pos: 4, Best: 1100, Average: 1200, Attempts: []int{1100, 1200, 1250, 1190, 1210}},
		{WCAID: "2015DDDD01", Name: "Dani", CountryISO2: "HU", EventID: "333bf", RoundTypeID: "c", Pos: 1, Best: 2500, Average: 2800, Attempts: []int{2500, 2900, 3000}},
		{WCAID: "2016EEEE01", Name: "Emese", CountryISO2: "HU", EventID: "333bf", RoundTypeID: "c", Pos: 3, Best: -1, Attempts: []int{-1, -1, -1}},
		{WCAID: "2017FFFF01", Name: "Feri", CountryISO2: "HU", EventID: "222", RoundTypeID: "f", Pos: 2, Best: 300, Average: 410},
		{WCAID: "2018GGGG01", Name: "Gabi", CountryISO2: "HU", EventID: "222", RoundTypeID: "f", Pos: 2, Best: 310, Average: 410},
	}

	events, recipients := podiumFromResults([]string{"333", "222", "333bf"}, results)

	assert.Equal(t, []Event{{"333", EventFinished}, {"222", EventFinished}, {"333bf", EventFinished}}, events)
	require.Len(t, recipients, 5)
	assert.Equal(t, Recipient{Name: "Béla", WCAID: "2012BBBB01", CountryISO2: "SK", EventID: "333", Place: 1, Result: "8.99", ResultType: "average"}, recipients[0],
		"a döntő számít, nem az első forduló, és a verseny versenyszám-sorrendje")
	assert.Equal(t, "Anna", recipients[1].Name)
	assert.Equal(t, 2, recipients[1].Place)
	assert.Equal(t, []string{"Feri", "Gabi"}, []string{recipients[2].Name, recipients[3].Name}, "holtverseny: mindketten 2. helyezettek")
	assert.Equal(t, 2, recipients[3].Place)
	assert.Equal(t, Recipient{Name: "Dani", WCAID: "2015DDDD01", CountryISO2: "HU", EventID: "333bf", Place: 1, Result: "25.00", ResultType: "single"}, recipients[4],
		"vakon a single rangsorol; a DNF-es 3. helyezett nem kap dobogós oklevelet")
}

func TestPodiumFromLive(t *testing.T) {
	first, second, third := 1, 2, 3
	live := &stats.LiveCompetition{Rounds: []stats.LiveRound{
		{EventID: "333", Number: 1, Open: true, Finished: true, Results: []stats.LiveResult{
			{Person: stats.Person{Name: "Anna", WCAID: "2010AAAA01"}, Ranking: &first, Best: 800, Average: 900},
		}},
		{EventID: "333", Number: 2, Open: true, Finished: true, Results: []stats.LiveResult{
			{Person: stats.Person{Name: "Béla", WCAID: "2012BBBB01", CountryISO2: "SK"}, Ranking: &first, Best: 701, Average: 899},
			{Person: stats.Person{Name: "Újonc Ubul", CountryISO2: "HU"}, Ranking: &second, Best: 750, Average: 950},
			{Person: stats.Person{Name: "Anna", WCAID: "2010AAAA01"}, Ranking: &third, Best: 800, Average: -1},
			{Person: stats.Person{Name: "Nem indult"}, Best: 0},
		}},
		{EventID: "222", Number: 1, Open: true, Results: []stats.LiveResult{
			{Person: stats.Person{Name: "Feri"}, Ranking: &first, Best: 300, Average: 400},
		}},
		{EventID: "pyram", Number: 1, Open: true, Finished: true},
		{EventID: "pyram", Number: 2},
	}}

	events, recipients := podiumFromLive([]string{"222", "333", "pyram", "sq1"}, live)

	assert.Equal(t, []Event{{"222", EventRunning}, {"333", EventFinished}, {"pyram", EventPending}, {"sq1", EventPending}}, events,
		"csak a lezárt döntő kész; amíg a döntő el sem indult, a versenyszám várakozik")
	require.Len(t, recipients, 3, "a futó döntőből még nincs oklevél")
	assert.Equal(t, Recipient{Name: "Béla", WCAID: "2012BBBB01", CountryISO2: "SK", EventID: "333", Place: 1, Result: "8.99", ResultType: "average"}, recipients[0])
	assert.Equal(t, Recipient{Name: "Újonc Ubul", CountryISO2: "HU", EventID: "333", Place: 2, Result: "9.50", ResultType: "average"}, recipients[1],
		"az újoncnak nincs WCA ID-ja, mégis kap oklevelet")
	assert.Equal(t, Recipient{Name: "Anna", WCAID: "2010AAAA01", EventID: "333", Place: 3, Result: "8.00", ResultType: "single"}, recipients[2],
		"DNF átlag mellett a single az eredmény")
}

func TestCompetitorsFromResults(t *testing.T) {
	results := []competition.Result{
		{WCAID: "2012BBBB01", Name: "béla", EventID: "333", Attempts: []int{900, 1000, 1100, 1000, 950}},
		{WCAID: "2010AAAA01", Name: "Anna", EventID: "333", Attempts: []int{-1, -1, -2, -2, -2}},
		{WCAID: "2010AAAA01", Name: "Anna", EventID: "222", Attempts: []int{300, 400, 500, 450, 350}},
		{WCAID: "2019HHHH01", Name: "Csak DNS", EventID: "333", Attempts: []int{-2, -2, -2, -2, -2}},
	}

	recipients := competitorsFromResults(results)

	assert.Equal(t, []Recipient{{Name: "Anna", WCAID: "2010AAAA01"}, {Name: "béla", WCAID: "2012BBBB01"}}, recipients,
		"egy ember egyszer szerepel; aki csak DNS-t kapott, nem versenyzett")
}

func TestCompetitorsFromLive(t *testing.T) {
	live := &stats.LiveCompetition{Rounds: []stats.LiveRound{
		{EventID: "333", Number: 1, Open: true, Results: []stats.LiveResult{
			{Person: stats.Person{Name: "Anna", WCAID: "2010AAAA01"}, Attempts: []int{820, 0, 0, 0, 0}},
			{Person: stats.Person{Name: "Újonc Ubul"}, Attempts: []int{-1, 0, 0, 0, 0}},
			{Person: stats.Person{Name: "Még nem jött"}, Attempts: []int{0, 0, 0, 0, 0}},
		}},
		{EventID: "222", Number: 1, Open: true, Results: []stats.LiveResult{
			{Person: stats.Person{Name: "Anna", WCAID: "2010AAAA01"}, Attempts: []int{300}},
		}},
	}}

	assert.Equal(t, []Recipient{{Name: "Anna", WCAID: "2010AAAA01"}, {Name: "Újonc Ubul"}}, competitorsFromLive(live))
}

func TestRegistrantsOf(t *testing.T) {
	registrations := competition.Registrations{Registrants: []competition.Registrant{
		{Name: "Zoli", WCAID: "2011ZZZZ01", CountryISO2: "HU"},
		{Name: "Újonc Ubul", CountryISO2: "HU"},
		{Name: "Anna", WCAID: "2010AAAA01", CountryISO2: "AT"},
	}}

	assert.Equal(t, []Recipient{
		{Name: "Anna", WCAID: "2010AAAA01", CountryISO2: "AT"},
		{Name: "Zoli", WCAID: "2011ZZZZ01", CountryISO2: "HU"},
		{Name: "Újonc Ubul", CountryISO2: "HU"},
	}, registrantsOf(registrations))
}
