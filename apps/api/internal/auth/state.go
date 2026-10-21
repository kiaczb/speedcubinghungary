package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	stateCookieName = "oauth_state"

	grantStateCookieName = "wca_grant_state"

	stateTTL = 10 * time.Minute

	stateNonceBytes = 32
)

func newStateNonce() string {
	buf := make([]byte, stateNonceBytes)
	rand.Read(buf)
	return hex.EncodeToString(buf)
}

func setStateCookie(w http.ResponseWriter, nonce, redirectPath string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    nonce + ":" + redirectPath,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int(stateTTL.Seconds()),
	})
}

func clearStateCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     stateCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	})
}

func verifyState(r *http.Request, state string) (string, bool) {
	cookie, err := r.Cookie(stateCookieName)
	if err != nil || cookie.Value == "" || state == "" {
		return "", false
	}

	nonce, redirectPath, found := strings.Cut(cookie.Value, ":")
	if !found || nonce == "" {
		return "", false
	}

	if subtle.ConstantTimeCompare([]byte(nonce), []byte(state)) != 1 {
		return "", false
	}

	return sanitizeRedirectPath(redirectPath), true
}

func sanitizeRedirectPath(path string) string {
	const fallback = "/"

	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return fallback
	}

	for _, r := range path {
		if !isSafePathRune(r) {
			return fallback
		}
	}

	parsed, err := url.Parse(path)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return fallback
	}

	return path
}

func isSafePathRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	}

	return strings.ContainsRune("-._~/%?&=@+:$!*'()#", r)
}

type grantState struct {
	userID       int
	redirectPath string
}

func setGrantStateCookie(w http.ResponseWriter, nonce string, userID int, redirectPath string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     grantStateCookieName,
		Value:    nonce + ":" + strconv.Itoa(userID) + ":" + redirectPath,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   int(stateTTL.Seconds()),
	})
}

func clearGrantStateCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     grantStateCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
		MaxAge:   -1,
	})
}

func verifyGrantState(r *http.Request, state string) (grantState, bool) {
	cookie, err := r.Cookie(grantStateCookieName)
	if err != nil || cookie.Value == "" || state == "" {
		return grantState{}, false
	}

	nonce, rest, found := strings.Cut(cookie.Value, ":")
	if !found || nonce == "" {
		return grantState{}, false
	}
	rawUserID, redirectPath, found := strings.Cut(rest, ":")
	if !found {
		return grantState{}, false
	}
	userID, err := strconv.Atoi(rawUserID)
	if err != nil || userID <= 0 {
		return grantState{}, false
	}

	if subtle.ConstantTimeCompare([]byte(nonce), []byte(state)) != 1 {
		return grantState{}, false
	}

	return grantState{userID: userID, redirectPath: sanitizeRedirectPath(redirectPath)}, true
}
