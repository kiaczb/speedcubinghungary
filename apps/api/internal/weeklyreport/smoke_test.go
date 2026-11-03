package weeklyreport

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSmoke_AgainstRealWCA(t *testing.T) {
	competitionID := os.Getenv("WEEKLY_REPORT_SMOKE")
	if competitionID == "" {
		t.Skip("set WEEKLY_REPORT_SMOKE=<competition id> to run against the real WCA API")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	wca := NewWCAClient()
	repo := newFakeRepo()

	records, err := wca.GetRecords(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, records.WorldRecords["333"].Single)
	assert.NotEmpty(t, records.ContinentalRecords["_Europe"])
	assert.NotEmpty(t, records.NationalRecords["Hungary"])

	listed, err := wca.ListCompetitions(ctx, time.Now().Add(-reportPeriod), time.Now())
	require.NoError(t, err)
	t.Logf("%d competitions in the last week", len(listed))

	s := NewService(repo, wca, &fakeMailer{})
	tally := newUnpostedTally()
	competition, ok := s.processCompetition(ctx, wcaCompetition{
		ID: competitionID, Name: competitionID, CountryISO2: "HU",
		StartDate: "2026-01-01", EndDate: "2026-01-01",
	}, newRecordBook(RecordTable{}), tally)
	require.True(t, ok, "no Hungarian competitor with a result")
	t.Logf("%d Hungarian-interest competitors, %d tallied as unposted", len(competition.Competitors), len(tally.names))

	ids := make([]string, 0)
	for _, c := range competition.Competitors {
		if c.WCAID != "" {
			ids = append(ids, c.WCAID)
		}
	}
	require.NotEmpty(t, ids)
	stats, err := wca.GetPersonStats(ctx, ids)
	require.NoError(t, err)
	assert.Len(t, stats, len(ids))
	assert.Greater(t, stats[ids[0]].CompetitionCount, 0)
	assert.Greater(t, stats[ids[0]].TotalSolves, 0)

	delegated, err := wca.CountDelegatedCompetitions(ctx, "2008BARA01", time.Now())
	require.NoError(t, err)
	t.Logf("2008BARA01 delegated %d competitions so far", delegated)
	assert.Greater(t, delegated, 50)

	_, err = wca.GetUpcomingCompetitions(ctx, ids[0])
	require.NoError(t, err)

	name, _, _ := s.estimateMilestoneCompetition(ctx, ids[0], 98)
	assert.False(t, strings.Contains(name, "nem elérhető"), name)
}
