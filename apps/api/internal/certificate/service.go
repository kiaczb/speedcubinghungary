package certificate

import (
	"context"
	"errors"
	"time"

	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/stats"
)

var ErrInvalidMode = errors.New("invalid mode: podium, competitors or registrants")

type CompetitionSource interface {
	GetCompetition(competitionID string) (competition.Competition, error)
	GetRegistrations(competitionID string) (competition.Registrations, error)
	GetResults(competitionID string) ([]competition.Result, error)
}

type LiveSource interface {
	LiveData(ctx context.Context, competitionID string) (*stats.LiveCompetition, time.Time, error)
}

type Service struct {
	repo         Repository
	competitions CompetitionSource
	live         LiveSource
	location     *time.Location
	now          func() time.Time
}

func NewService(repo Repository, competitions CompetitionSource, live LiveSource) *Service {
	location, err := time.LoadLocation("Europe/Budapest")
	if err != nil {
		location = time.UTC
	}
	return &Service{
		repo:         repo,
		competitions: competitions,
		live:         live,
		location:     location,
		now:          time.Now,
	}
}

func ParseMode(value string) (Mode, error) {
	switch mode := Mode(value); mode {
	case ModePodium, ModeCompetitors, ModeRegistrants:
		return mode, nil
	}
	return "", ErrInvalidMode
}

func (s *Service) GetRecipients(ctx context.Context, competitionID string, mode Mode) (*RecipientList, error) {
	comp, err := s.competitions.GetCompetition(competitionID)
	if err != nil {
		return nil, err
	}

	info := toInfo(comp)
	list := &RecipientList{
		Competition: info,
		Mode:        mode,
		Source:      SourceNone,
		Events:      []Event{},
		Recipients:  []Recipient{},
	}

	if mode == ModeRegistrants {
		registrations, err := s.competitions.GetRegistrations(competitionID)
		if err != nil {
			return nil, err
		}
		list.Source = SourceRegistrations
		list.Recipients = registrantsOf(registrations)
		return list, nil
	}

	switch s.resultSource(info) {
	case SourceResults:
		results, err := s.competitions.GetResults(competitionID)
		if err != nil {
			return nil, err
		}
		list.Source = SourceResults
		if mode == ModePodium {
			list.Events, list.Recipients = podiumFromResults(info.EventIDs, results)
		} else {
			list.Recipients = competitorsFromResults(results)
		}
	case SourceLive:
		live, fetchedAt, err := s.live.LiveData(ctx, competitionID)
		if err != nil {
			return nil, err
		}
		list.Source = SourceLive
		list.FetchedAt = &fetchedAt
		if mode == ModePodium {
			list.Events, list.Recipients = podiumFromLive(info.EventIDs, live)
		} else {
			list.Recipients = competitorsFromLive(live)
		}
	default:
		if mode == ModePodium {
			list.Events = pendingEvents(info.EventIDs)
		}
	}
	return list, nil
}

func (s *Service) resultSource(info CompetitionInfo) Source {
	if info.ResultsPostedAt != "" {
		return SourceResults
	}
	today := s.now().In(s.location).Format(time.DateOnly)
	if info.StartDate != "" && today >= info.StartDate {
		return SourceLive
	}
	return SourceNone
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
		ResultsPostedAt: c.ResultsPostedAt,
	}
}
