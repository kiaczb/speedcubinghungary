package allrounder

import (
	"encoding/json"
	"fmt"
	"net/http"
	"speedcubinghungary/internal/httperr"
	"strconv"
	"strings"
	"time"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) GetStandings(w http.ResponseWriter, r *http.Request) {
	standings, err := h.service.GetStandings(r.Context(), parseYear(r.URL.Query().Get("year")))
	if err != nil {
		httperr.Write(w, r, http.StatusInternalServerError, err)
		return
	}

	json.NewEncoder(w).Encode(standings)
}

func (h *Handler) Refresh(w http.ResponseWriter, r *http.Request) {
	standings, err := h.service.Refresh(r.Context(), parseYear(r.URL.Query().Get("year")))
	if err != nil {
		httperr.Write(w, r, http.StatusInternalServerError, err)
		return
	}

	json.NewEncoder(w).Encode(standings)
}

func (h *Handler) StreamUpdates(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	competitionWCAID, liveCompetitionID, year, live := h.service.LiveCompetition(r.Context(), parseYear(r.URL.Query().Get("year")))

	if live {
		closeViewer := h.service.Tracker().AddViewer(competitionWCAID, liveCompetitionID, year)
		defer closeViewer()
	}

	w.Header().Del("Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	events, unsubscribe := h.service.Broadcaster().Subscribe(strconv.Itoa(year))
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

func parseYear(raw string) int {
	year, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}

	return year
}
