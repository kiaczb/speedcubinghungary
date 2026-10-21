package stats

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/wcalive"

	"github.com/stretchr/testify/require"
)

func TestSmoke_ReportAgainstRealWCA(t *testing.T) {
	ids := os.Getenv("STATS_SMOKE")
	if ids == "" {
		t.Skip("STATS_SMOKE=<wcaId>[,<wcaId>…] nincs megadva")
	}

	service := NewService(competition.NewService(competition.NewWCARepository()), NewLiveRepository(wcalive.DefaultURL))
	waitForHistory := os.Getenv("STATS_SMOKE_HISTORY") != ""

	for _, id := range strings.Split(ids, ",") {
		started := time.Now()
		report, err := service.GetReport(context.Background(), id)
		require.NoError(t, err, id)
		t.Logf("%s: fázis %s, %v", id, report.Phase, time.Since(started).Round(time.Millisecond))

		if waitForHistory && report.History.Status == HistoryRunning {
			require.Eventually(t, func() bool {
				report, err = service.GetReport(context.Background(), id)
				require.NoError(t, err)
				t.Logf("  történet: %d/%d (hibás: %d)", report.History.Done, report.History.Total, report.History.Failed)
				return report.History.Status == HistoryReady
			}, 5*time.Minute, 5*time.Second)
		}

		out, err := json.MarshalIndent(report, "", "  ")
		require.NoError(t, err)
		t.Logf("%s", out)
	}
}
