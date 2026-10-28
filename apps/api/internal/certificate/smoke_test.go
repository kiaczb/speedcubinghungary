package certificate

import (
	"context"
	"os"
	"strings"
	"testing"

	"speedcubinghungary/internal/auth"
	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/stats"
	"speedcubinghungary/internal/wcalive"

	"github.com/stretchr/testify/require"
)

func TestSmoke_RecipientsAgainstRealWCA(t *testing.T) {
	ids := os.Getenv("CERTIFICATE_SMOKE")
	if ids == "" {
		t.Skip("CERTIFICATE_SMOKE=<wcaId>[,<wcaId>…] nincs megadva")
	}

	competitions := competition.NewService(competition.NewWCARepository())
	live := stats.NewService(competitions, stats.NewLiveRepository(wcalive.DefaultURL), noGrant{})
	service := NewService(nil, competitions, live)

	for _, id := range strings.Split(ids, ",") {
		for _, mode := range []Mode{ModePodium, ModeCompetitors, ModeRegistrants} {
			list, err := service.GetRecipients(context.Background(), id, mode)
			require.NoError(t, err, "%s %s", id, mode)
			t.Logf("%s %s: forrás %s, %d címzett", id, mode, list.Source, len(list.Recipients))
			if mode == ModePodium {
				for _, event := range list.Events {
					t.Logf("  %s: %s", event.EventID, event.Status)
				}
				for _, r := range list.Recipients {
					t.Logf("  %s %d. %s (%s) %s %s", r.EventID, r.Place, r.Name, r.WCAID, r.Result, r.ResultType)
				}
			}
		}
	}
}

type noGrant struct{}

func (noGrant) WCAAccessToken(context.Context, int) (string, error) { return "", auth.ErrNoGrant }
