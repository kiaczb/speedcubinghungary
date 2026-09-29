package prediction

import (
	"strings"
	"testing"
	"time"

	"speedcubinghungary/internal/types"

	"github.com/stretchr/testify/assert"
)

func validNewContest() (string, time.Time, []string, []NewExtraQuestion) {
	return "HungarianOpen2026",
		time.Now().Add(24 * time.Hour),
		[]string{"333", "222"},
		[]NewExtraQuestion{{Prompt: "Hány rekord születik?", Type: types.QUESTION_NUMBER}}
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
		questions []NewExtraQuestion
		wantErr   error
	}{
		{
			"empty prompt",
			[]NewExtraQuestion{{Prompt: "", Type: types.QUESTION_NUMBER}},
			ErrEmptyQuestionPrompt,
		},
		{
			"invalid type",
			[]NewExtraQuestion{{Prompt: "Ki nyer?", Type: types.PredictionQuestionType("banán")}},
			ErrInvalidQuestionType,
		},
		{
			"person type valid",
			[]NewExtraQuestion{{Prompt: "Ki nyeri a 3x3x3-at?", Type: types.QUESTION_PERSON}},
			nil,
		},
		{
			"time type valid",
			[]NewExtraQuestion{{Prompt: "Mi lesz a győztes átlaga?", Type: types.QUESTION_TIME}},
			nil,
		},
		{
			"text type valid",
			[]NewExtraQuestion{{Prompt: "Milyen új trükköt fogunk látni?", Type: types.QUESTION_TEXT}},
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

// A (contest_id, event_id) elsődleges kulcs: a duplikátum korábban csak az
// INSERT-nél, kulcssértésként bukott el — 500-ként, 400 helyett.
func TestValidateNewContest_RejectsDuplicateAndMalformedPodiumEvents(t *testing.T) {
	competitionID, deadline, _, questions := validNewContest()

	assert.ErrorIs(t, ValidateNewContest(competitionID, deadline, []string{"333", "333"}, questions),
		ErrDuplicatePodiumEvent)
	assert.ErrorIs(t, ValidateNewContest(competitionID, deadline, []string{"333; DROP"}, questions),
		ErrInvalidPodiumEvent)

	tooMany := make([]string, MaxPodiumEvents+1)
	for i := range tooMany {
		tooMany[i] = "e" + strings.Repeat("x", i%10) + string(rune('a'+i%26))
	}
	assert.ErrorIs(t, ValidateNewContest(competitionID, deadline, tooMany, questions),
		ErrTooManyPodiumEvents)
}

// A csak szóközből álló prompt nem üres a Go szemében, de a felületen üres
// kérdésként jelenne meg.
func TestValidateNewContest_RejectsBlankAndOversizedPrompts(t *testing.T) {
	competitionID, deadline, podiumEventIDs, _ := validNewContest()

	blank := []NewExtraQuestion{{Prompt: "   \t ", Type: types.QUESTION_TEXT}}
	assert.ErrorIs(t, ValidateNewContest(competitionID, deadline, podiumEventIDs, blank),
		ErrEmptyQuestionPrompt)

	long := []NewExtraQuestion{{Prompt: strings.Repeat("é", MaxPromptLength+1), Type: types.QUESTION_TEXT}}
	assert.ErrorIs(t, ValidateNewContest(competitionID, deadline, podiumEventIDs, long),
		ErrQuestionPromptLong)

	tooMany := make([]NewExtraQuestion, MaxExtraQuestions+1)
	for i := range tooMany {
		tooMany[i] = NewExtraQuestion{Prompt: "Kérdés", Type: types.QUESTION_TEXT}
	}
	assert.ErrorIs(t, ValidateNewContest(competitionID, deadline, podiumEventIDs, tooMany),
		ErrTooManyQuestions)
}

// Minden validációs hiba ErrValidation is — a handler erre az egyre nézve ad
// 400-at, így egy új szabály nem eshet ki a leképezésből.
func TestValidationErrorsShareOneRoot(t *testing.T) {
	for _, err := range []error{
		ErrInvalidContestID, ErrInvalidCompetitionID, ErrPastDeadline,
		ErrNoPodiumEvents, ErrTooManyPodiumEvents, ErrInvalidPodiumEvent,
		ErrDuplicatePodiumEvent, ErrTooManyQuestions, ErrInvalidQuestionType,
		ErrEmptyQuestionPrompt, ErrQuestionPromptLong,
	} {
		assert.ErrorIs(t, err, ErrValidation, err.Error())
	}
	assert.NotErrorIs(t, ErrContestNotFound, ErrValidation)
}
