package stats

import "time"

type Phase string

const (
	PhaseBefore Phase = "before"
	PhaseDuring Phase = "during"
	PhaseAfter  Phase = "after"
)

type Report struct {
	Competition  CompetitionInfo `json:"competition"`
	Phase        Phase           `json:"phase"`
	GeneratedAt  time.Time       `json:"generatedAt"`
	Field        *FieldStats     `json:"field,omitempty"`
	FieldError   string          `json:"fieldError,omitempty"`
	History      HistoryStats    `json:"history"`
	Live         *LiveStats      `json:"live,omitempty"`
	LiveError    string          `json:"liveError,omitempty"`
	Results      *ResultStats    `json:"results,omitempty"`
	ResultsError string          `json:"resultsError,omitempty"`
}

type CompetitionInfo struct {
	ID              string   `json:"id"`
	Name            string   `json:"name"`
	StartDate       string   `json:"startDate"`
	EndDate         string   `json:"endDate"`
	City            string   `json:"city"`
	CountryISO2     string   `json:"countryIso2"`
	URL             string   `json:"url"`
	EventIDs        []string `json:"eventIds"`
	CompetitorLimit *int     `json:"competitorLimit"`
	ResultsPostedAt string   `json:"resultsPostedAt,omitempty"`
}

type Person struct {
	Name        string `json:"name"`
	WCAID       string `json:"wcaId,omitempty"`
	CountryISO2 string `json:"countryIso2,omitempty"`
}

type Count struct {
	Key   string `json:"key"`
	Count int    `json:"count"`
}

type PersonValue struct {
	Person Person `json:"person"`
	Value  int    `json:"value"`
}

type FieldStats struct {
	Competitors      int            `json:"competitors"`
	Limit            *int           `json:"limit"`
	Newcomers        int            `json:"newcomers"`
	Foreigners       int            `json:"foreigners"`
	Countries        []Count        `json:"countries"`
	Genders          []Count        `json:"genders"`
	Events           []Count        `json:"events"`
	AverageEvents    float64        `json:"averageEvents"`
	AllEvents        []Person       `json:"allEvents"`
	TopRanked        []RankedResult `json:"topRanked"`
	RecordHolders    []RankedResult `json:"recordHolders"`
	Veterans         []PersonValue  `json:"veterans"`
	Anniversaries    []Anniversary  `json:"anniversaries"`
	Milestones       []PersonValue  `json:"milestones"`
	MostCompetitions []PersonValue  `json:"mostCompetitions"`
	MostMedals       []MedalHolder  `json:"mostMedals"`
	MostRecords      []RecordHolder `json:"mostRecords"`
	CareerSolves     int            `json:"careerSolves"`
	SummariesError   string         `json:"summariesError,omitempty"`
}

type RankedResult struct {
	Person          Person `json:"person"`
	EventID         string `json:"eventId"`
	Type            string `json:"type"`
	Display         string `json:"display"`
	Record          string `json:"record,omitempty"`
	WorldRanking    int    `json:"worldRanking"`
	NationalRanking int    `json:"nationalRanking"`
}

type Anniversary struct {
	Years     int      `json:"years"`
	StartYear int      `json:"startYear"`
	People    []Person `json:"people"`
}

type MedalHolder struct {
	Person Person `json:"person"`
	Gold   int    `json:"gold"`
	Silver int    `json:"silver"`
	Bronze int    `json:"bronze"`
}

type RecordHolder struct {
	Person      Person `json:"person"`
	World       int    `json:"world"`
	Continental int    `json:"continental"`
	National    int    `json:"national"`
}

type HistoryStatus string

const (
	HistoryUnavailable HistoryStatus = "unavailable"
	HistoryRunning     HistoryStatus = "running"
	HistoryReady       HistoryStatus = "ready"
)

type HistoryStats struct {
	Status         HistoryStatus `json:"status"`
	Done           int           `json:"done"`
	Total          int           `json:"total"`
	Failed         int           `json:"failed"`
	Comebacks      []Comeback    `json:"comebacks"`
	FirstInCountry []Person      `json:"firstInCountry"`
	BusiestYear    []PersonValue `json:"busiestYear"`
}

type Comeback struct {
	Person          Person          `json:"person"`
	LastCompetition CompetitionStub `json:"lastCompetition"`
	GapDays         int             `json:"gapDays"`
}

type CompetitionStub struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Date string `json:"date"`
}

type LiveStats struct {
	RoundsTotal    int           `json:"roundsTotal"`
	RoundsOpen     int           `json:"roundsOpen"`
	RoundsFinished int           `json:"roundsFinished"`
	ResultsEntered int           `json:"resultsEntered"`
	Attempts       int           `json:"attempts"`
	DNFs           int           `json:"dnfs"`
	Records        []RecordEntry `json:"records"`
	PRs            int           `json:"prs"`
	PRLeaders      []PersonValue `json:"prLeaders"`
	EventBests     []EventBest   `json:"eventBests"`
	FetchedAt      time.Time     `json:"fetchedAt"`
}

type RecordEntry struct {
	Person  Person `json:"person"`
	EventID string `json:"eventId"`
	Type    string `json:"type"`
	Record  string `json:"record"`
	Display string `json:"display"`
}

type EventBest struct {
	EventID string     `json:"eventId"`
	Single  *Highlight `json:"single,omitempty"`
	Average *Highlight `json:"average,omitempty"`
}

type Highlight struct {
	Person  Person `json:"person"`
	Display string `json:"display"`
}

type ResultStats struct {
	Winners             []EventWinner   `json:"winners"`
	Records             []RecordEntry   `json:"records"`
	Podiums             []MedalHolder   `json:"podiums"`
	CountryPodiums      []CountryMedals `json:"countryPodiums"`
	Solves              int             `json:"solves"`
	DNFs                int             `json:"dnfs"`
	SolvingCentiseconds int64           `json:"solvingCentiseconds"`
	Sub10Singles        []Person        `json:"sub10Singles"`
}

type EventWinner struct {
	EventID string `json:"eventId"`
	Person  Person `json:"person"`
	Best    string `json:"best"`
	Average string `json:"average,omitempty"`
}

type CountryMedals struct {
	CountryISO2 string `json:"countryIso2"`
	Gold        int    `json:"gold"`
	Silver      int    `json:"silver"`
	Bronze      int    `json:"bronze"`
}
