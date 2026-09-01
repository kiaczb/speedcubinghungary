package allrounder

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNeedsSync(t *testing.T) {
	tests := []struct {
		name  string
		state SyncState
		known bool
		want  bool
	}{
		{
			name:  "még soha nem szinkronizált verseny",
			known: false,
			want:  true,
		},
		{
			name:  "friss eredmény: nem kell újra",
			state: SyncState{FetchedAt: time.Now().Add(-time.Hour), ResultCount: 298},
			known: true,
			want:  false,
		},
		{
			name:  "régi eredmény: felülvizsgálat",
			state: SyncState{FetchedAt: time.Now().Add(-resultsTTL - time.Hour), ResultCount: 298},
			known: true,
			want:  true,
		},
		{
			name:  "még nincs eredmény, nemrég néztük",
			state: SyncState{FetchedAt: time.Now().Add(-time.Hour)},
			known: true,
			want:  false,
		},
		{
			name:  "még nincs eredmény, rég néztük",
			state: SyncState{FetchedAt: time.Now().Add(-pendingTTL - time.Minute)},
			known: true,
			want:  true,
		},
		{
			name:  "hibás letöltés: hamar újrapróbáljuk",
			state: SyncState{FetchedAt: time.Now().Add(-retryTTL - time.Minute), LastError: "boom"},
			known: true,
			want:  true,
		},
		{
			name:  "hibás letöltés, de épp most próbáltuk",
			state: SyncState{FetchedAt: time.Now(), LastError: "boom"},
			known: true,
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, needsSync(tt.state, tt.known))
		})
	}
}

func TestOldestSync(t *testing.T) {
	older := time.Date(2026, 8, 1, 0, 0, 0, 0, time.UTC)
	newer := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)

	assert.Nil(t, oldestSync(map[string]SyncState{}))

	oldest := oldestSync(map[string]SyncState{
		"C1": {FetchedAt: newer},
		"C2": {FetchedAt: older},
	})
	require.NotNil(t, oldest)
	assert.Equal(t, older, *oldest, "a táblázat ennél frissebb adatot nem tartalmaz")
}

type fakeRepository struct {
	mu      sync.Mutex
	results map[string][]Result
	states  map[string]SyncState
	failed  map[string]string
}

func newFakeRepository() *fakeRepository {
	return &fakeRepository{
		results: make(map[string][]Result),
		states:  make(map[string]SyncState),
		failed:  make(map[string]string),
	}
}

func (f *fakeRepository) ListResults(ids []string) (map[string][]Result, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make(map[string][]Result)
	for _, id := range ids {
		if stored, ok := f.results[id]; ok {
			out[id] = stored
		}
	}

	return out, nil
}

func (f *fakeRepository) ListSyncStates(ids []string) (map[string]SyncState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make(map[string]SyncState)
	for _, id := range ids {
		if state, ok := f.states[id]; ok {
			out[id] = state
		}
	}

	return out, nil
}

func (f *fakeRepository) ReplaceResults(id string, results []Result, fetchedAt time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.results[id] = results
	f.states[id] = SyncState{CompetitionWCAID: id, FetchedAt: fetchedAt, ResultCount: len(results)}

	return nil
}

func (f *fakeRepository) MarkSyncFailed(id string, fetchedAt time.Time, message string) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	state := f.states[id]
	state.CompetitionWCAID = id
	state.FetchedAt = fetchedAt
	state.LastError = message
	f.states[id] = state
	f.failed[id] = message

	return nil
}

func (f *fakeRepository) SaveCompetitionDates(id string, startDate, endDate time.Time) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	state := f.states[id]
	state.CompetitionWCAID = id
	state.StartDate = &startDate
	state.EndDate = &endDate
	f.states[id] = state

	return nil
}

func (f *fakeRepository) SaveLiveCompetitionID(id string, liveCompetitionID int) error {
	f.mu.Lock()
	defer f.mu.Unlock()

	state := f.states[id]
	state.CompetitionWCAID = id
	state.LiveCompetitionID = &liveCompetitionID
	f.states[id] = state

	return nil
}

func newTestService(t *testing.T, repo Repository, handler http.HandlerFunc) *Service {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	client := NewWCAClient()
	client.baseURL = server.URL

	live := NewLiveClient(server.URL)

	return NewService(repo, client, live, nil, NewBroadcaster())
}

func TestGetStandings_FetchesOnceThenServesFromCache(t *testing.T) {
	repo := newFakeRepository()

	var mu sync.Mutex
	requests := 0

	service := newTestService(t, repo, func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasSuffix(r.URL.Path, "/results") {
			w.Write([]byte(`{"start_date":"2026-01-10","end_date":"2026-01-10"}`))
			return
		}

		mu.Lock()
		requests++
		mu.Unlock()

		if !strings.Contains(r.URL.Path, "HungarianAllRounder12026") {
			http.Error(w, "not found", http.StatusNotFound)
			return
		}

		w.Write([]byte(`[
			{"pos":1,"best":700,"name":"Anna","event_id":"333","round_type_id":"f","wca_id":"2020AAAA01"},
			{"pos":2,"best":800,"name":"Bence","event_id":"333","round_type_id":"f","wca_id":"2020BBBB01"}
		]`))
	})

	series := ResolveSeries(2026)

	standings, err := service.GetStandings(context.Background(), 2026)
	require.NoError(t, err)

	assert.Equal(t, 2026, standings.Year)
	assert.Equal(t, series.Competitions, standings.Competitions, "minden állomás oszlopa megmarad")
	require.Len(t, standings.People, 2)
	assert.Equal(t, "Anna", standings.People[0].Name)
	assert.Equal(t, 50, standings.People[0].TotalPoints)
	require.NotNil(t, standings.UpdatedAt)

	mu.Lock()
	afterFirst := requests
	mu.Unlock()
	assert.Equal(t, len(series.Competitions), afterFirst, "az első kérés minden állomást lekér")

	_, err = service.GetStandings(context.Background(), 2026)
	require.NoError(t, err)

	mu.Lock()
	defer mu.Unlock()
	assert.Equal(t, afterFirst, requests, "a gyorsítótár lejártáig nincs újabb WCA hívás")
}

func TestGetStandings_KeepsCachedResultsWhenWCAFails(t *testing.T) {
	repo := newFakeRepository()
	repo.results["HungarianAllRounder12026"] = results(
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
	)
	repo.states["HungarianAllRounder12026"] = SyncState{
		CompetitionWCAID: "HungarianAllRounder12026",
		FetchedAt:        time.Now().Add(-resultsTTL - time.Hour),
		ResultCount:      1,
	}

	service := newTestService(t, repo, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "upstream down", http.StatusInternalServerError)
	})

	standings, err := service.GetStandings(context.Background(), 2026)

	require.NoError(t, err)
	require.Len(t, standings.People, 1)
	assert.Equal(t, "Anna", standings.People[0].Name)
	assert.NotEmpty(t, repo.failed["HungarianAllRounder12026"], "a hiba bekerül a szinkron állapotába")
	assert.Len(t, repo.results["HungarianAllRounder12026"], 1, "a meglévő eredmények megmaradnak")
}

func TestGetStandings_KeepsResultsOn404AfterTheyExisted(t *testing.T) {
	repo := newFakeRepository()
	repo.results["HungarianAllRounder12026"] = results(
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
	)
	repo.states["HungarianAllRounder12026"] = SyncState{
		CompetitionWCAID: "HungarianAllRounder12026",
		FetchedAt:        time.Now().Add(-resultsTTL - time.Hour),
		ResultCount:      1,
	}

	service := newTestService(t, repo, func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	standings, err := service.GetStandings(context.Background(), 2026)

	require.NoError(t, err)
	require.Len(t, standings.People, 1)
	assert.Len(t, repo.results["HungarianAllRounder12026"], 1)
}

func TestGetStandings_UnknownYearFallsBackToDefault(t *testing.T) {
	service := newTestService(t, newFakeRepository(), func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "not found", http.StatusNotFound)
	})

	standings, err := service.GetStandings(context.Background(), 1999)

	require.NoError(t, err)
	assert.Contains(t, AvailableYears(), standings.Year)
	assert.Equal(t, AvailableYears(), standings.AvailableYears)
}
