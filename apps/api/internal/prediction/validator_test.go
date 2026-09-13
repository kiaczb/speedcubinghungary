package prediction

import (
	"testing"
	"time"

	"speedcubinghungary/internal/types"

	"github.com/stretchr/testify/assert"
)

func validNewContest() (string, time.Time, []string, []ExtraQuestion) {
	return "HungarianOpen2026",
		time.Now().Add(24 * time.Hour),
		[]string{"333", "222"},
		[]ExtraQuestion{{Prompt: "Hány rekord születik?", Type: types.QUESTION_NUMBER}}
}

func TestValidateNewContest_Valid(t *testing.T) {
	competitionID, deadline, podiumEventIDs, questions := validNewContest()
	assert.NoError(t, ValidateNewContest(competitionID, deadline, podiumEventIDs, questions))
}

func TestValidateNewContest_CompetitionID(t *testing.T) {
	_, deadline, podiumEventIDs, questions := validNewContest()

	tests := []struct {
		name          string
		competitionID string
		wantErr       error
	}{
		{"empty", "", ErrInvalidCompetitionID},
		{"too long", string(make([]byte, 65)), ErrInvalidCompetitionID},
		{"invalid characters", "Hungarian Open 2026!", ErrInvalidCompetitionID},
		{"valid", "HungarianOpen2026", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNewContest(tt.competitionID, deadline, podiumEventIDs, questions)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestValidateNewContest_Deadline(t *testing.T) {
	competitionID, _, podiumEventIDs, questions := validNewContest()

	t.Run("past deadline", func(t *testing.T) {
		err := ValidateNewContest(competitionID, time.Now().Add(-time.Hour), podiumEventIDs, questions)
		assert.Equal(t, ErrPastDeadline, err)
	})

	t.Run("future deadline", func(t *testing.T) {
		err := ValidateNewContest(competitionID, time.Now().Add(time.Hour), podiumEventIDs, questions)
		assert.NoError(t, err)
	})
}

func TestValidateNewContest_PodiumEvents(t *testing.T) {
	competitionID, deadline, _, questions := validNewContest()

	err := ValidateNewContest(competitionID, deadline, nil, questions)
	assert.Equal(t, ErrNoPodiumEvents, err)
}

func TestValidateNewContest_ExtraQuestions(t *testing.T) {
	competitionID, deadline, podiumEventIDs, _ := validNewContest()

	tests := []struct {
		name      string
		questions []ExtraQuestion
		wantErr   error
	}{
		{
			"empty prompt",
			[]ExtraQuestion{{Prompt: "", Type: types.QUESTION_NUMBER}},
			ErrEmptyQuestionPrompt,
		},
		{
			"invalid type",
			[]ExtraQuestion{{Prompt: "Ki nyer?", Type: types.PredictionQuestionType("banán")}},
			ErrInvalidQuestionType,
		},
		{
			"person type valid",
			[]ExtraQuestion{{Prompt: "Ki nyeri a 3x3x3-at?", Type: types.QUESTION_PERSON}},
			nil,
		},
		{
			"time type valid",
			[]ExtraQuestion{{Prompt: "Mi lesz a győztes átlaga?", Type: types.QUESTION_TIME}},
			nil,
		},
		{
			"no extra questions is fine",
			nil,
			nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNewContest(competitionID, deadline, podiumEventIDs, tt.questions)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
