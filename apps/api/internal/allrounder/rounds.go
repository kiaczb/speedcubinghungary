package allrounder

import "sort"

type Placement struct {
	WCAID string
	Name  string
	Pos   int
	Best  int
}

type RoundResults struct {
	EventID     string
	RoundTypeID string
	Rank        int
	Placements  []Placement
}

type CompetitionRounds struct {
	CompetitionWCAID string
	Rounds           []RoundResults
	Live             bool
}

var roundRanks = map[string]int{
	"0": 0,
	"1": 1, "d": 1,
	"2": 2, "e": 2,
	"3": 3, "g": 3,
	"f": 4, "c": 4,
}

func roundRank(roundTypeID string) int {
	if rank, known := roundRanks[roundTypeID]; known {
		return rank
	}
	return 0
}

func RoundsFromResults(results []Result) []RoundResults {
	type key struct {
		eventID     string
		roundTypeID string
	}

	grouped := make(map[key][]Placement)
	for _, result := range results {
		if result.WCAID == "" {
			continue
		}

		k := key{eventID: result.EventID, roundTypeID: result.RoundTypeID}
		grouped[k] = append(grouped[k], Placement{
			WCAID: result.WCAID,
			Name:  result.Name,
			Pos:   result.Pos,
			Best:  result.Best,
		})
	}

	rounds := make([]RoundResults, 0, len(grouped))
	for k, placements := range grouped {
		rounds = append(rounds, RoundResults{
			EventID:     k.eventID,
			RoundTypeID: k.roundTypeID,
			Rank:        roundRank(k.roundTypeID),
			Placements:  placements,
		})
	}

	sortRounds(rounds)

	return rounds
}

func sortRounds(rounds []RoundResults) {
	for i := range rounds {
		placements := rounds[i].Placements
		sort.SliceStable(placements, func(a, b int) bool {
			if placements[a].Pos != placements[b].Pos {
				return placements[a].Pos < placements[b].Pos
			}
			return placements[a].WCAID < placements[b].WCAID
		})
	}

	sort.SliceStable(rounds, func(a, b int) bool {
		if rounds[a].EventID != rounds[b].EventID {
			return rounds[a].EventID < rounds[b].EventID
		}
		return rounds[a].Rank < rounds[b].Rank
	})
}
