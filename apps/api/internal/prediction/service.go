package prediction

import (
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"speedcubinghungary/internal/competition"

	"github.com/google/uuid"
)

type CompetitionLookup interface {
	GetCompetition(competitionID string) (competition.Competition, error)
}

type Service struct {
	repo         Repository
	competitions CompetitionLookup
}

func NewService(repo Repository, competitions CompetitionLookup) *Service {
	return &Service{repo: repo, competitions: competitions}
}

func (s *Service) GetContests() ([]Contest, error) {
	rows, err := s.repo.GetAll()
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return []Contest{}, nil
	}

	ids := make([]string, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
	}

	podiumEvents, err := s.repo.GetPodiumEvents(ids)
	if err != nil {
		return nil, err
	}
	extraQuestions, err := s.repo.GetExtraQuestions(ids)
	if err != nil {
		return nil, err
	}

	competitionsByID := s.competitionsByID(rows)

	contests := make([]Contest, 0, len(rows))
	for _, row := range rows {
		contests = append(contests, buildContest(row, competitionsByID[row.CompetitionWCAID], podiumEvents[row.ID], extraQuestions[row.ID]))
	}

	return contests, nil
}

func (s *Service) competitionsByID(rows []contestRow) map[string]competition.Competition {
	ids := make(map[string]bool, len(rows))
	for _, row := range rows {
		ids[row.CompetitionWCAID] = true
	}

	var (
		mu   sync.Mutex
		wg   sync.WaitGroup
		byID = make(map[string]competition.Competition, len(ids))
	)

	for id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()

			comp, err := s.competitions.GetCompetition(id)
			if err != nil {
				if errors.Is(err, competition.ErrNotFound) {
					log.Printf("[prediction] the WCA does not know competition %s", id)
				} else {
					log.Printf("[prediction] looking up competition %s failed: %v", id, err)
				}
				return
			}

			mu.Lock()
			byID[id] = comp
			mu.Unlock()
		}(id)
	}
	wg.Wait()

	return byID
}

func buildContest(row contestRow, comp competition.Competition, podiumEventIDs []string, questions []extraQuestionRow) Contest {
	name := comp.Name
	if name == "" {
		name = row.CompetitionWCAID
	}

	extraQuestions := make([]ExtraQuestion, 0, len(questions))
	for _, q := range questions {
		extraQuestions = append(extraQuestions, ExtraQuestion{ID: q.ID, Prompt: q.Prompt, Type: q.Type})
	}
	if podiumEventIDs == nil {
		podiumEventIDs = []string{}
	}

	return Contest{
		ID:                   row.ID,
		CompetitionID:        row.CompetitionWCAID,
		CompetitionName:      name,
		CompetitionCity:      comp.City,
		CompetitionCountry:   comp.Country,
		CompetitionStartDate: comp.StartDate,
		PredictionDeadline:   row.PredictionDeadline,
		Status:               statusFor(row.PredictionDeadline),
		ParticipantCount:     0,
		PodiumEventIDs:       podiumEventIDs,
		ExtraQuestions:       extraQuestions,
	}
}

func (s *Service) CreateContest(competitionID string, deadline time.Time, podiumEventIDs []string, questions []NewExtraQuestion, createdBy int) error {
	trimmed := make([]NewExtraQuestion, len(questions))
	for i, q := range questions {
		trimmed[i] = NewExtraQuestion{Prompt: strings.TrimSpace(q.Prompt), Type: q.Type}
	}

	if err := ValidateNewContest(competitionID, deadline, podiumEventIDs, trimmed); err != nil {
		return err
	}

	row := contestRow{
		ID:                 uuid.NewString(),
		CompetitionWCAID:   competitionID,
		PredictionDeadline: deadline,
		CreatedBy:          createdBy,
	}

	questionRows := make([]extraQuestionRow, len(trimmed))
	for i, q := range trimmed {
		questionRows[i] = extraQuestionRow{
			ID:      uuid.NewString(),
			Prompt:  q.Prompt,
			Type:    q.Type,
			Ordinal: i,
		}
	}

	return s.repo.Create(row, podiumEventIDs, questionRows)
}

func (s *Service) DeleteContest(id string) error {
	if err := validContestID(id); err != nil {
		return err
	}
	return s.repo.Delete(id)
}
