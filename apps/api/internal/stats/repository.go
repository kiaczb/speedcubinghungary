package stats

import (
	"context"
	"errors"
	"strconv"
	"sync"

	"speedcubinghungary/internal/wcalive"
)

var ErrNotOnLive = errors.New("competition is not on WCA Live")

const liveRoundConcurrency = 4

const liveCompetitionQuery = `query Competition($id: ID!) {
  competition(id: $id) {
    id
    competitionEvents {
      event { id }
      rounds { id number open finished }
    }
  }
}`

const liveRoundQuery = `query Round($id: ID!) {
  round(id: $id) {
    id
    results {
      ranking
      best
      average
      singleRecordTag
      averageRecordTag
      attempts { result }
      person { wcaId name country { iso2 } }
    }
  }
}`

type LiveCompetition struct {
	Rounds []LiveRound
}

type LiveRound struct {
	ID       string
	EventID  string
	Number   int
	Open     bool
	Finished bool
	Results  []LiveResult
}

type LiveResult struct {
	Person           Person
	Ranking          *int
	Best             int
	Average          int
	SingleRecordTag  string
	AverageRecordTag string
	Attempts         []int
}

type LiveRepository struct {
	client *wcalive.Client
}

func NewLiveRepository(baseURL string) *LiveRepository {
	return &LiveRepository{client: wcalive.NewClient(baseURL)}
}

func (r *LiveRepository) ResolveCompetitionID(ctx context.Context, competitionWCAID string) (int, error) {
	return r.client.ResolveCompetitionID(ctx, competitionWCAID)
}

func (r *LiveRepository) LoadCompetition(ctx context.Context, liveCompetitionID int) (*LiveCompetition, error) {
	var data struct {
		Competition *struct {
			CompetitionEvents []struct {
				Event struct {
					ID string `json:"id"`
				} `json:"event"`
				Rounds []struct {
					ID       string `json:"id"`
					Number   int    `json:"number"`
					Open     bool   `json:"open"`
					Finished bool   `json:"finished"`
				} `json:"rounds"`
			} `json:"competitionEvents"`
		} `json:"competition"`
	}
	if err := r.client.Query(ctx, liveCompetitionQuery, map[string]any{"id": strconv.Itoa(liveCompetitionID)}, &data); err != nil {
		return nil, err
	}
	if data.Competition == nil {
		return nil, ErrNotOnLive
	}

	competition := &LiveCompetition{}
	for _, ce := range data.Competition.CompetitionEvents {
		for _, round := range ce.Rounds {
			competition.Rounds = append(competition.Rounds, LiveRound{
				ID:       round.ID,
				EventID:  ce.Event.ID,
				Number:   round.Number,
				Open:     round.Open,
				Finished: round.Finished,
			})
		}
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error
	slots := make(chan struct{}, liveRoundConcurrency)
	for i := range competition.Rounds {
		round := &competition.Rounds[i]
		if !round.Open {
			continue
		}
		wg.Go(func() {
			slots <- struct{}{}
			defer func() { <-slots }()

			results, err := r.loadRound(ctx, round.ID)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				if firstErr == nil {
					firstErr = err
				}
				return
			}
			round.Results = results
		})
	}
	wg.Wait()
	if firstErr != nil {
		return nil, firstErr
	}
	return competition, nil
}

func (r *LiveRepository) loadRound(ctx context.Context, roundID string) ([]LiveResult, error) {
	var data struct {
		Round *struct {
			Results []struct {
				Ranking          *int    `json:"ranking"`
				Best             int     `json:"best"`
				Average          int     `json:"average"`
				SingleRecordTag  *string `json:"singleRecordTag"`
				AverageRecordTag *string `json:"averageRecordTag"`
				Attempts         []struct {
					Result int `json:"result"`
				} `json:"attempts"`
				Person struct {
					WCAID   *string `json:"wcaId"`
					Name    string  `json:"name"`
					Country struct {
						ISO2 string `json:"iso2"`
					} `json:"country"`
				} `json:"person"`
			} `json:"results"`
		} `json:"round"`
	}
	if err := r.client.Query(ctx, liveRoundQuery, map[string]any{"id": roundID}, &data); err != nil {
		return nil, err
	}
	if data.Round == nil {
		return nil, nil
	}

	results := make([]LiveResult, 0, len(data.Round.Results))
	for _, raw := range data.Round.Results {
		result := LiveResult{
			Person: Person{
				Name:        raw.Person.Name,
				WCAID:       derefString(raw.Person.WCAID),
				CountryISO2: raw.Person.Country.ISO2,
			},
			Ranking:          raw.Ranking,
			Best:             raw.Best,
			Average:          raw.Average,
			SingleRecordTag:  derefString(raw.SingleRecordTag),
			AverageRecordTag: derefString(raw.AverageRecordTag),
			Attempts:         make([]int, 0, len(raw.Attempts)),
		}
		for _, attempt := range raw.Attempts {
			result.Attempts = append(result.Attempts, attempt.Result)
		}
		results = append(results, result)
	}
	return results, nil
}

func derefString(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
