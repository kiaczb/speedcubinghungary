package special

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_GetReport_QueryParams(t *testing.T) {
	server := fakeWCALive(t,
		fakeRound{id: "1", eventID: "333", number: 1, results: []liveResult{
			result("1", 7, "Teszt Tamás", attempts(1000)),
		}},
		fakeRound{id: "2", eventID: "333", number: 2, results: []liveResult{
			result("2", 7, "Teszt Tamás", attempts(900)),
		}},
	)

	repo := new(mockRepo)
	repo.On("ListRecordedTimes", testCompetitionID).Return([]RecordedTime{}, nil)

	handler := NewHandler(newTestService(repo, server.URL, NewBroadcaster()))

	get := func(url string) (*httptest.ResponseRecorder, Report) {
		rec := httptest.NewRecorder()
		handler.GetReport(rec, httptest.NewRequest(http.MethodGet, url, nil))
		var report Report
		_ = json.Unmarshal(rec.Body.Bytes(), &report)
		return rec, report
	}

	t.Run("round=2 csak a második kört adja", func(t *testing.T) {
		rec, report := get("/special/report?competition=" + testCompetitionID + "&round=2")
		require.Equal(t, http.StatusOK, rec.Code)
		require.Len(t, report.Competitors, 1)
		require.Len(t, report.Competitors[0].Rounds, 1)
		assert.Equal(t, "333-r2", report.Competitors[0].Rounds[0].RoundWCIFID)
		assert.Equal(t, []int{1, 2}, report.AvailableRounds)
	})

	t.Run("kör nélkül minden kör", func(t *testing.T) {
		rec, report := get("/special/report?competition=" + testCompetitionID)
		require.Equal(t, http.StatusOK, rec.Code)
		assert.Len(t, report.Competitors[0].Rounds, 2)
		assert.Nil(t, report.Round)
		assert.Equal(t, testCompetitionID, report.Competition.CompetitionWCAID)
	})

	t.Run("érvénytelen kör 400", func(t *testing.T) {
		rec, _ := get("/special/report?competition=" + testCompetitionID + "&round=abc")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("verseny nélkül 400", func(t *testing.T) {
		rec, _ := get("/special/report?round=1")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("hibás alakú verseny-azonosító 400", func(t *testing.T) {
		rec, _ := get("/special/report?competition=..%2Fcompetitions%2F1")
		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestHandler_GetCompetitions(t *testing.T) {
	server := fakeWCALive(t)
	handler := NewHandler(newTestService(new(mockRepo), server.URL, NewBroadcaster()))

	rec := httptest.NewRecorder()
	handler.GetCompetitions(rec, httptest.NewRequest(http.MethodGet, "/special/competitions", nil))

	require.Equal(t, http.StatusOK, rec.Code)

	var competitions []Competition
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &competitions))
	require.Len(t, competitions, 1)
	assert.Equal(t, testCompetitionID, competitions[0].CompetitionWCAID)
}

func TestParseRoundParam(t *testing.T) {
	t.Run("üres = nincs szűrés", func(t *testing.T) {
		value, err := parseRoundParam("")
		require.NoError(t, err)
		assert.Nil(t, value)
	})

	t.Run("érvényes sorszám", func(t *testing.T) {
		value, err := parseRoundParam("2")
		require.NoError(t, err)
		require.NotNil(t, value)
		assert.Equal(t, 2, *value)
	})

	for _, invalid := range []string{"0", "-1", "abc", "1.5"} {
		t.Run("érvénytelen: "+invalid, func(t *testing.T) {
			_, err := parseRoundParam(invalid)
			assert.Error(t, err)
		})
	}
}
