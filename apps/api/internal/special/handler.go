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

// GetCompetitions handles GET /special/competitions — a verseny-választó
// forrása. A `?query=` névre keres a WCA Live-on; nélküle a magyar versenyeket
// adja (a közeli múlt + a jövőbeliek).
func (h *Handler) GetCompetitions(w http.ResponseWriter, r *http.Request) {
	competitions, err := h.service.Competitions(r.Context(), r.URL.Query().Get("query"))
	if err != nil {
		// A lista a WCA Live-ból jön, ezért az ő kiesése upstream hiba.
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	json.NewEncoder(w).Encode(competitions)
}

// GetReport handles GET /special/report. A `?competition=` a néző által
// választott verseny WCA azonosítója (kötelező), a `?round=N` szűrő pedig
// minden versenyszám N. körére szűkít (1 → 333-r1, 222-r1, …).
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

// writeSpecialError a versenyhez kötött hibák egységes leképezése. Ami a WCA
// Live-tól jön (ismeretlen verseny, elérhetetlen szolgáltatás), az upstream
// hiba — nem a kérés hibája.
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

// parseRoundParam üres stringre nil-t ad (nincs szűrés). A körsorszám WCIF
// szerint 1-alapú, ezért a 0 és a negatív érték érvénytelen.
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

// StreamUpdates handles GET /special/stream?competition=<wcaId>.
//
// Server-Sent Events: a választott verseny minden változásánál egy "update"
// eseményt küldünk, adat nélkül — a kliens erre újrakérdezi a riportot.
//
// A nyitott kapcsolat egyben azt is jelenti, hogy VAN nézője a versenynek:
// erre indul a WCA Live subscription, és az utolsó kapcsolat bezárásakor áll le.
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

	// A versenyt itt is a WCA Live-nál oldjuk fel: a figyelést nem indíthatja
	// el egy kitalált azonosító.
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

	// ⚠️ A globális middleware.JSON minden válaszra application/json-t állít.
	// SSE-nél ezt törölni kell, különben a böngésző nem eseményfolyamként
	// olvassa (ugyanaz a csapda, mint a /uploads/* fájlkiszolgálónál).
	w.Header().Del("Content-Type")
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	// Ha valaha proxy kerülne elé, ne pufferelje a folyamot.
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(http.StatusOK)

	events, unsubscribe := h.service.Broadcaster().Subscribe(competition.CompetitionWCAID)
	defer unsubscribe()

	// Azonnali kezdő esemény: a kliens tudja, hogy él a kapcsolat.
	fmt.Fprint(w, "event: ready\ndata: {}\n\n")
	flusher.Flush()

	// Életjel, hogy a köztes proxyk ne bontsák az üresjáratban álló kapcsolatot.
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

// RecordElapsed handles PUT /special/results/{liveResultId}/attempts/{n}/elapsed.
//
// A kulcs a WCA Live result id-ja és a kísérlet sorszáma — nem saját azonosító:
// a DNF-ek listája a friss WCA Live adatból jön, nálunk csak a rögzített idő él.
// A törzs elapsedCs mezője a solve TÉNYLEGES hossza századmásodpercben, emberi
// megfigyelésből; null érték törli a bejegyzést. A competitionWcaId azért kell,
// mert nincs szerveroldali "figyelt verseny" — a rögzítés is ahhoz a versenyhez
// tartozik, amit az admin épp néz.
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
			// Itt már nem a WCA Live a szűk keresztmetszet (az admin egy épp
			// nézett, tehát feloldott versenyhez rögzít), a maradék hiba a
			// saját tárolónké.
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
