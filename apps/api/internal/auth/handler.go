package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/middleware"
)

const authCookieName = "auth"

type Handler struct {
	service Service
	cfg     *config.Config
}

func NewHandler(s Service, cfg *config.Config) *Handler {
	return &Handler{service: s, cfg: cfg}
}

func (h *Handler) secureCookies() bool {
	return !h.cfg.IsDevelopment()
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	redirectPath := sanitizeRedirectPath(r.URL.Query().Get("redirect"))
	nonce := newStateNonce()

	setStateCookie(w, nonce, redirectPath, h.secureCookies())

	http.Redirect(w, r, h.service.AuthorizeURL(nonce), http.StatusFound)
}

func (h *Handler) LoginCallback(w http.ResponseWriter, r *http.Request) {
	clearStateCookie(w, h.secureCookies())

	redirectPath, ok := verifyState(r, r.URL.Query().Get("state"))
	if !ok {
		log.Printf("auth: elutasított login callback (hiányzó vagy nem egyező state)")
		h.failLogin(w, r)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		log.Printf("auth: login callback `code` nélkül (wca error=%q)", r.URL.Query().Get("error"))
		h.failLogin(w, r)
		return
	}

	jwtToken, err := h.service.LoginCallback(r.Context(), code)
	if err != nil {
		log.Printf("auth: sikertelen bejelentkezés: %v", err)
		h.failLogin(w, r)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    jwtToken,
		HttpOnly: true,
		Secure:   h.secureCookies(),
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   int(tokenLifetime.Seconds()),
	})

	http.Redirect(w, r, h.frontendURL(redirectPath), http.StatusFound)
}

func (h *Handler) failLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, h.frontendURL("/?login=failed"), http.StatusFound)
}

func (h *Handler) frontendURL(path string) string {
	return strings.TrimSuffix(h.cfg.FrontendURL, "/") + path
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r)
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp := map[string]interface{}{
		"user_id":   user.UserID,
		"wca_id":    user.WCAID,
		"name":      user.Name,
		"avatarURL": user.AvatarURL,
		"roles":     user.Roles,
	}

	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   h.secureCookies(),
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1,
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"logged out"}`))
}
