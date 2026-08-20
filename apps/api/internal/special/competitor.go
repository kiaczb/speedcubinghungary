package special

import (
	"sort"
	"strings"
)

type attemptKey struct {
	resultID      int
	attemptNumber int
}

func BuildCompetitors(results []ResultSnapshot, recorded []RecordedTime) []Competitor {
	recordedByAttempt := make(map[attemptKey]RecordedTime, len(recorded))
	for _, entry := range recorded {
		recordedByAttempt[attemptKey{entry.LiveResultID, entry.AttemptNumber}] = entry
	}

	byRegistrant := make(map[int]*Competitor)

	for _, result := range results {
		competitor, exists := byRegistrant[result.RegistrantID]
		if !exists {
			competitor = &Competitor{
				RegistrantID:    result.RegistrantID,
				CompetitorWCAID: result.CompetitorWCAID,
				Name:            result.CompetitorName,
				Rounds:          []CompetitorRound{},
			}
			byRegistrant[result.RegistrantID] = competitor
		}

		competitor.Rounds = append(competitor.Rounds, buildRound(result, recordedByAttempt))
	}

	competitors := make([]Competitor, 0, len(byRegistrant))
	for _, competitor := range byRegistrant {
		sort.Slice(competitor.Rounds, func(i, j int) bool {
			if competitor.Rounds[i].EventID != competitor.Rounds[j].EventID {
				return competitor.Rounds[i].EventID < competitor.Rounds[j].EventID
			}
			return competitor.Rounds[i].RoundNumber < competitor.Rounds[j].RoundNumber
		})

		for _, round := range competitor.Rounds {
			competitor.UsedCs += round.CompletedCs + round.RecordedDNFCs
			competitor.MissingDNFCount += round.MissingDNFCount
			for _, attempt := range round.Attempts {
				if attempt.Result == AttemptDNF {
					competitor.DNFCount++
				}
			}
		}

		competitor.Budgets = buildBudgets(competitor.Rounds)
		competitors = append(competitors, *competitor)
	}

	sort.Slice(competitors, func(i, j int) bool {
		return competitors[i].RegistrantID < competitors[j].RegistrantID
	})

	return competitors
}

func buildRound(result ResultSnapshot, recorded map[attemptKey]RecordedTime) CompetitorRound {
	round := CompetitorRound{
		LiveResultID:       result.LiveResultID,
		RoundWCIFID:        result.RoundWCIFID,
		EventID:            result.EventID,
		RoundNumber:        result.RoundNumber,
		Attempts:           make([]CompetitorAttempt, 0, len(result.Attempts)),
		Best:               result.Best,
		Average:            result.Average,
		TimeLimitCs:        result.TimeLimitCs,
		CumulativeRoundIDs: result.CumulativeRoundIDs,
	}
	if round.CumulativeRoundIDs == nil {
		round.CumulativeRoundIDs = []string{}
	}

	for index, value := range result.Attempts {
		attempt := CompetitorAttempt{
			AttemptNumber: index + 1,
			Result:        value,
		}

		if value > 0 {
			round.CompletedCs += value
		}

		if value == AttemptDNF {
			attempt.MaxElapsedCs = attemptBound(
				result.TimeLimitCs, result.CumulativeRoundIDs, result.Attempts, index,
			)

			if entry, ok := recorded[attemptKey{result.LiveResultID, index + 1}]; ok {
				elapsed := entry.ElapsedCs
				attempt.ElapsedCs = &elapsed
				round.RecordedDNFCs += elapsed
			} else {
				round.MissingDNFCount++
			}
		}

		round.Attempts = append(round.Attempts, attempt)
	}

	return round
}

func buildBudgets(rounds []CompetitorRound) []CumulativeBudget {
	budgets := make(map[string]*CumulativeBudget)
	order := make([]string, 0)

	for _, round := range rounds {
		if len(round.CumulativeRoundIDs) == 0 || round.TimeLimitCs == nil {
			continue
		}

		ids := append([]string(nil), round.CumulativeRoundIDs...)
		sort.Strings(ids)
		key := strings.Join(ids, "|")

		budget, exists := budgets[key]
		if !exists {
			budget = &CumulativeBudget{
				RoundWCIFIDs: ids,
				LimitCs:      *round.TimeLimitCs,
			}
			budgets[key] = budget
			order = append(order, key)
		}

		budget.CompletedCs += round.CompletedCs
		budget.RecordedDNFCs += round.RecordedDNFCs
		budget.MissingDNFCount += round.MissingDNFCount
	}

	out := make([]CumulativeBudget, 0, len(order))
	for _, key := range order {
		budget := budgets[key]
		budget.UsedCs = budget.CompletedCs + budget.RecordedDNFCs
		budget.RemainingCs = budget.LimitCs - budget.UsedCs
		out = append(out, *budget)
	}

	return out
}
