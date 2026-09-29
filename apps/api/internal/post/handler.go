package post

import (
	"encoding/json"
	"errors"
	"net/http"
	"speedcubinghungary/internal/httperr"
	"speedcubinghungary/internal/types"

	"github.com/go-chi/chi/v5"
)

// ErrInvalidImageOrder: az image_order olyan képre hivatkozik, ami nem ehhez a
// poszthoz tartozik, vagy nem annyi új képet vár, amennyi feltöltésre került.
var ErrInvalidImageOrder = errors.New("invalid image order")

// newImageMarker az image_order-ben egy ÚJ (ebben a kérésben feltöltött) kép
// helyét jelöli; a feltöltött fájlok sorrendben töltik ki ezeket a helyeket.
const newImageMarker = "new"

const maxUploadBytes = MaxImages*MaxImageSize + 1<<20

type Handler struct {
	service   *Service
	uploadDir string
}

func NewHandler(service *Service, uploadDir string) *Handler {
	return &Handler{service: service, uploadDir: uploadDir}
}

// errorStatus a poszt-műveletek hibáit fordítja HTTP státusszá: a kliens
// hibája 400/404, minden más 500.
func errorStatus(err error) int {
	switch {
	case errors.Is(err, ErrPostNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrInvalidID), errors.Is(err, ErrInvalidTag),
		errors.Is(err, ErrEmptyTitle), errors.Is(err, ErrShortContent),
		errors.Is(err, ErrTooManyImages), errors.Is(err, ErrImageTooLarge),
		errors.Is(err, ErrUnsupportedImage), errors.Is(err, ErrInvalidImageOrder):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

// postForm a létrehozás és a szerkesztés közös multipart törzse.
type postForm struct {
	tag     types.PostTag
	title   string
	content string
	// uploaded az ebben a kérésben kiírt képek útvonala, a beérkezés sorrendjében.
	uploaded []string
}

// parsePostForm beolvassa a multipart törzset, és kiírja a feltöltött képeket.
// Hibánál a választ is megírja.
func (h *Handler) parsePostForm(w http.ResponseWriter, r *http.Request) (postForm, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(MaxImageSize); err != nil {
		http.Error(w, "invalid multipart form", http.StatusBadRequest)
		return postForm{}, false
	}

	uploaded, err := SaveImages(r.MultipartForm.File["images"], h.uploadDir)
	if err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return postForm{}, false
	}

	return postForm{
		tag:      types.PostTag(r.FormValue("tag")),
		title:    r.FormValue("title"),
		content:  r.FormValue("content"),
		uploaded: uploaded,
	}, true
}

// resolveImageOrder az image_order-ből (a TELJES végső galéria: meglévő
// útvonalak és "new" jelölők) állítja elő a mentendő képlistát.
func resolveImageOrder(raw string, existing, uploaded []string) ([]string, error) {
	var order []string
	if err := json.Unmarshal([]byte(raw), &order); err != nil {
		return nil, ErrInvalidImageOrder
	}

	allowed := make(map[string]bool, len(existing))
	for _, path := range existing {
		allowed[path] = true
	}

	resolved := make([]string, 0, len(order))
	seen := make(map[string]bool, len(order))
	next := 0

	for _, item := range order {
		if item == newImageMarker {
			if next >= len(uploaded) {
				return nil, ErrInvalidImageOrder
			}
			resolved = append(resolved, uploaded[next])
			next++
			continue
		}
		// Idegen útvonallal más poszt képét lehetne törölni, egy duplikátum
		// pedig a törléskor árván maradó fájlra mutatna.
		if !allowed[item] || seen[item] {
			return nil, ErrInvalidImageOrder
		}
		seen[item] = true
		resolved = append(resolved, item)
	}

	// Minden feltöltött képnek helye kell legyen, különben árva fájl maradna.
	if next != len(uploaded) {
		return nil, ErrInvalidImageOrder
	}

	if len(resolved) > MaxImages {
		return nil, ErrTooManyImages
	}

	return resolved, nil
}

func (h *Handler) GetPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.service.GetPosts()
	if err != nil {
		httperr.Write(w, r, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(posts)
}

func (h *Handler) GetLatestPosts(w http.ResponseWriter, r *http.Request) {
	posts, err := h.service.GetLatestPosts()
	if err != nil {
		httperr.Write(w, r, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(posts)
}

func (h *Handler) GetPostByID(w http.ResponseWriter, r *http.Request) {
	post, err := h.service.GetPostByID(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}
	json.NewEncoder(w).Encode(post)
}

func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	form, ok := h.parsePostForm(w, r)
	if !ok {
		return
	}

	if err := h.service.CreatePost(form.tag, form.title, form.content, form.uploaded); err != nil {
		RemoveImages(form.uploaded, h.uploadDir)
		httperr.Write(w, r, errorStatus(err), err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// A régi képlista kell: ebből derül ki, mely fájlok esnek ki a galériából.
	existing, err := h.service.GetPostByID(id)
	if err != nil {
		httperr.Write(w, r, errorStatus(err), err)
		return
	}

	form, ok := h.parsePostForm(w, r)
	if !ok {
		return
	}

	imagePaths, err := resolveImageOrder(r.FormValue("image_order"), existing.ImagePaths, form.uploaded)
	if err == nil {
		err = h.service.UpdatePost(id, form.tag, form.title, form.content, imagePaths)
	}
	if err != nil {
		RemoveImages(form.uploaded, h.uploadDir)
		httperr.Write(w, r, errorStatus(err), err)
		return
	}

	kept := make(map[string]bool, len(imagePaths))
	for _, path := range imagePaths {
		kept[path] = true
	}
	for _, path := range existing.ImagePaths {
		if !kept[path] {
			RemoveImage(path, h.uploadDir)
		}
	}
}

func (h *Handler) DeletePosts(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	orphaned, err := h.service.DeletePosts(req.IDs)

	// Hibaágon is takarítunk: amit a DB visszaadott, annak a sora már törlődött.
	RemoveImages(orphaned, h.uploadDir)

	if err != nil {
		httperr.Write(w, r, errorStatus(err), err)
	}
}
