package special

import (
	"context"
	"errors"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"speedcubinghungary/internal/wcalive"
)

var (
	ErrInvalidCompetitionID = errors.New("invalid competition id")
	ErrInvalidElapsed       = errors.New("elapsed time must be between 1 and 8640000 centiseconds")
	ErrInvalidAttempt       = errors.New("invalid attempt reference")
)

const maxElapsedCs = 24 * 60 * 60 * 100

var competitionIDPattern = regexp.MustCompile(`^[A-Za-z0-9]{1,64}$`)

const (
	defaultCountryISO2      = "HU"
	competitionLookbackDays = 60
	competitionListLimit    = 500
	competitionSearchLimit  = 25

	competitionCacheTTL     = time.Hour
	competitionListCacheTTL = time.Minute
)

type Service struct {
	repo        Repository
	client      *LiveClient
	tracker     *Tracker
	broadcaster *wcalive.Broadcaster

	mu       sync.Mutex
	resolved map[string]cachedCompetition
	list     []Competition
	listedAt time.Time
}

type cachedCompetition struct {
	competition Competition
	at          time.Time
}

func NewService(repo Repository, client *LiveClient, tracker *Tracker, broadcaster *wcalive.Broadcaster) *Service {
	return &Service{
		repo:        repo,
		client:      client,
		tracker:     tracker,
		broadcaster: broadcaster,
		resolved:    make(map[string]cachedCompetition),
	}
}

func (s *Service) Broadcaster() *wcalive.Broadcaster { return s.broadcaster }

func (s *Service) Tracker() *Tracker { return s.tracker }

func (s *Service) Competitions(ctx context.Context, query string) ([]Competition, error) {
	query = strings.TrimSpace(query)

	if query != "" {
		found, err := s.client.ListCompetitions(ctx, query, time.Time{}, competitionSearchLimit)
		if err != nil {
			return nil, err
		}
		sortCompetitions(found, today())
		return found, nil
	}

	if cached, ok := s.cachedList(); ok {
		return cached, nil
	}

	from := time.Now().AddDate(0, 0, -competitionLookbackDays)
	all, err := s.client.ListCompetitions(ctx, "", from, competitionListLimit)
	if err != nil {
		return nil, err
	}

	local := make([]Competition, 0, 8)
	for _, competition := range all {
		if competition.CountryISO2 == defaultCountryISO2 {
			local = append(local, competition)
		}
	}
	sortCompetitions(local, today())

	s.mu.Lock()
	s.list = local
	s.listedAt = time.Now()
	s.mu.Unlock()

	return local, nil
}

func (s *Service) cachedList() ([]Competition, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.list == nil || time.Since(s.listedAt) > competitionListCacheTTL {
		return nil, false
	}
	return s.list, true
}

func (s *Service) ResolveCompetition(ctx context.Context, competitionWCAID string) (Competition, error) {
	competitionWCAID = strings.TrimSpace(competitionWCAID)
	if !competitionIDPattern.MatchString(competitionWCAID) {
		return Competition{}, ErrInvalidCompetitionID
	}

	s.mu.Lock()
	cached, known := s.resolved[competitionWCAID]
	s.mu.Unlock()

	if known && time.Since(cached.at) < competitionCacheTTL {
		return cached.competition, nil
	}

	liveID, err := s.client.ResolveCompetitionID(ctx, competitionWCAID)
	if err != nil {
		return Competition{}, err
	}

	competition, err := s.client.GetCompetitionInfo(ctx, liveID)
	if err != nil {
		return Competition{}, err
	}

	s.mu.Lock()
	s.resolved[competitionWCAID] = cachedCompetition{competition: *competition, at: time.Now()}
	s.mu.Unlock()

	return *competition, nil
}

type Report struct {
	Competition     Competition  `json:"competition"`
	AvailableRounds []int        `json:"availableRounds"`
	Round           *int         `json:"round"`
	Competitors     []Competitor `json:"competitors"`
}

func (s *Service) GetReport(ctx context.Context, competitionWCAID string, roundNumber *int) (*Report, error) {
	competition, err := s.ResolveCompetition(ctx, competitionWCAID)
	if err != nil {
		return nil, err
	}

	results, err := s.tracker.Results(ctx, competition)
	if err != nil {
		return nil, err
	}

	recorded, err := s.repo.ListRecordedTimes(competition.CompetitionWCAID)
	if err != nil {
		return nil, err
	}

	available := availableRoundNumbers(results)

	if roundNumber != nil {
		results = filterResultsByRound(results, *roundNumber)
	}

	return &Report{
		Competition:     competition,
		AvailableRounds: available,
		Round:           roundNumber,
		Competitors:     BuildCompetitors(results, recorded),
	}, nil
}

func availableRoundNumbers(results []ResultSnapshot) []int {
	seen := make(map[int]bool)
	for _, result := range results {
		seen[result.RoundNumber] = true
	}

	numbers := make([]int, 0, len(seen))
	for number := range seen {
		numbers = append(numbers, number)
	}
	sort.Ints(numbers)

	return numbers
}

func filterResultsByRound(results []ResultSnapshot, roundNumber int) []ResultSnapshot {
	filtered := make([]ResultSnapshot, 0, len(results))
	for _, result := range results {
		if result.RoundNumber == roundNumber {
			filtered = append(filtered, result)
		}
	}
	return filtered
}

func sortCompetitions(competitions []Competition, today string) {
	sort.SliceStable(competitions, func(i, j int) bool {
		left, right := competitions[i], competitions[j]

		leftRank, rightRank := competitionRank(left, today), competitionRank(right, today)
		if leftRank != rightRank {
			return leftRank < rightRank
		}

		if leftRank == rankPast {
			return left.StartDate > right.StartDate
		}
		return left.StartDate < right.StartDate
	})
}

const (
	rankOngoing = iota
	rankUpcoming
	rankPast
)

func competitionRank(competition Competition, today string) int {
	switch {
	case competition.StartDate <= today && today <= competition.EndDate:
		return rankOngoing
	case competition.StartDate > today:
		return rankUpcoming
	default:
		return rankPast
	}
}

func today() string {
	return time.Now().Format(dateLayout)
}

func (s *Service) RecordElapsed(
	ctx context.Context,
	competitionWCAID string,
	liveResultID, attemptNumber int,
	elapsedCs *int,
	note string,
	userID int,
) error {
	if liveResultID <= 0 || attemptNumber <= 0 {
		return ErrInvalidAttempt
	}

	competitionWCAID = strings.TrimSpace(competitionWCAID)
	if !competitionIDPattern.MatchString(competitionWCAID) {
		return ErrInvalidCompetitionID
	}

	if elapsedCs == nil {
		if err := s.repo.DeleteElapsed(competitionWCAID, liveResultID, attemptNumber); err != nil {
			return err
		}
		s.broadcaster.Notify(competitionWCAID)
		return nil
	}

	if *elapsedCs <= 0 || *elapsedCs > maxElapsedCs {
		return ErrInvalidElapsed
	}

	competition, err := s.ResolveCompetition(ctx, competitionWCAID)
	if err != nil {
		return err
	}
	if err := s.repo.EnsureCompetition(competition, userID); err != nil {
		return err
	}

	if err := s.repo.SetElapsed(
		competition.CompetitionWCAID, liveResultID, attemptNumber,
		*elapsedCs, strings.TrimSpace(note), userID, time.Now().UTC(),
	); err != nil {
		return err
	}

	s.broadcaster.Notify(competition.CompetitionWCAID)

	return nil
}
