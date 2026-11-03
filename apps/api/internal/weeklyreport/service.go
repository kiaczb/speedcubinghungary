package weeklyreport

import (
	"context"
	"errors"
	"fmt"
	"log"
	"slices"
	"sort"
	"sync"
	"time"
	_ "time/tzdata"

	"github.com/google/uuid"
)

var (
	competitionMilestones          = []int{100, 200}
	competitionMilestoneDifference = 3
	delegateMilestones             = []int{50, 100}
	delegateMilestoneDifference    = 3
	solveMilestoneStep             = 5000
	solveMilestoneDifference       = 250
)

const (
	competitionWorkers = 3
	reportPeriod       = 7 * 24 * time.Hour
)

type Mailer interface {
	Send(subject, htmlBody string) error
}

type Options struct {
	Force      bool
	SkipEmail  bool
	CountsOnly bool
	DryRun     bool
}

type Service struct {
	repo   Repository
	wca    WCA
	mailer Mailer
	now    func() time.Time
}

func NewService(repo Repository, wca WCA, mailer Mailer) *Service {
	budapest, err := time.LoadLocation("Europe/Budapest")
	if err != nil {
		budapest = time.UTC
	}
	return &Service{
		repo:   repo,
		wca:    wca,
		mailer: mailer,
		now:    func() time.Time { return time.Now().In(budapest) },
	}
}

func (s *Service) Run(ctx context.Context, opts Options) error {
	if err := s.repo.CheckSchema(); err != nil {
		return fmt.Errorf("database check: %w", err)
	}
	if opts.DryRun {
		log.Printf("heti riport: SZÁRAZ FUTÁS — az adatbázisba semmi nem íródik")
	}

	periodEnd := s.now()
	periodStart := periodEnd.Add(-reportPeriod)

	recent, err := s.wca.ListCompetitions(ctx, periodEnd.Add(-unpostedLookback), periodEnd)
	if err != nil {
		return fmt.Errorf("listing recent competitions: %w", err)
	}
	tally := newUnpostedTally()

	if opts.CountsOnly {
		s.tallyUnposted(ctx, recent, nil, periodEnd, tally)
		countsErr := s.syncCounts(ctx, periodEnd, nil, tally, hungarianDelegates(recent, periodEnd))
		return errors.Join(wrapErr("syncing counts", countsErr), wrapErr("saving records", s.saveRecords(ctx)))
	}

	records, err := s.repo.GetRecords()
	if err != nil {
		return fmt.Errorf("loading records: %w", err)
	}
	book := newRecordBook(records)

	started := time.Now()
	raw, err := s.wca.ListCompetitions(ctx, periodStart, periodEnd)
	if err != nil {
		return fmt.Errorf("listing competitions: %w", err)
	}
	log.Printf("heti riport: %d verseny %s és %s között", len(raw), periodStart.Format(time.DateOnly), periodEnd.Format(time.DateOnly))

	competitions := s.processCompetitions(ctx, raw, book, tally)
	log.Printf("heti riport: %d magyar érdekeltségű verseny feldolgozva %.1f s alatt", len(competitions), time.Since(started).Seconds())

	processed := make(map[string]bool, len(raw))
	for _, comp := range raw {
		processed[comp.ID] = true
	}
	s.tallyUnposted(ctx, recent, processed, periodEnd, tally)

	seen := map[string]string{}
	for _, comp := range competitions {
		for _, c := range comp.Competitors {
			if c.WCAID != "" {
				seen[c.WCAID] = c.Name
			}
		}
	}
	countsErr := s.syncCounts(ctx, periodEnd, seen, tally, hungarianDelegates(slices.Concat(recent, raw), periodEnd))
	if countsErr != nil {
		log.Printf("heti riport: a számlálók frissítése sikertelen, a mérföldkövek a régi értékekből készülnek: %v", countsErr)
	}

	data, err := s.buildReportData(ctx, competitions)
	if err != nil {
		return err
	}

	if err := s.publishPost(data, periodStart, periodEnd, opts.Force, opts.DryRun); err != nil {
		return fmt.Errorf("publishing report post: %w", err)
	}

	recordsErr := s.saveRecords(ctx)
	if recordsErr != nil {
		log.Printf("heti riport: a rekordok frissítése sikertelen: %v", recordsErr)
	}

	if opts.SkipEmail {
		log.Printf("heti riport: e-mail küldés kihagyva (-skip-email)")
	} else {
		if err := s.sendEmail(data, periodEnd); err != nil {
			return fmt.Errorf("sending email: %w", err)
		}
		log.Printf("heti riport: e-mail elküldve")
	}

	return errors.Join(wrapErr("syncing counts", countsErr), wrapErr("saving records", recordsErr))
}

func wrapErr(context string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", context, err)
}

func (s *Service) processCompetitions(ctx context.Context, raw []wcaCompetition, book *recordBook, tally *unpostedTally) []Competition {
	jobs := make(chan wcaCompetition)
	var mu sync.Mutex
	var wg sync.WaitGroup
	competitions := make([]Competition, 0)

	for range competitionWorkers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for comp := range jobs {
				competition, ok := s.processCompetition(ctx, comp, book, tally)
				if !ok {
					continue
				}
				mu.Lock()
				competitions = append(competitions, competition)
				mu.Unlock()
			}
		}()
	}
	for _, comp := range raw {
		jobs <- comp
	}
	close(jobs)
	wg.Wait()

	sort.SliceStable(competitions, func(i, j int) bool {
		if !competitions[i].From.Equal(competitions[j].From) {
			return competitions[i].From.Before(competitions[j].From)
		}
		return competitions[i].Name < competitions[j].Name
	})
	return competitions
}

func (s *Service) processCompetition(ctx context.Context, comp wcaCompetition, book *recordBook, tally *unpostedTally) (Competition, bool) {
	from, errFrom := time.Parse(time.DateOnly, comp.StartDate)
	to, errTo := time.Parse(time.DateOnly, comp.EndDate)
	if errFrom != nil || errTo != nil {
		log.Printf("heti riport: %s dátuma hibás (%q – %q), kihagyva", comp.ID, comp.StartDate, comp.EndDate)
		return Competition{}, false
	}

	wcif, err := s.wca.GetWCIF(ctx, comp.ID)
	if err != nil {
		log.Printf("heti riport: %s WCIF nem érhető el, kihagyva: %v", comp.ID, err)
		return Competition{}, false
	}
	if !comp.resultsPosted() {
		tally.addCompetition(wcif)
	}

	events := eventsFromWCIF(wcif.Events)
	results, err := s.wca.GetResults(ctx, comp.ID)
	switch {
	case err == nil && len(results) > 0:
		events = eventsFromResults(results, wcif.Persons)
		log.Printf("heti riport: %s: feltöltött eredmények (%d sor)", comp.ID, len(results))
	case err != nil && !errors.Is(err, ErrNotFound):
		log.Printf("heti riport: %s /results nem érhető el, a WCIF eredményei maradnak: %v", comp.ID, err)
	}

	competitors := make([]Competitor, 0)
	for _, person := range wcif.Persons {
		if c := processPerson(person, events, book); c != nil {
			competitors = append(competitors, *c)
		}
	}
	if len(competitors) == 0 {
		return Competition{}, false
	}
	log.Printf("heti riport: %s: %d magyar érdekeltségű versenyző", comp.ID, len(competitors))

	return Competition{
		ID:          comp.ID,
		Name:        comp.Name,
		CountryISO2: comp.CountryISO2,
		From:        from,
		To:          to,
		Hungarian:   comp.CountryISO2 == "HU",
		Competitors: competitors,
	}, true
}

func processPerson(person wcaWCIFPerson, events []eventResults, book *recordBook) *Competitor {
	if !isTrackedPerson(person.CountryISO2, person.WCAID) || person.Registration == nil || person.RegistrantID == nil {
		return nil
	}

	registered := make(map[string]bool, len(person.Registration.EventIDs))
	for _, id := range person.Registration.EventIDs {
		registered[id] = true
	}

	competitor := Competitor{Name: person.Name, WCAID: person.WCAID}
	competed := false

	for _, event := range events {
		if !registered[event.EventID] {
			continue
		}
		for _, result := range event.Results {
			if result.RegistrantID != *person.RegistrantID || !hasCompeted(result.Attempts) {
				continue
			}
			competed = true
			if badge, ok := book.check(person.WCAID, event.EventID, recordAverage, result.Average); ok {
				competitor.addRecord(event.EventID, Record{Type: recordAverage, Result: result.Average, Badge: badge})
			}
			if badge, ok := book.check(person.WCAID, event.EventID, recordSingle, result.Best); ok {
				competitor.addRecord(event.EventID, Record{Type: recordSingle, Result: result.Best, Badge: badge})
			}
		}
	}

	if !competed {
		return nil
	}
	return &competitor
}

func countSuccessfulAttempts(attempts []int) int {
	count := 0
	for _, a := range attempts {
		if a > 0 {
			count++
		}
	}
	return count
}

func eventsFromWCIF(events []wcaWCIFEvent) []eventResults {
	out := make([]eventResults, 0, len(events))
	for _, event := range events {
		er := eventResults{EventID: event.ID}
		for _, round := range event.Rounds {
			for _, r := range round.Results {
				attempts := make([]int, 0, len(r.Attempts))
				for _, a := range r.Attempts {
					attempts = append(attempts, a.Result)
				}
				er.Results = append(er.Results, personResult{
					RegistrantID: r.PersonID,
					Best:         r.Best,
					Average:      r.Average,
					Attempts:     attempts,
				})
			}
		}
		out = append(out, er)
	}
	return out
}

func eventsFromResults(results []wcaResult, persons []wcaWCIFPerson) []eventResults {
	byWCAID := map[string]int{}
	byName := map[string]int{}
	for _, p := range persons {
		if p.RegistrantID == nil {
			continue
		}
		if p.WCAID != "" {
			byWCAID[p.WCAID] = *p.RegistrantID
		}
		byName[p.Name] = *p.RegistrantID
	}

	index := map[string]int{}
	out := make([]eventResults, 0)
	for _, r := range results {
		i, ok := index[r.EventID]
		if !ok {
			i = len(out)
			index[r.EventID] = i
			out = append(out, eventResults{EventID: r.EventID})
		}

		registrantID, found := byWCAID[r.WCAID]
		if !found {
			registrantID, found = byName[r.Name]
		}
		if !found {
			continue
		}
		out[i].Results = append(out[i].Results, personResult{
			RegistrantID: registrantID,
			Best:         r.Best,
			Average:      r.Average,
			Attempts:     r.Attempts,
		})
	}
	return out
}

func (s *Service) buildReportData(ctx context.Context, competitions []Competition) (ReportData, error) {
	competitors, err := s.repo.CompetitorsNearMilestones(competitionMilestones, competitionMilestoneDifference)
	if err != nil {
		return ReportData{}, fmt.Errorf("reading competitor milestones: %w", err)
	}
	delegates, err := s.repo.DelegatesNearMilestones(delegateMilestones, delegateMilestoneDifference)
	if err != nil {
		return ReportData{}, fmt.Errorf("reading delegate milestones: %w", err)
	}
	solvers, err := s.repo.SolversNearMilestones(solveMilestoneStep, solveMilestoneDifference)
	if err != nil {
		return ReportData{}, fmt.Errorf("reading solve milestones: %w", err)
	}

	data := ReportData{
		Competitions:         competitions,
		CompetitorMilestones: make([]CompetitorMilestone, 0, len(competitors)),
		DelegateMilestones:   make([]DelegateMilestone, 0, len(delegates)),
		SolveMilestones:      make([]SolveMilestone, 0, len(solvers)),
	}
	for _, p := range competitors {
		name, date, achieved := s.estimateMilestoneCompetition(ctx, p.WCAID, p.CompetitionCount)
		data.CompetitorMilestones = append(data.CompetitorMilestones, CompetitorMilestone{
			WCAID:                p.WCAID,
			Name:                 p.Name,
			CompetitionCount:     p.CompetitionCount,
			EstimatedCompetition: name,
			EstimatedDate:        date,
			Achieved:             achieved,
		})
	}
	for _, p := range delegates {
		data.DelegateMilestones = append(data.DelegateMilestones, DelegateMilestone{
			WCAID:          p.WCAID,
			Name:           p.Name,
			DelegatedCount: p.DelegatedCount,
		})
	}
	for _, p := range solvers {
		data.SolveMilestones = append(data.SolveMilestones, solveMilestone(p))
	}
	return data, nil
}

func (s *Service) estimateMilestoneCompetition(ctx context.Context, wcaID string, count int) (string, string, bool) {
	if wcaID == "" {
		return "Ismeretlen", "", false
	}

	next := 0
	for _, milestone := range competitionMilestones {
		if count <= milestone {
			next = milestone
			break
		}
	}
	if next == 0 {
		return "Már elérte a maximumot", "", false
	}
	needed := next - count

	upcoming, err := s.wca.GetUpcomingCompetitions(ctx, wcaID)
	if err != nil {
		log.Printf("heti riport: %s közelgő versenyei nem kérhetők le: %v", wcaID, err)
		return "WCA API nem elérhető", "", false
	}
	sort.SliceStable(upcoming, func(i, j int) bool { return upcoming[i].StartDate < upcoming[j].StartDate })

	if needed == 0 {
		if len(upcoming) == 0 {
			return "Elérte a mérföldkövet! 🎉", "", true
		}
		return upcoming[0].Name, upcomingDate(upcoming[0]), true
	}
	if len(upcoming) >= needed {
		target := upcoming[needed-1]
		return target.Name, upcomingDate(target), false
	}
	return fmt.Sprintf("Nincs elég regisztrált verseny (még %d kellene)", needed), "", false
}

func upcomingDate(c wcaUpcomingCompetition) string {
	from, errFrom := time.Parse(time.DateOnly, c.StartDate)
	to, errTo := time.Parse(time.DateOnly, c.EndDate)
	if errFrom != nil || errTo != nil {
		return ""
	}
	return FormatDateRange(from, to)
}

func solveMilestone(p Person) SolveMilestone {
	remainder := p.SolveCount % solveMilestoneStep
	lastCrossed := p.SolveCount - remainder

	m := SolveMilestone{WCAID: p.WCAID, Name: p.Name, SolveCount: p.SolveCount}
	if remainder <= solveMilestoneDifference && lastCrossed > 0 {
		m.Milestone = lastCrossed
		m.Achieved = true
		return m
	}
	m.Milestone = lastCrossed + solveMilestoneStep
	m.Remaining = m.Milestone - p.SolveCount
	return m
}

func (s *Service) publishPost(data ReportData, periodStart, periodEnd time.Time, replace, dryRun bool) error {
	content, err := RenderPostContent(data)
	if err != nil {
		return err
	}
	week := WeekStart(periodEnd)

	written, err := s.repo.PublishReport(ReportPost{
		ID:         uuid.NewString(),
		Title:      PostTitle(periodStart, periodEnd),
		Content:    content,
		ReportWeek: week,
	}, replace)
	if err != nil {
		return err
	}

	switch {
	case dryRun:
	case written && replace:
		log.Printf("heti riport: poszt létrehozva/felülírva (%s heti)", week.Format(time.DateOnly))
	case written:
		log.Printf("heti riport: poszt létrehozva (%s heti)", week.Format(time.DateOnly))
	default:
		log.Printf("heti riport: a %s heti poszt már létezik, kihagyva (felülírás: -force)", week.Format(time.DateOnly))
	}
	return nil
}

func (s *Service) saveRecords(ctx context.Context) error {
	records, err := s.wca.GetRecords(ctx)
	if err != nil {
		return err
	}

	regions := map[string]map[string]RecordValues{
		regionWorld:   records.WorldRecords,
		regionEurope:  records.ContinentalRecords["_Europe"],
		regionHungary: records.NationalRecords["Hungary"],
	}
	saved := 0
	for region, events := range regions {
		for eventID, values := range events {
			if err := s.repo.UpsertRecord(region, eventID, values); err != nil {
				return fmt.Errorf("%s %s: %w", region, eventID, err)
			}
			saved++
		}
	}
	log.Printf("heti riport: %d rekord a WCA-tól feldolgozva", saved)
	return nil
}

func (s *Service) sendEmail(data ReportData, now time.Time) error {
	body, err := RenderEmail(data)
	if err != nil {
		return err
	}
	return s.mailer.Send(EmailSubject(now), body)
}
