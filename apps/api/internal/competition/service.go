package competition

import (
	"errors"
	"regexp"
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

var ErrInvalidID = errors.New("invalid id")

const upstreamCacheTTL = 15 * time.Minute

type cacheEntry[T any] struct {
	value T
	at    time.Time
}

type Service struct {
	repo Repository

	mu           sync.Mutex
	upcoming     map[string]cacheEntry[[]Competition]
	competitions map[string]cacheEntry[Competition]
	psychSheets  map[string]cacheEntry[[]PsychSheetEntry]
	competitors  map[string]cacheEntry[[]Competitor]
	flight       singleflight.Group
	nextSweep    time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:         repo,
		upcoming:     make(map[string]cacheEntry[[]Competition]),
		competitions: make(map[string]cacheEntry[Competition]),
		psychSheets:  make(map[string]cacheEntry[[]PsychSheetEntry]),
		competitors:  make(map[string]cacheEntry[[]Competitor]),
	}
}

func (s *Service) GetUpcomingHungarian() ([]Competition, error) {
	return cached(s, s.upcoming, "upcoming:HU", func() ([]Competition, error) {
		raw, err := s.repo.GetUpcomingByCountry("HU")
		if err != nil {
			return nil, err
		}

		competitions := make([]Competition, 0, len(raw))
		for _, c := range raw {
			competitions = append(competitions, toCompetition(c))
		}
		return competitions, nil
	})
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

func toCompetition(c wcaCompetition) Competition {
	return Competition{
		ID:        c.ID,
		Name:      c.Name,
		StartDate: c.StartDate,
		EndDate:   c.EndDate,
		City:      c.City,
		Country:   countryName(c.CountryISO2),
		URL:       c.URL,
		EventIDs:  c.EventIDs,
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
	if !idPattern.MatchString(competitionID) {
		return nil, ErrInvalidID
	}

	return cached(s, s.competitors, "competitors:"+competitionID, func() ([]Competitor, error) {
		wcif, err := s.repo.GetWCIF(competitionID)
		if err != nil {
			return nil, err
		}

		competitors := make([]Competitor, 0, len(wcif.Persons))
		for _, p := range wcif.Persons {
			if p.Registration == nil || p.Registration.Status != "accepted" || !p.Registration.IsCompeting {
				continue
			}
			competitors = append(competitors, Competitor{Name: p.Name, WCAID: p.WCAID})
		}
		sort.Slice(competitors, func(i, j int) bool {
			return strings.ToLower(competitors[i].Name) < strings.ToLower(competitors[j].Name)
		})
		return competitors, nil
	})
}

func cached[T any](s *Service, cache map[string]cacheEntry[T], key string, fetch func() (T, error)) (T, error) {
	s.mu.Lock()
	entry, ok := cache[key]
	s.mu.Unlock()
	if ok && time.Since(entry.at) < upstreamCacheTTL {
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

	sweep(s.upcoming, now)
	sweep(s.competitions, now)
	sweep(s.psychSheets, now)
	sweep(s.competitors, now)
}

func sweep[T any](cache map[string]cacheEntry[T], now time.Time) {
	for key, entry := range cache {
		if now.Sub(entry.at) >= upstreamCacheTTL {
			delete(cache, key)
		}
	}
}
