package middleware

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/types"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testSecret = "test-secret-that-is-at-least-32-chars-long"

func testCfg() *config.Config {
	return &config.Config{JWTSecret: testSecret}
}

func validClaims() UserClaims {
	now := time.Now()
	return UserClaims{
		UserID: 1,
		WCAID:  "2019ABCD01",
		Name:   "John Doe",
		Roles:  []types.Role{types.ROLE_USER},
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		},
	}
}

func signHS256(t *testing.T, claims UserClaims, secret string) string {
	t.Helper()
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	require.NoError(t, err)
	return token
}

func requestWithToken(token string) *http.Request {
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	if token != "" {
		req.AddCookie(&http.Cookie{Name: "auth", Value: token})
	}
	return req
}

type stubSessions struct {
	session Session
	err     error
}

func (s stubSessions) GetSession(context.Context, int) (Session, error) {
	return s.session, s.err
}

func currentSession() stubSessions {
	return stubSessions{session: Session{Version: 0, Roles: []types.Role{types.ROLE_USER}}}
}

func run(cfg *config.Config, req *http.Request) (*httptest.ResponseRecorder, *UserClaims) {
	return runWith(cfg, currentSession(), req)
}

func runWith(cfg *config.Config, sessions SessionStore, req *http.Request) (*httptest.ResponseRecorder, *UserClaims) {
	var seen *UserClaims

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seen = GetUserFromContext(r)
		w.WriteHeader(http.StatusOK)
	})

	w := httptest.NewRecorder()
	AuthMiddlewareWithConfig(cfg, sessions)(next).ServeHTTP(w, req)

	return w, seen
}

func TestAuthMiddleware_RolesComeFromTheDatabaseNotTheToken(t *testing.T) {
	claims := validClaims()
	claims.Roles = []types.Role{types.ROLE_USER, types.ROLE_ADMIN}

	w, seen := runWith(testCfg(), currentSession(), requestWithToken(signHS256(t, claims, testSecret)))

	require.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, []types.Role{types.ROLE_USER}, seen.Roles,
		"az elvett admin jog azonnal hat, nem csak a 12 órás token lejártakor")
}

func TestAuthMiddleware_RejectsRevokedOrUnknownSessions(t *testing.T) {
	token := signHS256(t, validClaims(), testSecret)

	tests := []struct {
		name     string
		sessions stubSessions
		expected int
	}{
		{
			name:     "kijelentkezés után nagyobb a verzió",
			sessions: stubSessions{session: Session{Version: 1, Roles: []types.Role{types.ROLE_USER}}},
			expected: http.StatusUnauthorized,
		},
		{name: "törölt felhasználó", sessions: stubSessions{err: ErrSessionNotFound}, expected: http.StatusUnauthorized},
		{name: "adatbázishiba", sessions: stubSessions{err: errors.New("pq: connection refused")}, expected: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, seen := runWith(testCfg(), tt.sessions, requestWithToken(token))

			assert.Equal(t, tt.expected, w.Code)
			assert.Nil(t, seen, "a védett handler nem futhatott le")
			assert.NotContains(t, w.Body.String(), "pq:", "a belső hiba szövege nem jut ki")
		})
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	w, claims := run(testCfg(), requestWithToken(signHS256(t, validClaims(), testSecret)))

	assert.Equal(t, http.StatusOK, w.Code)
	require.NotNil(t, claims)
	assert.Equal(t, 1, claims.UserID)
	assert.Equal(t, "2019ABCD01", claims.WCAID)
	assert.Contains(t, claims.Roles, types.ROLE_USER)
}

func TestAuthMiddleware_Rejects(t *testing.T) {
	expired := validClaims()
	expired.ExpiresAt = jwt.NewNumericDate(time.Now().Add(-time.Minute))

	noExpiry := validClaims()
	noExpiry.ExpiresAt = nil

	notYetValid := validClaims()
	notYetValid.NotBefore = jwt.NewNumericDate(time.Now().Add(time.Hour))

	tests := []struct {
		name  string
		token string
	}{
		{name: "nincs süti", token: ""},
		{name: "üres token", token: " "},
		{name: "szemét token", token: "not-a-jwt"},
		{name: "lejárt token", token: signHS256(t, expired, testSecret)},
		{name: "lejárat nélküli token", token: signHS256(t, noExpiry, testSecret)},
		{name: "még nem érvényes token", token: signHS256(t, notYetValid, testSecret)},
		{name: "másik titokkal aláírva", token: signHS256(t, validClaims(), "another-secret-at-least-32-chars-long")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w, claims := run(testCfg(), requestWithToken(tt.token))

			assert.Equal(t, http.StatusUnauthorized, w.Code)
			assert.Nil(t, claims, "a védett handler nem futhatott le")
		})
	}
}

func TestAuthMiddleware_RejectsNoneAlgorithm(t *testing.T) {
	claims := validClaims()
	claims.Roles = []types.Role{types.ROLE_ADMIN}

	token, err := jwt.NewWithClaims(jwt.SigningMethodNone, claims).
		SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	w, seen := run(testCfg(), requestWithToken(token))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, seen)
}

func TestAuthMiddleware_RejectsRSAToken(t *testing.T) {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	claims := validClaims()
	claims.Roles = []types.Role{types.ROLE_ADMIN}

	token, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(key)
	require.NoError(t, err)

	w, seen := run(testCfg(), requestWithToken(token))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Nil(t, seen)
}

func TestRBACMiddleware(t *testing.T) {
	tests := []struct {
		name     string
		roles    []types.Role
		withUser bool
		expected int
	}{
		{name: "admin átmegy", roles: []types.Role{types.ROLE_ADMIN}, withUser: true, expected: http.StatusOK},
		{
			name:     "több szerepkör közül az egyik illeszkedik",
			roles:    []types.Role{types.ROLE_USER, types.ROLE_ADMIN},
			withUser: true,
			expected: http.StatusOK,
		},
		{name: "sima user tiltva", roles: []types.Role{types.ROLE_USER}, withUser: true, expected: http.StatusForbidden},
		{name: "üres szerepkörlista", roles: []types.Role{}, withUser: true, expected: http.StatusUnauthorized},
		{name: "nincs bejelentkezve", withUser: false, expected: http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest(http.MethodPost, "/posts", nil)
			if tt.withUser {
				claims := validClaims()
				claims.Roles = tt.roles
				req, _ = requestWithClaims(req, &claims)
			}

			w := httptest.NewRecorder()
			RBACMiddleware(types.ROLE_ADMIN)(next).ServeHTTP(w, req)

			assert.Equal(t, tt.expected, w.Code)
			assert.Equal(t, tt.expected == http.StatusOK, called)
		})
	}
}

func requestWithClaims(req *http.Request, claims *UserClaims) (*http.Request, *UserClaims) {
	return req.WithContext(context.WithValue(req.Context(), userContextKey, claims)), claims
}
