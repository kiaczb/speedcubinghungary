package competition

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const wcaAPIBaseURL = "https://www.worldcubeassociation.org/api/v0/competitions"

// maxPages caps how many pages we follow when paginating the WCA API, guarding
// against an unexpected upstream response that never returns an empty page.
const maxPages = 20

const (
	// maxConcurrentUpstream bounds how many requests are in flight to the WCA
	// API at once, across every caller. The WCA API starts answering 429 at
	// roughly 15 parallel requests, and a prediction page asks for one psych
	// sheet per podium event — a dozen of them on a cold cache used to go out
	// at once. This is also what protects the WCA from us when someone
	// deliberately requests ids we have never cached: it holds regardless of
	// who is asking or from which IP, which a per-IP limit cannot (see
	// middleware.RateLimit for why we don't have one here).
	maxConcurrentUpstream = 6

	// upstreamQueueTimeout is how long a request waits for a free slot before
	// giving up. Without it a flood of cache misses would queue goroutines
	// without bound; with it they fail as an upstream error (502) instead.
	upstreamQueueTimeout = 15 * time.Second

	// maxUpstreamBody caps how much of a response we read. A public WCIF of a
	// large competition is a few MB; anything far beyond that is not a
	// response we want to hold in memory.
	maxUpstreamBody = 16 << 20
)

var errUpstreamBusy = errors.New("too many concurrent WCA API requests")

// ErrNotFound is returned when the WCA API answers 404 — the id is well-formed
// but the WCA does not know it (or has nothing published under it yet).
var ErrNotFound = errors.New("not found on the WCA API")

// Repository defines read access to competition data.
type Repository interface {
	GetCompetition(competitionID string) (wcaCompetition, error)
	GetUpcomingByCountry(countryISO2 string) ([]wcaCompetition, error)
	GetPsychSheet(competitionID, eventID string) (wcaPsychSheet, error)
	GetWCIF(competitionID string) (wcaWCIF, error)
}

// WCARepository fetches competition data from the public WCA API.
type WCARepository struct {
	client  *http.Client
	baseURL string
	// slots is a semaphore shared by every request this repository makes.
	slots chan struct{}
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

// GetCompetition fetches one competition by id, whatever its date — unlike
// GetUpcomingByCountry, which only sees competitions that have not ended yet.
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

// getJSON is the only place this repository talks to the WCA API. It holds an
// upstream slot for the whole request, body included — releasing it before the
// body is read would let the real concurrency exceed the bound.
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
	case resp.StatusCode != http.StatusOK:
		return fmt.Errorf("WCA API returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(io.LimitReader(resp.Body, maxUpstreamBody)).Decode(out)
}
