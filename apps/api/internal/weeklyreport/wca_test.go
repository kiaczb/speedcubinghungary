package weeklyreport

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestWCAClient(t *testing.T, handler http.HandlerFunc) *WCAClient {
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	c := newWCAClient(server.URL)
	c.backoff = func(int) time.Duration { return 0 }
	return c
}

func TestWCAClient_RetriesRateLimitedRequests(t *testing.T) {
	var calls atomic.Int32
	c := newTestWCAClient(t, func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		fmt.Fprint(w, `[{"person": {"wca_id": "2015REKA01"}, "competition_count": 11, "total_solves": 4800}]`)
	})

	stats, err := c.GetPersonStats(context.Background(), []string{"2015REKA01"})

	require.NoError(t, err)
	assert.Equal(t, map[string]PersonStats{"2015REKA01": {CompetitionCount: 11, TotalSolves: 4800}}, stats)
	assert.Equal(t, int32(3), calls.Load())
}

func TestWCAClient_GetPersonStatsBatchesByHundred(t *testing.T) {
	var batches []int
	c := newTestWCAClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/persons", r.URL.Path)
		ids := strings.Split(r.URL.Query().Get("wca_ids"), ",")
		batches = append(batches, len(ids))
		assert.Equal(t, fmt.Sprint(len(ids)), r.URL.Query().Get("per_page"))
		items := make([]string, 0, len(ids))
		for _, id := range ids {
			items = append(items, fmt.Sprintf(`{"person": {"wca_id": %q}, "competition_count": 1, "total_solves": 1}`, id))
		}
		fmt.Fprint(w, "["+strings.Join(items, ",")+"]")
	})

	ids := make([]string, 230)
	for i := range ids {
		ids[i] = fmt.Sprintf("2020TEST%02d", i)
	}
	stats, err := c.GetPersonStats(context.Background(), ids)

	require.NoError(t, err)
	assert.Equal(t, []int{100, 100, 30}, batches)
	assert.Len(t, stats, 230)
}

func TestWCAClient_CountDelegatedCompetitionsCountsStartedOnes(t *testing.T) {
	c := newTestWCAClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "2008BARA01", r.URL.Query().Get("delegate"))
		switch r.URL.Query().Get("page") {
		case "1":
			fmt.Fprint(w, `[
				{"id": "Old", "start_date": "2010-05-01"},
				{"id": "Cancelled", "start_date": "2026-09-01", "cancelled_at": "2026-08-01T00:00:00Z"},
				{"id": "HungarianOpen2026", "start_date": "2026-10-23", "results_posted_at": null}
			]`)
		case "2":
			fmt.Fprint(w, `[{"id": "Future", "start_date": "2026-11-28"}]`)
		default:
			fmt.Fprint(w, `[]`)
		}
	})

	n, err := c.CountDelegatedCompetitions(context.Background(), "2008BARA01", time.Date(2026, 10, 28, 8, 0, 0, 0, time.UTC))

	require.NoError(t, err)
	assert.Equal(t, 2, n, "started ones count even before their results are posted; cancelled and future ones don't")
}

func TestWCAClient_GivesUpAfterMaxAttempts(t *testing.T) {
	var calls atomic.Int32
	c := newTestWCAClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	})

	_, err := c.GetWCIF(context.Background(), "MiskolcOpen2026")

	assert.Error(t, err)
	assert.Equal(t, int32(maxAttempts), calls.Load())
}

func TestWCAClient_NotFoundIsNotRetried(t *testing.T) {
	var calls atomic.Int32
	c := newTestWCAClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNotFound)
	})

	_, err := c.GetResults(context.Background(), "MiskolcOpen2026")

	assert.ErrorIs(t, err, ErrNotFound)
	assert.Equal(t, int32(1), calls.Load())
}

func TestWCAClient_ListCompetitionsPagesUntilEmpty(t *testing.T) {
	var queries []string
	c := newTestWCAClient(t, func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.RawQuery)
		switch r.URL.Query().Get("page") {
		case "1":
			fmt.Fprint(w, `[{"id":"A"},{"id":"B"}]`)
		case "2":
			fmt.Fprint(w, `[{"id":"C"}]`)
		default:
			fmt.Fprint(w, `[]`)
		}
	})

	start := time.Date(2026, 7, 20, 8, 0, 0, 0, time.UTC)
	competitions, err := c.ListCompetitions(context.Background(), start, start.Add(reportPeriod))

	require.NoError(t, err)
	assert.Len(t, competitions, 3)
	require.Len(t, queries, 3)
	assert.Contains(t, queries[0], "start=2026-07-20")
	assert.Contains(t, queries[0], "end=2026-07-27")
}

func TestWCAClient_ParsesWCIFAttemptsAndRegistrants(t *testing.T) {
	c := newTestWCAClient(t, func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/competitions/MiskolcOpen2026/wcif/public", r.URL.Path)
		fmt.Fprint(w, `{
			"persons": [{"name": "Új Újonc", "wcaId": null, "registrantId": 9, "countryIso2": "HU", "registration": {"eventIds": ["333"]}},
			            {"name": "Nem elfogadott", "wcaId": null, "registrantId": null, "countryIso2": "HU", "registration": null}],
			"events": [{"id": "333", "rounds": [{"results": [{"personId": 9, "best": 900, "average": -1,
				"attempts": [{"result": 900, "reconstruction": null}, {"result": -1}]}]}]}]
		}`)
	})

	wcif, err := c.GetWCIF(context.Background(), "MiskolcOpen2026")

	require.NoError(t, err)
	require.Len(t, wcif.Persons, 2)
	assert.Equal(t, 9, *wcif.Persons[0].RegistrantID)
	assert.Nil(t, wcif.Persons[1].RegistrantID)
	events := eventsFromWCIF(wcif.Events)
	assert.Equal(t, []int{900, -1}, events[0].Results[0].Attempts)
}
