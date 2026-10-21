package prediction

import (
	"errors"
	"testing"
	"time"

	"speedcubinghungary/internal/competition"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubRepository struct {
	rows      []contestRow
	deletedID string
	created   *contestRow
}

func (s *stubRepository) GetAll() ([]contestRow, error) { return s.rows, nil }
func (s *stubRepository) GetPodiumEvents([]string) (map[string][]string, error) {
	return map[string][]string{}, nil
}
func (s *stubRepository) GetExtraQuestions([]string) (map[string][]extraQuestionRow, error) {
	return map[string][]extraQuestionRow{}, nil
}
func (s *stubRepository) Create(row contestRow, _ []string, _ []extraQuestionRow) error {
	s.created = &row
	return nil
}
func (s *stubRepository) Delete(id string) error {
	s.deletedID = id
	return nil
}

type stubLookup struct {
	known map[string]competition.Competition
	err   error
}

func (s stubLookup) GetCompetition(id string) (competition.Competition, error) {
	if s.err != nil {
		return competition.Competition{}, s.err
	}
	comp, ok := s.known[id]
	if !ok {
		return competition.Competition{}, competition.ErrNotFound
	}
	return comp, nil
}

func TestGetContests_DescribesACompetitionThatIsAlreadyOver(t *testing.T) {
	repo := &stubRepository{rows: []contestRow{{
		ID:                 "c1",
		CompetitionWCAID:   "BudapestOpen2025",
		PredictionDeadline: time.Now().Add(-48 * time.Hour),
	}}}
	lookup := stubLookup{known: map[string]competition.Competition{
		"BudapestOpen2025": {
			ID: "BudapestOpen2025", Name: "Budapest Open 2025",
			City: "Budapest", Country: "Hungary", StartDate: "2025-03-01",
		},
	}}

	contests, err := NewService(repo, lookup).GetContests()
	require.NoError(t, err)
	require.Len(t, contests, 1)

	assert.Equal(t, "Budapest Open 2025", contests[0].CompetitionName)
	assert.Equal(t, "Budapest", contests[0].CompetitionCity)
	assert.Equal(t, "2025-03-01", contests[0].CompetitionStartDate)
}

func TestGetContests_FallsBackToTheIDWhenTheLookupFails(t *testing.T) {
	repo := &stubRepository{rows: []contestRow{{ID: "c1", CompetitionWCAID: "Valami2026"}}}

	contests, err := NewService(repo, stubLookup{err: errors.New("WCA API returned status 503")}).GetContests()
	require.NoError(t, err)
	require.Len(t, contests, 1)

	assert.Equal(t, "Valami2026", contests[0].CompetitionName)
}

func TestDeleteContest_RejectsANonUUIDBeforeTheDatabase(t *testing.T) {
	repo := &stubRepository{}

	err := NewService(repo, stubLookup{}).DeleteContest("nem-uuid")

	assert.ErrorIs(t, err, ErrValidation)
	assert.Empty(t, repo.deletedID, "az adatbázisig el sem juthat")
}

func TestCreateContest_RecordsTheCreatorAndTrimsPrompts(t *testing.T) {
	repo := &capturingRepository{}

	err := NewService(repo, stubLookup{}).CreateContest(
		"HungarianOpen2026", time.Now().Add(time.Hour), []string{"333"},
		[]NewExtraQuestion{{Prompt: "  Hány rekord születik?  ", Type: "number"}},
		42,
	)
	require.NoError(t, err)

	assert.Equal(t, 42, repo.row.CreatedBy)
	require.Len(t, repo.questions, 1)
	assert.Equal(t, "Hány rekord születik?", repo.questions[0].Prompt)
}

type capturingRepository struct {
	stubRepository
	row       contestRow
	questions []extraQuestionRow
}

func (c *capturingRepository) Create(row contestRow, _ []string, questions []extraQuestionRow) error {
	c.row, c.questions = row, questions
	return nil
}
