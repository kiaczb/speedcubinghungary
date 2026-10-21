package stats

import (
	"errors"
	"log"
	"sort"
	"sync"
	"time"

	"speedcubinghungary/internal/competition"
)

const (
	historyWorkers          = 3
	historyRetryAfter       = 5 * time.Minute
	historyRateLimitBackoff = 30 * time.Second
	historyMaxAttempts      = 3
	historyRetention        = 24 * time.Hour
	comebackMinDays         = 365
	comebackShown           = 10
)

type historyJob struct {
	histories map[string][]competition.PersonCompetition
	failed    map[string]bool
	running   bool
	lastRun   time.Time
}

type historyTracker struct {
	mu    sync.Mutex
	jobs  map[string]*historyJob
	fetch func(wcaID string) ([]competition.PersonCompetition, error)
	sleep func(time.Duration)
}

func newHistoryTracker(fetch func(string) ([]competition.PersonCompetition, error)) *historyTracker {
	return &historyTracker{
		jobs:  make(map[string]*historyJob),
		fetch: fetch,
		sleep: time.Sleep,
	}
}

type historySnapshot struct {
	histories map[string][]competition.PersonCompetition
	done      int
	failed    int
	running   bool
}

func (t *historyTracker) snapshot(competitionID string, wcaIDs []string) historySnapshot {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	t.pruneLocked(now)

	job, ok := t.jobs[competitionID]
	if !ok {
		job = &historyJob{
			histories: make(map[string][]competition.PersonCompetition),
			failed:    make(map[string]bool),
		}
		t.jobs[competitionID] = job
	}

	snapshot := historySnapshot{
		histories: make(map[string][]competition.PersonCompetition, len(wcaIDs)),
		running:   job.running,
	}
	pending := make([]string, 0)
	for _, id := range wcaIDs {
		if history, ok := job.histories[id]; ok {
			snapshot.histories[id] = history
			snapshot.done++
			continue
		}
		if job.failed[id] {
			snapshot.failed++
			if now.Sub(job.lastRun) < historyRetryAfter {
				continue
			}
		}
		pending = append(pending, id)
	}

	if !job.running && len(pending) > 0 {
		job.running = true
		snapshot.running = true
		go t.run(job, pending)
	}
	return snapshot
}

func (t *historyTracker) pruneLocked(now time.Time) {
	for id, job := range t.jobs {
		if !job.running && !job.lastRun.IsZero() && now.Sub(job.lastRun) > historyRetention {
			delete(t.jobs, id)
		}
	}
}

func (t *historyTracker) run(job *historyJob, wcaIDs []string) {
	work := make(chan string)
	var wg sync.WaitGroup
	for range historyWorkers {
		wg.Go(func() {
			for id := range work {
				history, err := t.fetchWithRetry(id)
				t.mu.Lock()
				if err != nil {
					job.failed[id] = true
				} else {
					job.histories[id] = history
					delete(job.failed, id)
				}
				t.mu.Unlock()
				if err != nil {
					log.Printf("stats: %s versenytörténete nem kérhető le: %v", id, err)
				}
			}
		})
	}
	for _, id := range wcaIDs {
		work <- id
	}
	close(work)
	wg.Wait()

	t.mu.Lock()
	job.running = false
	job.lastRun = time.Now()
	t.mu.Unlock()
}

func (t *historyTracker) fetchWithRetry(wcaID string) ([]competition.PersonCompetition, error) {
	var err error
	for attempt := 1; attempt <= historyMaxAttempts; attempt++ {
		var history []competition.PersonCompetition
		history, err = t.fetch(wcaID)
		if err == nil {
			return history, nil
		}
		if !errors.Is(err, competition.ErrRateLimited) {
			return nil, err
		}
		if attempt < historyMaxAttempts {
			t.sleep(historyRateLimitBackoff)
		}
	}
	return nil, err
}

func computeHistory(info CompetitionInfo, registrants []competition.Registrant, histories map[string][]competition.PersonCompetition) ([]Comeback, []Person, []PersonValue) {
	comebacks := make([]Comeback, 0)
	firstInCountry := make([]Person, 0)
	busiest := make([]PersonValue, 0)

	start, err := time.Parse(time.DateOnly, info.StartDate)
	if err != nil {
		return comebacks, firstInCountry, busiest
	}
	year := info.StartDate[:4]

	for _, r := range registrants {
		history, ok := histories[r.WCAID]
		if !ok || r.WCAID == "" {
			continue
		}

		prior := make([]competition.PersonCompetition, 0, len(history))
		for _, c := range history {
			if c.ID != info.ID && c.StartDate < info.StartDate {
				prior = append(prior, c)
			}
		}
		if len(prior) == 0 {
			continue
		}
		sort.SliceStable(prior, func(i, j int) bool { return prior[i].StartDate < prior[j].StartDate })
		person := personOf(r)

		last := prior[len(prior)-1]
		lastDay := last.EndDate
		if lastDay == "" {
			lastDay = last.StartDate
		}
		if lastEnd, err := time.Parse(time.DateOnly, lastDay); err == nil {
			if gap := int(start.Sub(lastEnd).Hours() / 24); gap >= comebackMinDays {
				comebacks = append(comebacks, Comeback{
					Person:          person,
					LastCompetition: CompetitionStub{ID: last.ID, Name: last.Name, Date: last.StartDate},
					GapDays:         gap,
				})
			}
		}

		seenCountry := false
		thisYear := 0
		for _, c := range prior {
			if c.CountryISO2 == info.CountryISO2 {
				seenCountry = true
			}
			if len(c.StartDate) >= 4 && c.StartDate[:4] == year {
				thisYear++
			}
		}
		if info.CountryISO2 != "" && !seenCountry {
			firstInCountry = append(firstInCountry, person)
		}
		if thisYear > 0 {
			busiest = append(busiest, PersonValue{Person: person, Value: thisYear})
		}
	}

	sort.SliceStable(comebacks, func(i, j int) bool { return comebacks[i].GapDays > comebacks[j].GapDays })
	sortPeople(firstInCountry)
	sort.SliceStable(busiest, func(i, j int) bool { return busiest[i].Value > busiest[j].Value })

	return comebacks[:min(comebackShown, len(comebacks))], firstInCountry, busiest[:min(topCount, len(busiest))]
}
