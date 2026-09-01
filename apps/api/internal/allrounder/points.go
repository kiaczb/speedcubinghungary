package allrounder

import "sort"

type competitionResults struct {
	CompetitionWCAID string
	Results          []Result
}

type scoredResult struct {
	EventResult
	scanIndex int
}

type personAccumulator struct {
	wcaID  string
	name   string
	scored []scoredResult
}

func ComputeStandingsLegacy(competitions []competitionResults, pointsTable []int, countingResults int) []PersonStandings {
	people := make(map[string]*personAccumulator)
	competitionOrder := make(map[string]int, len(competitions))
	for index, competition := range competitions {
		competitionOrder[competition.CompetitionWCAID] = index
	}

	seen := make(map[string]bool)
	scanIndex := 0

	for _, competition := range competitions {
		invalidResultCount := 0
		prevEvent := ""

		for i := len(competition.Results) - 1; i >= 0; i-- {
			result := competition.Results[i]

			if prevEvent != result.EventID {
				invalidResultCount = 0
				prevEvent = result.EventID
			}

			if result.Best < 0 {
				invalidResultCount++
				continue
			}

			if result.Pos > len(pointsTable)+invalidResultCount {
				continue
			}

			if result.WCAID == "" {
				continue
			}

			key := result.WCAID + "|" + competition.CompetitionWCAID + "|" + result.EventID
			if seen[key] {
				continue
			}
			seen[key] = true

			pointsIndex := result.Pos - 1
			if result.Pos > len(pointsTable) {
				pointsIndex = result.Pos - invalidResultCount - 1
			}
			if pointsIndex < 0 || pointsIndex >= len(pointsTable) {
				continue
			}

			person, known := people[result.WCAID]
			if !known {
				person = &personAccumulator{wcaID: result.WCAID, name: result.Name}
				people[result.WCAID] = person
			}

			person.scored = append(person.scored, scoredResult{
				EventResult: EventResult{
					CompetitionWCAID: competition.CompetitionWCAID,
					EventID:          result.EventID,
					Ranking:          result.Pos,
					Points:           pointsTable[pointsIndex],
				},
				scanIndex: scanIndex,
			})
			scanIndex++
		}
	}

	return finalizeAll(people, countingResults, competitionOrder)
}

func ComputeStandingsByRound(competitions []CompetitionRounds, pointsTable []int, countingResults int) []PersonStandings {
	people := make(map[string]*personAccumulator)
	competitionOrder := make(map[string]int, len(competitions))
	for index, competition := range competitions {
		competitionOrder[competition.CompetitionWCAID] = index
	}

	scanIndex := 0

	for _, competition := range competitions {
		for _, event := range groupRoundsByEvent(competition.Rounds) {
			awarded := make(map[string]bool)

			for _, round := range event {
				for _, placement := range round.Placements {
					if awarded[placement.WCAID] {
						continue
					}

					points := pointsForPlacement(placement, pointsTable)
					if points == 0 {
						continue
					}
					awarded[placement.WCAID] = true

					person, known := people[placement.WCAID]
					if !known {
						person = &personAccumulator{wcaID: placement.WCAID, name: placement.Name}
						people[placement.WCAID] = person
					}

					person.scored = append(person.scored, scoredResult{
						EventResult: EventResult{
							CompetitionWCAID: competition.CompetitionWCAID,
							EventID:          round.EventID,
							Ranking:          placement.Pos,
							Points:           points,
						},
						scanIndex: scanIndex,
					})
					scanIndex++
				}
			}
		}
	}

	return finalizeAll(people, countingResults, competitionOrder)
}

func pointsForPlacement(placement Placement, pointsTable []int) int {
	if placement.Best <= 0 {
		return 0
	}
	if placement.Pos < 1 || placement.Pos > len(pointsTable) {
		return 0
	}

	return pointsTable[placement.Pos-1]
}

func groupRoundsByEvent(rounds []RoundResults) [][]RoundResults {
	byEvent := make(map[string][]RoundResults)
	order := make([]string, 0)

	for _, round := range rounds {
		if _, known := byEvent[round.EventID]; !known {
			order = append(order, round.EventID)
		}
		byEvent[round.EventID] = append(byEvent[round.EventID], round)
	}

	sort.Strings(order)

	events := make([][]RoundResults, 0, len(order))
	for _, eventID := range order {
		eventRounds := byEvent[eventID]
		sort.SliceStable(eventRounds, func(i, j int) bool {
			return eventRounds[i].Rank > eventRounds[j].Rank
		})
		events = append(events, eventRounds)
	}

	return events
}

func finalizeAll(
	people map[string]*personAccumulator,
	countingResults int,
	competitionOrder map[string]int,
) []PersonStandings {
	standings := make([]PersonStandings, 0, len(people))
	for _, person := range people {
		standing := person.finalize(countingResults, competitionOrder)
		if standing.TotalPoints > 0 {
			standings = append(standings, standing)
		}
	}

	sort.Slice(standings, func(i, j int) bool {
		if standings[i].TotalPoints != standings[j].TotalPoints {
			return standings[i].TotalPoints > standings[j].TotalPoints
		}
		return standings[i].Name < standings[j].Name
	})

	return standings
}

func (p *personAccumulator) finalize(countingResults int, competitionOrder map[string]int) PersonStandings {
	sort.Slice(p.scored, func(i, j int) bool {
		if p.scored[i].Points != p.scored[j].Points {
			return p.scored[i].Points > p.scored[j].Points
		}
		return p.scored[i].scanIndex < p.scored[j].scanIndex
	})

	counting := p.scored
	if len(counting) > countingResults {
		counting = counting[:countingResults]
	}

	standing := PersonStandings{
		WCAID:   p.wcaID,
		Name:    p.name,
		Results: make([]EventResult, 0, len(counting)),
	}

	for _, scored := range counting {
		standing.TotalPoints += scored.Points
		standing.Results = append(standing.Results, scored.EventResult)
	}

	sort.Slice(standing.Results, func(i, j int) bool {
		left, right := standing.Results[i], standing.Results[j]
		if left.CompetitionWCAID != right.CompetitionWCAID {
			return competitionOrder[left.CompetitionWCAID] < competitionOrder[right.CompetitionWCAID]
		}
		return left.EventID < right.EventID
	})

	return standing
}
