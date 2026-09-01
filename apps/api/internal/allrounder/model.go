package allrounder

import "time"

type Result struct {
	Ordinal     int
	WCAID       string
	Name        string
	EventID     string
	RoundTypeID string
	Pos         int
	Best        int
}

type SyncState struct {
	CompetitionWCAID string
	FetchedAt        time.Time
	ResultCount      int
	LastError        string

	StartDate         *time.Time
	EndDate           *time.Time
	LiveCompetitionID *int
}

type CompetitionInfo struct {
	StartDate time.Time
	EndDate   time.Time
}

type wcaResult struct {
	Pos         int    `json:"pos"`
	Best        int    `json:"best"`
	Name        string `json:"name"`
	EventID     string `json:"event_id"`
	RoundTypeID string `json:"round_type_id"`
	WCAID       string `json:"wca_id"`
}

type EventResult struct {
	CompetitionWCAID string `json:"competitionWcaId"`
	EventID          string `json:"eventId"`
	Ranking          int    `json:"ranking"`
	Points           int    `json:"points"`
}

type PersonStandings struct {
	WCAID       string        `json:"wcaId"`
	Name        string        `json:"name"`
	TotalPoints int           `json:"totalPoints"`
	Results     []EventResult `json:"results"`
}

type Standings struct {
	Year            int                 `json:"year"`
	AvailableYears  []int               `json:"availableYears"`
	PointsTable     []int               `json:"pointsTable"`
	CountingResults int                 `json:"countingResults"`
	PodiumCount     int                 `json:"podiumCount"`
	Competitions    []SeriesCompetition `json:"competitions"`
	People          []PersonStandings   `json:"people"`
	UpdatedAt       *time.Time          `json:"updatedAt"`
	Live            *LiveLeg            `json:"live,omitempty"`
}

type LiveLeg struct {
	CompetitionWCAID string `json:"competitionWcaId"`
	Label            string `json:"label"`
	RoundCount       int    `json:"roundCount"`
}
