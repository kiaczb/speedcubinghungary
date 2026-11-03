package weeklyreport

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"speedcubinghungary/internal/wcaformat"
)

const unpostedLookback = 8 * 7 * 24 * time.Hour

type unpostedTally struct {
	mu           sync.Mutex
	competitions map[string]int
	solves       map[string]int
	names        map[string]string
}

func newUnpostedTally() *unpostedTally {
	return &unpostedTally{competitions: map[string]int{}, solves: map[string]int{}, names: map[string]string{}}
}

func (t *unpostedTally) addCompetition(wcif wcaWCIF) {
	byRegistrant := map[int]wcaWCIFPerson{}
	for _, p := range wcif.Persons {
		if p.RegistrantID != nil && p.WCAID != "" && isTrackedPerson(p.CountryISO2, p.WCAID) {
			byRegistrant[*p.RegistrantID] = p
		}
	}

	competed := map[string]bool{}
	solves := map[string]int{}
	for _, event := range eventsFromWCIF(wcif.Events) {
		for _, result := range event.Results {
			p, ok := byRegistrant[result.RegistrantID]
			if !ok || !hasCompeted(result.Attempts) {
				continue
			}
			competed[p.WCAID] = true
			solves[p.WCAID] += countSuccessfulAttempts(result.Attempts)
		}
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	for _, p := range byRegistrant {
		if !competed[p.WCAID] {
			continue
		}
		t.competitions[p.WCAID]++
		t.solves[p.WCAID] += solves[p.WCAID]
		t.names[p.WCAID] = p.Name
	}
}

func hasCompeted(attempts []int) bool {
	for _, a := range attempts {
		if a > 0 || a == wcaformat.DNF {
			return true
		}
	}
	return false
}

func (s *Service) tallyUnposted(ctx context.Context, competitions []wcaCompetition, skip map[string]bool, until time.Time, tally *unpostedTally) {
	last := until.Format(time.DateOnly)
	for _, comp := range competitions {
		if skip[comp.ID] || comp.resultsPosted() || comp.CancelledAt != nil || comp.StartDate > last {
			continue
		}
		wcif, err := s.wca.GetWCIF(ctx, comp.ID)
		if err != nil {
			log.Printf("heti riport: %s (eredmény még nincs feltöltve) WCIF nem érhető el, a számlálókból kimarad: %v", comp.ID, err)
			continue
		}
		tally.addCompetition(wcif)
		log.Printf("heti riport: %s: feltöltetlen eredmény, a WCA Live adatai számítanak", comp.ID)
	}
}

func hungarianDelegates(competitions []wcaCompetition, until time.Time) map[string]string {
	last := until.Format(time.DateOnly)
	delegates := map[string]string{}
	for _, comp := range competitions {
		if comp.CancelledAt != nil || comp.StartDate > last {
			continue
		}
		for _, d := range comp.Delegates {
			if d.WCAID != "" && isTrackedPerson(d.CountryISO2, d.WCAID) {
				delegates[d.WCAID] = d.Name
			}
		}
	}
	return delegates
}

func (s *Service) syncCounts(ctx context.Context, until time.Time, seen map[string]string, tally *unpostedTally, delegates map[string]string) error {
	tracked, err := s.repo.TrackedPeople()
	if err != nil {
		return fmt.Errorf("reading tracked people: %w", err)
	}

	names := map[string]string{}
	delegateNames := map[string]string{}
	for _, p := range tracked {
		names[p.WCAID] = p.Name
		if p.DelegatedCount > 0 {
			delegateNames[p.WCAID] = p.Name
		}
	}
	for id, name := range seen {
		names[id] = name
	}
	for id, name := range tally.names {
		names[id] = name
	}
	for id, name := range delegates {
		delegateNames[id] = name
	}

	ids := sortedKeys(names)
	stats, err := s.wca.GetPersonStats(ctx, ids)
	if err != nil {
		return fmt.Errorf("fetching person stats: %w", err)
	}

	counts := make([]PersonCounts, 0, len(ids))
	for _, id := range ids {
		official, ok := stats[id]
		if !ok {
			continue
		}
		counts = append(counts, PersonCounts{
			WCAID:            id,
			Name:             names[id],
			CompetitionCount: official.CompetitionCount + tally.competitions[id],
			SolveCount:       official.TotalSolves + tally.solves[id],
		})
	}
	if missing := len(ids) - len(counts); missing > 0 {
		log.Printf("heti riport: %d WCA ID nem található a WCA-n, a számlálójuk nem változott", missing)
	}
	if err := s.repo.SaveCounts(counts); err != nil {
		return fmt.Errorf("saving counts: %w", err)
	}
	log.Printf("heti riport: %d versenyző verseny- és kirakásszáma kiszámolva (%d versenyzőnél feltöltetlen eredmény is számít)", len(counts), len(tally.names))

	delegated := make([]DelegatedCount, 0, len(delegateNames))
	for _, id := range sortedKeys(delegateNames) {
		n, err := s.wca.CountDelegatedCompetitions(ctx, id, until)
		if err != nil {
			log.Printf("heti riport: %s delegált versenyei nem kérhetők le: %v", id, err)
			continue
		}
		delegated = append(delegated, DelegatedCount{WCAID: id, Name: delegateNames[id], Count: n})
	}
	if err := s.repo.SaveDelegatedCounts(delegated); err != nil {
		return fmt.Errorf("saving delegated counts: %w", err)
	}
	log.Printf("heti riport: %d delegált száma kiszámolva", len(delegated))
	return nil
}

func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
