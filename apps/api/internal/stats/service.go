package stats

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"time"
	_ "time/tzdata"

	"speedcubinghungary/internal/auth"
	"speedcubinghungary/internal/competition"
)

const liveCacheTTL = time.Minute

const birthdateCacheTTL = 15 * time.Minute

type CompetitionSource interface {
	GetCompetition(competitionID string) (competition.Competition, error)
	SearchCompetitions(query string) ([]competition.Competition, error)
	GetRegistrations(competitionID string) (competition.Registrations, error)
	GetPersonSummaries(wcaIDs []string) (map[string]competition.PersonSummary, error)
	GetPersonCompetitions(wcaID string) ([]competition.PersonCompetition, error)
	GetResults(competitionID string) ([]competition.Result, error)
	GetBirthdates(competitionID, accessToken string) (map[int]string, error)
}

type TokenSource interface {
	WCAAccessToken(ctx context.Context, userID int) (string, error)
}

type cachedBirthdates struct {
	birthdates map[int]string
	status     AgeStatus
	at         time.Time
}

type LiveSource interface {
	ResolveCompetitionID(ctx context.Context, competitionWCAID string) (int, error)
	LoadCompetition(ctx context.Context, liveCompetitionID int) (*LiveCompetition, error)
}

type cachedLive struct {
	data *LiveCompetition
	at   time.Time
}

type Service struct {
	competitions CompetitionSource
	live         LiveSource
	tokens       TokenSource
	history      *historyTracker
	location     *time.Location
	now          func() time.Time

	liveMu    sync.Mutex
	liveIDs   map[string]int
	liveCache map[string]cachedLive

	birthdatesMu sync.Mutex
	birthdates   map[string]cachedBirthdates
}

func NewService(competitions CompetitionSource, live LiveSource, tokens TokenSource) *Service {
	location, err := time.LoadLocation("Europe/Budapest")
	if err != nil {
		location = time.UTC
	}
	return &Service{
		competitions: competitions,
		live:         live,
		tokens:       tokens,
		history:      newHistoryTracker(competitions.GetPersonCompetitions),
		location:     location,
		now:          time.Now,
		liveIDs:      make(map[string]int),
		liveCache:    make(map[string]cachedLive),
		birthdates:   make(map[string]cachedBirthdates),
	}
}

func (s *Service) SearchCompetitions(query string) ([]CompetitionInfo, error) {
	competitions, err := s.competitions.SearchCompetitions(query)
	if err != nil {
		return nil, err
	}
	infos := make([]CompetitionInfo, 0, len(competitions))
	for _, c := range competitions {
		infos = append(infos, toInfo(c))
	}
	return infos, nil
}

func toInfo(c competition.Competition) CompetitionInfo {
	eventIDs := c.EventIDs
	if eventIDs == nil {
		eventIDs = []string{}
	}
	return CompetitionInfo{
		ID:              c.ID,
		Name:            c.Name,
		StartDate:       c.StartDate,
		EndDate:         c.EndDate,
		City:            c.City,
		CountryISO2:     c.CountryISO2,
		URL:             c.URL,
		EventIDs:        eventIDs,
		CompetitorLimit: c.CompetitorLimit,
		ResultsPostedAt: c.ResultsPostedAt,
	}
}

func phaseFor(info CompetitionInfo, now time.Time) Phase {
	if info.ResultsPostedAt != "" {
		return PhaseAfter
	}
	if info.StartDate != "" && now.Format(time.DateOnly) >= info.StartDate {
		return PhaseDuring
	}
	return PhaseBefore
}

func (s *Service) GetReport(ctx context.Context, competitionID string, userID int) (*Report, error) {
	comp, err := s.competitions.GetCompetition(competitionID)
	if err != nil {
		return nil, err
	}

	info := toInfo(comp)
	now := s.now()
	phase := phaseFor(info, now.In(s.location))
	report := &Report{
		Competition: info,
		Phase:       phase,
		GeneratedAt: now,
		History:     emptyHistory(HistoryUnavailable),
		Ages:        emptyAges(AgesUnavailable),
	}

	var wg sync.WaitGroup
	var registrations competition.Registrations
	var registrationsErr error
	wg.Go(func() {
		registrations, registrationsErr = s.competitions.GetRegistrations(competitionID)
	})
	if info.ResultsPostedAt != "" {
		wg.Go(func() {
			results, err := s.competitions.GetResults(competitionID)
			if err != nil {
				log.Printf("stats: %s eredményei nem kérhetők le: %v", competitionID, err)
				report.ResultsError = "A hivatalos eredmények most nem kérhetők le a WCA-tól."
				return
			}
			report.Results = computeResults(info, results)
		})
	}
	if phase != PhaseBefore {
		wg.Go(func() {
			live, fetchedAt, err := s.liveData(ctx, competitionID)
			if err != nil {
				log.Printf("stats: %s WCA Live adatai nem kérhetők le: %v", competitionID, err)
				report.LiveError = "A verseny nincs a WCA Live-on, vagy az most nem érhető el."
				return
			}
			report.Live = computeLive(info, live, fetchedAt)
		})
	}
	wg.Wait()

	if registrationsErr != nil {
		log.Printf("stats: %s nevezései nem kérhetők le: %v", competitionID, registrationsErr)
		report.FieldError = "A nevezési adatok most nem kérhetők le a WCA-tól."
		return report, nil
	}

	ids := wcaIDs(registrations)
	summaries, err := s.competitions.GetPersonSummaries(ids)
	report.Field = computeField(info, phase, registrations, summaries)
	if err != nil {
		log.Printf("stats: %s versenyzőinek összesítője nem kérhető le: %v", competitionID, err)
		report.Field.SummariesError = "A versenyzők WCA-összesítője (versenyszám, érmek, rekordok) most nem kérhető le."
	}
	report.History = s.historyStats(info, registrations, ids)
	report.Ages = s.ageStats(ctx, info, registrations, userID)
	return report, nil
}

func (s *Service) ageStats(ctx context.Context, info CompetitionInfo, registrations competition.Registrations, userID int) AgeStats {
	birthdates, status := s.birthdatesFor(ctx, info.ID, userID)
	if status != AgesReady {
		return emptyAges(status)
	}
	return computeAges(info, registrations.Registrants, birthdates)
}

func (s *Service) birthdatesFor(ctx context.Context, competitionID string, userID int) (map[int]string, AgeStatus) {
	token, err := s.tokens.WCAAccessToken(ctx, userID)
	if errors.Is(err, auth.ErrNoGrant) {
		return nil, AgesNoGrant
	}
	if err != nil {
		log.Printf("stats: a WCA-token nem kérhető le (felhasználó %d): %v", userID, err)
		return nil, AgesUnavailable
	}

	key := fmt.Sprintf("%d:%s", userID, competitionID)
	now := time.Now()
	s.birthdatesMu.Lock()
	for k, entry := range s.birthdates {
		if now.Sub(entry.at) >= birthdateCacheTTL {
			delete(s.birthdates, k)
		}
	}
	if entry, ok := s.birthdates[key]; ok {
		s.birthdatesMu.Unlock()
		return entry.birthdates, entry.status
	}
	s.birthdatesMu.Unlock()

	birthdates, err := s.competitions.GetBirthdates(competitionID, token)
	status := AgesReady
	switch {
	case errors.Is(err, competition.ErrForbidden):
		status = AgesForbidden
	case err != nil:
		log.Printf("stats: %s privát WCIF-je nem kérhető le: %v", competitionID, err)
		return nil, AgesUnavailable
	}

	s.birthdatesMu.Lock()
	s.birthdates[key] = cachedBirthdates{birthdates: birthdates, status: status, at: now}
	s.birthdatesMu.Unlock()
	return birthdates, status
}

func wcaIDs(registrations competition.Registrations) []string {
	ids := make([]string, 0, len(registrations.Registrants))
	for _, r := range registrations.Registrants {
		if r.WCAID != "" {
			ids = append(ids, r.WCAID)
		}
	}
	return ids
}

func emptyHistory(status HistoryStatus) HistoryStats {
	return HistoryStats{
		Status:         status,
		Comebacks:      []Comeback{},
		FirstInCountry: []Person{},
		BusiestYear:    []PersonValue{},
	}
}

func (s *Service) historyStats(info CompetitionInfo, registrations competition.Registrations, ids []string) HistoryStats {
	snapshot := s.history.snapshot(info.ID, ids)

	stats := emptyHistory(HistoryReady)
	if snapshot.running {
		stats.Status = HistoryRunning
	}
	stats.Done = snapshot.done
	stats.Failed = snapshot.failed
	stats.Total = len(ids)
	stats.Comebacks, stats.FirstInCountry, stats.BusiestYear = computeHistory(info, registrations.Registrants, snapshot.histories)
	return stats
}

func (s *Service) liveData(ctx context.Context, competitionID string) (*LiveCompetition, time.Time, error) {
	s.liveMu.Lock()
	defer s.liveMu.Unlock()

	now := time.Now()
	for id, entry := range s.liveCache {
		if now.Sub(entry.at) >= liveCacheTTL {
			delete(s.liveCache, id)
		}
	}
	if entry, ok := s.liveCache[competitionID]; ok {
		return entry.data, entry.at, nil
	}

	liveID, ok := s.liveIDs[competitionID]
	if !ok {
		resolved, err := s.live.ResolveCompetitionID(ctx, competitionID)
		if err != nil {
			return nil, time.Time{}, fmt.Errorf("%w: %v", ErrNotOnLive, err)
		}
		liveID = resolved
		s.liveIDs[competitionID] = liveID
	}

	data, err := s.live.LoadCompetition(ctx, liveID)
	if err != nil {
		return nil, time.Time{}, err
	}
	s.liveCache[competitionID] = cachedLive{data: data, at: now}
	return data, now, nil
}
