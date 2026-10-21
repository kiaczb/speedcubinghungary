package competition

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const wcaAPIBaseURL = "https://www.worldcubeassociation.org/api/v0/competitions"

const maxPages = 20

const (
	maxConcurrentUpstream = 6

	upstreamQueueTimeout = 15 * time.Second

	maxUpstreamBody = 16 << 20
)

var errUpstreamBusy = errors.New("too many concurrent WCA API requests")

var ErrNotFound = errors.New("not found on the WCA API")

var ErrRateLimited = errors.New("rate limited by the WCA API")

type Repository interface {
	GetCompetition(competitionID string) (wcaCompetition, error)
	GetUpcomingByCountry(countryISO2 string) ([]wcaCompetition, error)
	GetPsychSheet(competitionID, eventID string) (wcaPsychSheet, error)
	GetWCIF(competitionID string) (wcaWCIF, error)
	GetRegistrationCount(competitionID string) (int, error)
	GetPersons(wcaIDs []string) ([]wcaPersonSummary, error)
	GetPersonCompetitions(wcaID string) ([]wcaCompetition, error)
	GetResults(competitionID string) ([]wcaResult, error)
	SearchCompetitions(params url.Values) ([]wcaCompetition, error)
}

type WCARepository struct {
	client  *http.Client
	baseURL string
	slots   chan struct{}
}

func NewWCARepository() *WCARepository {
	return newWCARepository(wcaAPIBaseURL)
}

func newWCARepository(baseURL string) *WCARepository {
	return &WCARepository{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: baseURL,
		slots:   make(chan struct{}, maxConcurrentUpstream),
	}
}

func (r *WCARepository) GetUpcomingByCountry(countryISO2 string) ([]wcaCompetition, error) {
	today := time.Now().Format("2006-01-02")
	all := make([]wcaCompetition, 0)

	for page := 1; page <= maxPages; page++ {
		q := url.Values{}
		q.Set("country_iso2", countryISO2)
		q.Set("start", today)
		q.Set("sort", "start_date")
		q.Set("page", fmt.Sprintf("%d", page))

		var pageItems []wcaCompetition
		if err := r.getJSON(fmt.Sprintf("%s?%s", r.baseURL, q.Encode()), &pageItems); err != nil {
			return nil, err
		}

		if len(pageItems) == 0 {
			break
		}
		all = append(all, pageItems...)
	}

	return all, nil
}

func (r *WCARepository) GetCompetition(competitionID string) (wcaCompetition, error) {
	var competition wcaCompetition
	err := r.getJSON(fmt.Sprintf("%s/%s", r.baseURL, url.PathEscape(competitionID)), &competition)
	return competition, err
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

func (r *WCARepository) GetRegistrationCount(competitionID string) (int, error) {
	var registrations []wcaRegistration
	if err := r.getJSON(fmt.Sprintf("%s/%s/registrations", r.baseURL, url.PathEscape(competitionID)), &registrations); err != nil {
		return 0, err
	}
	return len(registrations), nil
}

func (r *WCARepository) GetPersons(wcaIDs []string) ([]wcaPersonSummary, error) {
	q := url.Values{}
	q.Set("wca_ids", strings.Join(wcaIDs, ","))
	q.Set("per_page", fmt.Sprintf("%d", len(wcaIDs)))

	var persons []wcaPersonSummary
	err := r.getJSON(fmt.Sprintf("%s/persons?%s", r.apiRoot(), q.Encode()), &persons)
	return persons, err
}

func (r *WCARepository) GetPersonCompetitions(wcaID string) ([]wcaCompetition, error) {
	var competitions []wcaCompetition
	err := r.getJSON(fmt.Sprintf("%s/persons/%s/competitions", r.apiRoot(), url.PathEscape(wcaID)), &competitions)
	return competitions, err
}

func (r *WCARepository) GetResults(competitionID string) ([]wcaResult, error) {
	var results []wcaResult
	err := r.getJSON(fmt.Sprintf("%s/%s/results", r.baseURL, url.PathEscape(competitionID)), &results)
	return results, err
}

func (r *WCARepository) SearchCompetitions(params url.Values) ([]wcaCompetition, error) {
	var competitions []wcaCompetition
	err := r.getJSON(fmt.Sprintf("%s?%s", r.baseURL, params.Encode()), &competitions)
	return competitions, err
}

func (r *WCARepository) apiRoot() string {
	return strings.TrimSuffix(r.baseURL, "/competitions")
}

func (r *WCARepository) getJSON(reqURL string, out any) error {
	select {
	case r.slots <- struct{}{}:
		defer func() { <-r.slots }()
	case <-time.After(upstreamQueueTimeout):
		return errUpstreamBusy
	}

	resp, err := r.client.Get(reqURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case resp.StatusCode == http.StatusTooManyRequests:
		return ErrRateLimited
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("WCA API returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(io.LimitReader(resp.Body, maxUpstreamBody)).Decode(out)
}
