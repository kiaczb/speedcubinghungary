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
