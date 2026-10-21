package prediction

import (
	"time"

	"speedcubinghungary/internal/types"
)

type PredictionStatus string

const (
	StatusOpen   PredictionStatus = "open"
	StatusClosed PredictionStatus = "closed"
)

type ExtraQuestion struct {
	ID     string                       `json:"id"`
	Prompt string                       `json:"prompt"`
	Type   types.PredictionQuestionType `json:"type"`
}

type NewExtraQuestion struct {
	Prompt string                       `json:"prompt"`
	Type   types.PredictionQuestionType `json:"type"`
}

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

func statusFor(deadline time.Time) PredictionStatus {
	if time.Now().After(deadline) {
		return StatusClosed
	}
	return StatusOpen
}
