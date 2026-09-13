package learn

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

// A kérés-törzsek közös alakja. Az `Order` mutató, hogy meg lehessen
// különböztetni a "0-t küldött" és a "nem küldött" esetet — utóbbinál marad a
// 0 alapérték, de a különbség itt látszik, nem a mezőnév mögött rejtve.
type eventRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Order       *int   `json:"order"`
}

type caseRequest struct {
	EventID     string `json:"eventId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Order       *int   `json:"order"`
}

type algorithmRequest struct {
	CaseID      string `json:"caseId"`
	Name        string `json:"name"`
	Notation    string `json:"notation"`
	Description string `json:"description"`
	VideoURL    string `json:"videoUrl"`
	Order       *int   `json:"order"`
}

// --- Olvasás (publikus) ------------------------------------------------------

func (h *Handler) GetEvents(w http.ResponseWriter, r *http.Request) {
	events, err := h.service.GetEvents()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(events)
}

func (h *Handler) GetEvent(w http.ResponseWriter, r *http.Request) {
	detail, err := h.service.GetEventDetail(chi.URLParam(r, "id"))
	if err != nil {
		writeLearnError(w, err)
		return
	}

	json.NewEncoder(w).Encode(detail)
}

func (h *Handler) GetCase(w http.ResponseWriter, r *http.Request) {
	detail, err := h.service.GetCaseDetail(chi.URLParam(r, "id"))
	if err != nil {
		writeLearnError(w, err)
		return
	}

	json.NewEncoder(w).Encode(detail)
}

func (h *Handler) GetAlgorithm(w http.ResponseWriter, r *http.Request) {
	detail, err := h.service.GetAlgorithmDetail(chi.URLParam(r, "id"))
	if err != nil {
		writeLearnError(w, err)
		return
	}

	json.NewEncoder(w).Encode(detail)
}

func (h *Handler) GetTree(w http.ResponseWriter, r *http.Request) {
	tree, err := h.service.GetTree()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(tree)
}

// --- Írás (admin) ------------------------------------------------------------

func (h *Handler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	var req eventRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	event, err := h.service.CreateEvent(req.Name, req.Description, orderOf(req.Order))
	if err != nil {
		writeLearnError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(event)
}

func (h *Handler) UpdateEvent(w http.ResponseWriter, r *http.Request) {
	var req eventRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	event, err := h.service.UpdateEvent(
		chi.URLParam(r, "id"), req.Name, req.Description, orderOf(req.Order),
	)
	if err != nil {
		writeLearnError(w, err)
		return
	}

	json.NewEncoder(w).Encode(event)
}

func (h *Handler) DeleteEvent(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteEvent(chi.URLParam(r, "id")); err != nil {
		writeLearnError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateCase(w http.ResponseWriter, r *http.Request) {
	var req caseRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	c, err := h.service.CreateCase(req.EventID, req.Name, req.Description, orderOf(req.Order))
	if err != nil {
		writeLearnError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(c)
}

func (h *Handler) UpdateCase(w http.ResponseWriter, r *http.Request) {
	var req caseRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	c, err := h.service.UpdateCase(
		chi.URLParam(r, "id"), req.EventID, req.Name, req.Description, orderOf(req.Order),
	)
	if err != nil {
		writeLearnError(w, err)
		return
	}

	json.NewEncoder(w).Encode(c)
}

func (h *Handler) DeleteCase(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteCase(chi.URLParam(r, "id")); err != nil {
		writeLearnError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) CreateAlgorithm(w http.ResponseWriter, r *http.Request) {
	var req algorithmRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	a, err := h.service.CreateAlgorithm(
		req.CaseID, req.Name, req.Notation, req.Description, req.VideoURL, orderOf(req.Order),
	)
	if err != nil {
		writeLearnError(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(a)
}

func (h *Handler) UpdateAlgorithm(w http.ResponseWriter, r *http.Request) {
	var req algorithmRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	a, err := h.service.UpdateAlgorithm(
		chi.URLParam(r, "id"), req.CaseID, req.Name, req.Notation,
		req.Description, req.VideoURL, orderOf(req.Order),
	)
	if err != nil {
		writeLearnError(w, err)
		return
	}

	json.NewEncoder(w).Encode(a)
}

func (h *Handler) DeleteAlgorithm(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteAlgorithm(chi.URLParam(r, "id")); err != nil {
		writeLearnError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// --- Segédek -----------------------------------------------------------------

func decodeJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return false
	}
	return true
}

func orderOf(order *int) int {
	if order == nil {
		return 0
	}
	return *order
}

func writeLearnError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrEventNotFound),
		errors.Is(err, ErrCaseNotFound),
		errors.Is(err, ErrAlgorithmNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case IsValidationError(err):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
