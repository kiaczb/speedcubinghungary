package special

const (
	AttemptSkipped = 0
	AttemptDNF     = -1
	AttemptDNS     = -2
)

type Competition struct {
	CompetitionWCAID  string `json:"competitionWcaId"`
	LiveCompetitionID int    `json:"liveCompetitionId"`
	Name              string `json:"name"`
	StartDate         string `json:"startDate"`
	EndDate           string `json:"endDate"`
	CountryISO2       string `json:"countryIso2,omitempty"`
}

// RecordedTime egy DNF-kísérlet emberileg rögzített tényleges ideje. A jegyzet,
// a rögzítő és az időpont csak auditnak kerül a DB-be, a riport nem használja.
type RecordedTime struct {
	LiveResultID  int
	AttemptNumber int
	ElapsedCs     int
}

type ResultSnapshot struct {
	LiveResultID int

	EventID     string
	RoundNumber int
	RoundWCIFID string

	RegistrantID    int
	CompetitorWCAID string
	CompetitorName  string

	Attempts []int
	Best     *int
	Average  *int

	TimeLimitCs        *int
	CumulativeRoundIDs []string
}

type CompetitorAttempt struct {
	AttemptNumber int  `json:"attemptNumber"`
	Result        int  `json:"result"`
	ElapsedCs     *int `json:"elapsedCs,omitempty"`
	MaxElapsedCs  *int `json:"maxElapsedCs,omitempty"`
}

type CompetitorRound struct {
	LiveResultID       int                 `json:"liveResultId"`
	RoundWCIFID        string              `json:"roundWcifId"`
	EventID            string              `json:"eventId"`
	RoundNumber        int                 `json:"roundNumber"`
	Attempts           []CompetitorAttempt `json:"attempts"`
	Best               *int                `json:"best,omitempty"`
	Average            *int                `json:"average,omitempty"`
	TimeLimitCs        *int                `json:"timeLimitCs,omitempty"`
	CumulativeRoundIDs []string            `json:"cumulativeRoundIds"`
	CompletedCs        int                 `json:"completedCs"`
	RecordedDNFCs      int                 `json:"recordedDnfCs"`
	MissingDNFCount    int                 `json:"missingDnfCount"`
}

type CumulativeBudget struct {
	RoundWCIFIDs    []string `json:"roundWcifIds"`
	LimitCs         int      `json:"limitCs"`
	CompletedCs     int      `json:"completedCs"`
	RecordedDNFCs   int      `json:"recordedDnfCs"`
	UsedCs          int      `json:"usedCs"`
	RemainingCs     int      `json:"remainingCs"`
	MissingDNFCount int      `json:"missingDnfCount"`
}

type Competitor struct {
	RegistrantID    int                `json:"registrantId"`
	CompetitorWCAID string             `json:"competitorWcaId,omitempty"`
	Name            string             `json:"name"`
	Rounds          []CompetitorRound  `json:"rounds"`
	Budgets         []CumulativeBudget `json:"budgets"`
	UsedCs          int                `json:"usedCs"`
	DNFCount        int                `json:"dnfCount"`
	MissingDNFCount int                `json:"missingDnfCount"`
}

type liveCompetition struct {
	ID                string                `json:"id"`
	WCAID             string                `json:"wcaId"`
	Name              string                `json:"name"`
	StartDate         string                `json:"startDate"`
	EndDate           string                `json:"endDate"`
	Venues            []liveVenue           `json:"venues"`
	CompetitionEvents []liveCompetitionEvnt `json:"competitionEvents"`
}

type liveCompetitionEvnt struct {
	Event  liveEvent   `json:"event"`
	Rounds []liveRound `json:"rounds"`
}

type liveEvent struct {
	ID string `json:"id"`
}

type liveRound struct {
	ID        string         `json:"id"`
	Number    int            `json:"number"`
	Name      string         `json:"name"`
	Open      bool           `json:"open"`
	Finished  bool           `json:"finished"`
	Active    bool           `json:"active"`
	Format    *liveFormat    `json:"format"`
	TimeLimit *liveTimeLimit `json:"timeLimit"`
	Cutoff    *liveCutoff    `json:"cutoff"`
	Results   []liveResult   `json:"results"`

	CompetitionEvent *liveCompEventRef `json:"competitionEvent"`
}

type liveCompEventRef struct {
	Event liveEvent `json:"event"`
}

type liveFormat struct {
	ID               string `json:"id"`
	NumberOfAttempts int    `json:"numberOfAttempts"`
}

type liveTimeLimit struct {
	Centiseconds           int      `json:"centiseconds"`
	CumulativeRoundWCIFIDs []string `json:"cumulativeRoundWcifIds"`
}

type liveCutoff struct {
	AttemptResult    int `json:"attemptResult"`
	NumberOfAttempts int `json:"numberOfAttempts"`
}

type liveResult struct {
	ID        string        `json:"id"`
	EnteredAt *string       `json:"enteredAt"`
	Best      *int          `json:"best"`
	Average   *int          `json:"average"`
	Attempts  []liveAttempt `json:"attempts"`
	Person    livePerson    `json:"person"`
}

type liveAttempt struct {
	Result int `json:"result"`
}

type livePerson struct {
	ID           string `json:"id"`
	RegistrantID *int   `json:"registrantId"`
	WCAID        string `json:"wcaId"`
	Name         string `json:"name"`
}

type liveVenue struct {
	Country liveCountry `json:"country"`
}

type liveCountry struct {
	ISO2 string `json:"iso2"`
	Name string `json:"name"`
}
