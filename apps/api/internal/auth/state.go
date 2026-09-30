package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	stateCookieName = "oauth_state"

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
