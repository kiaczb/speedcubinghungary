package special

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"speedcubinghungary/internal/middleware"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetCompetitions(w http.ResponseWriter, r *http.Request) {
	competitions, err := h.service.Competitions(r.Context(), r.URL.Query().Get("query"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	json.NewEncoder(w).Encode(competitions)
}

func (h *Handler) GetReport(w http.ResponseWriter, r *http.Request) {
	competitionWCAID := r.URL.Query().Get("competition")
	if strings.TrimSpace(competitionWCAID) == "" {
		http.Error(w, "competition is required", http.StatusBadRequest)
		return
	}

	roundNumber, err := parseRoundParam(r.URL.Query().Get("round"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	report, err := h.service.GetReport(r.Context(), competitionWCAID, roundNumber)
	if err != nil {
		writeSpecialError(w, err)
		return
	}

	json.NewEncoder(w).Encode(report)
}

func writeSpecialError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidCompetitionID):
		http.Error(w, err.Error(), http.StatusBadRequest)
	case errors.Is(err, ErrTooManyCompetitions):
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	default:
		http.Error(w, err.Error(), http.StatusBadGateway)
	}
}

func parseRoundParam(raw string) (*int, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	number, err := strconv.Atoi(raw)
	if err != nil || number < 1 {
		return nil, fmt.Errorf("invalid round number %q", raw)
	}

	return &number, nil
}

func (h *Handler) StreamUpdates(w http.ResponseWriter, r *http.Request) {
	competitionWCAID := r.URL.Query().Get("competition")
	if strings.TrimSpace(competitionWCAID) == "" {
		http.Error(w, "competition is required", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	competition, err := h.service.ResolveCompetition(r.Context(), competitionWCAID)
	if err != nil {
		writeSpecialError(w, err)
		return
	}

	closeViewer, err := h.service.Tracker().AddViewer(competition)
	if err != nil {
		writeSpecialError(w, err)
		return
	}
	defer closeViewer()

	w.Header().Del("Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	events, unsubscribe := h.service.Broadcaster().Subscribe(competition.CompetitionWCAID)
	defer unsubscribe()

	fmt.Fprint(w, "event: ready\ndata: {}\n\n")
	flusher.Flush()

	keepAlive := time.NewTicker(25 * time.Second)
	defer keepAlive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return

		case <-events:
			fmt.Fprint(w, "event: update\ndata: {}\n\n")
			flusher.Flush()

		case <-keepAlive.C:
			fmt.Fprint(w, ": keep-alive\n\n")
			flusher.Flush()
		}
	}
}

func (h *Handler) RecordElapsed(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	liveResultID, err := strconv.Atoi(chi.URLParam(r, "liveResultId"))
	if err != nil {
		http.Error(w, "invalid result id", http.StatusBadRequest)
		return
	}

	attemptNumber, err := strconv.Atoi(chi.URLParam(r, "attemptNumber"))
	if err != nil {
		http.Error(w, "invalid attempt number", http.StatusBadRequest)
		return
	}

	var req struct {
		CompetitionWCAID string `json:"competitionWcaId"`
		ElapsedCs        *int   `json:"elapsedCs"`
		Note             string `json:"note"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = h.service.RecordElapsed(
		r.Context(), req.CompetitionWCAID,
		liveResultID, attemptNumber, req.ElapsedCs, req.Note, user.UserID,
	)
	if err != nil {
		switch {
		case errors.Is(err, ErrInvalidElapsed), errors.Is(err, ErrInvalidAttempt),
			errors.Is(err, ErrInvalidCompetitionID):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
