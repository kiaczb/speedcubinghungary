package competition

// Competition is the simplified competition representation returned to the
// frontend. JSON keys are camelCase to match the web client conventions.
type Competition struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	StartDate string   `json:"startDate"`
	EndDate   string   `json:"endDate"`
	City      string   `json:"city"`
	Country   string   `json:"country"`
	URL       string   `json:"url"`
	EventIDs  []string `json:"eventIds"`
}

// wcaCompetition mirrors the relevant subset of fields returned by the WCA API.
// It is an internal type used only for decoding the upstream response.
type wcaCompetition struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	StartDate   string   `json:"start_date"`
	EndDate     string   `json:"end_date"`
	City        string   `json:"city"`
	CountryISO2 string   `json:"country_iso2"`
	URL         string   `json:"url"`
	EventIDs    []string `json:"event_ids"`
}

// PsychSheetEntry is one competitor's position on a single event's psych
// sheet (the WCA's PR-ranked start list). JSON keys are camelCase to match
// the web client conventions.
type PsychSheetEntry struct {
	Name     string `json:"name"`
	WCAID    string `json:"wcaId"`
	Position int    `json:"position"`
}

// wcaPsychSheet mirrors the relevant subset of fields returned by the WCA
// API's psych sheet endpoint. It is an internal type used only for decoding
// the upstream response.
type wcaPsychSheet struct {
	SortedRankings []wcaPsychSheetEntry `json:"sorted_rankings"`
}

type wcaPsychSheetEntry struct {
	Name  string `json:"name"`
	WCAID string `json:"wca_id"`
	Pos   int    `json:"pos"`
}

// Competitor is one accepted, competing registrant of a competition. JSON
// keys are camelCase to match the web client conventions. WCAID is empty
// for newcomers.
type Competitor struct {
	Name  string `json:"name"`
	WCAID string `json:"wcaId"`
}

// wcaWCIF mirrors the relevant subset of the WCA public WCIF document
// (/competitions/{id}/wcif/public). Internal decoding type only.
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
