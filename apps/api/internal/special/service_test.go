package special

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type mockRepo struct {
	mock.Mock
}

func (m *mockRepo) EnsureCompetition(c Competition, userID int) error {
	return m.Called(c, userID).Error(0)
}

func (m *mockRepo) ListRecordedTimes(competitionWCAID string) ([]RecordedTime, error) {
	args := m.Called(competitionWCAID)
	return args.Get(0).([]RecordedTime), args.Error(1)
}

func (m *mockRepo) SetElapsed(
	competitionWCAID string,
	liveResultID, attemptNumber, elapsedCs int,
	note string,
	userID int,
	now time.Time,
) error {
	return m.Called(competitionWCAID, liveResultID, attemptNumber, elapsedCs, note, userID).Error(0)
}

func (m *mockRepo) DeleteElapsed(liveResultID, attemptNumber int) error {
	return m.Called(liveResultID, attemptNumber).Error(0)
}

type fakeRound struct {
	id        string
	eventID   string
	number    int
	timeLimit *liveTimeLimit
	results   []liveResult
}

func result(id string, registrantID int, name string, atts []liveAttempt) liveResult {
	return liveResult{
		ID:       id,
		Attempts: atts,
		Person:   livePerson{RegistrantID: &registrantID, Name: name},
	}
}

func fakeWCALive(t *testing.T, rounds ...fakeRound) *httptest.Server {
	t.Helper()

	byID := make(map[string]fakeRound, len(rounds))
	for _, round := range rounds {
		byID[round.id] = round
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/link/competitions/") {
			id := strings.TrimPrefix(r.URL.Path, "/link/competitions/")
			if id != "SpeedcubingHungary2026" {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			http.Redirect(w, r, "/competitions/11151", http.StatusFound)
			return
		}

		var req struct {
			Query     string         `json:"query"`
			Variables map[string]any `json:"variables"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		w.Header().Set("Content-Type", "application/json")

		if strings.Contains(req.Query, "competitions(filter:") {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{"competitions": []any{
					map[string]any{
						"id": "11151", "wcaId": "SpeedcubingHungary2026",
						"name":      "Speedcubing Hungary 2026",
						"startDate": "2026-08-14", "endDate": "2026-08-15",
						"venues": []any{map[string]any{
							"country": map[string]any{"iso2": "HU", "name": "Hungary"},
						}},
					},
					map[string]any{
						"id": "11188", "wcaId": "UFROalCubo2026",
						"name":      "UFRO al Cubo 2026",
						"startDate": "2026-08-22", "endDate": "2026-08-22",
						"venues": []any{map[string]any{
							"country": map[string]any{"iso2": "CL", "name": "Chile"},
						}},
					},
				}},
			})
			return
		}

		if strings.Contains(req.Query, "competition(id:") {
			events := make(map[string][]any)
			order := make([]string, 0)
			for _, round := range rounds {
				if _, seen := events[round.eventID]; !seen {
					order = append(order, round.eventID)
				}
				events[round.eventID] = append(events[round.eventID], map[string]any{
					"id": round.id, "number": round.number, "open": true,
					"finished": false, "active": true, "timeLimit": round.timeLimit,
				})
			}

			competitionEvents := make([]any, 0, len(order))
			for _, eventID := range order {
				competitionEvents = append(competitionEvents, map[string]any{
					"event":  map[string]any{"id": eventID},
					"rounds": events[eventID],
				})
			}

			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"competition": map[string]any{
						"id":                "11151",
						"wcaId":             "SpeedcubingHungary2026",
						"name":              "Speedcubing Hungary 2026",
						"startDate":         "2026-08-14",
						"endDate":           "2026-08-15",
						"competitionEvents": competitionEvents,
					},
				},
			})
			return
		}

		if strings.Contains(req.Query, "round(id:") {
			id, _ := req.Variables["id"].(string)
			round, known := byID[id]
			if !known {
				_, _ = w.Write([]byte(`{"errors":[{"message":"round not found"}]}`))
				return
			}

			_ = json.NewEncoder(w).Encode(map[string]any{
				"data": map[string]any{
					"round": liveRound{
						ID:               round.id,
						Number:           round.number,
						Open:             true,
						TimeLimit:        round.timeLimit,
						Results:          round.results,
						CompetitionEvent: &liveCompEventRef{Event: liveEvent{ID: round.eventID}},
					},
				},
			})
			return
		}

		w.WriteHeader(http.StatusBadRequest)
	}))

	t.Cleanup(server.Close)
	return server
}

func newTestService(repo Repository, liveURL string, broadcaster *Broadcaster) *Service {
	client := NewLiveClient(liveURL)
	tracker := NewTracker(client, NewSocketClient(liveURL), broadcaster)
	return NewService(repo, client, tracker, broadcaster)
}

const testCompetitionID = "SpeedcubingHungary2026"

func testCompetition() Competition {
	return Competition{
		CompetitionWCAID:  testCompetitionID,
		LiveCompetitionID: 11151,
		Name:              "Speedcubing Hungary 2026",
		StartDate:         "2026-08-14",
		EndDate:           "2026-08-15",
	}
}

func TestService_ResolveCompetition(t *testing.T) {
	server := fakeWCALive(t)

	competition, err := newTestService(new(mockRepo), server.URL, NewBroadcaster()).
		ResolveCompetition(context.Background(), "  SpeedcubingHungary2026  ")

	require.NoError(t, err)
	assert.Equal(t, testCompetition(), competition)
}

func TestService_ResolveCompetition_Validation(t *testing.T) {
	server := fakeWCALive(t)
	service := newTestService(new(mockRepo), server.URL, NewBroadcaster())

	for _, invalid := range []string{"", "   ", "../competitions/1", "Speedcubing Hungary", "a/b"} {
		t.Run("érvénytelen: "+invalid, func(t *testing.T) {
			_, err := service.ResolveCompetition(context.Background(), invalid)
			assert.ErrorIs(t, err, ErrInvalidCompetitionID)
		})
	}

	t.Run("ismeretlen verseny", func(t *testing.T) {
		_, err := service.ResolveCompetition(context.Background(), "NincsIlyen2026")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "not found on WCA Live")
	})
}

func TestService_ResolveCompetition_Caches(t *testing.T) {
	server := fakeWCALive(t)
	service := newTestService(new(mockRepo), server.URL, NewBroadcaster())

	_, err := service.ResolveCompetition(context.Background(), testCompetitionID)
	require.NoError(t, err)

	server.Close()

	competition, err := service.ResolveCompetition(context.Background(), testCompetitionID)
	require.NoError(t, err, "a második feloldás a gyorsítótárból jön")
	assert.Equal(t, testCompetition(), competition)
}

func TestService_Competitions(t *testing.T) {
	server := fakeWCALive(t)
	service := newTestService(new(mockRepo), server.URL, NewBroadcaster())

	t.Run("keresés nélkül csak a magyar versenyek", func(t *testing.T) {
		competitions, err := service.Competitions(context.Background(), "")
		require.NoError(t, err)
		require.Len(t, competitions, 1)
		assert.Equal(t, testCompetitionID, competitions[0].CompetitionWCAID)
		assert.Equal(t, "HU", competitions[0].CountryISO2)
	})

	t.Run("keresésre az országtól függetlenül minden találat", func(t *testing.T) {
		competitions, err := service.Competitions(context.Background(), "cubo")
		require.NoError(t, err)
		assert.Len(t, competitions, 2)
	})
}

func TestSortCompetitions(t *testing.T) {
	competitions := []Competition{
		{CompetitionWCAID: "Regi", StartDate: "2026-07-01", EndDate: "2026-07-01"},
		{CompetitionWCAID: "Jovobeli", StartDate: "2026-09-01", EndDate: "2026-09-01"},
		{CompetitionWCAID: "Fut", StartDate: "2026-08-14", EndDate: "2026-08-15"},
		{CompetitionWCAID: "Regebbi", StartDate: "2026-06-01", EndDate: "2026-06-01"},
	}

	sortCompetitions(competitions, "2026-08-14")

	order := make([]string, 0, len(competitions))
	for _, competition := range competitions {
		order = append(order, competition.CompetitionWCAID)
	}
	assert.Equal(t, []string{"Fut", "Jovobeli", "Regi", "Regebbi"}, order)
}

func TestService_RecordElapsed(t *testing.T) {
	server := fakeWCALive(t)

	repo := new(mockRepo)
	repo.On("EnsureCompetition", testCompetition(), 7).Return(nil)
	repo.On("SetElapsed", testCompetitionID, 900, 2, 3742, "bíró jegyezte", 7).Return(nil)

	service := newTestService(repo, server.URL, NewBroadcaster())

	elapsed := 3742
	require.NoError(t, service.RecordElapsed(
		context.Background(), testCompetitionID, 900, 2, &elapsed, "  bíró jegyezte  ", 7,
	))

	repo.AssertExpectations(t)
}

func TestService_RecordElapsed_DeletesOnNil(t *testing.T) {
	repo := new(mockRepo)
	repo.On("DeleteElapsed", 900, 2).Return(nil)

	service := newTestService(repo, "http://example.invalid", NewBroadcaster())

	require.NoError(t, service.RecordElapsed(
		context.Background(), testCompetitionID, 900, 2, nil, "", 7,
	))

	repo.AssertExpectations(t)
	repo.AssertNotCalled(t, "SetElapsed", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestService_RecordElapsed_NotifiesSubscribers(t *testing.T) {
	repo := new(mockRepo)
	repo.On("DeleteElapsed", 900, 2).Return(nil)

	broadcaster := NewBroadcaster()
	events, unsubscribe := broadcaster.Subscribe(testCompetitionID)
	defer unsubscribe()

	otherEvents, unsubscribeOther := broadcaster.Subscribe("MasikVerseny2026")
	defer unsubscribeOther()

	service := newTestService(repo, "http://example.invalid", broadcaster)

	require.NoError(t, service.RecordElapsed(
		context.Background(), testCompetitionID, 900, 2, nil, "", 7,
	))

	select {
	case <-events:
	default:
		t.Fatal("a rögzítésről értesíteni kell a verseny nyitott felületeit")
	}

	select {
	case <-otherEvents:
		t.Fatal("másik verseny nézőit nem szabad értesíteni")
	default:
	}
}

func TestService_RecordElapsed_Validation(t *testing.T) {
	service := newTestService(new(mockRepo), "http://example.invalid", NewBroadcaster())
	ctx := context.Background()

	t.Run("időtartam", func(t *testing.T) {
		for _, invalid := range []int{0, -1, -3742, maxElapsedCs + 1} {
			value := invalid
			err := service.RecordElapsed(ctx, testCompetitionID, 900, 2, &value, "", 7)
			assert.ErrorIs(t, err, ErrInvalidElapsed, "elutasítjuk: %d", invalid)
		}
	})

	t.Run("kísérlet-hivatkozás", func(t *testing.T) {
		elapsed := 3742
		assert.ErrorIs(t, service.RecordElapsed(ctx, testCompetitionID, 0, 2, &elapsed, "", 7), ErrInvalidAttempt)
		assert.ErrorIs(t, service.RecordElapsed(ctx, testCompetitionID, 900, 0, &elapsed, "", 7), ErrInvalidAttempt)
	})

	t.Run("verseny-azonosító", func(t *testing.T) {
		elapsed := 3742
		assert.ErrorIs(t, service.RecordElapsed(ctx, "  ", 900, 2, &elapsed, "", 7), ErrInvalidCompetitionID)
	})
}

func TestService_GetReport(t *testing.T) {
	server := fakeWCALive(t, fakeRound{
		id: "147464", eventID: "333bf", number: 1,
		timeLimit: &liveTimeLimit{Centiseconds: 60000, CumulativeRoundWCIFIDs: []string{"333bf-r1"}},
		results: []liveResult{
			result("900", 7, "Kumulatív Kata", attempts(20000, AttemptDNF, AttemptDNS)),
		},
	})

	repo := new(mockRepo)
	repo.On("ListRecordedTimes", testCompetitionID).
		Return([]RecordedTime{recorded(900, 2, 3742)}, nil)

	report, err := newTestService(repo, server.URL, NewBroadcaster()).
		GetReport(context.Background(), testCompetitionID, nil)

	require.NoError(t, err)
	assert.Equal(t, testCompetitionID, report.Competition.CompetitionWCAID)

	require.Len(t, report.Competitors, 1)
	competitor := report.Competitors[0]
	assert.Equal(t, "Kumulatív Kata", competitor.Name)
	require.Len(t, competitor.Rounds, 1)
	require.Len(t, competitor.Rounds[0].Attempts, 3)
	assert.Equal(t, 23742, competitor.UsedCs)

	assert.Equal(t, 900, competitor.Rounds[0].LiveResultID)

	dnfAttempt := competitor.Rounds[0].Attempts[1]
	assert.Equal(t, AttemptDNF, dnfAttempt.Result)
	require.NotNil(t, dnfAttempt.ElapsedCs)
	assert.Equal(t, 3742, *dnfAttempt.ElapsedCs)

	require.Len(t, competitor.Budgets, 1)
	assert.Equal(t, 36258, competitor.Budgets[0].RemainingCs)
}

func TestService_GetReport_FiltersByRoundNumber(t *testing.T) {
	server := fakeWCALive(t,
		fakeRound{id: "1", eventID: "333", number: 1, results: []liveResult{
			result("1", 7, "Teszt Tamás", attempts(1000)),
		}},
		fakeRound{id: "2", eventID: "222", number: 1, results: []liveResult{
			result("2", 7, "Teszt Tamás", attempts(500)),
		}},
		fakeRound{id: "3", eventID: "333", number: 2, results: []liveResult{
			result("3", 7, "Teszt Tamás", attempts(900)),
		}},
	)

	repo := new(mockRepo)
	repo.On("ListRecordedTimes", testCompetitionID).Return([]RecordedTime{}, nil)

	service := newTestService(repo, server.URL, NewBroadcaster())

	t.Run("első kör", func(t *testing.T) {
		round := 1
		report, err := service.GetReport(context.Background(), testCompetitionID, &round)
		require.NoError(t, err)

		require.Len(t, report.Competitors, 1)
		rounds := report.Competitors[0].Rounds
		require.Len(t, rounds, 2, "minden versenyszám első köre")
		assert.ElementsMatch(t,
			[]string{"333-r1", "222-r1"},
			[]string{rounds[0].RoundWCIFID, rounds[1].RoundWCIFID},
		)
		assert.Equal(t, 1500, report.Competitors[0].UsedCs)

		assert.Equal(t, []int{1, 2}, report.AvailableRounds)
		require.NotNil(t, report.Round)
		assert.Equal(t, 1, *report.Round)
	})

	t.Run("második kör", func(t *testing.T) {
		round := 2
		report, err := service.GetReport(context.Background(), testCompetitionID, &round)
		require.NoError(t, err)

		require.Len(t, report.Competitors, 1)
		require.Len(t, report.Competitors[0].Rounds, 1)
		assert.Equal(t, "333-r2", report.Competitors[0].Rounds[0].RoundWCIFID)
		assert.Equal(t, 900, report.Competitors[0].UsedCs)
	})

	t.Run("szűrés nélkül minden kör", func(t *testing.T) {
		report, err := service.GetReport(context.Background(), testCompetitionID, nil)
		require.NoError(t, err)

		require.Len(t, report.Competitors, 1)
		assert.Len(t, report.Competitors[0].Rounds, 3)
		assert.Nil(t, report.Round)
	})
}

func TestService_GetReport_UnknownCompetition(t *testing.T) {
	server := fakeWCALive(t)
	repo := new(mockRepo)

	_, err := newTestService(repo, server.URL, NewBroadcaster()).
		GetReport(context.Background(), "NincsIlyen2026", nil)

	assert.Error(t, err)
	repo.AssertNotCalled(t, "ListRecordedTimes", mock.Anything)
}

func TestLiveClient_GetRound(t *testing.T) {
	server := fakeWCALive(t, fakeRound{
		id: "147464", eventID: "333", number: 1,
		results: []liveResult{result("4973030", 37, "Dávid Szabó", attempts(696, AttemptDNF, 674))},
	})

	round, err := NewLiveClient(server.URL).GetRound(context.Background(), 147464)

	require.NoError(t, err)
	assert.Equal(t, "147464", round.ID)
	require.Len(t, round.Results, 1)
	assert.Equal(t, AttemptDNF, round.Results[0].Attempts[1].Result)
	require.NotNil(t, round.CompetitionEvent)
	assert.Equal(t, "333", round.CompetitionEvent.Event.ID)
}

func TestLiveClient_GraphQLErrorIsSurfaced(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"errors":[{"message":"round not found"}]}`))
	}))
	defer server.Close()

	_, err := NewLiveClient(server.URL).GetRound(context.Background(), 1)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "round not found")
}
