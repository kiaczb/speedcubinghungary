package post

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"speedcubinghungary/internal/httperr"
	"speedcubinghungary/internal/types"
	"strings"

	"github.com/go-chi/chi/v5"
)

// ErrInvalidImageOrder: az image_order olyan képre hivatkozik, ami nem ehhez a
// poszthoz tartozik, vagy nem annyi új képet vár, amennyi feltöltésre került.
var ErrInvalidImageOrder = errors.New("invalid image order")

// newImageMarker az image_order-ben egy ÚJ (ebben a kérésben feltöltött) kép
// helyét jelöli; a feltöltött fájlok sorrendben töltik ki ezeket a helyeket.
const newImageMarker = "new"

type Handler struct {
	service   *Service
	uploadDir string
}

func NewHandler(service *Service, uploadDir string) *Handler {
	return &Handler{service: service, uploadDir: uploadDir}
}

const maxUploadBytes = MaxImages*MaxImageSize + 1<<20

func imageHeaders(form *multipart.Form) []*multipart.FileHeader {
	if form == nil {
		return nil
	}
	headers := append([]*multipart.FileHeader{}, form.File["images"]...)
	return append(headers, form.File["image"]...)
}

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
		// Duplikátum nem lehet: két azonos útvonalból a törléskor az egyik
		// árván maradó fájlra mutatna.
		if !allowed[item] || seen[item] {
			return nil, ErrInvalidImageOrder
		}
		seen[item] = true
		resolved = append(resolved, item)
	}

	// Minden feltöltött képnek helye kell legyen a sorrendben, különben a fájl
	// kiírásra kerülne, de semmi sem hivatkozna rá.
	if next != len(uploaded) {
		return nil, ErrInvalidImageOrder
	}

	if len(resolved) > MaxImages {
		return nil, ErrTooManyImages
	}

	return resolved, nil
}

// uploadStatus a képfeltöltés hibáját fordítja HTTP státusszá.
func uploadStatus(err error) int {
	if errors.Is(err, ErrUnsupportedImage) || errors.Is(err, ErrTooManyImages) ||
		errors.Is(err, ErrImageTooLarge) {
		return http.StatusBadRequest
	}
	return http.StatusInternalServerError
}

// readStatus a poszt-olvasás hibáit fordítja HTTP státusszá. A nem létező (vagy
// nem is UUID alakú) azonosító a KLIENS hibája — korábban mindkettő 500-at
// adott, és a nyers adatbázis-hibával együtt ment ki.
func readStatus(err error) int {
	switch {
	case errors.Is(err, ErrPostNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrInvalidID):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
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
	id := chi.URLParam(r, "id")
	post, err := h.service.GetPostByID(id)
	if err != nil {
		httperr.Write(w, r, readStatus(err), err)
		return
	}
	json.NewEncoder(w).Encode(post)
}
func (h *Handler) CreatePost(w http.ResponseWriter, r *http.Request) {
	var (
		tag        types.PostTag
		title      string
		content    string
		imagePaths []string
	)

	// Kétféle kérést fogadunk: multipart/form-data (képfeltöltéssel, ezt küldi az
	// admin modal) és a régi application/json (kép nélkül, visszafelé kompatibilis).
	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		// A teljes kérés törzsét korlátozzuk, ne csak a memóriában tartott részt.
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)

		if err := r.ParseMultipartForm(MaxImageSize); err != nil {
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}

		tag = types.PostTag(r.FormValue("tag"))
		title = r.FormValue("title")
		content = r.FormValue("content")

		// A képek opcionálisak — kép nélkül is létrehozható a poszt.
		headers := imageHeaders(r.MultipartForm)
		if len(headers) > MaxImages {
			http.Error(w, ErrTooManyImages.Error(), http.StatusBadRequest)
			return
		}

		var err error
		imagePaths, err = SaveImages(headers, h.uploadDir)
		if err != nil {
			httperr.Write(w, r, uploadStatus(err), err)
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

	if err := h.service.CreatePost(tag, title, content, imagePaths); err != nil {
		// Ne maradjon árva fájl a volume-on, ha a poszt mentése elbukott.
		RemoveImages(imagePaths, h.uploadDir)

		// Validációs hiba a kliens hibája (400), minden más szerverhiba (500).
		status := http.StatusInternalServerError
		if errors.Is(err, ErrInvalidID) || errors.Is(err, ErrInvalidTag) ||
			errors.Is(err, ErrEmptyTitle) || errors.Is(err, ErrShortContent) ||
			errors.Is(err, ErrTooManyImages) {
			status = http.StatusBadRequest
		}
		httperr.Write(w, r, status, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (h *Handler) UpdatePost(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, "missing post id", http.StatusBadRequest)
		return
	}

	// A meglévő posztot előbb lekérjük: kell a régi képlistája, hogy csere/
	// eltávolítás után ne maradjon árva fájl a volume-on, illetve hogy képek
	// nélküli szerkesztéskor megtartsuk a jelenlegi galériát.
	existing, err := h.service.GetPostByID(id)
	if err != nil {
		httperr.Write(w, r, readStatus(err), err)
		return
	}

	var (
		tag        types.PostTag
		title      string
		content    string
		imagePaths = existing.ImagePaths // alapból marad a jelenlegi galéria
		uploaded   []string              // az ebben a kérésben feltöltött képek (rollbackhez)
	)

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)

		if err := r.ParseMultipartForm(MaxImageSize); err != nil {
			http.Error(w, "invalid multipart form", http.StatusBadRequest)
			return
		}

		tag = types.PostTag(r.FormValue("tag"))
		title = r.FormValue("title")
		content = r.FormValue("content")

		headers := imageHeaders(r.MultipartForm)
		if len(headers) > MaxImages {
			http.Error(w, ErrTooManyImages.Error(), http.StatusBadRequest)
			return
		}

		uploaded, err = SaveImages(headers, h.uploadDir)
		if err != nil {
			httperr.Write(w, r, uploadStatus(err), err)
			return
		}

		if raw := r.FormValue("image_order"); raw != "" {
			// A kliens a teljes végső sorrendet küldi: ez egyszerre fedi le a
			// törlést, az átrendezést és az új képek beszúrását.
			imagePaths, err = resolveImageOrder(raw, existing.ImagePaths, uploaded)
			if err != nil {
				RemoveImages(uploaded, h.uploadDir)
				httperr.Write(w, r, http.StatusBadRequest, err)
				return
			}
		} else {
			// Sorrend nélkül a régi viselkedés: remove_image=true kiüríti a
			// galériát, az új képek pedig a meglévők végére kerülnek.
			if r.FormValue("remove_image") == "true" {
				imagePaths = nil
			}
			imagePaths = append(append([]string{}, imagePaths...), uploaded...)
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

	if err := h.service.UpdatePost(id, tag, title, content, imagePaths); err != nil {
		// Ha ebben a kérésben töltöttünk fel képet, de a mentés elbukott,
		// takarítsuk el, hogy ne maradjon árva fájl.
		RemoveImages(uploaded, h.uploadDir)

		status := http.StatusInternalServerError
		switch {
		case errors.Is(err, ErrPostNotFound):
			status = http.StatusNotFound
		case errors.Is(err, ErrInvalidID) || errors.Is(err, ErrInvalidTag) ||
			errors.Is(err, ErrEmptyTitle) || errors.Is(err, ErrShortContent) ||
			errors.Is(err, ErrTooManyImages):
			status = http.StatusBadRequest
		}
		httperr.Write(w, r, status, err)
		return
	}

	// Sikeres mentés után a galériából kiesett képek már nem hivatkozottak —
	// töröljük őket a volume-ról.
	kept := make(map[string]bool, len(imagePaths))
	for _, path := range imagePaths {
		kept[path] = true
	}
	for _, path := range existing.ImagePaths {
		if !kept[path] {
			RemoveImage(path, h.uploadDir)
		}
	}

	w.WriteHeader(http.StatusOK)
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
		httperr.Write(w, r, readStatus(err), err)
		return
	}
}
