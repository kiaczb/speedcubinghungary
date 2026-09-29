package wcalive

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const DefaultURL = "https://live.worldcubeassociation.org"

var competitionIDFromPath = regexp.MustCompile(`/competitions/(\d+)`)

type Client struct {
	client  *http.Client
	baseURL string
}

func NewClient(baseURL string) *Client {
	return &Client{
		client: &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		baseURL: normalizeBaseURL(baseURL),
	}
}

func normalizeBaseURL(baseURL string) string {
	if baseURL == "" {
		baseURL = DefaultURL
	}
	return strings.TrimSuffix(baseURL, "/")
}

type graphQLRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphQLError struct {
	Message string `json:"message"`
}

type graphQLResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphQLError  `json:"errors"`
}

func (c *Client) Query(ctx context.Context, doc string, variables map[string]any, out any) error {
	body, err := json.Marshal(graphQLRequest{Query: doc, Variables: variables})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/api", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("WCA Live GraphQL returned status %d", resp.StatusCode)
	}

	var parsed graphQLResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return err
	}
	if len(parsed.Errors) > 0 {
		return fmt.Errorf("WCA Live GraphQL error: %s", parsed.Errors[0].Message)
	}

	return json.Unmarshal(parsed.Data, out)
}

func (c *Client) ResolveCompetitionID(ctx context.Context, competitionWCAID string) (int, error) {
	url := fmt.Sprintf("%s/link/competitions/%s", c.baseURL, competitionWCAID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return 0, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound && resp.StatusCode != http.StatusMovedPermanently {
		return 0, fmt.Errorf("competition %q not found on WCA Live (status %d)", competitionWCAID, resp.StatusCode)
	}

	match := competitionIDFromPath.FindStringSubmatch(resp.Header.Get("Location"))
	if match == nil {
		return 0, fmt.Errorf("unexpected WCA Live redirect for %q", competitionWCAID)
	}

	return strconv.Atoi(match[1])
}
