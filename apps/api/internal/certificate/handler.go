package certificate

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"

	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/httperr"
	"speedcubinghungary/internal/middleware"

	"github.com/go-chi/chi/v5"
)

const maxUploadBytes = MaxAssetSize + 1<<20

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetRecipients(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	mode, err := ParseMode(query.Get("mode"))
	if err != nil {
		httperr.Write(w, r, http.StatusBadRequest, err)
		return
	}

	list, err := h.service.GetRecipients(r.Context(), query.Get("competition"), mode)
	if err != nil {
		writeError(w, r, err)
		return
	}
	json.NewEncoder(w).Encode(list)
}

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	templates, err := h.service.ListTemplates()
	if err != nil {
		httperr.Write(w, r, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(templates)
}

func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	template, err := h.service.GetTemplate(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, storeStatus(err), err)
		return
	}
	json.NewEncoder(w).Encode(template)
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var input TemplateInput
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxDocumentBytes+1<<16)).Decode(&input); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	template, err := h.service.CreateTemplate(input, user.UserID)
	if err != nil {
		httperr.Write(w, r, storeStatus(err), err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(template)
}

type updateTemplateRequest struct {
	TemplateInput
	Version int `json:"version"`
}

func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req updateTemplateRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, MaxDocumentBytes+1<<16)).Decode(&req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	template, err := h.service.UpdateTemplate(chi.URLParam(r, "id"), req.TemplateInput, req.Version, user.UserID)
	if err != nil {
		httperr.Write(w, r, storeStatus(err), err)
		return
	}
	json.NewEncoder(w).Encode(template)
}

func (h *Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteTemplate(chi.URLParam(r, "id")); err != nil {
		httperr.Write(w, r, storeStatus(err), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) ListAssets(w http.ResponseWriter, r *http.Request) {
	assets, err := h.service.ListAssets()
	if err != nil {
		httperr.Write(w, r, http.StatusInternalServerError, err)
		return
	}
	json.NewEncoder(w).Encode(assets)
}

func (h *Handler) UploadAsset(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxUploadBytes)
	if err := r.ParseMultipartForm(maxUploadBytes); err != nil {
		httperr.Write(w, r, http.StatusBadRequest, ErrAssetTooLarge)
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()
	if header.Size > MaxAssetSize {
		httperr.Write(w, r, http.StatusBadRequest, ErrAssetTooLarge)
		return
	}

	data, err := io.ReadAll(io.LimitReader(file, MaxAssetSize+1))
	if err != nil {
		httperr.Write(w, r, http.StatusBadRequest, err)
		return
	}

	asset, err := h.service.CreateAsset(header.Filename, data, user.UserID)
	if err != nil {
		httperr.Write(w, r, storeStatus(err), err)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(asset)
}

func (h *Handler) GetAsset(w http.ResponseWriter, r *http.Request) {
	asset, err := h.service.GetAssetData(chi.URLParam(r, "id"))
	if err != nil {
		httperr.Write(w, r, storeStatus(err), err)
		return
	}
	w.Header().Set("Content-Type", asset.Mime)
	w.Header().Set("Content-Length", strconv.Itoa(len(asset.Data)))
	w.Header().Set("Cache-Control", "private, max-age=31536000, immutable")
	w.Write(asset.Data)
}

func (h *Handler) DeleteAsset(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteAsset(chi.URLParam(r, "id")); err != nil {
		httperr.Write(w, r, storeStatus(err), err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func storeStatus(err error) int {
	switch {
	case errors.Is(err, ErrValidation):
		return http.StatusBadRequest
	case errors.Is(err, ErrTemplateNotFound), errors.Is(err, ErrAssetNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrVersionConflict), errors.Is(err, ErrAssetInUse):
		return http.StatusConflict
	}
	return http.StatusInternalServerError
}

func writeError(w http.ResponseWriter, r *http.Request, err error) {
	status := http.StatusBadGateway
	switch {
	case errors.Is(err, competition.ErrInvalidID):
		status = http.StatusBadRequest
	case errors.Is(err, competition.ErrNotFound):
		status = http.StatusNotFound
	}
	httperr.Write(w, r, status, err)
}
