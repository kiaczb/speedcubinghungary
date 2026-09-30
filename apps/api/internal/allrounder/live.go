package allrounder

import (
	"context"
	"sort"
	"strconv"

	"speedcubinghungary/internal/wcalive"
)

type LiveClient struct {
	*wcalive.Client
}

func NewLiveClient(baseURL string) *LiveClient {
	return &LiveClient{Client: wcalive.NewClient(baseURL)}
}

const competitionRoundsQuery = `query Competition($id: ID!) {
  competition(id: $id) {
    id
    competitionEvents {
      event { id }
      rounds { id number open }
    }
  }
}`

const roundRankingQuery = `query Round($id: ID!) {
  round(id: $id) {
    id
    number
    open
    competitionEvent { event { id } }
    results {
      ranking
      best
      attempts { result }
      person { wcaId name }
    }
  }
}`

const RoundRankingSubscription = `subscription RoundUpdated($id: ID!) {
  roundUpdated(id: $id) {
    id
    results {
      ranking
      best
      attempts { result }
      person { wcaId name }
    }
  }
}`

type liveCompetitionRounds struct {
	ID                string `json:"id"`
	CompetitionEvents []struct {
		Event struct {
			ID string `json:"id"`
		} `json:"event"`
		Rounds []struct {
			ID     string `json:"id"`
			Number int    `json:"number"`
			Open   bool   `json:"open"`
		} `json:"rounds"`
	} `json:"competitionEvents"`
}

type liveRound struct {
	ID               string `json:"id"`
	Number           int    `json:"number"`
	Open             bool   `json:"open"`
	CompetitionEvent *struct {
		Event struct {
			ID string `json:"id"`
		} `json:"event"`
	} `json:"competitionEvent"`
	Results []liveResult `json:"results"`
}

type liveResult struct {
	Ranking  *int `json:"ranking"`
	Best     *int `json:"best"`
	Attempts []struct {
		Result int `json:"result"`
	} `json:"attempts"`
	Person struct {
		WCAID string `json:"wcaId"`
		Name  string `json:"name"`
	} `json:"person"`
}

type roundIdentity struct {
	LiveRoundID int
	EventID     string
	RoundNumber int
}

func (c *LiveClient) OpenRounds(ctx context.Context, liveCompetitionID int) ([]roundIdentity, error) {
	var data struct {
		Competition *liveCompetitionRounds `json:"competition"`
	}

	err := c.Query(ctx, competitionRoundsQuery, map[string]any{
		"id": strconv.Itoa(liveCompetitionID),
	}, &data)
	if err != nil {
		return nil, err
	}
	if data.Competition == nil {
		return nil, ErrCompetitionUnknown
	}

	identities := make([]roundIdentity, 0)
	for _, competitionEvent := range data.Competition.CompetitionEvents {
		for _, round := range competitionEvent.Rounds {
			if !round.Open {
				continue
			}
			roundID, err := strconv.Atoi(round.ID)
			if err != nil {
				continue
			}

			identities = append(identities, roundIdentity{
				LiveRoundID: roundID,
				EventID:     competitionEvent.Event.ID,
				RoundNumber: round.Number,
			})
		}
	}

	return identities, nil
}

func (c *LiveClient) GetRound(ctx context.Context, liveRoundID int) (*liveRound, error) {
	var data struct {
		Round *liveRound `json:"round"`
	}

	err := c.Query(ctx, roundRankingQuery, map[string]any{
		"id": strconv.Itoa(liveRoundID),
	}, &data)
	if err != nil {
		return nil, err
	}
	if data.Round == nil {
		return nil, ErrCompetitionUnknown
	}

	return data.Round, nil
}

func roundResultsFromLive(identity roundIdentity, round liveRound) RoundResults {
	placements := make([]Placement, 0, len(round.Results))

	for _, result := range round.Results {
		if result.Ranking == nil || len(result.Attempts) == 0 {
			continue
		}
		if result.Person.WCAID == "" {
			continue
		}

		best := 0
		if result.Best != nil {
			best = *result.Best
		}
		if best == 0 {
			continue
		}

		placements = append(placements, Placement{
			WCAID: result.Person.WCAID,
			Name:  result.Person.Name,
			Pos:   *result.Ranking,
			Best:  best,
		})
	}

	sort.SliceStable(placements, func(a, b int) bool {
		if placements[a].Pos != placements[b].Pos {
			return placements[a].Pos < placements[b].Pos
		}
		return placements[a].WCAID < placements[b].WCAID
	})

	return RoundResults{
		EventID:    identity.EventID,
		Rank:       identity.RoundNumber,
		Placements: placements,
	}
}
