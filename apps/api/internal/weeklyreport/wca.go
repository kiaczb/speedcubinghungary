package weeklyreport

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const wcaAPIBaseURL = "https://www.worldcubeassociation.org/api/v0"

const (
	maxConcurrentUpstream = 6
	maxAttempts           = 5
	maxUpstreamBody       = 64 << 20
	maxCompetitionPages   = 50
	personsBatchSize      = 100
)

var ErrNotFound = errors.New("not found on the WCA API")

type WCA interface {
	ListCompetitions(ctx context.Context, start, end time.Time) ([]wcaCompetition, error)
	GetWCIF(ctx context.Context, competitionID string) (wcaWCIF, error)
	GetResults(ctx context.Context, competitionID string) ([]wcaResult, error)
	GetUpcomingCompetitions(ctx context.Context, wcaID string) ([]wcaUpcomingCompetition, error)
	GetPersonStats(ctx context.Context, wcaIDs []string) (map[string]PersonStats, error)
	CountDelegatedCompetitions(ctx context.Context, wcaID string, until time.Time) (int, error)
	GetRecords(ctx context.Context) (wcaRecords, error)
}

type WCAClient struct {
	client  *http.Client
	baseURL string
	slots   chan struct{}
	backoff func(attempt int) time.Duration
}

func NewWCAClient() *WCAClient {
	return newWCAClient(wcaAPIBaseURL)
}

func newWCAClient(baseURL string) *WCAClient {
	return &WCAClient{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: baseURL,
		slots:   make(chan struct{}, maxConcurrentUpstream),
		backoff: func(attempt int) time.Duration {
			return time.Duration(1<<(2*attempt)) * time.Second
		},
	}
}

func (c *WCAClient) ListCompetitions(ctx context.Context, start, end time.Time) ([]wcaCompetition, error) {
	all := make([]wcaCompetition, 0)
	for page := 1; page <= maxCompetitionPages; page++ {
		q := url.Values{}
		q.Set("sort", "start_date")
		q.Set("start", start.Format(time.DateOnly))
		q.Set("end", end.Format(time.DateOnly))
		q.Set("page", fmt.Sprintf("%d", page))

		var items []wcaCompetition
		if err := c.getJSON(ctx, "/competitions?"+q.Encode(), &items); err != nil {
			return nil, err
		}
		if len(items) == 0 {
			return all, nil
		}
		all = append(all, items...)
	}
	return nil, fmt.Errorf("competition list has more than %d pages", maxCompetitionPages)
}

func (c *WCAClient) GetWCIF(ctx context.Context, competitionID string) (wcaWCIF, error) {
	var wcif wcaWCIF
	err := c.getJSON(ctx, fmt.Sprintf("/competitions/%s/wcif/public", url.PathEscape(competitionID)), &wcif)
	return wcif, err
}

func (c *WCAClient) GetResults(ctx context.Context, competitionID string) ([]wcaResult, error) {
	var results []wcaResult
	err := c.getJSON(ctx, fmt.Sprintf("/competitions/%s/results", url.PathEscape(competitionID)), &results)
	return results, err
}

func (c *WCAClient) GetUpcomingCompetitions(ctx context.Context, wcaID string) ([]wcaUpcomingCompetition, error) {
	var user struct {
		UpcomingCompetitions []wcaUpcomingCompetition `json:"upcoming_competitions"`
	}
	err := c.getJSON(ctx, fmt.Sprintf("/users/%s?upcoming_competitions=true", url.PathEscape(wcaID)), &user)
	return user.UpcomingCompetitions, err
}

func (c *WCAClient) GetPersonStats(ctx context.Context, wcaIDs []string) (map[string]PersonStats, error) {
	stats := make(map[string]PersonStats, len(wcaIDs))
	for start := 0; start < len(wcaIDs); start += personsBatchSize {
		chunk := wcaIDs[start:min(start+personsBatchSize, len(wcaIDs))]
		q := url.Values{}
		q.Set("wca_ids", strings.Join(chunk, ","))
		q.Set("per_page", fmt.Sprintf("%d", len(chunk)))

		var persons []struct {
			Person struct {
				WCAID string `json:"wca_id"`
			} `json:"person"`
			CompetitionCount int `json:"competition_count"`
			TotalSolves      int `json:"total_solves"`
		}
		if err := c.getJSON(ctx, "/persons?"+q.Encode(), &persons); err != nil {
			return nil, err
		}
		for _, p := range persons {
			stats[p.Person.WCAID] = PersonStats{CompetitionCount: p.CompetitionCount, TotalSolves: p.TotalSolves}
		}
	}
	return stats, nil
}

func (c *WCAClient) CountDelegatedCompetitions(ctx context.Context, wcaID string, until time.Time) (int, error) {
	last := until.Format(time.DateOnly)
	count := 0
	for page := 1; page <= maxCompetitionPages; page++ {
		q := url.Values{}
		q.Set("delegate", wcaID)
		q.Set("per_page", "100")
		q.Set("page", fmt.Sprintf("%d", page))

		var items []wcaCompetition
		if err := c.getJSON(ctx, "/competitions?"+q.Encode(), &items); err != nil {
			return 0, err
		}
		if len(items) == 0 {
			return count, nil
		}
		for _, comp := range items {
			if comp.CancelledAt == nil && comp.StartDate != "" && comp.StartDate <= last {
				count++
			}
		}
	}
	return 0, fmt.Errorf("delegated competition list of %s has more than %d pages", wcaID, maxCompetitionPages)
}

func (c *WCAClient) GetRecords(ctx context.Context) (wcaRecords, error) {
	var records wcaRecords
	err := c.getJSON(ctx, "/records", &records)
	return records, err
}

type retryableError struct{ err error }

func (e retryableError) Error() string { return e.err.Error() }
func (e retryableError) Unwrap() error { return e.err }

func (c *WCAClient) getJSON(ctx context.Context, path string, out any) error {
	var err error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			select {
			case <-time.After(c.backoff(attempt - 1)):
			case <-ctx.Done():
				return ctx.Err()
			}
		}

		err = c.getOnce(ctx, c.baseURL+path, out)
		var retryable retryableError
		if !errors.As(err, &retryable) {
			return err
		}
		log.Printf("heti riport: WCA %s: %v (%d/%d)", path, err, attempt+1, maxAttempts)
	}
	return fmt.Errorf("%s: giving up after %d attempts: %w", path, maxAttempts, err)
}

func (c *WCAClient) getOnce(ctx context.Context, reqURL string, out any) error {
	select {
	case c.slots <- struct{}{}:
		defer func() { <-c.slots }()
	case <-ctx.Done():
		return ctx.Err()
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return retryableError{err}
	}
	defer resp.Body.Close()

	switch {
	case resp.StatusCode == http.StatusOK:
	case resp.StatusCode == http.StatusNotFound:
		return ErrNotFound
	case resp.StatusCode == http.StatusTooManyRequests, resp.StatusCode >= 500:
		return retryableError{fmt.Errorf("WCA API returned status %d", resp.StatusCode)}
	default:
		return fmt.Errorf("WCA API returned status %d", resp.StatusCode)
	}

	return json.NewDecoder(io.LimitReader(resp.Body, maxUpstreamBody)).Decode(out)
}
