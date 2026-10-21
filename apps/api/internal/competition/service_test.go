package competition

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepository struct{ err error }

func (s stubRepository) GetCompetition(string) (wcaCompetition, error) {
	return wcaCompetition{}, s.err
}
func (s stubRepository) GetUpcomingByCountry(string) ([]wcaCompetition, error) { return nil, s.err }
func (s stubRepository) GetPsychSheet(string, string) (wcaPsychSheet, error) {
	return wcaPsychSheet{}, s.err
}
func (s stubRepository) GetWCIF(string) (wcaWCIF, error)          { return wcaWCIF{}, s.err }
func (s stubRepository) GetRegistrationCount(string) (int, error) { return 0, s.err }
func (s stubRepository) GetPersons([]string) ([]wcaPersonSummary, error) { return nil, s.err }
func (s stubRepository) GetPersonCompetitions(string) ([]wcaCompetition, error) {
	return nil, s.err
}
func (s stubRepository) GetResults(string) ([]wcaResult, error) { return nil, s.err }
func (s stubRepository) SearchCompetitions(url.Values) ([]wcaCompetition, error) {
	return nil, s.err
}

type upcomingRepository struct {
	stubRepository
	competitions []wcaCompetition
	counts       map[string]int
	mu           sync.Mutex
	asked        []string
	listCalls    int
}

func (r *upcomingRepository) GetUpcomingByCountry(string) ([]wcaCompetition, error) {
	r.mu.Lock()
	r.listCalls++
	r.mu.Unlock()
	return r.competitions, nil
}

func (r *upcomingRepository) GetRegistrationCount(competitionID string) (int, error) {
	r.mu.Lock()
	r.asked = append(r.asked, competitionID)
	r.mu.Unlock()

	r.mu.Lock()
	count, ok := r.counts[competitionID]
	r.mu.Unlock()
	if !ok {
		return 0, errors.New("status 503")
	}
	return count, nil
}

func TestRegistrationOpen(t *testing.T) {
	now := time.Date(2026, 10, 10, 12, 0, 0, 0, time.UTC)
	competition := func(opens, closes string) Competition {
		return Competition{RegistrationOpen: opens, RegistrationClose: closes}
	}

	cases := map[string]struct {
		competition Competition
		want        bool
	}{
		"nyitva":            {competition("2026-10-01T18:00:00.000Z", "2026-10-20T22:59:00.000Z"), true},
		"még nem nyílt meg": {competition("2026-10-14T18:00:00.000Z", "2026-11-06T22:59:00.000Z"), false},
		"már lezárult":      {competition("2026-08-12T18:00:00.000Z", "2026-10-09T21:59:00.000Z"), false},
		"pont most nyílik":  {competition("2026-10-10T12:00:00Z", "2026-10-20T22:59:00Z"), true},
		"pont most zárul":   {competition("2026-10-01T18:00:00Z", "2026-10-10T12:00:00Z"), false},
		"hiányzó időpontok": {competition("", ""), false},
		"értelmezhetetlen":  {competition("holnap", "jövő héten"), false},
	}

	for name, tc := range cases {
		assert.Equal(t, tc.want, registrationOpen(tc.competition, now), name)
	}
}

func TestService_CountsRegistrationsOnlyWhileRegistrationIsOpen(t *testing.T) {
	future := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)
	past := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	limit := 40

	repo := &upcomingRepository{
		competitions: []wcaCompetition{
			{ID: "Nyitott2026", RegistrationOpen: past, RegistrationClose: future, CompetitorLimit: &limit},
			{ID: "Hibas2026", RegistrationOpen: past, RegistrationClose: future},
			{ID: "Kesobbi2026", RegistrationOpen: future, RegistrationClose: future},
			{ID: "Lezart2026", RegistrationOpen: past, RegistrationClose: past},
		},
		counts: map[string]int{"Nyitott2026": 12},
	}

	competitions, err := NewService(repo).GetUpcomingHungarian()
	require.NoError(t, err, "a létszám lekérésének hibája nem ronthatja el a listát")
	require.Len(t, competitions, 4)

	assert.ElementsMatch(t, []string{"Nyitott2026", "Hibas2026"}, repo.asked,
		"csak a nyitott nevezésű versenyek létszámát kérjük le")

	require.NotNil(t, competitions[0].AcceptedCount)
	assert.Equal(t, 12, *competitions[0].AcceptedCount)
	assert.Equal(t, &limit, competitions[0].CompetitorLimit)
	assert.Nil(t, competitions[1].AcceptedCount, "hiba esetén a létszám ismeretlen marad")
	assert.Nil(t, competitions[2].AcceptedCount)
	assert.Nil(t, competitions[3].AcceptedCount)
}

func TestService_SweepsExpiredCacheEntries(t *testing.T) {
	service := NewService(stubRepository{})

	stale := time.Now().Add(-2 * upstreamCacheTTL)
	service.psychSheets["psych-sheet:Regi/333"] = cacheEntry[[]PsychSheetEntry]{at: stale}
	service.registrations["registrations:Regi"] = cacheEntry[Registrations]{at: stale}

	_, err := service.GetPsychSheet("Uj2026", "333")
	require.NoError(t, err)

	assert.NotContains(t, service.psychSheets, "psych-sheet:Regi/333")
	assert.NotContains(t, service.registrations, "registrations:Regi")
	assert.Contains(t, service.psychSheets, "psych-sheet:Uj2026/333", "a friss bejegyzés marad")
}

func TestHandler_MapsWCANotFoundTo404(t *testing.T) {
	cases := map[error]int{
		ErrNotFound:              http.StatusNotFound,
		errors.New("status 503"): http.StatusBadGateway,
	}

	for upstreamErr, want := range cases {
		router := chi.NewRouter()
		router.Get("/competitions/{id}/competitors", NewHandler(NewService(stubRepository{err: upstreamErr})).GetCompetitors)

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/competitions/NemLetezo2099/competitors", nil))

		assert.Equal(t, want, recorder.Code, upstreamErr.Error())
	}
}

func TestService_RefreshesRegistrationCountsFasterThanTheList(t *testing.T) {
	past := time.Now().Add(-24 * time.Hour).UTC().Format(time.RFC3339)
	future := time.Now().Add(24 * time.Hour).UTC().Format(time.RFC3339)

	repo := &upcomingRepository{
		competitions: []wcaCompetition{
			{ID: "Nyitott2026", RegistrationOpen: past, RegistrationClose: future},
		},
		counts: map[string]int{"Nyitott2026": 12},
	}
	service := NewService(repo)

	first, err := service.GetUpcomingHungarian()
	require.NoError(t, err)
	require.NotNil(t, first[0].AcceptedCount)
	assert.Equal(t, 12, *first[0].AcceptedCount)

	repo.mu.Lock()
	repo.counts["Nyitott2026"] = 19
	repo.mu.Unlock()
	service.mu.Lock()
	entry := service.registered["registered:Nyitott2026"]
	entry.at = entry.at.Add(-2 * registrationCountTTL)
	service.registered["registered:Nyitott2026"] = entry
	service.mu.Unlock()

	second, err := service.GetUpcomingHungarian()
	require.NoError(t, err)
	require.NotNil(t, second[0].AcceptedCount)
	assert.Equal(t, 19, *second[0].AcceptedCount, "a létszám a lista cache-étől függetlenül frissül")
	assert.Equal(t, 12, *first[0].AcceptedCount, "a korábban visszaadott szelet nem változik utólag")
	assert.Equal(t, 1, repo.listCalls, "a versenylistát közben nem kérjük le újra")

	service.mu.Lock()
	listed := service.upcoming["upcoming:HU"]
	service.mu.Unlock()
	assert.Nil(t, listed.value[0].AcceptedCount, "a cache-elt lista nem tárol létszámot")
}

func TestService_KeepsPersonDataAcrossTheShortSweep(t *testing.T) {
	service := NewService(stubRepository{})

	hourOld := time.Now().Add(-time.Hour)
	service.persons["2016SZAN01"] = cacheEntry[PersonSummary]{value: PersonSummary{WCAID: "2016SZAN01"}, at: hourOld}
	service.histories["history:2016SZAN01"] = cacheEntry[[]PersonCompetition]{at: hourOld}
	service.persons["2000REGI01"] = cacheEntry[PersonSummary]{at: time.Now().Add(-2 * personDataTTL)}

	_, err := service.GetPsychSheet("Uj2026", "333")
	require.NoError(t, err)

	assert.Contains(t, service.persons, "2016SZAN01", "a 24 órás személyadat nem eshet ki a 15 perces söprésnél")
	assert.Contains(t, service.histories, "history:2016SZAN01")
	assert.NotContains(t, service.persons, "2000REGI01")
}

type personsRepository struct {
	stubRepository
	mu     sync.Mutex
	asked  [][]string
	people map[string]int
}

func (r *personsRepository) GetPersons(wcaIDs []string) ([]wcaPersonSummary, error) {
	r.mu.Lock()
	r.asked = append(r.asked, append([]string(nil), wcaIDs...))
	r.mu.Unlock()

	summaries := make([]wcaPersonSummary, 0, len(wcaIDs))
	for _, id := range wcaIDs {
		if count, ok := r.people[id]; ok {
			summary := wcaPersonSummary{CompetitionCount: count}
			summary.Person.WCAID = id
			summaries = append(summaries, summary)
		}
	}
	return summaries, nil
}

func TestService_GetPersonSummaries_FetchesOnlyMissingValidIDs(t *testing.T) {
	repo := &personsRepository{people: map[string]int{"2016SZAN01": 80, "2019SIMO07": 12}}
	service := NewService(repo)
	service.persons["2016SZAN01"] = cacheEntry[PersonSummary]{value: PersonSummary{WCAID: "2016SZAN01", CompetitionCount: 79}, at: time.Now()}

	summaries, err := service.GetPersonSummaries([]string{"2016SZAN01", "2019SIMO07", "2019SIMO07", "", "nem-wca-id"})
	require.NoError(t, err)

	assert.Equal(t, [][]string{{"2019SIMO07"}}, repo.asked, "a cache-ben lévőt, a duplikátumot és az érvénytelen azonosítót nem kérjük le")
	assert.Equal(t, 79, summaries["2016SZAN01"].CompetitionCount, "a cache-elt érték szolgál ki")
	assert.Equal(t, 12, summaries["2019SIMO07"].CompetitionCount)
	assert.Len(t, summaries, 2)
}
