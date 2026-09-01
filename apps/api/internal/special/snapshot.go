package special

import (
	"fmt"
	"strconv"
)

type RoundIdentity struct {
	LiveRoundID int
	EventID     string
	RoundNumber int
}

func RoundWCIFID(eventID string, roundNumber int) string {
	return fmt.Sprintf("%s-r%d", eventID, roundNumber)
}

func snapshotResults(id RoundIdentity, round liveRound) []ResultSnapshot {
	roundWCIF := RoundWCIFID(id.EventID, id.RoundNumber)

	var limitCs *int
	cumulative := []string{}
	if round.TimeLimit != nil {
		cs := round.TimeLimit.Centiseconds
		limitCs = &cs
		if round.TimeLimit.CumulativeRoundWCIFIDs != nil {
			cumulative = round.TimeLimit.CumulativeRoundWCIFIDs
		}
	}

	snapshots := make([]ResultSnapshot, 0, len(round.Results))

	for _, result := range round.Results {
		if result.Person.RegistrantID == nil {
			continue
		}
		resultID, err := strconv.Atoi(result.ID)
		if err != nil {
			continue
		}
		if len(result.Attempts) == 0 {
			continue
		}

		attempts := make([]int, 0, len(result.Attempts))
		for _, attempt := range result.Attempts {
			attempts = append(attempts, attempt.Result)
		}

		snapshots = append(snapshots, ResultSnapshot{
			LiveResultID:       resultID,
			EventID:            id.EventID,
			RoundNumber:        id.RoundNumber,
			RoundWCIFID:        roundWCIF,
			RegistrantID:       *result.Person.RegistrantID,
			CompetitorWCAID:    result.Person.WCAID,
			CompetitorName:     result.Person.Name,
			Attempts:           attempts,
			Best:               result.Best,
			Average:            result.Average,
			TimeLimitCs:        limitCs,
			CumulativeRoundIDs: cumulative,
		})
	}

	return snapshots
}

func attemptBound(limitCs *int, cumulativeRoundIDs []string, attempts []int, idx int) *int {
	if limitCs == nil {
		return nil
	}

	if len(cumulativeRoundIDs) == 0 {
		cs := *limitCs
		return &cs
	}

	used := 0
	for j := 0; j < idx && j < len(attempts); j++ {
		if attempts[j] > 0 {
			used += attempts[j]
		}
	}

	remaining := *limitCs - used
	if remaining < 0 {
		remaining = 0
	}
	return &remaining
}
