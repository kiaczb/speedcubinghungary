package stats

import (
	"errors"
	"sync"
	"testing"
	"time"

	"speedcubinghungary/internal/competition"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestComputeHistory_ComebacksCountFromTheLastCompetitionNotTheFirst(t *testing.T) {
	registrants := []competition.Registrant{
		{Name: "Visszatérő Vera", WCAID: "2010VISS01", CountryISO2: "HU"},
		{Name: "Rendszeres Rezső", WCAID: "2015REND01", CountryISO2: "HU"},
		{Name: "Vendég Viktor", WCAID: "2018VEND01", CountryISO2: "AT"},
		{Name: "Új Ubul", CountryISO2: "HU"},
	}
	histories := map[string][]competition.PersonCompetition{
		"2010VISS01": {
			{ID: "Elso2010", Name: "Első 2010", StartDate: "2010-05-01", EndDate: "2010-05-02", CountryISO2: "HU"},
			{ID: "Utolso2014", Name: "Utolsó 2014", StartDate: "2014-03-08", EndDate: "2014-03-09", CountryISO2: "HU"},
		},
		"2015REND01": {
			{ID: "Tavaszi2026", StartDate: "2026-04-11", EndDate: "2026-04-11", CountryISO2: "HU"},
			{ID: "Nyari2026", StartDate: "2026-07-04", EndDate: "2026-07-05", CountryISO2: "HU"},
			{ID: "Teszt2026", StartDate: "2026-10-23", EndDate: "2026-10-25", CountryISO2: "HU"},
		},
		"2018VEND01": {
			{ID: "Becs2025", StartDate: "2025-11-01", EndDate: "2025-11-02", CountryISO2: "AT"},
		},
	}

	comebacks, firstInCountry, busiest := computeHistory(testInfo(), registrants, histories)

	require.Len(t, comebacks, 1)
	assert.Equal(t, "Visszatérő Vera", comebacks[0].Person.Name)
	assert.Equal(t, CompetitionStub{ID: "Utolso2014", Name: "Utolsó 2014", Date: "2014-03-08"}, comebacks[0].LastCompetition)
	assert.Equal(t, 4611, comebacks[0].GapDays, "2014-03-09 → 2026-10-23")

	assert.Equal(t, []Person{{Name: "Vendég Viktor", WCAID: "2018VEND01", CountryISO2: "AT"}}, firstInCountry)

	require.Len(t, busiest, 1)
	assert.Equal(t, PersonValue{Person: Person{Name: "Rendszeres Rezső", WCAID: "2015REND01", CountryISO2: "HU"}, Value: 2}, busiest[0],
		"maga a verseny nem számít bele az idei versenyekbe")
}

func TestHistoryTracker_FetchesOncePerPersonAndRetriesRateLimits(t *testing.T) {
	var mu sync.Mutex
	calls := map[string]int{}
	tracker := newHistoryTracker(func(wcaID string) ([]competition.PersonCompetition, error) {
		mu.Lock()
		defer mu.Unlock()
		calls[wcaID]++
		if wcaID == "2019LASS01" && calls[wcaID] == 1 {
			return nil, competition.ErrRateLimited
		}
		if wcaID == "2019HIBA01" {
			return nil, errors.New("status 500")
		}
		return []competition.PersonCompetition{{ID: "X2025", StartDate: "2025-01-01"}}, nil
	})
	var slept []time.Duration
	tracker.sleep = func(d time.Duration) { slept = append(slept, d) }

	ids := []string{"2015ELSO01", "2019LASS01", "2019HIBA01"}
	first := tracker.snapshot("Teszt2026", ids)
	assert.True(t, first.running, "az első lekérés elindítja a háttérfeladatot")

	require.Eventually(t, func() bool {
		return !tracker.snapshot("Teszt2026", ids).running
	}, 2*time.Second, 10*time.Millisecond)

	done := tracker.snapshot("Teszt2026", ids)
	assert.Equal(t, 2, done.done)
	assert.Equal(t, 1, done.failed)
	assert.False(t, done.running, "a hibás versenyzőt nem kérjük újra azonnal")
	assert.Len(t, done.histories, 2)

	mu.Lock()
	assert.Equal(t, 1, calls["2015ELSO01"])
	assert.Equal(t, 2, calls["2019LASS01"], "429 után újrapróbálja")
	assert.Equal(t, 1, calls["2019HIBA01"], "más hibánál nem próbálkozik újra")
	mu.Unlock()
	assert.Equal(t, []time.Duration{historyRateLimitBackoff}, slept)
}
