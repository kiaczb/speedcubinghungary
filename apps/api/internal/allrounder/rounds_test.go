package allrounder

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func rounds(rows ...Result) []RoundResults {
	return RoundsFromResults(results(rows...))
}

func TestRoundsFromResults_GroupsByEventAndRound(t *testing.T) {
	got := rounds(
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "1", Pos: 2, Best: 900},
		Result{WCAID: "2020BBBB01", Name: "Bence", EventID: "333", RoundTypeID: "1", Pos: 1, Best: 800},
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "222", RoundTypeID: "c", Pos: 1, Best: 300},
	)

	require.Len(t, got, 3)

	assert.Equal(t, "222", got[0].EventID)
	assert.Equal(t, 4, got[0].Rank, "a kombinált döntő ('c') ugyanaz a kör, mint az 'f'")

	assert.Equal(t, "333", got[1].EventID)
	assert.Equal(t, 1, got[1].Rank)
	assert.Equal(t, "333", got[2].EventID)
	assert.Equal(t, 4, got[2].Rank)
}

func TestRoundsFromResults_SortsPlacementsByPosition(t *testing.T) {
	got := rounds(
		Result{WCAID: "2020CCCC01", Name: "Cili", EventID: "333", RoundTypeID: "1", Pos: 8, Best: 1500},
		Result{WCAID: "2020DDDD01", Name: "Dani", EventID: "333", RoundTypeID: "1", Pos: 25, Best: 2500},
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "1", Pos: 1, Best: 700},
		Result{WCAID: "2020BBBB01", Name: "Bence", EventID: "333", RoundTypeID: "1", Pos: 13, Best: 1900},
	)

	require.Len(t, got, 1)
	positions := make([]int, 0, len(got[0].Placements))
	for _, placement := range got[0].Placements {
		positions = append(positions, placement.Pos)
	}
	assert.Equal(t, []int{1, 8, 13, 25}, positions)
}

func TestRoundsFromResults_SkipsRowsWithoutWCAID(t *testing.T) {
	got := rounds(
		Result{WCAID: "", Name: "Névtelen", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 2, Best: 800},
	)

	require.Len(t, got, 1)
	require.Len(t, got[0].Placements, 1, "WCA-azonosító nélkül nem tudjuk, kihez tartozik")
	assert.Equal(t, "2020AAAA01", got[0].Placements[0].WCAID)
}

func TestRoundRank_MapsCombinedAndPlainRoundsTogether(t *testing.T) {
	assert.Equal(t, roundRank("1"), roundRank("d"))
	assert.Equal(t, roundRank("2"), roundRank("e"))
	assert.Equal(t, roundRank("3"), roundRank("g"))
	assert.Equal(t, roundRank("f"), roundRank("c"))

	assert.Less(t, roundRank("1"), roundRank("2"))
	assert.Less(t, roundRank("3"), roundRank("f"))
	assert.Less(t, roundRank("0"), roundRank("1"), "a kvalifikációs kör a legkorábbi")
}
