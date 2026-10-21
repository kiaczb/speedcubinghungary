package stats

import (
	"encoding/json"
	"errors"
	"net/http"

	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/httperr"
	"speedcubinghungary/internal/middleware"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) SearchCompetitions(w http.ResponseWriter, r *http.Request) {
	competitions, err := h.service.SearchCompetitions(r.URL.Query().Get("query"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	json.NewEncoder(w).Encode(competitions)
}

func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	report, err := h.service.GetReport(r.Context(), chi.URLParam(r, "id"), user.UserID)
	if err != nil {
		writeError(w, r, err)
		return
	}
	json.NewEncoder(w).Encode(report)
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusBadGateway
	switch {
	case errors.Is(err, competition.ErrInvalidID), errors.Is(err, competition.ErrInvalidQuery):
		status = http.StatusBadRequest
	case errors.Is(err, competition.ErrNotFound):
		status = http.StatusNotFound
	}
	httperr.Write(w, r, status, err)
}
