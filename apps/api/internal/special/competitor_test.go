package special

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func snapshot(resultID, registrantID int, name, eventID string, roundNumber int, atts []int) ResultSnapshot {
	return ResultSnapshot{
		LiveResultID:       resultID,
		EventID:            eventID,
		RoundNumber:        roundNumber,
		RoundWCIFID:        RoundWCIFID(eventID, roundNumber),
		RegistrantID:       registrantID,
		CompetitorName:     name,
		Attempts:           atts,
		CumulativeRoundIDs: []string{},
	}
}

func withLimit(result ResultSnapshot, limitCs int) ResultSnapshot {
	result.TimeLimitCs = &limitCs
	return result
}

func withCumulative(result ResultSnapshot, limitCs int, roundIDs ...string) ResultSnapshot {
	result.TimeLimitCs = &limitCs
	result.CumulativeRoundIDs = roundIDs
	return result
}

func recorded(resultID, attemptNumber, elapsedCs int) RecordedTime {
	return RecordedTime{
		LiveResultID:  resultID,
		AttemptNumber: attemptNumber,
		ElapsedCs:     elapsedCs,
	}
}

func TestBuildCompetitors_AttachesRecordedDurations(t *testing.T) {
	competitors := BuildCompetitors(
		[]ResultSnapshot{
			withLimit(
				snapshot(101, 7, "Teszt Elek", "333", 1, []int{696, AttemptDNF, 674, AttemptDNS, AttemptSkipped}),
				60000,
			),
		},
		[]RecordedTime{recorded(101, 2, 3742)},
	)

	require.Len(t, competitors, 1)
	competitor := competitors[0]
	assert.Equal(t, 7, competitor.RegistrantID)
	assert.Equal(t, 1, competitor.DNFCount)
	assert.Zero(t, competitor.MissingDNFCount)

	require.Len(t, competitor.Rounds, 1)
	round := competitor.Rounds[0]
	require.Len(t, round.Attempts, 5)

	assert.Equal(t, 101, round.LiveResultID)

	assert.Equal(t, 696, round.Attempts[0].Result)
	assert.Nil(t, round.Attempts[0].ElapsedCs, "sikeres kísérlethez nem tartozik rögzített idő")
	assert.Nil(t, round.Attempts[0].MaxElapsedCs, "korlát csak DNF-re értelmes")

	assert.Equal(t, AttemptDNF, round.Attempts[1].Result)
	require.NotNil(t, round.Attempts[1].ElapsedCs)
	assert.Equal(t, 3742, *round.Attempts[1].ElapsedCs)
	require.NotNil(t, round.Attempts[1].MaxElapsedCs)
	assert.Equal(t, 60000, *round.Attempts[1].MaxElapsedCs)

	assert.Equal(t, AttemptDNS, round.Attempts[3].Result)
	assert.Nil(t, round.Attempts[3].ElapsedCs, "a DNS nem DNF")
	assert.Nil(t, round.Attempts[3].MaxElapsedCs)

	assert.Equal(t, 696+674, round.CompletedCs)
	assert.Equal(t, 3742, round.RecordedDNFCs)
	assert.Equal(t, 696+674+3742, competitor.UsedCs)
}

func TestBuildCompetitors_CountsMissingDurations(t *testing.T) {
	competitors := BuildCompetitors(
		[]ResultSnapshot{
			snapshot(101, 7, "Teszt Elek", "333", 1, []int{AttemptDNF, 800, AttemptDNF}),
		},
		nil,
	)

	require.Len(t, competitors, 1)
	assert.Equal(t, 2, competitors[0].DNFCount)
	assert.Equal(t, 2, competitors[0].MissingDNFCount)
	assert.Equal(t, 800, competitors[0].UsedCs, "az ismeretlen DNF idő nem növelheti a felhasználtat")
}

func TestBuildCompetitors_MatchesRecordedTimePerAttempt(t *testing.T) {
	competitors := BuildCompetitors(
		[]ResultSnapshot{
			snapshot(101, 7, "Teszt Elek", "333", 1, []int{AttemptDNF, AttemptDNF}),
		},
		[]RecordedTime{recorded(101, 2, 5000)},
	)

	require.Len(t, competitors, 1)
	round := competitors[0].Rounds[0]

	assert.Nil(t, round.Attempts[0].ElapsedCs)
	require.NotNil(t, round.Attempts[1].ElapsedCs)
	assert.Equal(t, 5000, *round.Attempts[1].ElapsedCs)
	assert.Equal(t, 1, competitors[0].MissingDNFCount)
}

func TestBuildCompetitors_IgnoresRecordedTimeFromAnotherResult(t *testing.T) {
	competitors := BuildCompetitors(
		[]ResultSnapshot{snapshot(101, 7, "Teszt Elek", "333", 1, []int{AttemptDNF})},
		[]RecordedTime{recorded(999, 1, 5000)},
	)

	require.Len(t, competitors, 1)
	assert.Nil(t, competitors[0].Rounds[0].Attempts[0].ElapsedCs)
	assert.Equal(t, 1, competitors[0].MissingDNFCount)
}

func TestBuildCompetitors_SeparatesCompetitors(t *testing.T) {
	competitors := BuildCompetitors([]ResultSnapshot{
		snapshot(102, 2, "Második", "333", 1, []int{900, AttemptDNF}),
		snapshot(101, 1, "Első", "333", 1, []int{AttemptDNF, 800}),
		snapshot(103, 1, "Első", "222", 1, []int{300, 350}),
	}, nil)

	require.Len(t, competitors, 2)

	assert.Equal(t, 1, competitors[0].RegistrantID)
	assert.Equal(t, 2, competitors[1].RegistrantID)

	require.Len(t, competitors[0].Rounds, 2)
	assert.Equal(t, "222-r1", competitors[0].Rounds[0].RoundWCIFID)
	assert.Equal(t, "333-r1", competitors[0].Rounds[1].RoundWCIFID)

	require.Len(t, competitors[1].Rounds, 1)
	assert.Equal(t, 1, competitors[1].DNFCount)
}

func TestBuildCompetitors_CrossRoundCumulativeBudget(t *testing.T) {
	rounds := []string{"333bf-r1", "444bf-r1", "555bf-r1"}

	competitors := BuildCompetitors(
		[]ResultSnapshot{
			withCumulative(snapshot(101, 7, "Vak Vince", "333bf", 1, []int{6000, AttemptDNF, 5500}), 60000, rounds...),
			withCumulative(snapshot(102, 7, "Vak Vince", "444bf", 1, []int{12000, AttemptDNS, AttemptDNS}), 60000, rounds...),
			withCumulative(snapshot(103, 7, "Vak Vince", "555bf", 1, []int{20000, AttemptSkipped, AttemptSkipped}), 60000, rounds...),
		},
		[]RecordedTime{recorded(101, 2, 4000)},
	)

	require.Len(t, competitors, 1)
	competitor := competitors[0]

	require.Len(t, competitor.Budgets, 1, "a három kör egyetlen közös keretet oszt")
	budget := competitor.Budgets[0]

	assert.Equal(t, []string{"333bf-r1", "444bf-r1", "555bf-r1"}, budget.RoundWCIFIDs)
	assert.Equal(t, 60000, budget.LimitCs)
	assert.Equal(t, 6000+5500+12000+20000, budget.CompletedCs)
	assert.Equal(t, 4000, budget.RecordedDNFCs)
	assert.Equal(t, 47500, budget.UsedCs)
	assert.Equal(t, 12500, budget.RemainingCs)
	assert.Zero(t, budget.MissingDNFCount)
}

func TestBuildCompetitors_BudgetMarksEstimate(t *testing.T) {
	competitors := BuildCompetitors(
		[]ResultSnapshot{
			withCumulative(snapshot(101, 7, "Vak Vince", "333bf", 1, []int{6000, AttemptDNF}), 60000, "333bf-r1"),
		},
		nil,
	)

	require.Len(t, competitors[0].Budgets, 1)
	budget := competitors[0].Budgets[0]

	assert.Equal(t, 6000, budget.UsedCs)
	assert.Equal(t, 54000, budget.RemainingCs)
	assert.Equal(t, 1, budget.MissingDNFCount, "amíg van rögzítetlen DNF, a maradék felső becslés")
}

func TestBuildCompetitors_NoBudgetWithoutCumulativeLimit(t *testing.T) {
	competitors := BuildCompetitors(
		[]ResultSnapshot{withLimit(snapshot(101, 7, "Teszt Elek", "333", 1, []int{696, AttemptDNF}), 60000)},
		nil,
	)

	require.Len(t, competitors, 1)
	assert.Empty(t, competitors[0].Budgets)
}
