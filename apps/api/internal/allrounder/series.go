package allrounder

import (
	"sort"
	"time"
)

type SeriesCompetition struct {
	CompetitionWCAID string `json:"competitionWcaId"`
	Label            string `json:"label"`
}

type Series struct {
	Year            int
	Competitions    []SeriesCompetition
	PointsTable     []int
	CountingResults int
	PodiumCount     int
	LegacyScoring   bool
}

var standardPointsTable = []int{
	50, 43, 38, 34, 30, 27, 24, 21, 18, 16, 14, 12, 10, 8, 6, 5, 4, 3, 2, 1,
}

var seriesByYear = map[int]Series{
	2024: {
		Year: 2024,
		Competitions: []SeriesCompetition{
			{"HungarianAllRounder12024", "The Saga Begins"},
			{"HungarianAllRounder22024", "Second Chance"},
			{"HungarianAllRounder32024", "A Kind of Magic"},
			{"HungarianAllRounder42024", "Holiday"},
			{"HungarianAllRounder52024", "With or Without You"},
			{"HungarianAllRounder62024", "Shape of You"},
			{"HungarianAllRounder72024", "Egy Elfelejtett Szó"},
			{"HungarianAllRounder82024", "Iskolatáska"},
			{"HungarianAllRounder92024", "November Rain"},
			{"HungarianAllRounder102024", "See You Again"},
		},
		PointsTable:     standardPointsTable,
		CountingResults: 30,
		PodiumCount:     10,
		LegacyScoring:   true,
	},
	2025: {
		Year: 2025,
		Competitions: []SeriesCompetition{
			{"HungarianAllRounder1Avengers2025", "Avengers"},
			{"HungarianAllRounder2Frozen2025", "Frozen"},
			{"HungarianAllRounder3Minions2025", "Minions"},
			{"HungarianAllRounder4Macskafo2025", "Macskafogó"},
			{"HungarianAllRounder5StarWars2025", "Star Wars"},
			{"HungarianAllRounder62025", "Tom & Jerry"},
			{"HungarianAllRounder72025", "Néma filmek"},
			{"HungarianAllRounder82025", "Egri csillagok"},
			{"HungarianAllRounder92025", "Kis vakondok"},
			{"HungarianAllRounder102025", "Pityke őrmester"},
			{"HungarianAllRounder112025", "Pom Pom meséi"},
			{"HungarianAllRounder122025", "Reszkessetek, betörők!"},
		},
		PointsTable:     standardPointsTable,
		CountingResults: 50,
		PodiumCount:     10,
		LegacyScoring:   true,
	},
	2026: {
		Year: 2026,
		Competitions: []SeriesCompetition{
			{"HungarianAllRounder12026", "All Rounder 1"},
			{"HungarianAllRounder22026", "All Rounder 2"},
			{"HungarianAllRounder32026", "All Rounder 3"},
			{"HungarianAllRounder42026", "All Rounder 4"},
			{"HungarianAllRounder52026", "All Rounder 5"},
			{"HungarianAllRounder62026", "All Rounder 6"},
			{"HungarianAllRounder72026", "All Rounder 7"},
			{"HungarianAllRounder82026", "All Rounder 8"},
			{"HungarianAllRounder92026", "All Rounder 9"},
			{"HungarianAllRounder102026", "All Rounder 10"},
		},
		PointsTable:     standardPointsTable,
		CountingResults: 50,
		PodiumCount:     10,
	},
}

func AvailableYears() []int {
	years := make([]int, 0, len(seriesByYear))
	for year := range seriesByYear {
		years = append(years, year)
	}
	sort.Sort(sort.Reverse(sort.IntSlice(years)))

	return years
}

func ResolveSeries(year int) Series {
	if series, ok := seriesByYear[year]; ok {
		return series
	}
	return seriesByYear[defaultYear()]
}

func defaultYear() int {
	current := time.Now().Year()
	if _, ok := seriesByYear[current]; ok {
		return current
	}

	return AvailableYears()[0]
}

func (s Series) competitionIDs() []string {
	ids := make([]string, 0, len(s.Competitions))
	for _, competition := range s.Competitions {
		ids = append(ids, competition.CompetitionWCAID)
	}

	return ids
}
