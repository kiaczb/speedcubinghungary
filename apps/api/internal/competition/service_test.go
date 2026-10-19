package competition

import (
	"errors"
	"net/http"
	"net/http/httptest"
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
	service.competitors["competitors:Regi"] = cacheEntry[[]Competitor]{at: stale}

	_, err := service.GetPsychSheet("Uj2026", "333")
	require.NoError(t, err)

	assert.NotContains(t, service.psychSheets, "psych-sheet:Regi/333")
	assert.NotContains(t, service.competitors, "competitors:Regi")
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
