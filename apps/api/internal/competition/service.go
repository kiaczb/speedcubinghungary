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

// countryNames maps ISO 3166-1 alpha-2 codes to human-readable country names.
var countryNames = map[string]string{
	"HU": "Hungary",
}

// idPattern matches a WCA competition or event id. Both end up directly in
// the upstream request path (GetPsychSheet, GetWCIF), so they are validated
// before use rather than trusted from the client.
var idPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,64}$`)

// ErrInvalidID is returned when a client-supplied competition or event id
// fails idPattern.
var ErrInvalidID = errors.New("invalid id")

// upstreamCacheTTL bounds how long a psych sheet / registrant list is served
// from memory. The WCA API rate-limits bursts at roughly 15 requests, and a
// prediction page with a dozen podium events would exceed that on every load
// without this cache.
const upstreamCacheTTL = 15 * time.Minute

type cacheEntry[T any] struct {
	value T
	at    time.Time
}

type Service struct {
	repo Repository

	mu           sync.Mutex
	competitions map[string]cacheEntry[Competition]
	psychSheets  map[string]cacheEntry[[]PsychSheetEntry]
	competitors  map[string]cacheEntry[[]Competitor]
	// flight collapses concurrent misses for the same key into one upstream
	// request — React StrictMode alone doubles every fetch in development.
	flight singleflight.Group
	// nextSweep is when the caches are next swept for expired entries.
	nextSweep time.Time
}

func NewService(repo Repository) *Service {
	return &Service{
		repo:         repo,
		competitions: make(map[string]cacheEntry[Competition]),
		psychSheets:  make(map[string]cacheEntry[[]PsychSheetEntry]),
		competitors:  make(map[string]cacheEntry[[]Competitor]),
	}
}

// GetUpcomingHungarian returns the simplified list of upcoming Hungarian
// competitions, already sorted by start date by the upstream API.
func (s *Service) GetUpcomingHungarian() ([]Competition, error) {
	raw, err := s.repo.GetUpcomingByCountry("HU")
	if err != nil {
		return nil, err
	}

	competitions := make([]Competition, 0, len(raw))
	for _, c := range raw {
		competitions = append(competitions, toCompetition(c))
	}

	return competitions, nil
}

// GetCompetition returns one competition by id, whatever its date. Use this —
// not GetUpcomingHungarian — to describe a competition you already know the id
// of: the upcoming list drops a competition once it is over, which is exactly
// when a prediction contest about it is most interesting. ErrNotFound means
// the WCA does not know the id.
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

// countryName resolves an ISO2 code to a readable name, falling back to the
// code itself when it is not known.
func countryName(iso2 string) string {
	if name, ok := countryNames[iso2]; ok {
		return name
	}
	return iso2
}

// GetPsychSheet returns the psych-sheet (PR-ranked start list) for one event
// of one competition, ordered by position.
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

// GetCompetitors returns every accepted, competing registrant of a
// competition, sorted by name. Sourced from the public WCIF because the
// /registrations endpoint only carries user ids, not names.
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

// cached serves key from cache while it is fresh, otherwise runs fetch once
// for all concurrent callers and stores the result. Failures are not cached,
// so a transient upstream error is retried on the next request.
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

// sweepLocked drops expired entries from every cache, at most once per TTL.
// The keys are client-supplied ids, so without it every id anyone ever asked
// about would stay in memory for the life of the process.
func (s *Service) sweepLocked() {
	now := time.Now()
	if now.Before(s.nextSweep) {
		return
	}
	s.nextSweep = now.Add(upstreamCacheTTL)

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
