package allrounder

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"
)

const wcaAPIBaseURL = "https://www.worldcubeassociation.org/api/v0"

var ErrNoResults = errors.New("no published results for competition")

type WCAClient struct {
	client  *http.Client
	baseURL string
}

func NewWCAClient() *WCAClient {
	return &WCAClient{
		client:  &http.Client{Timeout: 30 * time.Second},
		baseURL: wcaAPIBaseURL,
	}
}

var ErrCompetitionUnknown = errors.New("competition is not known to the WCA")

func (c *WCAClient) FetchCompetitionInfo(ctx context.Context, competitionWCAID string) (*CompetitionInfo, error) {
	url := fmt.Sprintf("%s/competitions/%s", c.baseURL, competitionWCAID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrCompetitionUnknown
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WCA API returned status %d for %s", resp.StatusCode, competitionWCAID)
	}

	var raw struct {
		StartDate string `json:"start_date"`
		EndDate   string `json:"end_date"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	startDate, err := time.Parse(dateLayout, raw.StartDate)
	if err != nil {
		return nil, fmt.Errorf("unexpected start_date %q for %s", raw.StartDate, competitionWCAID)
	}
	endDate, err := time.Parse(dateLayout, raw.EndDate)
	if err != nil {
		return nil, fmt.Errorf("unexpected end_date %q for %s", raw.EndDate, competitionWCAID)
	}

	return &CompetitionInfo{StartDate: startDate, EndDate: endDate}, nil
}

const dateLayout = "2006-01-02"

func (c *WCAClient) FetchResults(ctx context.Context, competitionWCAID string) ([]Result, error) {
	url := fmt.Sprintf("%s/competitions/%s/results", c.baseURL, competitionWCAID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNoResults
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WCA API returned status %d for %s", resp.StatusCode, competitionWCAID)
	}

	var raw []wcaResult
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}

	results := make([]Result, 0, len(raw))
	for index, item := range raw {
		results = append(results, Result{
			Ordinal:     index,
			WCAID:       item.WCAID,
			Name:        item.Name,
			EventID:     item.EventID,
			RoundTypeID: item.RoundTypeID,
			Pos:         item.Pos,
			Best:        item.Best,
		})
	}

	return results, nil
}
