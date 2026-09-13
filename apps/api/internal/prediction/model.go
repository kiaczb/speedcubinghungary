package prediction

import (
	"time"

	"speedcubinghungary/internal/types"
)

// PredictionStatus a leadási határidőből számolt állapot — nem DB oszlop.
// A kiértékelés (evaluated) még nem része ennek a csomagnak.
type PredictionStatus string

const (
	StatusOpen   PredictionStatus = "open"
	StatusClosed PredictionStatus = "closed"
)

// ExtraQuestion egy admin által megadott szabad kérdés (pl. "Hány rekord
// születik?"). A JSON kulcsok a frontend app/tools/prediction/types.ts
// ExtraQuestion típusával egyeznek.
type ExtraQuestion struct {
	ID     string                       `json:"id"`
	Prompt string                       `json:"prompt"`
	Type   types.PredictionQuestionType `json:"type"`
}

// Contest egy tippverseny — egy WCA versenyhez tartozó dobogó-tippek +
// extra kérdések konfigurációja. A JSON kulcsok a frontend
// PredictionContest típusával egyeznek (app/tools/prediction/types.ts).
type Contest struct {
	ID                   string           `json:"id"`
	CompetitionID        string           `json:"competitionId"`
	CompetitionName      string           `json:"competitionName"`
	CompetitionCity      string           `json:"competitionCity"`
	CompetitionCountry   string           `json:"competitionCountry"`
	CompetitionStartDate string           `json:"competitionStartDate"`
	PredictionDeadline   time.Time        `json:"predictionDeadline"`
	Status               PredictionStatus `json:"status"`
	ParticipantCount     int              `json:"participantCount"`
	PodiumEventIDs       []string         `json:"podiumEventIds"`
	ExtraQuestions       []ExtraQuestion  `json:"extraQuestions"`
}

// statusFor a leadási határidőből számolja az állapotot.
func statusFor(deadline time.Time) PredictionStatus {
	if time.Now().After(deadline) {
		return StatusClosed
	}
	return StatusOpen
}
