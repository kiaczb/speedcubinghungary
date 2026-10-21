package prediction

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"speedcubinghungary/internal/types"

	"github.com/google/uuid"
)

var ErrValidation = errors.New("invalid prediction contest")

type validationError string

func (e validationError) Error() string        { return string(e) }
func (e validationError) Is(target error) bool { return target == ErrValidation }

const (
	MaxPodiumEvents   = 20
	MaxExtraQuestions = 20
	MaxPromptLength   = 300
)

var (
	ErrInvalidContestID     error = validationError("invalid prediction contest id")
	ErrInvalidCompetitionID error = validationError("invalid competition id")
	ErrPastDeadline         error = validationError("prediction deadline must be in the future")
	ErrNoPodiumEvents       error = validationError("at least one podium event is required")
	ErrTooManyPodiumEvents  error = validationError(fmt.Sprintf("at most %d podium events are allowed", MaxPodiumEvents))
	ErrInvalidPodiumEvent   error = validationError("invalid podium event id")
	ErrDuplicatePodiumEvent error = validationError("a podium event is listed more than once")
	ErrTooManyQuestions     error = validationError(fmt.Sprintf("at most %d extra questions are allowed", MaxExtraQuestions))
	ErrInvalidQuestionType  error = validationError("invalid extra question type")
	ErrEmptyQuestionPrompt  error = validationError("extra question prompt cannot be empty")
	ErrQuestionPromptLong   error = validationError(fmt.Sprintf("extra question prompt must be at most %d characters", MaxPromptLength))

	ErrContestNotFound = errors.New("prediction contest not found")
)

var competitionIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,64}$`)

var eventIDPattern = regexp.MustCompile(`^[a-z0-9]{1,16}$`)

func isValidQuestionType(t types.PredictionQuestionType) bool {
	switch t {
	case types.QUESTION_PERSON, types.QUESTION_TIME, types.QUESTION_NUMBER, types.QUESTION_TEXT:
		return true
	}
	return false
}

func validContestID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidContestID
	}
	return nil
}

func ValidateNewContest(competitionID string, deadline time.Time, podiumEventIDs []string, questions []NewExtraQuestion) error {
	if !competitionIDPattern.MatchString(competitionID) {
		return ErrInvalidCompetitionID
	}
	if !deadline.After(time.Now()) {
		return ErrPastDeadline
	}

	if len(podiumEventIDs) == 0 {
		return ErrNoPodiumEvents
	}
	if len(podiumEventIDs) > MaxPodiumEvents {
		return ErrTooManyPodiumEvents
	}
	seen := make(map[string]bool, len(podiumEventIDs))
	for _, eventID := range podiumEventIDs {
		if !eventIDPattern.MatchString(eventID) {
			return ErrInvalidPodiumEvent
		}
		if seen[eventID] {
			return ErrDuplicatePodiumEvent
		}
		seen[eventID] = true
	}

	if len(questions) > MaxExtraQuestions {
		return ErrTooManyQuestions
	}
	for _, q := range questions {
		if strings.TrimSpace(q.Prompt) == "" {
			return ErrEmptyQuestionPrompt
		}
		if utf8.RuneCountInString(q.Prompt) > MaxPromptLength {
			return ErrQuestionPromptLong
		}
		if !isValidQuestionType(q.Type) {
			return ErrInvalidQuestionType
		}
	}

	return nil
}
