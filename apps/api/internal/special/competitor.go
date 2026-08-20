package special

import (
	"sort"
	"strings"
)

// attemptKey egy konkrét kísérletet azonosít: a WCA Live result id-ja
// (verseny + kör + versenyző) és a kísérlet 1-alapú sorszáma.
type attemptKey struct {
	resultID      int
	attemptNumber int
}

// BuildCompetitors összeállítja a versenyzői nézetet a WCA Live friss
// eredményeiből és a saját, rögzített DNF-időinkből.
//
// A felhasznált idő a Szabályzat A1a3 szerint áll össze: a sikeres kísérletek
// eredménye + a DNF kísérletek ELTELT ideje. Utóbbi csak akkor ismert, ha
// valaki rögzítette; a hiányzókat a MissingDNFCount jelzi, és amíg az nem
// nulla, a RemainingCs csak felső becslés.
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

	// A rajtszám versenyzőnként egyedi, ezért ez teljes (determinisztikus)
	// rendezés — a map bejárási sorrendje nem szivárog ki.
	sort.Slice(competitors, func(i, j int) bool {
		return competitors[i].RegistrantID < competitors[j].RegistrantID
	})

	return competitors
}

// buildRound egy kör sorát állítja össze, a DNF kísérletekhez hozzákötve a
// rögzített tényleges időt és a felső korlátot.
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

		// Csak a -1 (DNF) kap rögzíthető időt:
		//   - -2 (DNS): a versenyző el sem indult, nincs eltelt idő (Szabályzat 9f2),
		//   - 0 (skipped): cutoff miatt kimaradt vagy még meg sem történt kísérlet.
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

// buildBudgets a kumulatív kereteket állítja össze. A keret azonosítója a
// cumulativeRoundIds halmaz — az ugyanabba a keretbe tartozó körök összeadódnak.
func buildBudgets(rounds []CompetitorRound) []CumulativeBudget {
	budgets := make(map[string]*CumulativeBudget)
	order := make([]string, 0)

	for _, round := range rounds {
		if len(round.CumulativeRoundIDs) == 0 || round.TimeLimitCs == nil {
			continue
		}

		// A keret azonosítója a körök rendezett halmaza — így a sorrendtől
		// függetlenül ugyanaz a keret ugyanoda kerül.
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
