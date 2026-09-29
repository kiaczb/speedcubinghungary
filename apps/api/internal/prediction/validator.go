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

// ErrValidation minden validációs hiba közös gyökere: a handler EGYETLEN
// errors.Is-szel dönti el, hogy a hiba a kliensé (400) — nem egy hosszú, minden
// új hibával bővítendő láncból, amiből egy kifelejtett elem 500-at csinálna.
var ErrValidation = errors.New("invalid prediction contest")

type validationError string

func (e validationError) Error() string        { return string(e) }
func (e validationError) Is(target error) bool { return target == ErrValidation }

const (
	// MaxPodiumEvents: a WCA-nak 17 versenyszáma van, ennyi bőven elég.
	MaxPodiumEvents = 20
	// MaxExtraQuestions és MaxPromptLength: egy tippverseny űrlapja maradjon
	// kitölthető, és egy kérdés ne lehessen tetszőleges méretű szöveg a DB-ben.
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

// competitionIDPattern: ugyanaz a korlát, mint a special csomagban a WCA
// versenyazonosítóra (internal/special/service.go) — a kliens adja meg, és
// egy backend-generált URL-be kerül, tehát validálni kell.
var competitionIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,64}$`)

// eventIDPattern: a WCA versenyszám-azonosítók kisbetűs alfanumerikusak
// ("333", "333bf", "minx", "sq1").
var eventIDPattern = regexp.MustCompile(`^[a-z0-9]{1,16}$`)

func isValidQuestionType(t types.PredictionQuestionType) bool {
	switch t {
	case types.QUESTION_PERSON, types.QUESTION_TIME, types.QUESTION_NUMBER, types.QUESTION_TEXT:
		return true
	}
	return false
}

// validContestID: a prediction_contests.id UUID oszlop, a nem UUID alakú
// azonosítót a Postgres típushibával utasítaná el, ami 500 lenne.
func validContestID(id string) error {
	if _, err := uuid.Parse(id); err != nil {
		return ErrInvalidContestID
	}
	return nil
}

// ValidateNewContest a létrehozási kérést ellenőrzi, mielőtt bármit írnánk.
// A kérdések promptját a hívó már levágott (TrimSpace) formában adja át.
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
	if len(podiumEventIDs) > MaxPodiumEvents {
		return ErrTooManyPodiumEvents
	}
	seen := make(map[string]bool, len(podiumEventIDs))
	for _, eventID := range podiumEventIDs {
		if !eventIDPattern.MatchString(eventID) {
			return ErrInvalidPodiumEvent
		}
		// A (contest_id, event_id) az elsődleges kulcs: egy duplikátum
		// kulcssértésként, 500-ként jönne vissza az INSERT-ből.
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
