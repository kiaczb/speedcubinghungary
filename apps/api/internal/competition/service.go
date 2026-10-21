package competition

import (
	"errors"
	"log"
	"net/url"
	"regexp"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/singleflight"
)

var countryNames = map[string]string{
	"HU": "Hungary",
}

var idPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,64}$`)

var wcaIDPattern = regexp.MustCompile(`^\d{4}[A-Z]{4}\d{2}$`)

var ErrInvalidID = errors.New("invalid id")

var ErrInvalidQuery = errors.New("invalid search query")

const upstreamCacheTTL = 15 * time.Minute

const registrationCountTTL = time.Minute

const (
	personDataTTL        = 24 * time.Hour
	personBatchSize      = 100
	searchLookback       = 365 * 24 * time.Hour
	maxSearchQueryLength = 100
)

type cacheEntry[T any] struct {
	value T
	at    time.Time
}

type Service struct {
	repo Repository

	mu            sync.Mutex
	upcoming      map[string]cacheEntry[[]Competition]
	competitions  map[string]cacheEntry[Competition]
	psychSheets   map[string]cacheEntry[[]PsychSheetEntry]
	registrations map[string]cacheEntry[Registrations]
	registered    map[string]cacheEntry[int]
	persons       map[string]cacheEntry[PersonSummary]
	histories     map[string]cacheEntry[[]PersonCompetition]
	results       map[string]cacheEntry[[]Result]
	searches      map[string]cacheEntry[[]Competition]
	flight        singleflight.Group
	nextSweep     time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:          repo,
		upcoming:      make(map[string]cacheEntry[[]Competition]),
		competitions:  make(map[string]cacheEntry[Competition]),
		psychSheets:   make(map[string]cacheEntry[[]PsychSheetEntry]),
		registrations: make(map[string]cacheEntry[Registrations]),
		registered:    make(map[string]cacheEntry[int]),
		persons:       make(map[string]cacheEntry[PersonSummary]),
		histories:     make(map[string]cacheEntry[[]PersonCompetition]),
		results:       make(map[string]cacheEntry[[]Result]),
		searches:      make(map[string]cacheEntry[[]Competition]),
	}
}

func (s *Service) GetUpcomingHungarian() ([]Competition, error) {
	listed, err := cached(s, s.upcoming, "upcoming:HU", func() ([]Competition, error) {
		raw, err := s.repo.GetUpcomingByCountry("HU")
		if err != nil {
			return nil, err
		}
		return toCompetitions(raw), nil
	})
	if err != nil {
		return nil, err
	}

	competitions := slices.Clone(listed)
	s.addAcceptedCounts(competitions, time.Now())
	return competitions, nil
}

func (s *Service) addAcceptedCounts(competitions []Competition, now time.Time) {
	var wg sync.WaitGroup
	for i := range competitions {
		competition := &competitions[i]
		if !registrationOpen(*competition, now) {
			continue
		}
		wg.Go(func() {
			count, err := cachedFor(s, s.registered, "registered:"+competition.ID, registrationCountTTL, func() (int, error) {
				return s.repo.GetRegistrationCount(competition.ID)
			})
			if err != nil {
				log.Printf("competition: a %s nevezési létszáma nem kérhető le: %v", competition.ID, err)
				return
			}
			competition.AcceptedCount = &count
		})
	}
	wg.Wait()
}

func registrationOpen(competition Competition, now time.Time) bool {
	opens, err := time.Parse(time.RFC3339, competition.RegistrationOpen)
	if err != nil {
		return false
	}
	closes, err := time.Parse(time.RFC3339, competition.RegistrationClose)
	if err != nil {
		return false
	}
	return !now.Before(opens) && now.Before(closes)
}

func (s *Service) GetCompetition(competitionID string) (Competition, error) {
	if !idPattern.MatchString(competitionID) {
		return Competition{}, ErrInvalidID
	}

	return cached(s, s.competitions, "competition:"+competitionID, func() (Competition, error) {
		raw, err := s.repo.GetCompetition(competitionID)
		if err != nil {
			return Competition{}, err
		}
		return toCompetition(raw), nil
	})
}

func (s *Service) SearchCompetitions(query string) ([]Competition, error) {
	query = strings.TrimSpace(query)
	if len(query) > maxSearchQueryLength {
		return nil, ErrInvalidQuery
	}

	return cached(s, s.searches, "search:"+strings.ToLower(query), func() ([]Competition, error) {
		params := url.Values{}
		if query == "" {
			params.Set("country_iso2", "HU")
			params.Set("start", time.Now().Add(-searchLookback).Format("2006-01-02"))
			params.Set("sort", "start_date")
			params.Set("per_page", "100")
		} else {
			params.Set("q", query)
			params.Set("sort", "-start_date")
			params.Set("per_page", "25")
		}

		raw, err := s.repo.SearchCompetitions(params)
		if err != nil {
			return nil, err
		}
		return toCompetitions(raw), nil
	})
}

func toCompetitions(raw []wcaCompetition) []Competition {
	competitions := make([]Competition, 0, len(raw))
	for _, c := range raw {
		competitions = append(competitions, toCompetition(c))
	}
	return competitions
}

func toCompetition(c wcaCompetition) Competition {
	return Competition{
		ID:          c.ID,
		Name:        c.Name,
		StartDate:   c.StartDate,
		EndDate:     c.EndDate,
		City:        c.City,
		Country:     countryName(c.CountryISO2),
		CountryISO2: c.CountryISO2,
		URL:         c.URL,
		EventIDs:    c.EventIDs,

		RegistrationOpen:  c.RegistrationOpen,
		RegistrationClose: c.RegistrationClose,
		CompetitorLimit:   c.CompetitorLimit,
		ResultsPostedAt:   c.ResultsPostedAt,
	}
}

func countryName(iso2 string) string {
	if name, ok := countryNames[iso2]; ok {
		return name
	}
	return iso2
}

func (s *Service) GetPsychSheet(competitionID, eventID string) ([]PsychSheetEntry, error) {
	if !idPattern.MatchString(competitionID) || !idPattern.MatchString(eventID) {
		return nil, ErrInvalidID
	}

	return cached(s, s.psychSheets, "psych-sheet:"+competitionID+"/"+eventID, func() ([]PsychSheetEntry, error) {
		raw, err := s.repo.GetPsychSheet(competitionID, eventID)
		if err != nil {
			return nil, err
		}

		entries := make([]PsychSheetEntry, 0, len(raw.SortedRankings))
		for _, e := range raw.SortedRankings {
			entries = append(entries, PsychSheetEntry{
				Name:     e.Name,
				WCAID:    e.WCAID,
				Position: e.Pos,
			})
		}
		return entries, nil
	})
}

func (s *Service) GetCompetitors(competitionID string) ([]Competitor, error) {
	registrations, err := s.GetRegistrations(competitionID)
	if err != nil {
		return nil, err
	}

	competitors := make([]Competitor, 0, len(registrations.Registrants))
	for _, r := range registrations.Registrants {
		competitors = append(competitors, Competitor{Name: r.Name, WCAID: r.WCAID})
	}
	sort.Slice(competitors, func(i, j int) bool {
		return strings.ToLower(competitors[i].Name) < strings.ToLower(competitors[j].Name)
	})
	return competitors, nil
}

func (s *Service) GetRegistrations(competitionID string) (Registrations, error) {
	if !idPattern.MatchString(competitionID) {
		return Registrations{}, ErrInvalidID
	}

	return cached(s, s.registrations, "registrations:"+competitionID, func() (Registrations, error) {
		wcif, err := s.repo.GetWCIF(competitionID)
		if err != nil {
			return Registrations{}, err
		}

		registrations := Registrations{
			EventIDs:    make([]string, 0, len(wcif.Events)),
			Registrants: make([]Registrant, 0, len(wcif.Persons)),
		}
		for _, e := range wcif.Events {
			registrations.EventIDs = append(registrations.EventIDs, e.ID)
		}
		for _, p := range wcif.Persons {
			if p.Registration == nil || p.Registration.Status != "accepted" || !p.Registration.IsCompeting {
				continue
			}
			registrations.Registrants = append(registrations.Registrants, Registrant{
				Name:          p.Name,
				WCAID:         p.WCAID,
				CountryISO2:   p.CountryISO2,
				Gender:        p.Gender,
				Roles:         p.Roles,
				EventIDs:      p.Registration.EventIDs,
				PersonalBests: p.PersonalBests,
			})
		}
		return registrations, nil
	})
}

func (s *Service) GetPersonSummaries(wcaIDs []string) (map[string]PersonSummary, error) {
	summaries := make(map[string]PersonSummary, len(wcaIDs))
	missing := make([]string, 0)
	seen := make(map[string]bool, len(wcaIDs))

	now := time.Now()
	s.mu.Lock()
	for _, id := range wcaIDs {
		if seen[id] || !wcaIDPattern.MatchString(id) {
			continue
		}
		seen[id] = true
		if entry, ok := s.persons[id]; ok && now.Sub(entry.at) < personDataTTL {
			summaries[id] = entry.value
			continue
		}
		missing = append(missing, id)
	}
	s.mu.Unlock()

	for start := 0; start < len(missing); start += personBatchSize {
		chunk := missing[start:min(start+personBatchSize, len(missing))]
		raw, err := s.repo.GetPersons(chunk)
		if err != nil {
			return summaries, err
		}

		s.mu.Lock()
		s.sweepLocked()
		for _, p := range raw {
			summary := PersonSummary{
				WCAID:            p.Person.WCAID,
				CompetitionCount: p.CompetitionCount,
				Medals:           p.Medals,
				Records:          p.Records,
				TotalSolves:      p.TotalSolves,
			}
			s.persons[summary.WCAID] = cacheEntry[PersonSummary]{value: summary, at: time.Now()}
			summaries[summary.WCAID] = summary
		}
		s.mu.Unlock()
	}
	return summaries, nil
}

func (s *Service) GetPersonCompetitions(wcaID string) ([]PersonCompetition, error) {
	if !wcaIDPattern.MatchString(wcaID) {
		return nil, ErrInvalidID
	}

	return cachedFor(s, s.histories, "history:"+wcaID, personDataTTL, func() ([]PersonCompetition, error) {
		raw, err := s.repo.GetPersonCompetitions(wcaID)
		if err != nil {
			return nil, err
		}

		competitions := make([]PersonCompetition, 0, len(raw))
		for _, c := range raw {
			competitions = append(competitions, PersonCompetition{
				ID:          c.ID,
				Name:        c.Name,
				StartDate:   c.StartDate,
				EndDate:     c.EndDate,
				CountryISO2: c.CountryISO2,
			})
		}
		sort.SliceStable(competitions, func(i, j int) bool {
			return competitions[i].StartDate < competitions[j].StartDate
		})
		return competitions, nil
	})
}

func (s *Service) GetResults(competitionID string) ([]Result, error) {
	if !idPattern.MatchString(competitionID) {
		return nil, ErrInvalidID
	}

	return cached(s, s.results, "results:"+competitionID, func() ([]Result, error) {
		raw, err := s.repo.GetResults(competitionID)
		if err != nil {
			return nil, err
		}

		results := make([]Result, 0, len(raw))
		for _, r := range raw {
			results = append(results, Result{
				WCAID:         r.WCAID,
				Name:          r.Name,
				CountryISO2:   r.CountryISO2,
				EventID:       r.EventID,
				RoundTypeID:   r.RoundTypeID,
				Pos:           r.Pos,
				Best:          r.Best,
				Average:       r.Average,
				Attempts:      r.Attempts,
				SingleRecord:  stringValue(r.RegionalSingleRecord),
				AverageRecord: stringValue(r.RegionalAverageRecord),
			})
		}
		return results, nil
	})
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func cached[T any](s *Service, cache map[string]cacheEntry[T], key string, fetch func() (T, error)) (T, error) {
	return cachedFor(s, cache, key, upstreamCacheTTL, fetch)
}

func cachedFor[T any](s *Service, cache map[string]cacheEntry[T], key string, ttl time.Duration, fetch func() (T, error)) (T, error) {
	s.mu.Lock()
	entry, ok := cache[key]
	s.mu.Unlock()
	if ok && time.Since(entry.at) < ttl {
		return entry.value, nil
	}

	value, err, _ := s.flight.Do(key, func() (any, error) {
		fetched, err := fetch()
		if err != nil {
			return nil, err
		}
		s.mu.Lock()
		s.sweepLocked()
		cache[key] = cacheEntry[T]{value: fetched, at: time.Now()}
		s.mu.Unlock()
		return fetched, nil
	})
	if err != nil {
		var zero T
		return zero, err
	}
	return value.(T), nil
}

func (s *Service) sweepLocked() {
	now := time.Now()
	if now.Before(s.nextSweep) {
		return
	}
	s.nextSweep = now.Add(upstreamCacheTTL)

	sweep(s.upcoming, now, upstreamCacheTTL)
	sweep(s.competitions, now, upstreamCacheTTL)
	sweep(s.psychSheets, now, upstreamCacheTTL)
	sweep(s.registrations, now, upstreamCacheTTL)
	sweep(s.registered, now, upstreamCacheTTL)
	sweep(s.results, now, upstreamCacheTTL)
	sweep(s.searches, now, upstreamCacheTTL)
	sweep(s.persons, now, personDataTTL)
	sweep(s.histories, now, personDataTTL)
}

func sweep[T any](cache map[string]cacheEntry[T], now time.Time, ttl time.Duration) {
	for key, entry := range cache {
		if now.Sub(entry.at) >= ttl {
			delete(cache, key)
		}
	}
}
