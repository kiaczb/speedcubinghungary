package stats

import (
	"sort"

	"speedcubinghungary/internal/competition"
)

var roundOrder = map[string]int{
	"0": 0, "h": 0,
	"1": 1, "d": 1,
	"2": 2, "e": 2,
	"3": 3, "g": 3,
	"b": 4,
	"f": 5, "c": 5,
}

const sub10 = 1000

func computeResults(info CompetitionInfo, results []competition.Result) *ResultStats {
	stats := &ResultStats{
		Winners:        []EventWinner{},
		Records:        []RecordEntry{},
		Podiums:        []MedalHolder{},
		CountryPodiums: []CountryMedals{},
		Sub10Singles:   []Person{},
	}

	eventOrder := make(map[string]int, len(info.EventIDs))
	for i, id := range info.EventIDs {
		eventOrder[id] = i
	}

	finalRound := map[string]int{}
	for _, r := range results {
		if order := roundOrder[r.RoundTypeID]; order > finalRound[r.EventID] {
			finalRound[r.EventID] = order
		}
	}

	medals := map[string]*MedalHolder{}
	countries := map[string]*CountryMedals{}
	sub10People := map[string]bool{}

	for _, r := range results {
		person := Person{Name: r.Name, WCAID: r.WCAID, CountryISO2: r.CountryISO2}

		for _, attempt := range r.Attempts {
			switch {
			case attempt == resultDNF:
				stats.DNFs++
				stats.Solves++
			case attempt > 0:
				stats.Solves++
				if isTimedEvent(r.EventID) {
					stats.SolvingCentiseconds += int64(attempt)
				}
			}
		}

		if r.SingleRecord != "" {
			stats.Records = append(stats.Records, RecordEntry{
				Person: person, EventID: r.EventID, Type: "single", Record: r.SingleRecord,
				Display: FormatResult(r.EventID, r.Best, false),
			})
		}
		if r.AverageRecord != "" {
			stats.Records = append(stats.Records, RecordEntry{
				Person: person, EventID: r.EventID, Type: "average", Record: r.AverageRecord,
				Display: FormatResult(r.EventID, r.Average, true),
			})
		}

		if r.EventID == "333" && r.Best > 0 && r.Best < sub10 && !sub10People[personKey(person)] {
			sub10People[personKey(person)] = true
			stats.Sub10Singles = append(stats.Sub10Singles, person)
		}

		if roundOrder[r.RoundTypeID] != finalRound[r.EventID] || r.Pos < 1 || r.Pos > 3 || r.Best <= 0 {
			continue
		}

		if r.Pos == 1 {
			stats.Winners = append(stats.Winners, EventWinner{
				EventID: r.EventID,
				Person:  person,
				Best:    FormatResult(r.EventID, r.Best, false),
				Average: FormatResult(r.EventID, r.Average, true),
			})
		}

		key := personKey(person)
		if medals[key] == nil {
			medals[key] = &MedalHolder{Person: person}
		}
		if countries[r.CountryISO2] == nil {
			countries[r.CountryISO2] = &CountryMedals{CountryISO2: r.CountryISO2}
		}
		switch r.Pos {
		case 1:
			medals[key].Gold++
			countries[r.CountryISO2].Gold++
		case 2:
			medals[key].Silver++
			countries[r.CountryISO2].Silver++
		case 3:
			medals[key].Bronze++
			countries[r.CountryISO2].Bronze++
		}
	}

	sort.SliceStable(stats.Winners, func(i, j int) bool {
		return orderOf(eventOrder, stats.Winners[i].EventID) < orderOf(eventOrder, stats.Winners[j].EventID)
	})
	sortRecords(stats.Records, eventOrder)
	sortPeople(stats.Sub10Singles)

	for _, holder := range medals {
		stats.Podiums = append(stats.Podiums, *holder)
	}
	sortMedals(stats.Podiums)
	stats.Podiums = stats.Podiums[:min(topCount, len(stats.Podiums))]

	for _, country := range countries {
		stats.CountryPodiums = append(stats.CountryPodiums, *country)
	}
	sort.SliceStable(stats.CountryPodiums, func(i, j int) bool {
		a, b := stats.CountryPodiums[i], stats.CountryPodiums[j]
		if a.Gold != b.Gold {
			return a.Gold > b.Gold
		}
		if a.Silver != b.Silver {
			return a.Silver > b.Silver
		}
		if a.Bronze != b.Bronze {
			return a.Bronze > b.Bronze
		}
		return a.CountryISO2 < b.CountryISO2
	})
	return stats
}

func sortMedals(holders []MedalHolder) {
	sort.SliceStable(holders, func(i, j int) bool {
		a, b := holders[i], holders[j]
		if a.Gold != b.Gold {
			return a.Gold > b.Gold
		}
		if a.Silver != b.Silver {
			return a.Silver > b.Silver
		}
		if a.Bronze != b.Bronze {
			return a.Bronze > b.Bronze
		}
		return a.Person.Name < b.Person.Name
	})
}
