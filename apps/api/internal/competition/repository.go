package competition

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

const wcaAPIBaseURL = "https://www.worldcubeassociation.org/api/v0/competitions"

// maxPages caps how many pages we follow when paginating the WCA API, guarding
// against an unexpected upstream response that never returns an empty page.
const maxPages = 20

// Repository defines read access to competition data.
type Repository interface {
	GetUpcomingByCountry(countryISO2 string) ([]wcaCompetition, error)
	GetPsychSheet(competitionID, eventID string) (wcaPsychSheet, error)
	GetWCIF(competitionID string) (wcaWCIF, error)
}

// WCARepository fetches competition data from the public WCA API.
type WCARepository struct {
	client  *http.Client
	baseURL string
}

func NewWCARepository() *WCARepository {
	return &WCARepository{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: wcaAPIBaseURL,
	}
}

// GetUpcomingByCountry fetches every upcoming competition (start date >= today)
// for the given country, sorted by start date. It walks the WCA API pagination
// until a page comes back empty, accumulating all results.
func (r *WCARepository) GetUpcomingByCountry(countryISO2 string) ([]wcaCompetition, error) {
	today := time.Now().Format("2006-01-02")
	all := make([]wcaCompetition, 0)

	for page := 1; page <= maxPages; page++ {
		q := url.Values{}
		q.Set("country_iso2", countryISO2)
		q.Set("start", today)
		q.Set("sort", "start_date")
		q.Set("page", fmt.Sprintf("%d", page))

		reqURL := fmt.Sprintf("%s?%s", r.baseURL, q.Encode())

		resp, err := r.client.Get(reqURL)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			return nil, fmt.Errorf("WCA API returned status %d", resp.StatusCode)
		}

		var pageItems []wcaCompetition
		if err := json.NewDecoder(resp.Body).Decode(&pageItems); err != nil {
			resp.Body.Close()
			return nil, err
		}
		resp.Body.Close()

		if len(pageItems) == 0 {
			break
		}
		all = append(all, pageItems...)
	}

	return all, nil
}

func (r *WCARepository) GetPsychSheet(competitionID, eventID string) (wcaPsychSheet, error) {
	var sheet wcaPsychSheet
	err := r.getJSON(fmt.Sprintf("%s/%s/psych-sheet/%s", r.baseURL, url.PathEscape(competitionID), url.PathEscape(eventID)), &sheet)
	return sheet, err
}

func (r *WCARepository) GetWCIF(competitionID string) (wcaWCIF, error) {
	var wcif wcaWCIF
	err := r.getJSON(fmt.Sprintf("%s/%s/wcif/public", r.baseURL, url.PathEscape(competitionID)), &wcif)
	return wcif, err
}

func (r *WCARepository) getJSON(reqURL string, out any) error {
	resp, err := r.client.Get(reqURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WCA API returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(out)
}
