package weeklyreport

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func snapshotPeople(repo *fakeRepo) map[string]Person {
	out := map[string]Person{}
	for id, p := range repo.people {
		out[id] = *p
	}
	return out
}

func TestDryRun_WritesNothingButStillEmails(t *testing.T) {
	wca, repo := standardWeek()
	wca.personStats["2015REKA01"] = PersonStats{CompetitionCount: 97, TotalSolves: 4800}
	wca.upcoming["2015REKA01"] = []wcaUpcomingCompetition{{Name: "Later Open", StartDate: "2026-09-01", EndDate: "2026-09-02"}}
	before := snapshotPeople(repo)
	mailer := &fakeMailer{}

	s := NewService(NewDryRunRepository(repo), wca, WithSubjectPrefix(mailer, "[TESZT] "))
	s.now = newTestService(wca, repo, mailer).now
	require.NoError(t, s.Run(context.Background(), Options{DryRun: true}))

	assert.Empty(t, repo.posts, "no post")
	assert.Empty(t, repo.saved, "no records")
	assert.Equal(t, before, snapshotPeople(repo), "no counters")

	assert.Equal(t, 1, mailer.calls)
	assert.Equal(t, "[TESZT] 2026-07-27 Weekly Hungarian competitors dump", mailer.subject)
	assert.Contains(t, mailer.body, "Rekorder Réka")
	assert.Contains(t, mailer.body, "98 verseny", "milestones come from the computed, unsaved counts")
	assert.Contains(t, mailer.body, "Delegált Dénes", "the delegate milestone uses the computed 50")
}

func TestDryRun_CountsOnlyWritesNothing(t *testing.T) {
	wca, repo := standardWeek()
	before := snapshotPeople(repo)

	s := NewService(NewDryRunRepository(repo), wca, &fakeMailer{})
	require.NoError(t, s.Run(context.Background(), Options{CountsOnly: true, DryRun: true}))

	assert.Equal(t, before, snapshotPeople(repo))
	assert.Empty(t, repo.saved)
}
