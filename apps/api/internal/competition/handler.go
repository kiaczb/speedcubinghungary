package competition

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// GetUpcoming handles GET /competitions/upcoming.
func (h *Handler) GetUpcoming(w http.ResponseWriter, r *http.Request) {
	competitions, err := h.service.GetUpcomingHungarian()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(competitions)
}

// GetPsychSheet handles GET /competitions/{id}/psych-sheet/{eventId}.
func (h *Handler) GetPsychSheet(w http.ResponseWriter, r *http.Request) {
	entries, err := h.service.GetPsychSheet(chi.URLParam(r, "id"), chi.URLParam(r, "eventId"))
	if err != nil {
		writeError(w, err)
		return
	}
	json.NewEncoder(w).Encode(entries)
}

// GetCompetitors handles GET /competitions/{id}/competitors.
func (h *Handler) GetCompetitors(w http.ResponseWriter, r *http.Request) {
	competitors, err := h.service.GetCompetitors(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, err)
		return
	}
	json.NewEncoder(w).Encode(competitors)
}

// writeError maps a client-caused id error to 400 and everything else — the
// WCA API being down or rate-limiting us — to 502.
func writeError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrInvalidID) {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	http.Error(w, err.Error(), http.StatusBadGateway)
}
