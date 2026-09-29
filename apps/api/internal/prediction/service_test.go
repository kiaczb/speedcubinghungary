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

// stubLookup csak azokat a versenyeket ismeri, amik a mapben vannak — minden
// mást a WCA ErrNotFound-ja szerint utasít el, vagy a megadott hibával.
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

// ⚠️ A tippverseny a versenyt ID szerint kérdezi le. Korábban a KÖZELGŐ
// versenyek listájából töltötte a nevet — ami a versenyt a vége után kiejti,
// így a tippverseny pont az eredményhirdetéskor maradt név, dátum és helyszín
// nélkül. Itt a lookup egy már lezajlott versenyt is ismer.
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

// Egy WCA-kiesés nem ürítheti ki a listát: a tippverseny ott marad, a
// verseny azonosítójával mint névvel.
func TestGetContests_FallsBackToTheIDWhenTheLookupFails(t *testing.T) {
	repo := &stubRepository{rows: []contestRow{{ID: "c1", CompetitionWCAID: "Valami2026"}}}

	contests, err := NewService(repo, stubLookup{err: errors.New("WCA API returned status 503")}).GetContests()
	require.NoError(t, err)
	require.Len(t, contests, 1)

	assert.Equal(t, "Valami2026", contests[0].CompetitionName)
}

// A prediction_contests.id UUID: egy nem UUID azonosító korábban lejutott a
// Postgresig, és típushibaként 500-at adott.
func TestDeleteContest_RejectsANonUUIDBeforeTheDatabase(t *testing.T) {
	repo := &stubRepository{}

	err := NewService(repo, stubLookup{}).DeleteContest("nem-uuid")

	assert.ErrorIs(t, err, ErrValidation)
	assert.Empty(t, repo.deletedID, "az adatbázisig el sem juthat")
}

// A created_by oszlopot korábban semmi nem töltötte; a prompt levágva kerül
// tárolásra, hogy a felületen ne legyenek láthatatlan szóközök.
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
