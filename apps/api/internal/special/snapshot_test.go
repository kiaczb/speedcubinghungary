package special

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func ptr(value int) *int { return &value }

func attempts(values ...int) []liveAttempt {
	out := make([]liveAttempt, 0, len(values))
	for _, value := range values {
		out = append(out, liveAttempt{Result: value})
	}
	return out
}

func identity333r1() RoundIdentity {
	return RoundIdentity{LiveRoundID: 147464, EventID: "333", RoundNumber: 1}
}

func TestRoundWCIFID(t *testing.T) {
	assert.Equal(t, "333-r1", RoundWCIFID("333", 1))
	assert.Equal(t, "333bf-r2", RoundWCIFID("333bf", 2))
}

func TestSnapshotResults(t *testing.T) {
	best, average := 517, 634

	round := liveRound{
		TimeLimit: &liveTimeLimit{Centiseconds: 60000, CumulativeRoundWCIFIDs: []string{}},
		Results: []liveResult{
			{
				ID:       "101",
				Best:     &best,
				Average:  &average,
				Attempts: attempts(696, AttemptDNF, 674),
				Person:   livePerson{RegistrantID: ptr(7), WCAID: "2020TEST01", Name: "Teszt Elek"},
			},
			{
				ID:       "102",
				Attempts: nil,
				Person:   livePerson{RegistrantID: ptr(8), Name: "Még Nem Volt"},
			},
			{
				ID:       "103",
				Attempts: attempts(500),
				Person:   livePerson{Name: "Nincs rajtszám"},
			},
		},
	}

	snapshots := snapshotResults(identity333r1(), round)

	require.Len(t, snapshots, 1)
	assert.Equal(t, 101, snapshots[0].LiveResultID)
	assert.Equal(t, "333-r1", snapshots[0].RoundWCIFID)
	assert.Equal(t, "333", snapshots[0].EventID)
	assert.Equal(t, 1, snapshots[0].RoundNumber)
	assert.Equal(t, 7, snapshots[0].RegistrantID)
	assert.Equal(t, []int{696, AttemptDNF, 674}, snapshots[0].Attempts)
	require.NotNil(t, snapshots[0].Best)
	assert.Equal(t, 517, *snapshots[0].Best)
	require.NotNil(t, snapshots[0].TimeLimitCs)
	assert.Equal(t, 60000, *snapshots[0].TimeLimitCs)
}

func TestSnapshotResults_IsDeterministic(t *testing.T) {
	round := liveRound{
		TimeLimit: &liveTimeLimit{Centiseconds: 60000},
		Results: []liveResult{{
			ID:       "101",
			Attempts: attempts(696, AttemptDNF),
			Person:   livePerson{RegistrantID: ptr(7), Name: "Teszt Elek"},
		}},
	}

	assert.Equal(t, snapshotResults(identity333r1(), round), snapshotResults(identity333r1(), round))
}

func TestAttemptBound_NoLimit(t *testing.T) {
	assert.Nil(t, attemptBound(nil, nil, []int{AttemptDNF}, 0))
}

func TestAttemptBound_PerAttemptLimit(t *testing.T) {
	bound := attemptBound(ptr(60000), nil, []int{1000, AttemptDNF}, 1)

	require.NotNil(t, bound)
	assert.Equal(t, 60000, *bound)
}

func TestAttemptBound_CumulativeSubtractsEarlierAttempts(t *testing.T) {
	bound := attemptBound(ptr(60000), []string{"333bf-r1"}, []int{20000, 15000, AttemptDNF}, 2)

	require.NotNil(t, bound)
	assert.Equal(t, 25000, *bound)
}

func TestAttemptBound_CumulativeIgnoresEarlierDNF(t *testing.T) {
	bound := attemptBound(ptr(60000), []string{"333bf-r1"}, []int{AttemptDNF, AttemptDNF}, 1)

	require.NotNil(t, bound)
	assert.Equal(t, 60000, *bound)
}

func TestAttemptBound_CumulativeClampsAtZero(t *testing.T) {
	bound := attemptBound(ptr(10000), []string{"333bf-r1"}, []int{9000, 6000, AttemptDNF}, 2)

	require.NotNil(t, bound)
	assert.Zero(t, *bound)
}
