package auth

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/middleware"
)

// authCookieName a saját munkamenet-sütink (HS256 JWT).
const authCookieName = "auth"

type Handler struct {
	service Service
	cfg     *config.Config
}

func NewHandler(s Service, cfg *config.Config) *Handler {
	return &Handler{service: s, cfg: cfg}
}

// secureCookies: fejlesztésen kívül minden sütink `Secure`. A DEPLOYMENT-et a
// config csomag kötelezővé teszi és értékkészletre ellenőrzi, tehát ez nem
// eshet vissza némán "development"-re egy hiányzó env miatt.
func (h *Handler) secureCookies() bool {
	return !h.cfg.IsDevelopment()
}

// Login indítja a WCA OAuth folyamatot: GET /login?redirect=/valahova
//
// A folyamat szándékosan a BACKENDEN indul, nem a böngészőben: így a CSRF
// nonce HttpOnly sütibe kerülhet, és a client_id / redirect_uri / scope hármas
// egyetlen helyen él.
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	redirectPath := sanitizeRedirectPath(r.URL.Query().Get("redirect"))
	nonce := newStateNonce()

	setStateCookie(w, nonce, redirectPath, h.secureCookies())

	http.Redirect(w, r, h.service.AuthorizeURL(nonce), http.StatusFound)
}

// LoginCallback a WCA átirányításának célpontja: GET /login/callback?code=…&state=…
func (h *Handler) LoginCallback(w http.ResponseWriter, r *http.Request) {
	// A state sütire a folyamat innentől nincs szükség, akárhogy is alakul —
	// ezért MINDEN ág előtt, egyszer töröljük.
	//
	// ⚠️ Nem `defer`-rel: minden ág `http.Redirect`-tel zárul, ami meghívja a
	// WriteHeader-t, és az AZUTÁN beállított Set-Cookie fejléc már nem kerül ki a
	// válaszba. A kérésből olvasott sütit (r.Cookie) ez nem érinti.
	clearStateCookie(w, h.secureCookies())

	// ⚠️ A CSRF-ellenőrzés a KÓD BEVÁLTÁSA ELŐTT fut: egy idegen `code`-ot ne is
	// próbáljunk beváltani, és semmiképp ne írjunk felhasználót a DB-be miatta.
	redirectPath, ok := verifyState(r, r.URL.Query().Get("state"))
	if !ok {
		// Nem feltétlenül támadás: lejárhatott a 10 perces ablak, a felhasználó
		// visszatölthette egy régi callback URL-t, vagy más böngészőben kezdte a
		// belépést. Mindenesetre itt nem folytatjuk.
		log.Printf("auth: elutasított login callback (hiányzó vagy nem egyező state)")
		h.failLogin(w, r)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		// Ide jut az is, aki a WCA hozzájárulási képernyőjén "elutasít"-ot
		// választ: olyankor a WCA `error=access_denied`-dal irányít vissza.
		log.Printf("auth: login callback `code` nélkül (wca error=%q)", r.URL.Query().Get("error"))
		h.failLogin(w, r)
		return
	}

	jwtToken, err := h.service.LoginCallback(r.Context(), code)
	if err != nil {
		// A részletes ok a szerver logjába megy: a WCA hibaüzenete (állapotkód,
		// upstream szöveg) nem tartozik a böngészőre.
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

// failLogin visszaviszi a felhasználót az oldalra, ahelyett hogy nyers
// hibaoldalt mutatna az API doménjén (és kiszivárogtatná az upstream hibát).
// A `login=failed` paramétert a frontend még nem dolgozza fel.
func (h *Handler) failLogin(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, h.frontendURL("/?login=failed"), http.StatusFound)
}

// frontendURL a FRONTEND_URL és egy (már szanitált) útvonal összefűzése.
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
	// A cookie "kiürítése" (érték üres, lejárat azonnal). Az attribútumoknak
	// egyezniük kell a kiállításkor használtakkal, különben a böngésző nem
	// ugyanazt a sütit törli.
	http.SetCookie(w, &http.Cookie{
		Name:     authCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   h.secureCookies(),
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1, // azonnali törlés
	})

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message":"logged out"}`))
}
