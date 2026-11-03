package weeklyreport

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeWCA struct {
	competitions   []wcaCompetition
	listErr        error
	wcifs          map[string]wcaWCIF
	results        map[string][]wcaResult
	upcoming       map[string][]wcaUpcomingCompetition
	personStats    map[string]PersonStats
	personStatsErr error
	delegated      map[string]int
	records        wcaRecords
	recordsErr     error
	windowStart    time.Time
	mu             sync.Mutex
	wcifRequests   map[string]int
}

func (f *fakeWCA) ListCompetitions(_ context.Context, start, end time.Time) ([]wcaCompetition, error) {
	if f.listErr != nil {
		return nil, f.listErr
	}
	if end.Sub(start) == reportPeriod {
		f.windowStart = start
	}
	from, to := start.Format(time.DateOnly), end.Format(time.DateOnly)
	out := make([]wcaCompetition, 0)
	for _, c := range f.competitions {
		if c.StartDate >= from && c.EndDate <= to {
			out = append(out, c)
		}
	}
	return out, nil
}

func (f *fakeWCA) GetWCIF(_ context.Context, id string) (wcaWCIF, error) {
	f.mu.Lock()
	if f.wcifRequests == nil {
		f.wcifRequests = map[string]int{}
	}
	f.wcifRequests[id]++
	f.mu.Unlock()
	wcif, ok := f.wcifs[id]
	if !ok {
		return wcaWCIF{}, errors.New("upstream down")
	}
	return wcif, nil
}

func (f *fakeWCA) GetResults(_ context.Context, id string) ([]wcaResult, error) {
	results, ok := f.results[id]
	if !ok {
		return nil, ErrNotFound
	}
	return results, nil
}

func (f *fakeWCA) GetUpcomingCompetitions(_ context.Context, wcaID string) ([]wcaUpcomingCompetition, error) {
	upcoming, ok := f.upcoming[wcaID]
	if !ok {
		return nil, errors.New("upstream down")
	}
	return upcoming, nil
}

func (f *fakeWCA) GetPersonStats(_ context.Context, wcaIDs []string) (map[string]PersonStats, error) {
	if f.personStatsErr != nil {
		return nil, f.personStatsErr
	}
	stats := map[string]PersonStats{}
	for _, id := range wcaIDs {
		if s, ok := f.personStats[id]; ok {
			stats[id] = s
		}
	}
	return stats, nil
}

func (f *fakeWCA) CountDelegatedCompetitions(_ context.Context, wcaID string, _ time.Time) (int, error) {
	n, ok := f.delegated[wcaID]
	if !ok {
		return 0, errors.New("upstream down")
	}
	return n, nil
}

func (f *fakeWCA) GetRecords(context.Context) (wcaRecords, error) {
	return f.records, f.recordsErr
}

type fakeRepo struct {
	mu        sync.Mutex
	records   RecordTable
	people    map[string]*Person
	saved     map[string]RecordValues
	posts     []ReportPost
	replaced  []bool
	schemaErr error
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{records: sampleRecords(), people: map[string]*Person{}, saved: map[string]RecordValues{}}
}

func (r *fakeRepo) CheckSchema() error { return r.schemaErr }

func (r *fakeRepo) GetRecords() (RecordTable, error) { return r.records, nil }

func (r *fakeRepo) UpsertRecord(region, eventID string, values RecordValues) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.saved[region+"/"+eventID] = values
	return nil
}

func (r *fakeRepo) person(wcaID, name string) *Person {
	p, ok := r.people[wcaID]
	if !ok {
		p = &Person{WCAID: wcaID}
		r.people[wcaID] = p
	}
	p.Name = name
	return p
}

func (r *fakeRepo) TrackedPeople() ([]Person, error) {
	out := make([]Person, 0)
	for _, p := range r.people {
		if p.CompetitionCount > 0 || p.DelegatedCount > 0 {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (r *fakeRepo) SaveCounts(counts []PersonCounts) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range counts {
		p := r.person(c.WCAID, c.Name)
		p.CompetitionCount = c.CompetitionCount
		p.SolveCount = c.SolveCount
	}
	return nil
}

func (r *fakeRepo) SaveDelegatedCounts(counts []DelegatedCount) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, c := range counts {
		r.person(c.WCAID, c.Name).DelegatedCount = c.Count
	}
	return nil
}

func (r *fakeRepo) near(value func(Person) int, milestones []int, difference int) []Person {
	out := make([]Person, 0)
	for _, p := range r.people {
		for _, m := range milestones {
			if v := value(*p); v >= m-difference && v <= m {
				out = append(out, *p)
				break
			}
		}
	}
	return out
}

func (r *fakeRepo) CompetitorsNearMilestones(milestones []int, difference int) ([]Person, error) {
	return r.near(func(p Person) int { return p.CompetitionCount }, milestones, difference), nil
}

func (r *fakeRepo) DelegatesNearMilestones(milestones []int, difference int) ([]Person, error) {
	return r.near(func(p Person) int { return p.DelegatedCount }, milestones, difference), nil
}

func (r *fakeRepo) SolversNearMilestones(step, difference int) ([]Person, error) {
	out := make([]Person, 0)
	for _, p := range r.people {
		rem := p.SolveCount % step
		if p.SolveCount > 0 && ((rem <= difference && p.SolveCount >= step) || rem >= step-difference) {
			out = append(out, *p)
		}
	}
	return out, nil
}

func (r *fakeRepo) PublishReport(post ReportPost, replace bool) (bool, error) {
	r.posts = append(r.posts, post)
	r.replaced = append(r.replaced, replace)
	return true, nil
}

type fakeMailer struct {
	subject string
	body    string
	calls   int
}

func (m *fakeMailer) Send(subject, body string) error {
	m.subject, m.body = subject, body
	m.calls++
	return nil
}

func registrantID(v int) *int { return &v }

func strPtr(s string) *string { return &s }

func person(id int, name, wcaID, country string, events ...string) wcaWCIFPerson {
	return wcaWCIFPerson{
		Name:         name,
		WCAID:        wcaID,
		RegistrantID: registrantID(id),
		CountryISO2:  country,
		Registration: &wcaWCIFRegistration{EventIDs: events},
	}
}

func wcifResult(personID, best, average int, attempts ...int) wcaWCIFResult {
	r := wcaWCIFResult{PersonID: personID, Best: best, Average: average}
	for _, a := range attempts {
		r.Attempts = append(r.Attempts, wcaWCIFAttempt{Result: a})
	}
	return r
}

func newTestService(wca *fakeWCA, repo *fakeRepo, mailer *fakeMailer) *Service {
	s := NewService(repo, wca, mailer)
	s.now = func() time.Time { return time.Date(2026, 7, 27, 8, 0, 0, 0, time.UTC) }
	return s
}

func standardWeek() (*fakeWCA, *fakeRepo) {
	wca := &fakeWCA{
		competitions: []wcaCompetition{
			{ID: "OldOpen2026", Name: "Old Open 2026", CountryISO2: "SK", StartDate: "2026-07-04", EndDate: "2026-07-05"},
			{ID: "PostedOpen2026", Name: "Posted Open 2026", CountryISO2: "HU", StartDate: "2026-07-11", EndDate: "2026-07-11", ResultsPostedAt: strPtr("2026-07-13T10:00:00Z")},
			{ID: "ViennaOpen2026", Name: "Vienna Open 2026", CountryISO2: "AT", StartDate: "2026-07-25", EndDate: "2026-07-26"},
			{
				ID: "MiskolcOpen2026", Name: "Miskolc Open 2026", CountryISO2: "HU", StartDate: "2026-07-25", EndDate: "2026-07-25",
				ResultsPostedAt: strPtr("2026-07-26T10:00:00Z"),
				Delegates: []wcaUser{
					{WCAID: "2010DENE01", Name: "Delegált Dénes", CountryISO2: "HU"},
					{WCAID: "2012UJDE01", Name: "Új Delegált", CountryISO2: "HU"},
					{WCAID: "2009FORE01", Name: "Foreign Delegate", CountryISO2: "AT"},
				},
			},
		},
		wcifs: map[string]wcaWCIF{
			"MiskolcOpen2026": {
				Persons: []wcaWCIFPerson{
					person(1, "Rekorder Réka", "2015REKA01", "HU", "333"),
					person(2, "Résztvevő Rezső", "2016REZS01", "HU", "333"),
					person(3, "Foreign Fred", "2014FRED01", "AT", "333"),
					person(4, "Delegált Dénes", "2010DENE01", "HU", "333"),
					person(5, "Nem Jelent Meg", "2019NEMJ01", "HU", "333"),
				},
				Events: []wcaWCIFEvent{{ID: "333", Rounds: []wcaWCIFRound{{Results: []wcaWCIFResult{
					wcifResult(1, 490, 590, 490, 600, 610, 580, 590),
					wcifResult(2, 900, 1000, 900, -1, 1000, 1100, 1000),
					wcifResult(3, 100, 200, 100, 200, 300, 200, 200),
					wcifResult(4, 1200, 1300, 1200, 1300, 1400, 1300, 1300),
					wcifResult(5, 0, 0, 0, 0, 0, 0, 0),
				}}}}},
			},
			"ViennaOpen2026": {
				Persons: []wcaWCIFPerson{
					person(7, "Külföldön Kálmán", "2018KALM01", "HU", "222"),
					person(8, "Delegált Dénes", "2010DENE01", "HU", "222"),
				},
				Events: []wcaWCIFEvent{{ID: "222", Rounds: []wcaWCIFRound{{Results: []wcaWCIFResult{
					wcifResult(7, 300, 400, 300, 400, -1, 500, 400),
					wcifResult(8, 350, 450, 350, 450, 500, 450, 450),
				}}}}},
			},
			"OldOpen2026": {
				Persons: []wcaWCIFPerson{person(1, "Rekorder Réka", "2015REKA01", "HU", "333")},
				Events: []wcaWCIFEvent{{ID: "333", Rounds: []wcaWCIFRound{{Results: []wcaWCIFResult{
					wcifResult(1, 600, 700, 600, 700, 800, -1, 700),
				}}}}},
			},
		},
		upcoming: map[string][]wcaUpcomingCompetition{},
		personStats: map[string]PersonStats{
			"2015REKA01": {CompetitionCount: 10, TotalSolves: 4800},
			"2016REZS01": {CompetitionCount: 3, TotalSolves: 100},
			"2010DENE01": {CompetitionCount: 150, TotalSolves: 9990},
			"2018KALM01": {CompetitionCount: 5, TotalSolves: 50},
			"2019SIMO07": {CompetitionCount: 38, TotalSolves: 2000},
		},
		delegated: map[string]int{"2010DENE01": 50, "2012UJDE01": 3},
		records: wcaRecords{
			WorldRecords:       map[string]RecordValues{"333": {Single: intPtr(300)}},
			ContinentalRecords: map[string]map[string]RecordValues{"_Europe": {"333": {Single: intPtr(310)}}},
			NationalRecords:    map[string]map[string]RecordValues{"Hungary": {"333": {Single: intPtr(490)}}},
		},
	}
	repo := newFakeRepo()
	repo.people["2010DENE01"] = &Person{WCAID: "2010DENE01", Name: "Delegált Dénes", CompetitionCount: 152, DelegatedCount: 48}
	repo.people["2019SIMO07"] = &Person{WCAID: "2019SIMO07", Name: "Simon Ádám", CompetitionCount: 8}
	return wca, repo
}

func TestRun_PublishesPostWithEveryHungarianAndEmailsOnlyRecords(t *testing.T) {
	wca, repo := standardWeek()
	mailer := &fakeMailer{}

	require.NoError(t, newTestService(wca, repo, mailer).Run(context.Background(), Options{}))

	require.Len(t, repo.posts, 1)
	post := repo.posts[0]
	assert.Equal(t, time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC), post.ReportWeek)
	assert.Equal(t, "Heti riport – 2026. 07. 20. – 2026. 07. 27.", post.Title)
	assert.False(t, repo.replaced[0])
	assert.Equal(t, time.Date(2026, 7, 20, 8, 0, 0, 0, time.UTC), wca.windowStart)

	var payload struct {
		Competitions []struct {
			Name        string `json:"name"`
			Competitors []struct {
				Name    string `json:"name"`
				Results []struct {
					Event, Type, Result, Badge string
				} `json:"results"`
			} `json:"competitors"`
		} `json:"competitions"`
	}
	require.NoError(t, json.Unmarshal([]byte(post.Content), &payload))
	require.Len(t, payload.Competitions, 2, "only this week's competitions are reported")
	assert.Equal(t, "Miskolc Open 2026", payload.Competitions[0].Name)
	assert.Equal(t, "Vienna Open 2026", payload.Competitions[1].Name)

	miskolc := payload.Competitions[0].Competitors
	require.Len(t, miskolc, 3, "nobody without a single attempt")
	assert.Equal(t, []string{"Rekorder Réka", "Résztvevő Rezső", "Delegált Dénes"}, []string{miskolc[0].Name, miskolc[1].Name, miskolc[2].Name})
	assert.Equal(t, "average", miskolc[0].Results[0].Type)
	assert.Equal(t, "5.90", miskolc[0].Results[0].Result)
	assert.Equal(t, "NR", miskolc[0].Results[1].Badge)
	assert.Empty(t, miskolc[1].Results)

	assert.Equal(t, 1, mailer.calls)
	assert.Equal(t, "2026-07-27 Weekly Hungarian competitors dump", mailer.subject)
	assert.Contains(t, mailer.body, "Rekorder Réka")
	assert.NotContains(t, mailer.body, "Résztvevő Rezső")
	assert.Contains(t, mailer.body, "Külföldön Kálmán")
	assert.NotContains(t, mailer.body, "Foreign Fred")

	assert.Equal(t, RecordValues{Single: intPtr(490)}, repo.saved["hungary/333"])
}

func TestRun_CountsAreOfficialPlusUnpostedCompetitions(t *testing.T) {
	wca, repo := standardWeek()

	require.NoError(t, newTestService(wca, repo, &fakeMailer{}).Run(context.Background(), Options{}))

	reka := repo.people["2015REKA01"]
	assert.Equal(t, 11, reka.CompetitionCount, "10 official + the unposted Old Open; Miskolc is already posted")
	assert.Equal(t, 4804, reka.SolveCount, "4800 official + 4 successful attempts at Old Open")

	denes := repo.people["2010DENE01"]
	assert.Equal(t, 151, denes.CompetitionCount, "the incremented 152 is replaced by 150 official + unposted Vienna")
	assert.Equal(t, 9995, denes.SolveCount)
	assert.Equal(t, 50, denes.DelegatedCount, "from the WCA, not incremented")

	assert.Equal(t, 6, repo.people["2018KALM01"].CompetitionCount)
	assert.Equal(t, 54, repo.people["2018KALM01"].SolveCount)
	assert.Equal(t, 38, repo.people["2019SIMO07"].CompetitionCount, "a stale count is corrected even without competing this week")
	assert.Equal(t, 3, repo.people["2012UJDE01"].DelegatedCount, "a Hungarian delegate new to the DB is added")

	assert.NotContains(t, repo.people, "2009FORE01")
	assert.NotContains(t, repo.people, "2014FRED01")
	assert.NotContains(t, repo.people, "2019NEMJ01")

	assert.Equal(t, 1, wca.wcifRequests["ViennaOpen2026"], "a WCIF downloaded for the report is reused")
	assert.Zero(t, wca.wcifRequests["PostedOpen2026"], "posted competitions outside the week are not downloaded")
}

func TestRun_CountsOnlyUpdatesCountersWithoutReporting(t *testing.T) {
	wca, repo := standardWeek()
	mailer := &fakeMailer{}

	require.NoError(t, newTestService(wca, repo, mailer).Run(context.Background(), Options{CountsOnly: true}))

	assert.Empty(t, repo.posts)
	assert.Zero(t, mailer.calls)
	assert.Equal(t, RecordValues{Single: intPtr(490)}, repo.saved["hungary/333"], "records are refreshed from the WCA too")
	assert.Equal(t, 151, repo.people["2010DENE01"].CompetitionCount)
	assert.Equal(t, 50, repo.people["2010DENE01"].DelegatedCount)
	assert.Equal(t, 38, repo.people["2019SIMO07"].CompetitionCount)
	assert.Equal(t, 11, repo.people["2015REKA01"].CompetitionCount, "found through the unposted Old Open")
	assert.NotContains(t, repo.people, "2016REZS01", "people only seen in posted results are added by a full run")
}

func TestRun_FailedStatsLookupKeepsCountsButStillReports(t *testing.T) {
	wca, repo := standardWeek()
	wca.personStatsErr = errors.New("upstream down")
	mailer := &fakeMailer{}

	err := newTestService(wca, repo, mailer).Run(context.Background(), Options{})

	assert.Error(t, err)
	assert.Len(t, repo.posts, 1)
	assert.Equal(t, 1, mailer.calls)
	assert.Equal(t, 152, repo.people["2010DENE01"].CompetitionCount)
	assert.Equal(t, 8, repo.people["2019SIMO07"].CompetitionCount)
}

func TestRun_FailedDelegateLookupSkipsOnlyThatDelegate(t *testing.T) {
	wca, repo := standardWeek()
	delete(wca.delegated, "2010DENE01")

	require.NoError(t, newTestService(wca, repo, &fakeMailer{}).Run(context.Background(), Options{}))

	assert.Equal(t, 48, repo.people["2010DENE01"].DelegatedCount)
	assert.Equal(t, 3, repo.people["2012UJDE01"].DelegatedCount)
}

func TestRun_ReportsMilestonesFromTheUpdatedCounts(t *testing.T) {
	wca, repo := standardWeek()
	wca.personStats["2015REKA01"] = PersonStats{CompetitionCount: 97, TotalSolves: 4800}
	wca.upcoming["2015REKA01"] = []wcaUpcomingCompetition{
		{Name: "Later Open", StartDate: "2026-09-01", EndDate: "2026-09-02"},
		{Name: "Budapest Open", StartDate: "2026-08-15", EndDate: "2026-08-15"},
	}

	require.NoError(t, newTestService(wca, repo, &fakeMailer{}).Run(context.Background(), Options{}))

	var payload struct {
		CompetitorMilestones []struct {
			Name, Estimated, EstimatedDate string
			Count                          int
		} `json:"competitorMilestones"`
		DelegateMilestones []struct {
			Name  string
			Count int
		} `json:"delegateMilestones"`
		SolveMilestones []struct {
			Name      string
			Count     int
			Milestone int
			Remaining int
			Achieved  bool
		} `json:"solveMilestones"`
	}
	require.NoError(t, json.Unmarshal([]byte(repo.posts[0].Content), &payload))

	require.Len(t, payload.CompetitorMilestones, 1)
	assert.Equal(t, 98, payload.CompetitorMilestones[0].Count)
	assert.Equal(t, "Later Open", payload.CompetitorMilestones[0].Estimated)
	assert.Equal(t, "2026. Sep. 01-02.", payload.CompetitorMilestones[0].EstimatedDate)

	require.Len(t, payload.DelegateMilestones, 1)
	assert.Equal(t, 50, payload.DelegateMilestones[0].Count)

	solves := map[string]int{}
	for _, m := range payload.SolveMilestones {
		solves[m.Name] = m.Remaining
	}
	assert.Equal(t, map[string]int{"Rekorder Réka": 196, "Delegált Dénes": 5}, solves)
}

func TestRun_UsesPostedResultsWhenWCIFHasNone(t *testing.T) {
	wca, repo := standardWeek()
	wcif := wca.wcifs["MiskolcOpen2026"]
	wcif.Events = []wcaWCIFEvent{{ID: "333"}}
	wcif.Persons = append(wcif.Persons, person(9, "Új Újonc", "", "HU", "333"))
	wca.wcifs["MiskolcOpen2026"] = wcif
	wca.results = map[string][]wcaResult{"MiskolcOpen2026": {
		{WCAID: "2015REKA01", Name: "Rekorder Réka", EventID: "333", Best: 480, Average: 700, Attempts: []int{480, 700, 700, 800, 700}},
		{WCAID: "2026UJUJ01", Name: "Új Újonc", EventID: "333", Best: 2000, Average: -1, Attempts: []int{2000, -1, -1}},
	}}

	require.NoError(t, newTestService(wca, repo, &fakeMailer{}).Run(context.Background(), Options{}))

	var payload struct {
		Competitions []struct {
			Competitors []struct {
				Name    string
				Results []struct{ Badge, Result string }
			}
		}
	}
	require.NoError(t, json.Unmarshal([]byte(repo.posts[0].Content), &payload))
	miskolc := payload.Competitions[0].Competitors
	require.Len(t, miskolc, 2)
	assert.Equal(t, "Rekorder Réka", miskolc[0].Name)
	assert.Equal(t, "4.80", miskolc[0].Results[0].Result)
	assert.Equal(t, "Új Újonc", miskolc[1].Name)
}

func TestRun_SkipsCompetitionWhoseWCIFIsUnavailable(t *testing.T) {
	wca, repo := standardWeek()
	delete(wca.wcifs, "ViennaOpen2026")

	require.NoError(t, newTestService(wca, repo, &fakeMailer{}).Run(context.Background(), Options{}))

	assert.NotContains(t, repo.posts[0].Content, "Vienna Open 2026")
	assert.Contains(t, repo.posts[0].Content, "Miskolc Open 2026")
}

func TestRun_FailsWithoutPublishingWhenTheCompetitionListIsUnavailable(t *testing.T) {
	wca, repo := standardWeek()
	wca.listErr = errors.New("upstream down")
	mailer := &fakeMailer{}

	err := newTestService(wca, repo, mailer).Run(context.Background(), Options{})

	assert.Error(t, err)
	assert.Empty(t, repo.posts)
	assert.Zero(t, mailer.calls)
}

func TestRun_FailsBeforeAnythingWhenTheSchemaIsMissing(t *testing.T) {
	wca, repo := standardWeek()
	repo.schemaErr = errors.New("missing column")

	assert.Error(t, newTestService(wca, repo, &fakeMailer{}).Run(context.Background(), Options{}))
	assert.Empty(t, repo.posts)
	assert.Equal(t, 8, repo.people["2019SIMO07"].CompetitionCount)
}

func TestRun_SkipEmailAndForce(t *testing.T) {
	wca, repo := standardWeek()
	mailer := &fakeMailer{}

	require.NoError(t, newTestService(wca, repo, mailer).Run(context.Background(), Options{Force: true, SkipEmail: true}))

	assert.Zero(t, mailer.calls)
	assert.True(t, repo.replaced[0])
}

func TestRun_FailedRecordSaveStillSendsTheEmailButFails(t *testing.T) {
	wca, repo := standardWeek()
	wca.recordsErr = errors.New("upstream down")
	mailer := &fakeMailer{}

	err := newTestService(wca, repo, mailer).Run(context.Background(), Options{})

	assert.Error(t, err)
	assert.Len(t, repo.posts, 1)
	assert.Equal(t, 1, mailer.calls)
}

func TestRun_IsIdempotentForCounts(t *testing.T) {
	wca, repo := standardWeek()
	s := newTestService(wca, repo, &fakeMailer{})

	require.NoError(t, s.Run(context.Background(), Options{SkipEmail: true}))
	first := *repo.people["2010DENE01"]
	require.NoError(t, s.Run(context.Background(), Options{SkipEmail: true, Force: true}))

	assert.Equal(t, first, *repo.people["2010DENE01"])
}

func TestHasCompeted(t *testing.T) {
	assert.True(t, hasCompeted([]int{-1, -2, 0}))
	assert.True(t, hasCompeted([]int{0, 900}))
	assert.False(t, hasCompeted([]int{-2, -2, 0}))
	assert.False(t, hasCompeted(nil))
}

func TestEstimateMilestoneCompetition(t *testing.T) {
	upcoming := []wcaUpcomingCompetition{
		{Name: "C", StartDate: "2026-10-01", EndDate: "2026-10-01"},
		{Name: "A", StartDate: "2026-08-01", EndDate: "2026-08-02"},
		{Name: "B", StartDate: "2026-09-01", EndDate: "2026-09-01"},
	}
	wca := &fakeWCA{upcoming: map[string][]wcaUpcomingCompetition{"2015X": upcoming, "2015EMPTY": nil}}
	s := newTestService(wca, newFakeRepo(), &fakeMailer{})

	cases := []struct {
		name     string
		wcaID    string
		count    int
		want     string
		wantDate string
		achieved bool
	}{
		{"nincs WCA ID", "", 98, "Ismeretlen", "", false},
		{"a második közelgő versennyel éri el", "2015X", 98, "B", "2026. Sep. 01.", false},
		{"pont most érte el", "2015X", 100, "A", "2026. Aug. 01-02.", true},
		{"elérte, nincs közelgő verseny", "2015EMPTY", 200, "Elérte a mérföldkövet! 🎉", "", true},
		{"nincs elég verseny", "2015X", 196, "Nincs elég regisztrált verseny (még 4 kellene)", "", false},
		{"túl a maximumon", "2015X", 201, "Már elérte a maximumot", "", false},
		{"WCA hiba", "2015DOWN", 99, "WCA API nem elérhető", "", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			name, date, achieved := s.estimateMilestoneCompetition(context.Background(), tc.wcaID, tc.count)
			assert.Equal(t, tc.want, name)
			assert.Equal(t, tc.wantDate, date)
			assert.Equal(t, tc.achieved, achieved)
		})
	}
}

func TestSolveMilestone(t *testing.T) {
	cases := []struct {
		count, milestone, remaining int
		achieved                    bool
	}{
		{5100, 5000, 0, true},
		{5250, 5000, 0, true},
		{4800, 5000, 200, false},
		{9900, 10000, 100, false},
		{100, 5000, 4900, false},
	}
	for _, tc := range cases {
		m := solveMilestone(Person{SolveCount: tc.count})
		assert.Equal(t, tc.milestone, m.Milestone, tc.count)
		assert.Equal(t, tc.remaining, m.Remaining, tc.count)
		assert.Equal(t, tc.achieved, m.Achieved, tc.count)
	}
}
