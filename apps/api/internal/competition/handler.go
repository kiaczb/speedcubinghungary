package competition

import (
	"encoding/json"
	"errors"
	"net/http"
	"speedcubinghungary/internal/httperr"

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
		writeError(w, r, err)
		return
	}
	json.NewEncoder(w).Encode(competitions)
}

// GetPsychSheet handles GET /competitions/{id}/psych-sheet/{eventId}.
func (h *Handler) GetPsychSheet(w http.ResponseWriter, r *http.Request) {
	entries, err := h.service.GetPsychSheet(chi.URLParam(r, "id"), chi.URLParam(r, "eventId"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	json.NewEncoder(w).Encode(entries)
}

// GetCompetitors handles GET /competitions/{id}/competitors.
func (h *Handler) GetCompetitors(w http.ResponseWriter, r *http.Request) {
	competitors, err := h.service.GetCompetitors(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	json.NewEncoder(w).Encode(competitors)
}

// writeError maps a malformed id to 400, an id the WCA does not know to 404,
// and everything else — the WCA API being down or rate-limiting us — to 502.
// The upstream error's text stays in the server log (httperr), because it
// carries WCA URLs and our own query parameters.
func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusBadGateway
	switch {
	case errors.Is(err, ErrInvalidID):
		status = http.StatusBadRequest
	case errors.Is(err, ErrNotFound):
		status = http.StatusNotFound
	}
	httperr.Write(w, r, status, err)
}
