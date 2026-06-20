package competition

import (
	"encoding/json"
	"net/http"
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
