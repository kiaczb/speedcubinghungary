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

func (h *Handler) GetUpcoming(w http.ResponseWriter, r *http.Request) {
	competitions, err := h.service.GetUpcomingHungarian()
	if err != nil {
		writeError(w, r, err)
		return
	}
	json.NewEncoder(w).Encode(competitions)
}

func (h *Handler) GetPsychSheet(w http.ResponseWriter, r *http.Request) {
	entries, err := h.service.GetPsychSheet(chi.URLParam(r, "id"), chi.URLParam(r, "eventId"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	json.NewEncoder(w).Encode(entries)
}

func (h *Handler) GetCompetitors(w http.ResponseWriter, r *http.Request) {
	competitors, err := h.service.GetCompetitors(chi.URLParam(r, "id"))
	if err != nil {
		writeError(w, r, err)
		return
	}
	json.NewEncoder(w).Encode(competitors)
}

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
