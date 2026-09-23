package prediction

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"speedcubinghungary/internal/types"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetContests(w http.ResponseWriter, r *http.Request) {
	contests, err := h.service.GetContests()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(contests)
}

type createContestRequest struct {
	CompetitionID      string   `json:"competitionId"`
	PredictionDeadline string   `json:"predictionDeadline"`
	PodiumEventIDs     []string `json:"podiumEventIds"`
	ExtraQuestions     []struct {
		Prompt string                       `json:"prompt"`
		Type   types.PredictionQuestionType `json:"type"`
	} `json:"extraQuestions"`
}

// CreateContest handles POST /predictions.
func (h *Handler) CreateContest(w http.ResponseWriter, r *http.Request) {
	var req createContestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	deadline, err := time.Parse(time.RFC3339, req.PredictionDeadline)
	if err != nil {
		http.Error(w, "invalid predictionDeadline", http.StatusBadRequest)
		return
	}

	questions := make([]NewExtraQuestion, len(req.ExtraQuestions))
	for i, q := range req.ExtraQuestions {
		questions[i] = NewExtraQuestion{Prompt: q.Prompt, Type: q.Type}
	}

	if err := h.service.CreateContest(req.CompetitionID, deadline, req.PodiumEventIDs, questions); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInvalidCompetitionID) || errors.Is(err, ErrPastDeadline) ||
			errors.Is(err, ErrNoPodiumEvents) || errors.Is(err, ErrInvalidQuestionType) ||
			errors.Is(err, ErrEmptyQuestionPrompt) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// DeleteContest handles DELETE /predictions/{id}.
func (h *Handler) DeleteContest(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing prediction id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteContest(id); err != nil {
		status := http.StatusInternalServerError
		if errors.Is(err, ErrContestNotFound) {
			status = http.StatusNotFound
		}
		http.Error(w, err.Error(), status)
		return
	}
}
