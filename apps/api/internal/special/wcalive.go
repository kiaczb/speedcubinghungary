package special

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"speedcubinghungary/internal/wcalive"
)

const DefaultWCALiveURL = wcalive.DefaultURL

type (
	SocketClient      = wcalive.SocketClient
	Broadcaster       = wcalive.Broadcaster
	SubscriptionEvent = wcalive.SubscriptionEvent
)

func NewSocketClient(baseURL string) *SocketClient { return wcalive.NewSocketClient(baseURL) }
func NewBroadcaster() *Broadcaster                 { return wcalive.NewBroadcaster() }

type LiveClient struct {
	*wcalive.Client
}

func NewLiveClient(baseURL string) *LiveClient {
	return &LiveClient{Client: wcalive.NewClient(baseURL)}
}

func (c *LiveClient) query(ctx context.Context, doc string, variables map[string]any, out any) error {
	return c.Client.Query(ctx, doc, variables, out)
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
