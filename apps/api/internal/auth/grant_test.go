package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"speedcubinghungary/internal/middleware"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type fakeGrantWCA struct {
	server *httptest.Server

	mu            sync.Mutex
	authorizeSeen url.Values
	tokenForms    []url.Values
	tokenStatus   int
}

func newFakeGrantWCA(t *testing.T) *fakeGrantWCA {
	t.Helper()

	wca := &fakeGrantWCA{tokenStatus: http.StatusOK}
	mux := http.NewServeMux()

	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		wca.mu.Lock()
		wca.authorizeSeen = r.URL.Query()
		wca.mu.Unlock()

		back, err := url.Parse(r.URL.Query().Get("redirect_uri"))
		require.NoError(t, err)
		query := back.Query()
		query.Set("code", "grant-code")
		query.Set("state", r.URL.Query().Get("state"))
		back.RawQuery = query.Encode()
		http.Redirect(w, r, back.String(), http.StatusFound)
	})

	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, r.ParseForm())
		wca.mu.Lock()
		wca.tokenForms = append(wca.tokenForms, r.Form)
		status := wca.tokenStatus
		wca.mu.Unlock()

		if status != http.StatusOK {
			w.WriteHeader(status)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  "uj-access",
			"refresh_token": "uj-refresh",
			"expires_in":    7200,
			"scope":         "public email manage_competitions",
		})
	})

	mux.HandleFunc("/api/v0/me", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]any{
			"me": map[string]any{"id": 42, "wca_id": "2019ABCD01", "name": "Admin Anna"},
		})
	})

	wca.server = httptest.NewServer(mux)
	t.Cleanup(wca.server.Close)
	return wca
}

func grantService(repo Repository, wca *fakeGrantWCA) *service {
	return &service{
		repo:         repo,
		cfg:          testConfig(),
		httpClient:   &http.Client{},
		authorizeURL: wca.server.URL + "/oauth/authorize",
		tokenURL:     wca.server.URL + "/oauth/token",
		meURL:        wca.server.URL + "/api/v0/me",
	}
}

func asAdmin(userID int, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		claims := &middleware.UserClaims{UserID: userID}
		next(w, r.WithContext(context.WithValue(r.Context(), middleware.UserClaimsContextKey, claims)))
	}
}

func runGrantFlow(t *testing.T, repo *MockRepo) (*url.URL, []*http.Cookie, *fakeGrantWCA, *service) {
	t.Helper()

	wca := newFakeGrantWCA(t)
	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(frontend.Close)

	svc := grantService(repo, wca)
	svc.cfg.FrontendURL = frontend.URL
	handler := NewHandler(svc, svc.cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/wca/authorize", asAdmin(7, handler.AuthorizeWCA))
	mux.HandleFunc("/login/callback", handler.LoginCallback)
	api := httptest.NewServer(mux)
	t.Cleanup(api.Close)
	svc.cfg.WCARedirectURI = api.URL + "/login/callback"

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{Jar: jar}

	resp, err := client.Get(api.URL + "/wca/authorize?redirect=" + url.QueryEscape("/admin/stats?competition=Teszt2026"))
	require.NoError(t, err)
	resp.Body.Close()

	apiURL, err := url.Parse(api.URL)
	require.NoError(t, err)
	return resp.Request.URL, jar.Cookies(apiURL), wca, svc
}

func TestGrantFlow_StoresEncryptedTokensForTheAdmin(t *testing.T) {
	repo := new(MockRepo)
	var saved WCAGrant
	repo.On("SaveGrant", mock.AnythingOfType("auth.WCAGrant"), "2019ABCD01").
		Run(func(args mock.Arguments) { saved = args.Get(0).(WCAGrant) }).
		Return(nil)

	final, cookies, wca, svc := runGrantFlow(t, repo)

	wca.mu.Lock()
	assert.Equal(t, "public email manage_competitions", wca.authorizeSeen.Get("scope"))
	wca.mu.Unlock()

	assert.Equal(t, "/admin/stats", final.Path)
	assert.Equal(t, "Teszt2026", final.Query().Get("competition"))
	assert.Empty(t, final.Query().Get("wca"))

	for _, c := range cookies {
		assert.NotEqual(t, authCookieName, c.Name, "az engedélyezés nem hoz létre új munkamenetet")
		assert.NotEqual(t, grantStateCookieName, c.Name, "a state sütit a callback törli")
	}

	repo.AssertExpectations(t)
	assert.Equal(t, 7, saved.UserID, "a state sütiben indító admin kapja az engedélyt")
	assert.NotContains(t, string(saved.AccessToken), "uj-access", "a token titkosítva kerül a DB-be")

	aead, err := svc.grantCipher()
	require.NoError(t, err)
	access, err := open(aead, saved.AccessToken, 7)
	require.NoError(t, err)
	assert.Equal(t, "uj-access", access)
	refresh, err := open(aead, saved.RefreshToken, 7)
	require.NoError(t, err)
	assert.Equal(t, "uj-refresh", refresh)

	_, err = open(aead, saved.AccessToken, 8)
	assert.Error(t, err, "más felhasználó sorába másolva nem nyitható ki")
}

func TestGrantFlow_RejectsAnotherWCAAccount(t *testing.T) {
	repo := new(MockRepo)
	repo.On("SaveGrant", mock.Anything, "2019ABCD01").Return(ErrGrantOwnerMismatch)

	final, _, _, _ := runGrantFlow(t, repo)

	assert.Equal(t, "/admin/stats", final.Path)
	assert.Equal(t, "failed", final.Query().Get("wca"))
}

func storedGrant(t *testing.T, svc *service, expiresAt time.Time) WCAGrant {
	t.Helper()
	grant, err := svc.sealGrant(7, &tokenResponse{AccessToken: "regi-access", ExpiresIn: 7200}, "regi-refresh")
	require.NoError(t, err)
	grant.ExpiresAt = expiresAt
	return grant
}

func TestWCAAccessToken_UsesTheStoredTokenWhileValid(t *testing.T) {
	wca := newFakeGrantWCA(t)
	repo := new(MockRepo)
	svc := grantService(repo, wca)
	repo.On("GetGrant", 7).Return(storedGrant(t, svc, time.Now().Add(time.Hour)), nil)

	token, err := svc.WCAAccessToken(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, "regi-access", token)

	wca.mu.Lock()
	assert.Empty(t, wca.tokenForms, "érvényes tokennél nem frissítünk")
	wca.mu.Unlock()
}

func TestWCAAccessToken_RefreshesAnExpiringToken(t *testing.T) {
	wca := newFakeGrantWCA(t)
	repo := new(MockRepo)
	svc := grantService(repo, wca)
	repo.On("GetGrant", 7).Return(storedGrant(t, svc, time.Now().Add(30*time.Second)), nil)
	var updated WCAGrant
	repo.On("UpdateGrant", mock.AnythingOfType("auth.WCAGrant")).
		Run(func(args mock.Arguments) { updated = args.Get(0).(WCAGrant) }).
		Return(nil)

	token, err := svc.WCAAccessToken(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, "uj-access", token)

	wca.mu.Lock()
	require.Len(t, wca.tokenForms, 1)
	assert.Equal(t, "refresh_token", wca.tokenForms[0].Get("grant_type"))
	assert.Equal(t, "regi-refresh", wca.tokenForms[0].Get("refresh_token"))
	wca.mu.Unlock()

	aead, err := svc.grantCipher()
	require.NoError(t, err)
	refresh, err := open(aead, updated.RefreshToken, 7)
	require.NoError(t, err)
	assert.Equal(t, "uj-refresh", refresh, "a WCA által rotált refresh token kerül mentésre")
	assert.True(t, updated.ExpiresAt.After(time.Now().Add(time.Hour)))
}

func TestWCAAccessToken_RevokedRefreshTokenDropsTheGrant(t *testing.T) {
	wca := newFakeGrantWCA(t)
	wca.tokenStatus = http.StatusBadRequest
	repo := new(MockRepo)
	svc := grantService(repo, wca)
	repo.On("GetGrant", 7).Return(storedGrant(t, svc, time.Now().Add(-time.Minute)), nil)
	repo.On("DeleteGrant", 7).Return(nil)

	_, err := svc.WCAAccessToken(context.Background(), 7)
	assert.ErrorIs(t, err, ErrNoGrant, "visszavont engedélynél újra kell engedélyezni")
	repo.AssertExpectations(t)
}

func TestVerifyGrantState_RejectsTamperedCookies(t *testing.T) {
	cases := map[string]string{
		"nem egyező nonce":     "masik:7:/admin/stats",
		"hiányzó felhasználó":  "abc:/admin/stats",
		"nem szám felhasználó": "abc:admin:/admin/stats",
		"nulla felhasználó":    "abc:0:/admin/stats",
		"negatív felhasználó":  "abc:-3:/admin/stats",
		"üres süti":            "",
	}

	for name, value := range cases {
		req := httptest.NewRequest(http.MethodGet, "/login/callback?state=abc", nil)
		req.AddCookie(&http.Cookie{Name: grantStateCookieName, Value: value})
		_, ok := verifyGrantState(req, "abc")
		assert.False(t, ok, name)
	}

	req := httptest.NewRequest(http.MethodGet, "/login/callback?state=abc", nil)
	req.AddCookie(&http.Cookie{Name: grantStateCookieName, Value: "abc:7://evil.example"})
	state, ok := verifyGrantState(req, "abc")
	require.True(t, ok)
	assert.Equal(t, grantState{userID: 7, redirectPath: "/"}, state, "a visszatérési útvonalat itt is tisztítjuk")
}
