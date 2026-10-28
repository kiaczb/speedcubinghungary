package certificate

import (
	"context"
	"errors"
	"testing"
	"time"

	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/stats"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubCompetitions struct {
	competition   competition.Competition
	registrations competition.Registrations
	results       []competition.Result
	resultsCalls  int
}

func (s *stubCompetitions) GetCompetition(string) (competition.Competition, error) {
	return s.competition, nil
}

func (s *stubCompetitions) GetRegistrations(string) (competition.Registrations, error) {
	return s.registrations, nil
}

func (s *stubCompetitions) GetResults(string) ([]competition.Result, error) {
	s.resultsCalls++
	return s.results, nil
}

type stubLive struct {
	data  *stats.LiveCompetition
	at    time.Time
	err   error
	calls int
}

func (s *stubLive) LiveData(context.Context, string) (*stats.LiveCompetition, time.Time, error) {
	s.calls++
	return s.data, s.at, s.err
}

func newTestService(competitions *stubCompetitions, live *stubLive, now time.Time) *Service {
	service := NewService(nil, competitions, live)
	service.now = func() time.Time { return now }
	return service
}

func testCompetition() competition.Competition {
	return competition.Competition{
		ID: "HungarianOpen2026", Name: "Hungarian Open 2026",
		StartDate: "2026-10-24", EndDate: "2026-10-25",
		EventIDs: []string{"333", "222"},
	}
}

func TestGetRecipients_ResultSourceFollowsTheCompetitionPhase(t *testing.T) {
	first := 1
	live := &stubLive{
		at: time.Date(2026, 10, 24, 14, 32, 0, 0, time.UTC),
		data: &stats.LiveCompetition{Rounds: []stats.LiveRound{
			{EventID: "333", Number: 1, Open: true, Finished: true, Results: []stats.LiveResult{
				{Person: stats.Person{Name: "Anna"}, Ranking: &first, Best: 800, Average: 900, Attempts: []int{800}},
			}},
		}},
	}

	t.Run("a verseny előtt nincs még eredmény", func(t *testing.T) {
		competitions := &stubCompetitions{competition: testCompetition()}
		service := newTestService(competitions, live, time.Date(2026, 10, 23, 21, 0, 0, 0, time.UTC))

		list, err := service.GetRecipients(context.Background(), "HungarianOpen2026", ModePodium)

		require.NoError(t, err)
		assert.Equal(t, SourceNone, list.Source)
		assert.Empty(t, list.Recipients)
		assert.Equal(t, []Event{{"333", EventPending}, {"222", EventPending}}, list.Events)
		assert.Zero(t, live.calls)
	})

	t.Run("budapesti idő szerint az első napon már a WCA Live a forrás", func(t *testing.T) {
		live.calls = 0
		competitions := &stubCompetitions{competition: testCompetition()}
		service := newTestService(competitions, live, time.Date(2026, 10, 23, 22, 30, 0, 0, time.UTC))

		list, err := service.GetRecipients(context.Background(), "HungarianOpen2026", ModeCompetitors)

		require.NoError(t, err)
		assert.Equal(t, SourceLive, list.Source)
		require.NotNil(t, list.FetchedAt)
		assert.Equal(t, live.at, *list.FetchedAt)
		assert.Equal(t, []Recipient{{Name: "Anna"}}, list.Recipients)
		assert.Empty(t, list.Events, "versenyszám-állapot csak dobogós módban van")
	})

	t.Run("a hivatalos eredmények felkerülése után a WCA API a forrás", func(t *testing.T) {
		live.calls = 0
		comp := testCompetition()
		comp.ResultsPostedAt = "2026-10-27T10:00:00.000Z"
		competitions := &stubCompetitions{competition: comp, results: []competition.Result{
			{WCAID: "2010AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 800, Average: 900, Attempts: []int{800}},
		}}
		service := newTestService(competitions, live, time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC))

		list, err := service.GetRecipients(context.Background(), "HungarianOpen2026", ModePodium)

		require.NoError(t, err)
		assert.Equal(t, SourceResults, list.Source)
		assert.Nil(t, list.FetchedAt)
		require.Len(t, list.Recipients, 1)
		assert.Equal(t, "9.00", list.Recipients[0].Result)
		assert.Zero(t, live.calls)
	})
}

func TestGetRecipients_RegistrantsComeFromTheRegistrationListInEveryPhase(t *testing.T) {
	comp := testCompetition()
	comp.ResultsPostedAt = "2026-10-27T10:00:00.000Z"
	competitions := &stubCompetitions{
		competition:   comp,
		registrations: competition.Registrations{Registrants: []competition.Registrant{{Name: "Anna", WCAID: "2010AAAA01", CountryISO2: "HU"}}},
	}
	service := newTestService(competitions, &stubLive{}, time.Date(2026, 10, 28, 12, 0, 0, 0, time.UTC))

	list, err := service.GetRecipients(context.Background(), "HungarianOpen2026", ModeRegistrants)

	require.NoError(t, err)
	assert.Equal(t, SourceRegistrations, list.Source)
	assert.Equal(t, []Recipient{{Name: "Anna", WCAID: "2010AAAA01", CountryISO2: "HU"}}, list.Recipients)
	assert.Zero(t, competitions.resultsCalls)
}

func TestGetRecipients_LiveFailureIsAnError(t *testing.T) {
	competitions := &stubCompetitions{competition: testCompetition()}
	service := newTestService(competitions, &stubLive{err: stats.ErrNotOnLive}, time.Date(2026, 10, 24, 12, 0, 0, 0, time.UTC))

	_, err := service.GetRecipients(context.Background(), "HungarianOpen2026", ModePodium)

	assert.True(t, errors.Is(err, stats.ErrNotOnLive))
}

func TestParseMode(t *testing.T) {
	for _, value := range []string{"podium", "competitors", "registrants"} {
		mode, err := ParseMode(value)
		require.NoError(t, err)
		assert.Equal(t, Mode(value), mode)
	}
	for _, value := range []string{"", "Podium", "all"} {
		_, err := ParseMode(value)
		assert.ErrorIs(t, err, ErrInvalidMode, value)
	}
}
