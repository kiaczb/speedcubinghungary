package prediction

import (
	"encoding/json"
	"errors"
	"net/http"
	"speedcubinghungary/internal/httperr"
	"speedcubinghungary/internal/middleware"
	"time"

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
		httperr.Write(w, r, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(contests)
}

type createContestRequest struct {
	CompetitionID      string             `json:"competitionId"`
	PredictionDeadline string             `json:"predictionDeadline"`
	PodiumEventIDs     []string           `json:"podiumEventIds"`
	ExtraQuestions     []NewExtraQuestion `json:"extraQuestions"`
}

func (h *Handler) CreateContest(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

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

	if err := h.service.CreateContest(req.CompetitionID, deadline, req.PodiumEventIDs, req.ExtraQuestions, user.UserID); err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) DeleteContest(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteContest(chi.URLParam(r, "id")); err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}
}

func errorStatus(err error) int {
	switch {
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, ErrContestNotFound):
		return http.StatusNotFound
	default:
		return http.StatusInternalServerError
	}
}
