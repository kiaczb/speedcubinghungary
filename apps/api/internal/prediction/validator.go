package prediction

import (
	"errors"
	"regexp"
	"time"

	"speedcubinghungary/internal/types"
)

var (
	ErrInvalidCompetitionID = errors.New("invalid competition id")
	ErrPastDeadline         = errors.New("prediction deadline must be in the future")
	ErrNoPodiumEvents       = errors.New("at least one podium event is required")
	ErrInvalidQuestionType  = errors.New("invalid extra question type")
	ErrEmptyQuestionPrompt  = errors.New("extra question prompt cannot be empty")
	ErrContestNotFound      = errors.New("prediction contest not found")
)

// competitionIDPattern: ugyanaz a korlát, mint a special csomagban a WCA
// versenyazonosítóra (internal/special/service.go) — a kliens adja meg, és
// egy backend-generált URL-be kerül, tehát validálni kell.
var competitionIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,64}$`)

func isValidQuestionType(t types.PredictionQuestionType) bool {
	switch t {
	case types.QUESTION_PERSON, types.QUESTION_TIME, types.QUESTION_NUMBER, types.QUESTION_TEXT:
		return true
	}
	return false
}

// ValidateNewContest a létrehozási kérést ellenőrzi, mielőtt bármit írnánk.
func ValidateNewContest(competitionID string, deadline time.Time, podiumEventIDs []string, questions []ExtraQuestion) error {
	if !competitionIDPattern.MatchString(competitionID) {
		return ErrInvalidCompetitionID
	}
	if !deadline.After(time.Now()) {
		return ErrPastDeadline
	}
	if len(podiumEventIDs) == 0 {
		return ErrNoPodiumEvents
	}
	for _, q := range questions {
		if q.Prompt == "" {
			return ErrEmptyQuestionPrompt
		}
		if !isValidQuestionType(q.Type) {
			return ErrInvalidQuestionType
		}
	}
	return nil
}
