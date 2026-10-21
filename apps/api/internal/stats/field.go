package stats

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"speedcubinghungary/internal/competition"
)

const (
	topCount = 5
)

var anniversaryYears = []int{30, 25, 20, 15, 10, 5}

func personOf(r competition.Registrant) Person {
	return Person{Name: r.Name, WCAID: r.WCAID, CountryISO2: r.CountryISO2}
}

func startYear(wcaID string) (int, bool) {
	if len(wcaID) < 4 {
		return 0, false
	}
	year, err := strconv.Atoi(wcaID[:4])
	return year, err == nil
}

func competitionYear(info CompetitionInfo) int {
	year, _ := startYear(info.StartDate)
	return year
}

func isMilestone(n int) bool {
	return n >= 50 && n%50 == 0
}

func computeField(info CompetitionInfo, phase Phase, registrations competition.Registrations, summaries map[string]competition.PersonSummary) *FieldStats {
	registrants := registrations.Registrants
	eventIDs := registrations.EventIDs
	if len(eventIDs) == 0 {
		eventIDs = info.EventIDs
	}

	field := &FieldStats{
		Competitors:      len(registrants),
		Limit:            info.CompetitorLimit,
		Countries:        []Count{},
		Genders:          []Count{},
		Events:           []Count{},
		AllEvents:        []Person{},
		TopRanked:        []RankedResult{},
		RecordHolders:    []RankedResult{},
		Veterans:         []PersonValue{},
		Anniversaries:    []Anniversary{},
		Milestones:       []PersonValue{},
		MostCompetitions: []PersonValue{},
		MostMedals:       []MedalHolder{},
		MostRecords:      []RecordHolder{},
	}

	countries := map[string]int{}
	genders := map[string]int{}
	events := map[string]int{}
	registeredEvents := 0
	eventOrder := make(map[string]int, len(eventIDs))
	for i, id := range eventIDs {
		eventOrder[id] = i
	}

	for _, r := range registrants {
		if r.WCAID == "" {
			field.Newcomers++
		}
		if info.CountryISO2 != "" && r.CountryISO2 != info.CountryISO2 {
			field.Foreigners++
		}
		countries[r.CountryISO2]++
		genders[r.Gender]++

		own := map[string]bool{}
		for _, id := range r.EventIDs {
			if _, ok := eventOrder[id]; ok && !own[id] {
				own[id] = true
				events[id]++
				registeredEvents++
			}
		}
		if len(eventIDs) > 1 && len(own) == len(eventIDs) {
			field.AllEvents = append(field.AllEvents, personOf(r))
		}
	}

	field.Countries = sortedCounts(countries)
	field.Genders = sortedCounts(genders)
	for _, id := range eventIDs {
		field.Events = append(field.Events, Count{Key: id, Count: events[id]})
	}
	if len(registrants) > 0 {
		field.AverageEvents = math.Round(float64(registeredEvents)/float64(len(registrants))*10) / 10
	}
	sortPeople(field.AllEvents)

	field.TopRanked = topRanked(registrants)
	field.RecordHolders = recordHolders(registrants, eventOrder)
	field.Veterans, field.Anniversaries = veteransAndAnniversaries(registrants, competitionYear(info))

	if summaries != nil {
		addSummaryStats(field, phase, registrants, summaries)
	}
	return field
}

func sortedCounts(counts map[string]int) []Count {
	result := make([]Count, 0, len(counts))
	for key, count := range counts {
		result = append(result, Count{Key: key, Count: count})
	}
	sort.Slice(result, func(i, j int) bool {
		if result[i].Count != result[j].Count {
			return result[i].Count > result[j].Count
		}
		return result[i].Key < result[j].Key
	})
	return result
}

func sortPeople(people []Person) {
	sort.Slice(people, func(i, j int) bool {
		return strings.ToLower(people[i].Name) < strings.ToLower(people[j].Name)
	})
}

func recordOf(pb competition.PersonalBest) string {
	switch {
	case pb.WorldRanking == 1:
		return "WR"
	case pb.ContinentalRanking == 1:
		return "CR"
	case pb.NationalRanking == 1:
		return "NR"
	}
	return ""
}

func rankedResult(r competition.Registrant, pb competition.PersonalBest) RankedResult {
	return RankedResult{
		Person:          personOf(r),
		EventID:         pb.EventID,
		Type:            pb.Type,
		Display:         FormatResult(pb.EventID, pb.Best, pb.Type == "average"),
		Record:          recordOf(pb),
		WorldRanking:    pb.WorldRanking,
		NationalRanking: pb.NationalRanking,
	}
}

func topRanked(registrants []competition.Registrant) []RankedResult {
	best := make([]RankedResult, 0)
	for _, r := range registrants {
		var top *competition.PersonalBest
		for i := range r.PersonalBests {
			pb := &r.PersonalBests[i]
			if pb.WorldRanking <= 0 {
				continue
			}
			if top == nil || pb.WorldRanking < top.WorldRanking {
				top = pb
			}
		}
		if top != nil {
			best = append(best, rankedResult(r, *top))
		}
	}
	sort.SliceStable(best, func(i, j int) bool { return best[i].WorldRanking < best[j].WorldRanking })
	return best[:min(topCount, len(best))]
}

func recordHolders(registrants []competition.Registrant, eventOrder map[string]int) []RankedResult {
	holders := make([]RankedResult, 0)
	for _, r := range registrants {
		for _, pb := range r.PersonalBests {
			if pb.NationalRanking == 1 {
				holders = append(holders, rankedResult(r, pb))
			}
		}
	}
	sort.SliceStable(holders, func(i, j int) bool {
		a, b := holders[i], holders[j]
		if ao, bo := orderOf(eventOrder, a.EventID), orderOf(eventOrder, b.EventID); ao != bo {
			return ao < bo
		}
		if a.Type != b.Type {
			return a.Type == "single"
		}
		return a.Person.Name < b.Person.Name
	})
	return holders
}

func orderOf(eventOrder map[string]int, eventID string) int {
	if order, ok := eventOrder[eventID]; ok {
		return order
	}
	return len(eventOrder)
}

func veteransAndAnniversaries(registrants []competition.Registrant, year int) ([]PersonValue, []Anniversary) {
	veterans := make([]PersonValue, 0)
	byYears := map[int][]Person{}
	for _, r := range registrants {
		start, ok := startYear(r.WCAID)
		if !ok {
			continue
		}
		veterans = append(veterans, PersonValue{Person: personOf(r), Value: start})
		if year > 0 {
			byYears[year-start] = append(byYears[year-start], personOf(r))
		}
	}
	sort.SliceStable(veterans, func(i, j int) bool {
		if veterans[i].Value != veterans[j].Value {
			return veterans[i].Value < veterans[j].Value
		}
		return veterans[i].Person.WCAID < veterans[j].Person.WCAID
	})

	anniversaries := make([]Anniversary, 0)
	for _, years := range anniversaryYears {
		if people := byYears[years]; len(people) > 0 {
			sortPeople(people)
			anniversaries = append(anniversaries, Anniversary{Years: years, StartYear: year - years, People: people})
		}
	}
	return veterans[:min(topCount, len(veterans))], anniversaries
}

func addSummaryStats(field *FieldStats, phase Phase, registrants []competition.Registrant, summaries map[string]competition.PersonSummary) {
	competitionsCount := make([]PersonValue, 0)
	for _, r := range registrants {
		summary, ok := summaries[r.WCAID]
		if !ok {
			continue
		}
		person := personOf(r)
		field.CareerSolves += summary.TotalSolves

		nth := summary.CompetitionCount
		if phase != PhaseAfter {
			nth++
		}
		if isMilestone(nth) {
			field.Milestones = append(field.Milestones, PersonValue{Person: person, Value: nth})
		}

		competitionsCount = append(competitionsCount, PersonValue{Person: person, Value: summary.CompetitionCount})
		if summary.Medals.Total > 0 {
			field.MostMedals = append(field.MostMedals, MedalHolder{
				Person: person,
				Gold:   summary.Medals.Gold,
				Silver: summary.Medals.Silver,
				Bronze: summary.Medals.Bronze,
			})
		}
		if summary.Records.Total > 0 {
			field.MostRecords = append(field.MostRecords, RecordHolder{
				Person:      person,
				World:       summary.Records.World,
				Continental: summary.Records.Continental,
				National:    summary.Records.National,
			})
		}
	}

	sort.SliceStable(field.Milestones, func(i, j int) bool { return field.Milestones[i].Value > field.Milestones[j].Value })
	sort.SliceStable(competitionsCount, func(i, j int) bool { return competitionsCount[i].Value > competitionsCount[j].Value })
	field.MostCompetitions = competitionsCount[:min(topCount, len(competitionsCount))]

	sort.SliceStable(field.MostMedals, func(i, j int) bool {
		a, b := field.MostMedals[i], field.MostMedals[j]
		if a.Gold != b.Gold {
			return a.Gold > b.Gold
		}
		if a.Silver != b.Silver {
			return a.Silver > b.Silver
		}
		return a.Bronze > b.Bronze
	})
	field.MostMedals = field.MostMedals[:min(topCount, len(field.MostMedals))]

	sort.SliceStable(field.MostRecords, func(i, j int) bool {
		a, b := field.MostRecords[i], field.MostRecords[j]
		if a.World != b.World {
			return a.World > b.World
		}
		if a.Continental != b.Continental {
			return a.Continental > b.Continental
		}
		return a.National > b.National
	})
	field.MostRecords = field.MostRecords[:min(topCount, len(field.MostRecords))]
}
