package allrounder

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmokeLive_AgainstRealWCALive(t *testing.T) {
	competitionWCAID := os.Getenv("WCA_LIVE_SMOKE")
	if competitionWCAID == "" {
		t.Skip("set WCA_LIVE_SMOKE=<wcaId> to run against the real WCA Live")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	client := NewLiveClient("")

	liveCompetitionID, err := client.ResolveCompetitionID(ctx, competitionWCAID)
	require.NoError(t, err)
	t.Logf("%s -> WCA Live #%d", competitionWCAID, liveCompetitionID)

	identities, err := client.OpenRounds(ctx, liveCompetitionID)
	require.NoError(t, err)
	require.NotEmpty(t, identities)
	t.Logf("megnyitott körök: %d", len(identities))

	round, err := client.GetRound(ctx, identities[0].LiveRoundID)
	require.NoError(t, err)

	converted := roundResultsFromLive(identities[0], *round)
	t.Logf("%s r%d: %d nyers sor -> %d értékelhető helyezés",
		converted.EventID, identities[0].RoundNumber, len(round.Results), len(converted.Placements))

	ranked := 0
	for _, result := range round.Results {
		if result.Ranking != nil {
			ranked++
		}
	}
	assert.NotZero(t, ranked)

	for i := 1; i < len(converted.Placements); i++ {
		assert.LessOrEqual(t, converted.Placements[i-1].Pos, converted.Placements[i].Pos)
	}
}
