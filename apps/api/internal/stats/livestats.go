package stats

import (
	"sort"
	"time"
)

func computeLive(info CompetitionInfo, live *LiveCompetition, fetchedAt time.Time) *LiveStats {
	stats := &LiveStats{
		Records:    []RecordEntry{},
		PRLeaders:  []PersonValue{},
		EventBests: []EventBest{},
		FetchedAt:  fetchedAt,
	}

	eventOrder := make(map[string]int, len(info.EventIDs))
	for i, id := range info.EventIDs {
		eventOrder[id] = i
	}

	prs := map[string]*PersonValue{}
	bests := map[string]*EventBest{}
	bestValues := map[string][2]int{}
	events := make([]string, 0)

	for _, round := range live.Rounds {
		stats.RoundsTotal++
		if round.Open {
			stats.RoundsOpen++
		}
		if round.Finished {
			stats.RoundsFinished++
		}

		for _, result := range round.Results {
			entered := false
			for _, attempt := range result.Attempts {
				switch {
				case attempt == resultDNF:
					stats.DNFs++
					stats.Attempts++
					entered = true
				case attempt > 0:
					stats.Attempts++
					entered = true
				}
			}
			if !entered {
				continue
			}
			stats.ResultsEntered++

			for _, tagged := range []struct {
				tag     string
				kind    string
				value   int
				average bool
			}{
				{result.SingleRecordTag, "single", result.Best, false},
				{result.AverageRecordTag, "average", result.Average, true},
			} {
				if tagged.tag == "" {
					continue
				}
				stats.PRs++
				key := personKey(result.Person)
				if prs[key] == nil {
					prs[key] = &PersonValue{Person: result.Person}
				}
				prs[key].Value++
				if tagged.tag != "PR" {
					stats.Records = append(stats.Records, RecordEntry{
						Person:  result.Person,
						EventID: round.EventID,
						Type:    tagged.kind,
						Record:  tagged.tag,
						Display: FormatResult(round.EventID, tagged.value, tagged.average),
					})
				}
			}

			if bests[round.EventID] == nil {
				bests[round.EventID] = &EventBest{EventID: round.EventID}
				events = append(events, round.EventID)
			}
			best := bests[round.EventID]
			values := bestValues[round.EventID]
			if result.Best > 0 && (values[0] == 0 || result.Best < values[0]) {
				values[0] = result.Best
				best.Single = &Highlight{Person: result.Person, Display: FormatResult(round.EventID, result.Best, false)}
			}
			if result.Average > 0 && (values[1] == 0 || result.Average < values[1]) {
				values[1] = result.Average
				best.Average = &Highlight{Person: result.Person, Display: FormatResult(round.EventID, result.Average, true)}
			}
			bestValues[round.EventID] = values
		}
	}

	for _, pr := range prs {
		stats.PRLeaders = append(stats.PRLeaders, *pr)
	}
	sort.SliceStable(stats.PRLeaders, func(i, j int) bool {
		a, b := stats.PRLeaders[i], stats.PRLeaders[j]
		if a.Value != b.Value {
			return a.Value > b.Value
		}
		return a.Person.Name < b.Person.Name
	})
	stats.PRLeaders = stats.PRLeaders[:min(topCount, len(stats.PRLeaders))]

	sort.SliceStable(events, func(i, j int) bool {
		return orderOf(eventOrder, events[i]) < orderOf(eventOrder, events[j])
	})
	for _, id := range events {
		stats.EventBests = append(stats.EventBests, *bests[id])
	}
	sortRecords(stats.Records, eventOrder)
	return stats
}

func personKey(person Person) string {
	if person.WCAID != "" {
		return person.WCAID
	}
	return "name:" + person.Name
}

func sortRecords(records []RecordEntry, eventOrder map[string]int) {
	sort.SliceStable(records, func(i, j int) bool {
		a, b := records[i], records[j]
		if ao, bo := orderOf(eventOrder, a.EventID), orderOf(eventOrder, b.EventID); ao != bo {
			return ao < bo
		}
		return a.Type == "single" && b.Type != "single"
	})
}
