package competition

type Competition struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	StartDate string   `json:"startDate"`
	EndDate   string   `json:"endDate"`
	City      string   `json:"city"`
	Country   string   `json:"country"`
	URL       string   `json:"url"`
	EventIDs  []string `json:"eventIds"`

	RegistrationOpen  string `json:"registrationOpen"`
	RegistrationClose string `json:"registrationClose"`
	CompetitorLimit   *int   `json:"competitorLimit"`
	AcceptedCount     *int   `json:"acceptedCount"`
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

type wcaWCIF struct {
	Persons []wcaWCIFPerson `json:"persons"`
}

type wcaWCIFPerson struct {
	Name         string               `json:"name"`
	WCAID        string               `json:"wcaId"`
	Registration *wcaWCIFRegistration `json:"registration"`
}

type wcaWCIFRegistration struct {
	Status      string `json:"status"`
	IsCompeting bool   `json:"isCompeting"`
}
