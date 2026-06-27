package auth

import (
	"encoding/json"
	"net/http"
	"os"
	"speedcubinghungary/internal/middleware"
	"strings"
)

type Handler struct {
	service Service
}

func NewHandler(s Service) *Handler {
	return &Handler{service: s}
}

func (h *Handler) LoginCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "missing code", http.StatusBadRequest)
		return
	}

	state := r.URL.Query().Get("state")

	if state == "" || !strings.HasPrefix(state, "/") || strings.HasPrefix(state, "//") {
		state = "/"
	}

	// JWT + user létrehozás, WCA user adat lekérése
	jwtToken, err := h.service.LoginCallback(code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Saját JWT cookie - HttpOnly
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    jwtToken,
		HttpOnly: true,
		Secure:   os.Getenv("DEPLOYMENT") != "development",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   3600,
	})

	redirectURL := os.Getenv("FRONTEND_URL") + state
	http.Redirect(
		w,
		r,
		redirectURL,
		http.StatusFound,
	)
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp := map[string]interface{}{
		"user_id": user.UserID,
		"wca_id":  user.WCAID,
		"name":    user.Name,
		//"avatarURL": user.AvatarURL,
		"roles": user.Roles,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	// A cookie “kiürítése” (érték üres, lejárat azonnal)
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    "",
		HttpOnly: true,
		Secure:   os.Getenv("DEPLOYMENT") != "development",
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1, // azonnali törlés
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"logged out"}`))
}
