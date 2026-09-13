package learn

import (
	"strings"

	"github.com/google/uuid"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// --- Olvasás (publikus) ------------------------------------------------------

func (s *Service) GetEvents() ([]Event, error) {
	return s.repo.ListEvents()
}

func (s *Service) GetEventDetail(id string) (EventDetail, error) {
	if !isUUID(id) {
		return EventDetail{}, ErrEventNotFound
	}

	event, err := s.repo.GetEvent(id)
	if err != nil {
		return EventDetail{}, err
	}

	cases, err := s.repo.ListCases(id)
	if err != nil {
		return EventDetail{}, err
	}

	return EventDetail{Event: event, Cases: cases}, nil
}

func (s *Service) GetCaseDetail(id string) (CaseDetail, error) {
	if !isUUID(id) {
		return CaseDetail{}, ErrCaseNotFound
	}

	c, err := s.repo.GetCase(id)
	if err != nil {
		return CaseDetail{}, err
	}

	event, err := s.repo.GetEvent(c.EventID)
	if err != nil {
		return CaseDetail{}, err
	}

	algorithms, err := s.repo.ListAlgorithms(id)
	if err != nil {
		return CaseDetail{}, err
	}

	return CaseDetail{Event: event, Case: c, Algorithms: algorithms}, nil
}

// GetAlgorithmDetail: a tanulónézethez a szülőket ÉS az eset többi
// algoritmusát is visszaadjuk, hogy a nézet egyetlen kérésből felépüljön, és
// tudjon köztük lapozni.
func (s *Service) GetAlgorithmDetail(id string) (AlgorithmDetail, error) {
	if !isUUID(id) {
		return AlgorithmDetail{}, ErrAlgorithmNotFound
	}

	algorithm, err := s.repo.GetAlgorithm(id)
	if err != nil {
		return AlgorithmDetail{}, err
	}

	c, err := s.repo.GetCase(algorithm.CaseID)
	if err != nil {
		return AlgorithmDetail{}, err
	}

	event, err := s.repo.GetEvent(c.EventID)
	if err != nil {
		return AlgorithmDetail{}, err
	}

	siblings, err := s.repo.ListAlgorithms(algorithm.CaseID)
	if err != nil {
		return AlgorithmDetail{}, err
	}

	return AlgorithmDetail{
		Event:     event,
		Case:      c,
		Algorithm: algorithm,
		Siblings:  siblings,
	}, nil
}

// GetTree: a teljes hierarchia az admin felületnek, három lekérdezésből
// összefűzve (nem szintenként N+1 kérésből).
func (s *Service) GetTree() ([]EventNode, error) {
	events, err := s.repo.ListEvents()
	if err != nil {
		return nil, err
	}

	cases, err := s.repo.ListAllCases()
	if err != nil {
		return nil, err
	}

	algorithms, err := s.repo.ListAllAlgorithms()
	if err != nil {
		return nil, err
	}

	algorithmsByCase := make(map[string][]Algorithm, len(cases))
	for _, a := range algorithms {
		algorithmsByCase[a.CaseID] = append(algorithmsByCase[a.CaseID], a)
	}

	casesByEvent := make(map[string][]CaseNode, len(events))
	for _, c := range cases {
		node := CaseNode{Case: c, Algorithms: algorithmsByCase[c.ID]}
		if node.Algorithms == nil {
			node.Algorithms = make([]Algorithm, 0)
		}
		casesByEvent[c.EventID] = append(casesByEvent[c.EventID], node)
	}

	tree := make([]EventNode, 0, len(events))
	for _, e := range events {
		node := EventNode{Event: e, Cases: casesByEvent[e.ID]}
		if node.Cases == nil {
			node.Cases = make([]CaseNode, 0)
		}
		tree = append(tree, node)
	}

	return tree, nil
}

// --- Írás (admin) ------------------------------------------------------------

func (s *Service) CreateEvent(name, description string, order int) (Event, error) {
	event := Event{
		ID:          uuid.NewString(),
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Order:       order,
	}
	if err := ValidateEvent(event); err != nil {
		return Event{}, err
	}
	if err := s.repo.CreateEvent(event); err != nil {
		return Event{}, err
	}

	return event, nil
}

func (s *Service) UpdateEvent(id, name, description string, order int) (Event, error) {
	if !isUUID(id) {
		return Event{}, ErrEventNotFound
	}

	event := Event{
		ID:          id,
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Order:       order,
	}
	if err := ValidateEvent(event); err != nil {
		return Event{}, err
	}
	if err := s.repo.UpdateEvent(event); err != nil {
		return Event{}, err
	}

	return event, nil
}

func (s *Service) DeleteEvent(id string) error {
	if !isUUID(id) {
		return ErrEventNotFound
	}
	return s.repo.DeleteEvent(id)
}

func (s *Service) CreateCase(eventID, name, description string, order int) (Case, error) {
	// A szülőt előre ellenőrizzük: így hiányzó versenyszámra 404-et adunk, nem
	// egy nyers FK-hibából lett 500-at.
	if err := s.requireEvent(eventID); err != nil {
		return Case{}, err
	}

	c := Case{
		ID:          uuid.NewString(),
		EventID:     eventID,
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Order:       order,
	}
	if err := ValidateCase(c); err != nil {
		return Case{}, err
	}
	if err := s.repo.CreateCase(c); err != nil {
		return Case{}, err
	}

	return c, nil
}

func (s *Service) UpdateCase(id, eventID, name, description string, order int) (Case, error) {
	if !isUUID(id) {
		return Case{}, ErrCaseNotFound
	}
	if err := s.requireEvent(eventID); err != nil {
		return Case{}, err
	}

	c := Case{
		ID:          id,
		EventID:     eventID,
		Name:        strings.TrimSpace(name),
		Description: strings.TrimSpace(description),
		Order:       order,
	}
	if err := ValidateCase(c); err != nil {
		return Case{}, err
	}
	if err := s.repo.UpdateCase(c); err != nil {
		return Case{}, err
	}

	return c, nil
}

func (s *Service) DeleteCase(id string) error {
	if !isUUID(id) {
		return ErrCaseNotFound
	}
	return s.repo.DeleteCase(id)
}

func (s *Service) CreateAlgorithm(caseID, name, notation, description, videoURL string, order int) (Algorithm, error) {
	if err := s.requireCase(caseID); err != nil {
		return Algorithm{}, err
	}

	a := Algorithm{
		ID:          uuid.NewString(),
		CaseID:      caseID,
		Name:        strings.TrimSpace(name),
		Notation:    strings.TrimSpace(notation),
		Description: strings.TrimSpace(description),
		VideoURL:    strings.TrimSpace(videoURL),
		Order:       order,
	}
	if err := ValidateAlgorithm(a); err != nil {
		return Algorithm{}, err
	}
	if err := s.repo.CreateAlgorithm(a); err != nil {
		return Algorithm{}, err
	}

	return a, nil
}

func (s *Service) UpdateAlgorithm(id, caseID, name, notation, description, videoURL string, order int) (Algorithm, error) {
	if !isUUID(id) {
		return Algorithm{}, ErrAlgorithmNotFound
	}
	if err := s.requireCase(caseID); err != nil {
		return Algorithm{}, err
	}

	a := Algorithm{
		ID:          id,
		CaseID:      caseID,
		Name:        strings.TrimSpace(name),
		Notation:    strings.TrimSpace(notation),
		Description: strings.TrimSpace(description),
		// Üres videó-URL = a videó eltávolítása (a repóban NULLIF-fal NULL lesz).
		VideoURL: strings.TrimSpace(videoURL),
		Order:    order,
	}
	if err := ValidateAlgorithm(a); err != nil {
		return Algorithm{}, err
	}
	if err := s.repo.UpdateAlgorithm(a); err != nil {
		return Algorithm{}, err
	}

	return a, nil
}

func (s *Service) DeleteAlgorithm(id string) error {
	if !isUUID(id) {
		return ErrAlgorithmNotFound
	}
	return s.repo.DeleteAlgorithm(id)
}

func (s *Service) requireEvent(eventID string) error {
	if !isUUID(eventID) {
		return ErrEventNotFound
	}
	_, err := s.repo.GetEvent(eventID)
	return err
}

func (s *Service) requireCase(caseID string) error {
	if !isUUID(caseID) {
		return ErrCaseNotFound
	}
	_, err := s.repo.GetCase(caseID)
	return err
}

// isUUID: az id-k UUID oszlopok, és a Postgres a formailag hibás értékre
// típushibát dob — az 500 lenne a 404 helyett. Ezért itt szűrjük.
func isUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
