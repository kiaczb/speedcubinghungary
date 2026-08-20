package special

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

const DefaultWCALiveURL = "https://live.worldcubeassociation.org"

var competitionIDFromPath = regexp.MustCompile(`/competitions/(\d+)`)

type LiveClient struct {
	client  *http.Client
	baseURL string
}

func NewLiveClient(baseURL string) *LiveClient {
	if baseURL == "" {
		baseURL = DefaultWCALiveURL
	}
	return &LiveClient{
		client: &http.Client{
			Timeout: 15 * time.Second,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		baseURL: strings.TrimSuffix(baseURL, "/"),
	}
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

func (c *LiveClient) query(ctx context.Context, doc string, variables map[string]any, out any) error {
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

func (c *LiveClient) ResolveCompetitionID(ctx context.Context, competitionWCAID string) (int, error) {
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

const competitionQuery = `query Competition($id: ID!) {
  competition(id: $id) {
    id
    wcaId
    name
    startDate
    endDate
    competitionEvents {
      event { id }
      rounds {
        id
        number
        name
        open
        finished
        active
        format { id numberOfAttempts }
        timeLimit { centiseconds cumulativeRoundWcifIds }
        cutoff { attemptResult numberOfAttempts }
      }
    }
  }
}`

func (c *LiveClient) GetCompetition(ctx context.Context, liveCompetitionID int) (*liveCompetition, error) {
	var data struct {
		Competition *liveCompetition `json:"competition"`
	}

	err := c.query(ctx, competitionQuery, map[string]any{
		"id": strconv.Itoa(liveCompetitionID),
	}, &data)
	if err != nil {
		return nil, err
	}
	if data.Competition == nil {
		return nil, fmt.Errorf("competition %d not found on WCA Live", liveCompetitionID)
	}

	return data.Competition, nil
}

const RoundQuery = `query Round($id: ID!) {
  round(id: $id) {
    id
    number
    name
    open
    finished
    active
    format { id numberOfAttempts }
    timeLimit { centiseconds cumulativeRoundWcifIds }
    cutoff { attemptResult numberOfAttempts }
    competitionEvent { event { id } }
    results {
      id
      enteredAt
      best
      average
      attempts { result }
      person { id registrantId wcaId name }
    }
  }
}`

const RoundSubscription = `subscription RoundUpdated($id: ID!) {
  roundUpdated(id: $id) {
    id
    results {
      id
      enteredAt
      best
      average
      attempts { result }
      person { id registrantId wcaId name }
    }
  }
}`

func (c *LiveClient) GetRound(ctx context.Context, liveRoundID int) (*liveRound, error) {
	var data struct {
		Round *liveRound `json:"round"`
	}

	err := c.query(ctx, RoundQuery, map[string]any{
		"id": strconv.Itoa(liveRoundID),
	}, &data)
	if err != nil {
		return nil, err
	}
	if data.Round == nil {
		return nil, fmt.Errorf("round %d not found on WCA Live", liveRoundID)
	}

	return data.Round, nil
}

const competitionsQuery = `query Competitions($filter: String, $from: Date, $limit: Int) {
  competitions(filter: $filter, from: $from, limit: $limit) {
    id
    wcaId
    name
    startDate
    endDate
    venues { country { iso2 name } }
  }
}`

func (c *LiveClient) ListCompetitions(ctx context.Context, filter string, from time.Time, limit int) ([]Competition, error) {
	variables := map[string]any{"limit": limit}
	if filter != "" {
		variables["filter"] = filter
	}
	if !from.IsZero() {
		variables["from"] = from.Format(dateLayout)
	}

	var data struct {
		Competitions []liveCompetition `json:"competitions"`
	}
	if err := c.query(ctx, competitionsQuery, variables, &data); err != nil {
		return nil, err
	}

	competitions := make([]Competition, 0, len(data.Competitions))
	for _, live := range data.Competitions {
		competition, err := toCompetition(live)
		if err != nil {
			continue
		}
		competitions = append(competitions, competition)
	}

	return competitions, nil
}

const competitionInfoQuery = `query CompetitionInfo($id: ID!) {
  competition(id: $id) {
    id
    wcaId
    name
    startDate
    endDate
    venues { country { iso2 name } }
  }
}`

func (c *LiveClient) GetCompetitionInfo(ctx context.Context, liveCompetitionID int) (*Competition, error) {
	var data struct {
		Competition *liveCompetition `json:"competition"`
	}

	err := c.query(ctx, competitionInfoQuery, map[string]any{
		"id": strconv.Itoa(liveCompetitionID),
	}, &data)
	if err != nil {
		return nil, err
	}
	if data.Competition == nil {
		return nil, fmt.Errorf("competition %d not found on WCA Live", liveCompetitionID)
	}

	competition, err := toCompetition(*data.Competition)
	if err != nil {
		return nil, err
	}

	return &competition, nil
}

const dateLayout = "2006-01-02"

func toCompetition(live liveCompetition) (Competition, error) {
	liveID, err := strconv.Atoi(live.ID)
	if err != nil {
		return Competition{}, fmt.Errorf("unexpected WCA Live competition id %q", live.ID)
	}

	competition := Competition{
		CompetitionWCAID:  live.WCAID,
		LiveCompetitionID: liveID,
		Name:              live.Name,
		StartDate:         live.StartDate,
		EndDate:           live.EndDate,
	}
	if len(live.Venues) > 0 {
		competition.CountryISO2 = live.Venues[0].Country.ISO2
	}

	return competition, nil
}
