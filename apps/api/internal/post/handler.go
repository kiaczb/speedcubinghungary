package post

import (
	"encoding/json"
	"net/http"
	"speedcubinghungary/internal/types"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.service.GetPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(posts)
}
func (h *Handler) GetLatestPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.service.GetLatestPosts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(posts)
}
func (h *Handler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	post, err := h.service.GetPostByID(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(post)
}
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Tag     types.PostTag `json:"tag"`
		Title   string        `json:"title"`
		Content string        `json:"content"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}
	err := h.service.CreatePost(req.Tag, req.Title, req.Content)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) DeletePosts(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil || len(req.IDs) == 0 {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	err = h.service.DeletePosts(req.IDs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
