package competition

import (
	"errors"
	"net/http"
	"net/http/httptest"
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
func (s stubRepository) GetWCIF(string) (wcaWCIF, error) { return wcaWCIF{}, s.err }

// A cache kulcsa a KLIENS által adott azonosító. Takarítás nélkül minden
// valaha lekérdezett id örökre a memóriában maradt, a lejártak is.
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

// A WCA 404-e (nem létező verseny) korábban 502-ként jutott ki — mintha a WCA
// esett volna ki, pedig csak rossz az azonosító.
func TestHandler_MapsWCANotFoundTo404(t *testing.T) {
	cases := map[error]int{
		ErrNotFound:               http.StatusNotFound,
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
