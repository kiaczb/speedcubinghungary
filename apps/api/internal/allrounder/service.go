package allrounder

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"speedcubinghungary/internal/wcalive"
)

const (
	resultsTTL   = 7 * 24 * time.Hour
	pendingTTL   = 6 * time.Hour
	retryTTL     = 15 * time.Minute
	standingsTTL = 5 * time.Minute

	fetchConcurrency = 4
	syncTimeout      = 2 * time.Minute
)

const liveGrace = 24 * time.Hour

type Service struct {
	repo        Repository
	client      *WCAClient
	live        *LiveClient
	tracker     *LiveTracker
	broadcaster *wcalive.Broadcaster

	mu     sync.Mutex
	states map[int]*yearState
}

type yearState struct {
	mu        sync.Mutex
	standings *Standings
	builtAt   time.Time
}

func NewService(repo Repository, client *WCAClient, live *LiveClient, tracker *LiveTracker, broadcaster *wcalive.Broadcaster) *Service {
	return &Service{
		repo:        repo,
		client:      client,
		live:        live,
		tracker:     tracker,
		broadcaster: broadcaster,
		states:      make(map[int]*yearState),
	}
}

func (s *Service) GetStandings(ctx context.Context, year int) (*Standings, error) {
	return s.standings(ctx, ResolveSeries(year), false)
}

func (s *Service) Refresh(ctx context.Context, year int) (*Standings, error) {
	return s.standings(ctx, ResolveSeries(year), true)
}

func (s *Service) standings(ctx context.Context, series Series, force bool) (*Standings, error) {
	state := s.yearState(series.Year)

	state.mu.Lock()
	defer state.mu.Unlock()

	if err := s.sync(ctx, series, force); err != nil {
		log.Printf("all-rounder: a %d-es évad szinkronja részben sikertelen: %v", series.Year, err)
	}

	states, err := s.repo.ListSyncStates(series.competitionIDs())
	if err != nil {
		return nil, err
	}

	live := s.liveLeg(ctx, series, states)

	if !force && live == nil && state.standings != nil && time.Since(state.builtAt) < standingsTTL {
		return state.standings, nil
	}

	standings, err := s.build(series, states, live)
	if err != nil {
		return nil, err
	}

	if live == nil {
		state.standings = standings
		state.builtAt = time.Now()
	} else {
		state.standings = nil
	}

	return standings, nil
}

func (s *Service) liveLeg(ctx context.Context, series Series, states map[string]SyncState) *CompetitionRounds {
	if series.LegacyScoring || s.tracker == nil {
		return nil
	}

	competition, state, found := currentLeg(series, states, time.Now())
	if !found {
		return nil
	}

	liveCompetitionID, err := s.resolveLiveCompetitionID(ctx, competition.CompetitionWCAID, state)
	if err != nil {
		log.Printf("all-rounder: %s nincs a WCA Live-on: %v", competition.CompetitionWCAID, err)
		return nil
	}

	rounds, err := s.tracker.Rounds(ctx, competition.CompetitionWCAID, liveCompetitionID, series.Year)
	if err != nil {
		log.Printf("all-rounder: a %s élő eredménye nem érhető el: %v", competition.CompetitionWCAID, err)
		return nil
	}
	if len(rounds) == 0 {
		return nil
	}

	return &CompetitionRounds{
		CompetitionWCAID: competition.CompetitionWCAID,
		Rounds:           rounds,
	}
}

func currentLeg(series Series, states map[string]SyncState, now time.Time) (SeriesCompetition, SyncState, bool) {
	for _, competition := range series.Competitions {
		state, known := states[competition.CompetitionWCAID]
		if !known {
			continue
		}
		if state.ResultCount > 0 {
			continue
		}
		if state.StartDate == nil || state.EndDate == nil {
			continue
		}

		if now.Before(state.StartDate.Add(-liveGrace)) {
			continue
		}
		if now.After(state.EndDate.Add(24*time.Hour + liveGrace)) {
			continue
		}

		return competition, state, true
	}

	return SeriesCompetition{}, SyncState{}, false
}

func (s *Service) resolveLiveCompetitionID(ctx context.Context, competitionWCAID string, state SyncState) (int, error) {
	if state.LiveCompetitionID != nil {
		return *state.LiveCompetitionID, nil
	}

	liveCompetitionID, err := s.live.ResolveCompetitionID(ctx, competitionWCAID)
	if err != nil {
		return 0, err
	}

	if err := s.repo.SaveLiveCompetitionID(competitionWCAID, liveCompetitionID); err != nil {
		log.Printf("all-rounder: a %s WCA Live azonosítója nem mentődött: %v", competitionWCAID, err)
	}

	return liveCompetitionID, nil
}

func (s *Service) LiveCompetition(ctx context.Context, year int) (string, int, int, bool) {
	series := ResolveSeries(year)
	if series.LegacyScoring {
		return "", 0, series.Year, false
	}

	states, err := s.repo.ListSyncStates(series.competitionIDs())
	if err != nil {
		return "", 0, series.Year, false
	}

	competition, state, found := currentLeg(series, states, time.Now())
	if !found {
		return "", 0, series.Year, false
	}

	liveCompetitionID, err := s.resolveLiveCompetitionID(ctx, competition.CompetitionWCAID, state)
	if err != nil {
		return "", 0, series.Year, false
	}

	return competition.CompetitionWCAID, liveCompetitionID, series.Year, true
}

func (s *Service) Tracker() *LiveTracker { return s.tracker }

func (s *Service) Broadcaster() *wcalive.Broadcaster { return s.broadcaster }

func (s *Service) yearState(year int) *yearState {
	s.mu.Lock()
	defer s.mu.Unlock()

	state, known := s.states[year]
	if !known {
		state = &yearState{}
		s.states[year] = state
	}

	return state
}

func (s *Service) sync(ctx context.Context, series Series, force bool) error {
	ids := series.competitionIDs()

	states, err := s.repo.ListSyncStates(ids)
	if err != nil {
		return err
	}

	type job struct {
		competitionWCAID string
		hadResults       bool
	}

	pending := make([]job, 0, len(ids))
	for _, id := range ids {
		state, known := states[id]
		if force || needsSync(state, known) {
			pending = append(pending, job{competitionWCAID: id, hadResults: state.ResultCount > 0})
		}
	}

	if len(pending) == 0 {
		return nil
	}

	syncCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), syncTimeout)
	defer cancel()

	var (
		wg       sync.WaitGroup
		mu       sync.Mutex
		failures []error
	)

	semaphore := make(chan struct{}, fetchConcurrency)

	for _, pendingJob := range pending {
		wg.Add(1)

		go func(pendingJob job) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := s.syncCompetition(syncCtx, pendingJob.competitionWCAID, pendingJob.hadResults); err != nil {
				mu.Lock()
				failures = append(failures, err)
				mu.Unlock()
			}
		}(pendingJob)
	}

	wg.Wait()

	return errors.Join(failures...)
}

func (s *Service) syncCompetition(ctx context.Context, competitionWCAID string, hadResults bool) error {
	now := time.Now().UTC()

	results, err := s.client.FetchResults(ctx, competitionWCAID)

	switch {
	case errors.Is(err, ErrNoResults) && !hadResults:
		return s.repo.ReplaceResults(competitionWCAID, nil, now)

	case errors.Is(err, ErrNoResults):
		err = errors.New("WCA API 404 for a competition that already had results")
		fallthrough

	case err != nil:
		if markErr := s.repo.MarkSyncFailed(competitionWCAID, now, err.Error()); markErr != nil {
			return markErr
		}
		return err
	}

	if err := s.repo.ReplaceResults(competitionWCAID, results, now); err != nil {
		return err
	}

	s.syncCompetitionDates(ctx, competitionWCAID)

	return nil
}

func (s *Service) syncCompetitionDates(ctx context.Context, competitionWCAID string) {
	info, err := s.client.FetchCompetitionInfo(ctx, competitionWCAID)
	if errors.Is(err, ErrCompetitionUnknown) {
		return
	}
	if err != nil {
		log.Printf("all-rounder: a %s dátuma nem érhető el: %v", competitionWCAID, err)
		return
	}

	if err := s.repo.SaveCompetitionDates(competitionWCAID, info.StartDate, info.EndDate); err != nil {
		log.Printf("all-rounder: a %s dátuma nem mentődött: %v", competitionWCAID, err)
	}
}

func needsSync(state SyncState, known bool) bool {
	if !known {
		return true
	}

	age := time.Since(state.FetchedAt)

	switch {
	case state.LastError != "":
		return age > retryTTL
	case state.ResultCount > 0:
		return age > resultsTTL
	default:
		return age > pendingTTL
	}
}

func (s *Service) build(series Series, states map[string]SyncState, live *CompetitionRounds) (*Standings, error) {
	resultsByCompetition, err := s.repo.ListResults(series.competitionIDs())
	if err != nil {
		return nil, err
	}

	standings := &Standings{
		Year:            series.Year,
		AvailableYears:  AvailableYears(),
		PointsTable:     series.PointsTable,
		CountingResults: series.CountingResults,
		PodiumCount:     series.PodiumCount,
		Competitions:    series.Competitions,
		People:          computePeople(series, resultsByCompetition, live),
		UpdatedAt:       oldestSync(states),
	}

	if live != nil {
		standings.Live = &LiveLeg{
			CompetitionWCAID: live.CompetitionWCAID,
			Label:            competitionLabel(series, live.CompetitionWCAID),
			RoundCount:       len(live.Rounds),
		}
	}

	return standings, nil
}

func competitionLabel(series Series, competitionWCAID string) string {
	for _, competition := range series.Competitions {
		if competition.CompetitionWCAID == competitionWCAID {
			return competition.Label
		}
	}
	return competitionWCAID
}

func computePeople(series Series, resultsByCompetition map[string][]Result, live *CompetitionRounds) []PersonStandings {
	if series.LegacyScoring {
		competitions := make([]competitionResults, 0, len(series.Competitions))
		for _, competition := range series.Competitions {
			results := resultsByCompetition[competition.CompetitionWCAID]
			if len(results) == 0 {
				continue
			}

			competitions = append(competitions, competitionResults{
				CompetitionWCAID: competition.CompetitionWCAID,
				Results:          results,
			})
		}

		return ComputeStandingsLegacy(competitions, series.PointsTable, series.CountingResults)
	}

	competitions := make([]CompetitionRounds, 0, len(series.Competitions))
	for _, competition := range series.Competitions {
		if live != nil && live.CompetitionWCAID == competition.CompetitionWCAID {
			competitions = append(competitions, *live)
			continue
		}

		results := resultsByCompetition[competition.CompetitionWCAID]
		if len(results) == 0 {
			continue
		}

		competitions = append(competitions, CompetitionRounds{
			CompetitionWCAID: competition.CompetitionWCAID,
			Rounds:           RoundsFromResults(results),
		})
	}

	return ComputeStandingsByRound(competitions, series.PointsTable, series.CountingResults)
}

func oldestSync(states map[string]SyncState) *time.Time {
	var oldest *time.Time

	for _, state := range states {
		if oldest == nil || state.FetchedAt.Before(*oldest) {
			fetchedAt := state.FetchedAt
			oldest = &fetchedAt
		}
	}

	return oldest
}
