package post

import (
	"encoding/json"
	"errors"
	"net/http"
	"speedcubinghungary/internal/types"
	"strings"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service   *Service
	uploadDir string
}

func NewHandler(service *Service, uploadDir string) *Handler {
	return &Handler{service: service, uploadDir: uploadDir}
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
	var (
		tag       types.PostTag
		title     string
		content   string
		imagePath string
	)

	// Kétféle kérést fogadunk: multipart/form-data (képfeltöltéssel, ezt küldi az
	// admin modal) és a régi application/json (kép nélkül, visszafelé kompatibilis).
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		// A teljes kérés törzsét korlátozzuk, ne csak a memóriában tartott részt.
		r.Body = http.MaxBytesReader(w, r.Body, MaxImageSize+1<<20)

		if err := r.ParseMultipartForm(MaxImageSize); err != nil {
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}

		tag = types.PostTag(r.FormValue("tag"))
		title = r.FormValue("title")
		content = r.FormValue("content")

		file, _, err := r.FormFile("image")
		switch {
		case err == nil:
			defer file.Close()
			imagePath, err = SaveImage(file, h.uploadDir)
			if err != nil {
				status := http.StatusInternalServerError
				if errors.Is(err, ErrUnsupportedImage) {
					status = http.StatusBadRequest
				}
				http.Error(w, err.Error(), status)
				return
			}
		case errors.Is(err, http.ErrMissingFile):
			// A kép opcionális — kép nélkül is létrehozható a poszt.
		default:
			http.Error(w, "invalid image upload", http.StatusBadRequest)
			return
		}
	} else {
		var req struct {
			Tag     types.PostTag `json:"tag"`
			Title   string        `json:"title"`
			Content string        `json:"content"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid json", http.StatusBadRequest)
			return
		}
		tag, title, content = req.Tag, req.Title, req.Content
	}

	if err := h.service.CreatePost(tag, title, content, imagePath); err != nil {
		// Ne maradjon árva fájl a volume-on, ha a poszt mentése elbukott.
		RemoveImage(imagePath, h.uploadDir)

		// Validációs hiba a kliens hibája (400), minden más szerverhiba (500).
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInvalidID) || errors.Is(err, ErrInvalidTag) ||
			errors.Is(err, ErrEmptyTitle) || errors.Is(err, ErrShortContent) {
			status = http.StatusBadRequest
		}
		http.Error(w, err.Error(), status)
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
