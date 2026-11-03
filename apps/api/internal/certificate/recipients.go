package certificate

import (
	"sort"
	"strings"

	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/stats"
	"speedcubinghungary/internal/wcaformat"
)

const podiumSize = 3

const (
	attemptDNF    = -1
	resultAverage = "average"
	resultSingle  = "single"
)

var rankedBySingle = map[string]bool{
	"333bf":  true,
	"444bf":  true,
	"555bf":  true,
	"333mbf": true,
	"333mbo": true,
}

func rankingResult(eventID string, best, average int) (string, string) {
	if !rankedBySingle[eventID] && average > 0 {
		return wcaformat.Result(eventID, average, true), resultAverage
	}
	return wcaformat.Result(eventID, best, false), resultSingle
}

func competed(attempts []int) bool {
	for _, attempt := range attempts {
		if attempt > 0 || attempt == attemptDNF {
			return true
		}
	}
	return false
}

func personKey(wcaID, name string) string {
	if wcaID != "" {
		return wcaID
	}
	return "name:" + name
}

func eventOrder(eventIDs []string) map[string]int {
	order := make(map[string]int, len(eventIDs))
	for i, id := range eventIDs {
		order[id] = i
	}
	return order
}

func orderedEvents(eventIDs []string, seen map[string]bool) []string {
	events := make([]string, 0, len(eventIDs))
	known := make(map[string]bool, len(eventIDs))
	for _, id := range eventIDs {
		known[id] = true
		events = append(events, id)
	}
	extra := make([]string, 0)
	for id := range seen {
		if !known[id] {
			extra = append(extra, id)
		}
	}
	sort.Strings(extra)
	return append(events, extra...)
}

func podiumFromResults(eventIDs []string, results []competition.Result) ([]Event, []Recipient) {
	finalRound := map[string]int{}
	for _, r := range results {
		order := stats.RoundOrder(r.RoundTypeID)
		if final, ok := finalRound[r.EventID]; !ok || order > final {
			finalRound[r.EventID] = order
		}
	}

	seen := make(map[string]bool, len(finalRound))
	for id := range finalRound {
		seen[id] = true
	}
	events := make([]Event, 0, len(eventIDs))
	for _, id := range orderedEvents(eventIDs, seen) {
		events = append(events, Event{EventID: id, Status: EventFinished})
	}

	recipients := make([]Recipient, 0)
	for _, r := range results {
		if stats.RoundOrder(r.RoundTypeID) != finalRound[r.EventID] || r.Pos < 1 || r.Pos > podiumSize || r.Best <= 0 {
			continue
		}
		result, resultType := rankingResult(r.EventID, r.Best, r.Average)
		recipients = append(recipients, Recipient{
			Name:        r.Name,
			WCAID:       r.WCAID,
			CountryISO2: r.CountryISO2,
			EventID:     r.EventID,
			Place:       r.Pos,
			Result:      result,
			ResultType:  resultType,
		})
	}
	sortPodium(recipients, eventIDs)
	return events, recipients
}

func podiumFromLive(eventIDs []string, live *stats.LiveCompetition) ([]Event, []Recipient) {
	finals := map[string]stats.LiveRound{}
	seen := map[string]bool{}
	for _, round := range live.Rounds {
		seen[round.EventID] = true
		if final, ok := finals[round.EventID]; !ok || round.Number > final.Number {
			finals[round.EventID] = round
		}
	}

	events := make([]Event, 0, len(eventIDs))
	recipients := make([]Recipient, 0)
	for _, id := range orderedEvents(eventIDs, seen) {
		final, ok := finals[id]
		status := EventPending
		switch {
		case ok && final.Finished:
			status = EventFinished
		case ok && final.Open:
			status = EventRunning
		}
		events = append(events, Event{EventID: id, Status: status})
		if status != EventFinished {
			continue
		}

		for _, r := range final.Results {
			if r.Ranking == nil || *r.Ranking < 1 || *r.Ranking > podiumSize || r.Best <= 0 {
				continue
			}
			result, resultType := rankingResult(id, r.Best, r.Average)
			recipients = append(recipients, Recipient{
				Name:        r.Person.Name,
				WCAID:       r.Person.WCAID,
				CountryISO2: r.Person.CountryISO2,
				EventID:     id,
				Place:       *r.Ranking,
				Result:      result,
				ResultType:  resultType,
			})
		}
	}
	sortPodium(recipients, eventIDs)
	return events, recipients
}

func pendingEvents(eventIDs []string) []Event {
	events := make([]Event, 0, len(eventIDs))
	for _, id := range eventIDs {
		events = append(events, Event{EventID: id, Status: EventPending})
	}
	return events
}

func competitorsFromResults(results []competition.Result) []Recipient {
	people := map[string]Recipient{}
	for _, r := range results {
		if !competed(r.Attempts) {
			continue
		}
		people[personKey(r.WCAID, r.Name)] = Recipient{Name: r.Name, WCAID: r.WCAID, CountryISO2: r.CountryISO2}
	}
	return sortedPeople(people)
}

func competitorsFromLive(live *stats.LiveCompetition) []Recipient {
	people := map[string]Recipient{}
	for _, round := range live.Rounds {
		for _, r := range round.Results {
			if !competed(r.Attempts) {
				continue
			}
			people[personKey(r.Person.WCAID, r.Person.Name)] = Recipient{
				Name:        r.Person.Name,
				WCAID:       r.Person.WCAID,
				CountryISO2: r.Person.CountryISO2,
			}
		}
	}
	return sortedPeople(people)
}

func registrantsOf(registrations competition.Registrations) []Recipient {
	people := make(map[string]Recipient, len(registrations.Registrants))
	for _, r := range registrations.Registrants {
		people[personKey(r.WCAID, r.Name)] = Recipient{Name: r.Name, WCAID: r.WCAID, CountryISO2: r.CountryISO2}
	}
	return sortedPeople(people)
}

func sortedPeople(people map[string]Recipient) []Recipient {
	recipients := make([]Recipient, 0, len(people))
	for _, person := range people {
		recipients = append(recipients, person)
	}
	sort.Slice(recipients, func(i, j int) bool {
		a, b := strings.ToLower(recipients[i].Name), strings.ToLower(recipients[j].Name)
		if a != b {
			return a < b
		}
		return recipients[i].WCAID < recipients[j].WCAID
	})
	return recipients
}

func sortPodium(recipients []Recipient, eventIDs []string) {
	order := eventOrder(eventIDs)
	position := func(eventID string) int {
		if i, ok := order[eventID]; ok {
			return i
		}
		return len(order)
	}
	sort.SliceStable(recipients, func(i, j int) bool {
		a, b := recipients[i], recipients[j]
		if pa, pb := position(a.EventID), position(b.EventID); pa != pb {
			return pa < pb
		}
		if a.EventID != b.EventID {
			return a.EventID < b.EventID
		}
		if a.Place != b.Place {
			return a.Place < b.Place
		}
		return strings.ToLower(a.Name) < strings.ToLower(b.Name)
	})
}
