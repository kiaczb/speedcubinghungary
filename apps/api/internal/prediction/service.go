package prediction

import (
	"time"

	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/types"

	"github.com/google/uuid"
)

// CompetitionLookup a competition csomag azon részét absztrahálja, amire a
// tippverseny-lista kiegészítéséhez (név, város, dátum) szükség van — nem az
// egész competition.Service-t vezetjük be függőségként.
type CompetitionLookup interface {
	GetUpcomingHungarian() ([]competition.Competition, error)
}

type Service struct {
	repo         Repository
	competitions CompetitionLookup
}

func NewService(repo Repository, competitions CompetitionLookup) *Service {
	return &Service{repo: repo, competitions: competitions}
}

// NewExtraQuestion egy létrehozáskor kapott extra kérdés — a service adja a
// szervergenerált ID-t és a sorrendet, a hívó csak a szöveget és a típust adja meg.
type NewExtraQuestion struct {
	Prompt string
	Type   types.PredictionQuestionType
}

// GetContests az összes tippversenyt adja vissza, a WCA verseny-metaadatokkal
// (név, város, dátum) feldúsítva. Az a verseny, amit a WCA API már nem ad
// vissza (pl. régi/nem közelgő), a saját id-jével jelenik meg névként —
// enélkül a lista egy hibás WCA-válasz miatt teljesen üres lenne.
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

	competitionsByID, err := s.competitionsByID()
	if err != nil {
		return nil, err
	}

	contests := make([]Contest, 0, len(rows))
	for _, row := range rows {
		contests = append(contests, buildContest(row, competitionsByID[row.CompetitionWCAID], podiumEvents[row.ID], extraQuestions[row.ID]))
	}

	return contests, nil
}

// competitionsByID a közelgő magyar versenyeket WCA id szerint indexeli. A
// hiba nem fatális a tippverseny-listára nézve (lásd GetContests) — a
// competition csomag maga is external API-tól függ.
func (s *Service) competitionsByID() (map[string]competition.Competition, error) {
	upcoming, err := s.competitions.GetUpcomingHungarian()
	if err != nil {
		return map[string]competition.Competition{}, nil
	}

	byID := make(map[string]competition.Competition, len(upcoming))
	for _, c := range upcoming {
		byID[c.ID] = c
	}
	return byID, nil
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
		ParticipantCount:     0, // a tippleadás még nincs implementálva
		PodiumEventIDs:       podiumEventIDs,
		ExtraQuestions:       extraQuestions,
	}
}

// CreateContest létrehoz egy tippversenyt. A verseny-metaadatokat (név,
// város, dátum) nem itt validáljuk — a competitionID a WCA-tól jön, a
// listázáskor derül ki, ha érvénytelen (lásd GetContests).
func (s *Service) CreateContest(competitionID string, deadline time.Time, podiumEventIDs []string, questions []NewExtraQuestion) error {
	extraQuestions := make([]ExtraQuestion, len(questions))
	for i, q := range questions {
		extraQuestions[i] = ExtraQuestion{Prompt: q.Prompt, Type: q.Type}
	}

	if err := ValidateNewContest(competitionID, deadline, podiumEventIDs, extraQuestions); err != nil {
		return err
	}

	row := contestRow{
		ID:                 uuid.NewString(),
		CompetitionWCAID:   competitionID,
		PredictionDeadline: deadline,
	}

	questionRows := make([]extraQuestionRow, len(questions))
	for i, q := range questions {
		questionRows[i] = extraQuestionRow{
			ID:      uuid.NewString(),
			Prompt:  q.Prompt,
			Type:    q.Type,
			Ordinal: i,
		}
	}

	return s.repo.Create(row, podiumEventIDs, questionRows)
}

// DeleteContest töröl egy tippversenyt. A dobogó-versenyszámok és az extra
// kérdések a DB-beli ON DELETE CASCADE miatt vele együtt törlődnek.
func (s *Service) DeleteContest(id string) error {
	return s.repo.Delete(id)
}
