package weeklyreport

import "time"

type Record struct {
	Type   string
	Result int
	Badge  string
}

type EventRecords struct {
	EventID string
	Records []Record
}

type Competitor struct {
	Name   string
	WCAID  string
	Events []EventRecords
}

func (c *Competitor) addRecord(eventID string, record Record) {
	for i := range c.Events {
		if c.Events[i].EventID == eventID {
			c.Events[i].Records = append(c.Events[i].Records, record)
			return
		}
	}
	c.Events = append(c.Events, EventRecords{EventID: eventID, Records: []Record{record}})
}

func (c Competitor) HasRecords() bool {
	return len(c.Events) > 0
}

type Competition struct {
	ID          string
	Name        string
	CountryISO2 string
	From        time.Time
	To          time.Time
	Hungarian   bool
	Competitors []Competitor
}

type CompetitorMilestone struct {
	WCAID                string
	Name                 string
	CompetitionCount     int
	EstimatedCompetition string
	EstimatedDate        string
	Achieved             bool
}

type DelegateMilestone struct {
	WCAID          string
	Name           string
	DelegatedCount int
}

type SolveMilestone struct {
	WCAID      string
	Name       string
	SolveCount int
	Milestone  int
	Remaining  int
	Achieved   bool
}

type ReportData struct {
	Competitions         []Competition
	CompetitorMilestones []CompetitorMilestone
	DelegateMilestones   []DelegateMilestone
	SolveMilestones      []SolveMilestone
}

type Person struct {
	WCAID            string
	Name             string
	CompetitionCount int
	DelegatedCount   int
	SolveCount       int
}

type RecordValues struct {
	Single  *int `json:"single"`
	Average *int `json:"average"`
}

type RecordTable map[string]map[string]RecordValues

type ReportPost struct {
	ID         string
	Title      string
	Content    string
	ReportWeek time.Time
}

type PersonCounts struct {
	WCAID            string
	Name             string
	CompetitionCount int
	SolveCount       int
}

type DelegatedCount struct {
	WCAID string
	Name  string
	Count int
}

type PersonStats struct {
	CompetitionCount int
	TotalSolves      int
}

type wcaCompetition struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	CountryISO2     string    `json:"country_iso2"`
	StartDate       string    `json:"start_date"`
	EndDate         string    `json:"end_date"`
	ResultsPostedAt *string   `json:"results_posted_at"`
	CancelledAt     *string   `json:"cancelled_at"`
	Delegates       []wcaUser `json:"delegates"`
}

func (c wcaCompetition) resultsPosted() bool {
	return c.ResultsPostedAt != nil && *c.ResultsPostedAt != ""
}

type wcaUser struct {
	WCAID       string `json:"wca_id"`
	Name        string `json:"name"`
	CountryISO2 string `json:"country_iso2"`
}

type wcaWCIF struct {
	Persons []wcaWCIFPerson `json:"persons"`
	Events  []wcaWCIFEvent  `json:"events"`
}

type wcaWCIFPerson struct {
	Name         string               `json:"name"`
	WCAID        string               `json:"wcaId"`
	RegistrantID *int                 `json:"registrantId"`
	CountryISO2  string               `json:"countryIso2"`
	Registration *wcaWCIFRegistration `json:"registration"`
}

type wcaWCIFRegistration struct {
	EventIDs []string `json:"eventIds"`
}

type wcaWCIFEvent struct {
	ID     string         `json:"id"`
	Rounds []wcaWCIFRound `json:"rounds"`
}

type wcaWCIFRound struct {
	Results []wcaWCIFResult `json:"results"`
}

type wcaWCIFResult struct {
	PersonID int              `json:"personId"`
	Best     int              `json:"best"`
	Average  int              `json:"average"`
	Attempts []wcaWCIFAttempt `json:"attempts"`
}

type wcaWCIFAttempt struct {
	Result int `json:"result"`
}

type wcaResult struct {
	WCAID    string `json:"wca_id"`
	Name     string `json:"name"`
	EventID  string `json:"event_id"`
	Best     int    `json:"best"`
	Average  int    `json:"average"`
	Attempts []int  `json:"attempts"`
}

type wcaUpcomingCompetition struct {
	Name      string `json:"name"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
}

type wcaRecords struct {
	WorldRecords       map[string]RecordValues            `json:"world_records"`
	ContinentalRecords map[string]map[string]RecordValues `json:"continental_records"`
	NationalRecords    map[string]map[string]RecordValues `json:"national_records"`
}

type eventResults struct {
	EventID string
	Results []personResult
}

type personResult struct {
	RegistrantID int
	Best         int
	Average      int
	Attempts     []int
}
