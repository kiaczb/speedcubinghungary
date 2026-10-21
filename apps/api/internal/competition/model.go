package competition

type Competition struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	StartDate   string   `json:"startDate"`
	EndDate     string   `json:"endDate"`
	City        string   `json:"city"`
	Country     string   `json:"country"`
	CountryISO2 string   `json:"countryIso2"`
	URL         string   `json:"url"`
	EventIDs    []string `json:"eventIds"`

	RegistrationOpen  string `json:"registrationOpen"`
	RegistrationClose string `json:"registrationClose"`
	CompetitorLimit   *int   `json:"competitorLimit"`
	AcceptedCount     *int   `json:"acceptedCount"`
	ResultsPostedAt   string `json:"resultsPostedAt,omitempty"`
}

type wcaCompetition struct {
	ID                string   `json:"id"`
	Name              string   `json:"name"`
	StartDate         string   `json:"start_date"`
	EndDate           string   `json:"end_date"`
	City              string   `json:"city"`
	CountryISO2       string   `json:"country_iso2"`
	URL               string   `json:"url"`
	EventIDs          []string `json:"event_ids"`
	RegistrationOpen  string   `json:"registration_open"`
	RegistrationClose string   `json:"registration_close"`
	CompetitorLimit   *int     `json:"competitor_limit"`
	ResultsPostedAt   string   `json:"results_posted_at"`
}

type wcaRegistration struct {
	ID int `json:"id"`
}

type PsychSheetEntry struct {
	Name     string `json:"name"`
	WCAID    string `json:"wcaId"`
	Position int    `json:"position"`
}

type wcaPsychSheet struct {
	SortedRankings []wcaPsychSheetEntry `json:"sorted_rankings"`
}

type wcaPsychSheetEntry struct {
	Name  string `json:"name"`
	WCAID string `json:"wca_id"`
	Pos   int    `json:"pos"`
}

type Competitor struct {
	Name  string `json:"name"`
	WCAID string `json:"wcaId"`
}

type Registrations struct {
	EventIDs    []string     `json:"eventIds"`
	Registrants []Registrant `json:"registrants"`
}

type Registrant struct {
	Name          string         `json:"name"`
	WCAID         string         `json:"wcaId"`
	CountryISO2   string         `json:"countryIso2"`
	Gender        string         `json:"gender"`
	Roles         []string       `json:"roles"`
	EventIDs      []string       `json:"eventIds"`
	PersonalBests []PersonalBest `json:"personalBests"`
}

type PersonalBest struct {
	EventID            string `json:"eventId"`
	Type               string `json:"type"`
	Best               int    `json:"best"`
	WorldRanking       int    `json:"worldRanking"`
	ContinentalRanking int    `json:"continentalRanking"`
	NationalRanking    int    `json:"nationalRanking"`
}

type wcaWCIF struct {
	Events  []wcaWCIFEvent  `json:"events"`
	Persons []wcaWCIFPerson `json:"persons"`
}

type wcaWCIFEvent struct {
	ID string `json:"id"`
}

type wcaWCIFPerson struct {
	Name          string               `json:"name"`
	WCAID         string               `json:"wcaId"`
	CountryISO2   string               `json:"countryIso2"`
	Gender        string               `json:"gender"`
	Roles         []string             `json:"roles"`
	Registration  *wcaWCIFRegistration `json:"registration"`
	PersonalBests []PersonalBest       `json:"personalBests"`
}

type wcaWCIFRegistration struct {
	Status      string   `json:"status"`
	IsCompeting bool     `json:"isCompeting"`
	EventIDs    []string `json:"eventIds"`
}

type PersonSummary struct {
	WCAID            string       `json:"wcaId"`
	CompetitionCount int          `json:"competitionCount"`
	Medals           Medals       `json:"medals"`
	Records          RecordCounts `json:"records"`
	TotalSolves      int          `json:"totalSolves"`
}

type Medals struct {
	Gold   int `json:"gold"`
	Silver int `json:"silver"`
	Bronze int `json:"bronze"`
	Total  int `json:"total"`
}

type RecordCounts struct {
	National    int `json:"national"`
	Continental int `json:"continental"`
	World       int `json:"world"`
	Total       int `json:"total"`
}

type wcaPersonSummary struct {
	Person struct {
		WCAID string `json:"wca_id"`
	} `json:"person"`
	CompetitionCount int          `json:"competition_count"`
	Medals           Medals       `json:"medals"`
	Records          RecordCounts `json:"records"`
	TotalSolves      int          `json:"total_solves"`
}

type PersonCompetition struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	StartDate   string `json:"startDate"`
	EndDate     string `json:"endDate"`
	CountryISO2 string `json:"countryIso2"`
}

type Result struct {
	WCAID         string `json:"wcaId"`
	Name          string `json:"name"`
	CountryISO2   string `json:"countryIso2"`
	EventID       string `json:"eventId"`
	RoundTypeID   string `json:"roundTypeId"`
	Pos           int    `json:"pos"`
	Best          int    `json:"best"`
	Average       int    `json:"average"`
	Attempts      []int  `json:"attempts"`
	SingleRecord  string `json:"singleRecord"`
	AverageRecord string `json:"averageRecord"`
}

type wcaResult struct {
	WCAID                 string  `json:"wca_id"`
	Name                  string  `json:"name"`
	CountryISO2           string  `json:"country_iso2"`
	EventID               string  `json:"event_id"`
	RoundTypeID           string  `json:"round_type_id"`
	Pos                   int     `json:"pos"`
	Best                  int     `json:"best"`
	Average               int     `json:"average"`
	Attempts              []int   `json:"attempts"`
	RegionalSingleRecord  *string `json:"regional_single_record"`
	RegionalAverageRecord *string `json:"regional_average_record"`
}
