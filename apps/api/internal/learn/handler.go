package learn

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"speedcubinghungary/internal/httperr"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func splitPath(raw string) []string {
	raw = strings.Trim(raw, "/")
	if raw == "" {
		return []string{}
	}
	return strings.Split(raw, "/")
}

func (h *Handler) GetLevel(w http.ResponseWriter, r *http.Request) {
	level, err := h.service.GetLevel(splitPath(r.URL.Query().Get("path")))
	if err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}
	json.NewEncoder(w).Encode(level)
}

func (h *Handler) GetTree(w http.ResponseWriter, r *http.Request) {
	tree, err := h.service.GetTree()
	if err != nil {
		httperr.Write(w, r, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(tree)
}

func (h *Handler) CreateNode(w http.ResponseWriter, r *http.Request) {
	var req NodeInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	node, err := h.service.CreateNode(req)
	if err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(node)
}

func (h *Handler) UpdateNode(w http.ResponseWriter, r *http.Request) {
	var req NodeInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	node, err := h.service.UpdateNode(chi.URLParam(r, "id"), req)
	if err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}
	json.NewEncoder(w).Encode(node)
}

func (h *Handler) DeleteNode(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteNode(chi.URLParam(r, "id")); err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}
}

type reorderNodesRequest struct {
	ParentID *string  `json:"parentId"`
	IDs      []string `json:"ids"`
}

func (h *Handler) ReorderNodes(w http.ResponseWriter, r *http.Request) {
	var req reorderNodesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.service.ReorderNodes(req.ParentID, req.IDs); err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}
}

func (h *Handler) CreateAlgorithm(w http.ResponseWriter, r *http.Request) {
	var req AlgorithmInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	algorithm, err := h.service.CreateAlgorithm(chi.URLParam(r, "id"), req)
	if err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(algorithm)
}

func (h *Handler) UpdateAlgorithm(w http.ResponseWriter, r *http.Request) {
	var req AlgorithmInput
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	algorithm, err := h.service.UpdateAlgorithm(chi.URLParam(r, "id"), req)
	if err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}
	json.NewEncoder(w).Encode(algorithm)
}

func (h *Handler) DeleteAlgorithm(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteAlgorithm(chi.URLParam(r, "id")); err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}
}

type reorderAlgorithmsRequest struct {
	IDs []string `json:"ids"`
}

func (h *Handler) ReorderAlgorithms(w http.ResponseWriter, r *http.Request) {
	var req reorderAlgorithmsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	if err := h.service.ReorderAlgorithms(chi.URLParam(r, "id"), req.IDs); err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}
}

func errorStatus(err error) int {
	switch {
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, ErrNodeNotFound),
		errors.Is(err, ErrParentNotFound),
		errors.Is(err, ErrAlgorithmNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}
