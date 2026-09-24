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

// A WCA OAuth folyamat CSRF-védelme.
//
// Az OAuth `state` paraméter NEM csak "hova térjünk vissza" adat: ez a folyamat
// CSRF-tokenje. Enélkül egy támadó a SAJÁT WCA-fiókjához tartozó `code`-ot
// játszhatja be az áldozat böngészőjébe (login CSRF), és az áldozat onnantól a
// támadó fiókjában dolgozik anélkül, hogy észrevenné.
//
// Ezért a folyamat a BACKENDEN indul (GET /login): itt generálunk egy véletlen
// nonce-ot, sütibe tesszük, és CSAK a nonce-ot küldjük ki `state`-ként. A
// visszatéréskor a kettőnek egyeznie kell.
//
// A visszairányítási útvonal szándékosan a SÜTIBEN utazik, nem a `state`-ben:
// a `state` a felhasználói ügynökön keresztül fordul meg, tehát a támadó
// tetszőlegesre írhatja. A süti HttpOnly és mi állítottuk ki, így az útvonal
// forrása végig a miénk marad.

const (
	// stateCookieName a nonce + visszairányítási útvonal rövid életű sütije.
	stateCookieName = "oauth_state"

	// stateTTL: ennyi ideje van a felhasználónak befejezni a WCA-s bejelentkezést.
	stateTTL = 10 * time.Minute

	// stateNonceBytes a nonce hossza bájtban (hex-ként kétszer ennyi karakter).
	stateNonceBytes = 32
)

// newStateNonce kriptográfiailag erős véletlen nonce-ot ad. A crypto/rand hibája
// Go 1.24+ alatt már nem fordulhat elő (a rand.Read sosem hibázik), de a hibát
// mégis felvezetjük: ha valaha mégis, NEM szabad kitalálható nonce-szal indítani
// a folyamatot.
func newStateNonce() (string, error) {
	buf := make([]byte, stateNonceBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}

// setStateCookie kiállítja a nonce + útvonal sütit.
//
// ⚠️ SameSite=Lax (nem Strict): a WCA-tól visszaérkező átirányítás egy
// KERESZT-OLDALI, felső szintű navigáció, amire a böngésző a Strict sütit NEM
// küldené el — a bejelentkezés minden alkalommal elbukna. A Lax pont ezt az
// esetet (top-level GET) engedi át.
//
// ⚠️ Path="/" (nem "/login"): test/prod alatt a Caddy a `WEB_DOMAIN/api` prefixet
// levágja, tehát a böngésző által látott útvonal `/api/login/callback`, míg az API
// `/login/callback`-et lát. Egy `/login`-ra szűkített süti ezért dev alatt működne,
// prodban viszont sosem érkezne vissza.
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

// clearStateCookie törli a state sütit. Az attribútumoknak egyezniük kell a
// kiállításkor használtakkal, különben a böngésző nem ugyanazt a sütit törli.
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

// verifyState összeveti a WCA-tól visszakapott `state`-et a sütivel, és
// visszaadja a sütiben tárolt visszairányítási útvonalat.
//
// Az összehasonlítás konstans idejű: a nonce titok, a `state` pedig támadó által
// vezérelt bemenet.
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

	// A süti a miénk, de az útvonalat a kiállításkor és itt is szanitáljuk —
	// egy elrontott /login ág így sem tud nyílt átirányítást okozni.
	return sanitizeRedirectPath(redirectPath), true
}

// sanitizeRedirectPath a bejelentkezés utáni visszairányítás célját szűri.
// Kizárólag a saját oldalunkon belüli, egyszerű útvonalat engedünk át; minden
// más "/"-re esik vissza.
//
// Amit ki kell zárni:
//   - abszolút URL ("https://evil.example") → nyílt átirányítás,
//   - protokoll-relatív útvonal ("//evil.example") → ugyanaz,
//   - visszaperjeles változat ("/\evil.example", "/\/evil.example") → egyes
//     böngészők a "\"-t "/"-ként normalizálják, tehát ez is protokoll-relatívvá
//     válhat,
//   - vezérlőkarakterek (CR/LF) → válaszfejléc-injektálás a Location fejlécben.
//
// Ezért nem tiltólistázunk, hanem ENGEDÉLYEZŐ listát használunk: a karakter vagy
// szerepel az URL-útvonalakban szokásos, biztonságos halmazban, vagy az egész
// útvonalat eldobjuk. Ez egyúttal a sütibe írhatóságot is garantálja (a szóköz,
// a vessző, a pontosvessző és az idézőjel a süti-értékben is tilos).
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

	// Végső ellenőrzés a szabványos parserrel: csak relatív, host nélküli,
	// séma nélküli útvonal maradhat.
	parsed, err := url.Parse(path)
	if err != nil || parsed.IsAbs() || parsed.Host != "" {
		return fallback
	}

	return path
}

// isSafePathRune: az alkalmazás útvonalaiban (pl. "/posts/<uuid>",
// "/tools/special?competition=Foo2026") előforduló karakterek. Szándékosan NEM
// tartalmazza a "\", a szóköz, a ",", a ";" és a '"' karaktereket.
func isSafePathRune(r rune) bool {
	switch {
	case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9':
		return true
	}

	return strings.ContainsRune("-._~/%?&=@+:$!*'()#", r)
}
