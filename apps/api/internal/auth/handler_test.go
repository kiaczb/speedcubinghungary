package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"speedcubinghungary/internal/middleware"
	"speedcubinghungary/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) AuthorizeURL(state string) string {
	args := m.Called(state)
	return args.String(0)
}

func (m *MockService) LoginCallback(ctx context.Context, code string) (string, error) {
	args := m.Called(code)
	return args.String(0), args.Error(1)
}

func newTestHandler(svc Service) *Handler {
	return NewHandler(svc, testConfig())
}

// cookieByName kikeresi a válaszban beállított sütit.
func cookieByName(resp *http.Response, name string) *http.Cookie {
	for _, c := range resp.Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

func TestHandler_Login_SetsStateCookieAndRedirects(t *testing.T) {
	mockSvc := new(MockService)
	mockSvc.On("AuthorizeURL", mock.AnythingOfType("string")).
		Return("https://www.worldcubeassociation.org/oauth/authorize?state=x")

	req := httptest.NewRequest(http.MethodGet, "/login?redirect=/posts", nil)
	w := httptest.NewRecorder()

	newTestHandler(mockSvc).Login(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusFound, resp.StatusCode)

	state := cookieByName(resp, stateCookieName)
	require.NotNil(t, state, "a CSRF nonce sütinek léteznie kell")
	assert.True(t, state.HttpOnly, "a nonce-ot JS nem olvashatja")
	// ⚠️ Lax, nem Strict: a WCA-tól visszaérkező kereszt-oldali navigációra a
	// Strict sütit a böngésző NEM küldené vissza, és a belépés mindig elbukna.
	assert.Equal(t, http.SameSiteLaxMode, state.SameSite)
	// ⚠️ Path "/": test/prod alatt a Caddy levágja az /api prefixet, a böngésző
	// viszont /api/login/callback-ot lát — egy /login-ra szűkített süti ott nem
	// érkezne vissza.
	assert.Equal(t, "/", state.Path)

	nonce, redirectPath, found := strings.Cut(state.Value, ":")
	assert.True(t, found)
	assert.Len(t, nonce, 64, "32 bájt hex-ként 64 karakter")
	assert.Equal(t, "/posts", redirectPath)

	// A kiküldött state a nonce, nem az útvonal.
	mockSvc.AssertCalled(t, "AuthorizeURL", nonce)
}

func TestHandler_Login_RejectsOpenRedirectTargets(t *testing.T) {
	// A `redirect` a felhasználótól jön, és a belépés végén Location fejlécbe
	// kerül — a saját oldalunkon kívülre mutató értéket el kell dobni.
	for _, target := range []string{
		"https://evil.example",
		"//evil.example",
		`/\evil.example`,
		`/\/evil.example`,
		"/valid\r\nX-Injected: 1",
		"javascript:alert(1)",
	} {
		t.Run(target, func(t *testing.T) {
			mockSvc := new(MockService)
			mockSvc.On("AuthorizeURL", mock.AnythingOfType("string")).Return("https://wca.example/authorize")

			req := httptest.NewRequest(http.MethodGet, "/login?redirect="+url.QueryEscape(target), nil)
			w := httptest.NewRecorder()

			newTestHandler(mockSvc).Login(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			state := cookieByName(resp, stateCookieName)
			require.NotNil(t, state)

			_, redirectPath, _ := strings.Cut(state.Value, ":")
			assert.Equal(t, "/", redirectPath, "a nem biztonságos célt gyökérre kell cserélni")
		})
	}
}

// startLogin lefuttatja a /login ágat, és visszaadja a state sütit, hogy a
// callback tesztek valódi (nem kézzel gyártott) nonce-szal dolgozhassanak.
func startLogin(t *testing.T, redirectPath string) *http.Cookie {
	t.Helper()

	mockSvc := new(MockService)
	mockSvc.On("AuthorizeURL", mock.AnythingOfType("string")).Return("https://wca.example/authorize")

	req := httptest.NewRequest(http.MethodGet, "/login?redirect="+url.QueryEscape(redirectPath), nil)
	w := httptest.NewRecorder()
	newTestHandler(mockSvc).Login(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	cookie := cookieByName(resp, stateCookieName)
	require.NotNil(t, cookie)

	return cookie
}

func TestHandler_LoginCallback_Success(t *testing.T) {
	stateCookie := startLogin(t, "/posts")
	nonce, _, _ := strings.Cut(stateCookie.Value, ":")

	mockSvc := new(MockService)
	mockSvc.On("LoginCallback", "dummy-code").Return("mock-jwt-token", nil)

	req := httptest.NewRequest(http.MethodGet, "/login/callback?code=dummy-code&state="+nonce, nil)
	req.AddCookie(stateCookie)
	w := httptest.NewRecorder()

	newTestHandler(mockSvc).LoginCallback(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	// A visszairányítás célja a SÜTIBŐL jön, nem a state-ből.
	assert.Equal(t, "http://localhost:3000/posts", resp.Header.Get("Location"))

	auth := cookieByName(resp, authCookieName)
	require.NotNil(t, auth)
	assert.Equal(t, "mock-jwt-token", auth.Value)
	assert.True(t, auth.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, auth.SameSite)
	assert.Equal(t, 3600, auth.MaxAge)

	// A state süti a folyamat végén törlődik.
	state := cookieByName(resp, stateCookieName)
	require.NotNil(t, state)
	assert.Equal(t, -1, state.MaxAge)

	mockSvc.AssertExpectations(t)
}

// ⚠️ EZ A TESZT ŐRZI A CSRF-VÉDELMET.
//
// A `state` a folyamat CSRF-tokenje. Enélkül (vagy nem egyező értékkel) a
// támadó a SAJÁT WCA-fiókjához tartozó code-ot játszhatná be az áldozat
// böngészőjébe, és az áldozat észrevétlenül a támadó fiókjában dolgozna.
// Egyik ág sem válthatja be a kódot, és nem állíthat ki auth sütit.
func TestHandler_LoginCallback_RejectsBadState(t *testing.T) {
	valid := startLogin(t, "/posts")
	validNonce, _, _ := strings.Cut(valid.Value, ":")

	tests := []struct {
		name   string
		state  string
		cookie *http.Cookie
	}{
		{name: "nincs state és nincs süti", state: "", cookie: nil},
		{name: "van süti, de nincs state", state: "", cookie: valid},
		{name: "van state, de nincs süti (CSRF)", state: validNonce, cookie: nil},
		{
			name:   "a state nem egyezik a sütivel",
			state:  "0000000000000000000000000000000000000000000000000000000000000000",
			cookie: valid,
		},
		{
			name:   "üres nonce a sütiben",
			state:  "",
			cookie: &http.Cookie{Name: stateCookieName, Value: ":/posts"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := new(MockService)

			req := httptest.NewRequest(http.MethodGet, "/login/callback?code=dummy-code&state="+tt.state, nil)
			if tt.cookie != nil {
				req.AddCookie(tt.cookie)
			}
			w := httptest.NewRecorder()

			newTestHandler(mockSvc).LoginCallback(w, req)

			resp := w.Result()
			defer resp.Body.Close()

			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "http://localhost:3000/?login=failed", resp.Header.Get("Location"))
			assert.Nil(t, cookieByName(resp, authCookieName), "nem születhet munkamenet")
			mockSvc.AssertNotCalled(t, "LoginCallback", mock.Anything)
		})
	}
}

func TestHandler_LoginCallback_UpstreamFailureDoesNotLeakDetails(t *testing.T) {
	stateCookie := startLogin(t, "/posts")
	nonce, _, _ := strings.Cut(stateCookie.Value, ":")

	mockSvc := new(MockService)
	mockSvc.On("LoginCallback", "dummy-code").
		Return("", assert.AnError)

	req := httptest.NewRequest(http.MethodGet, "/login/callback?code=dummy-code&state="+nonce, nil)
	req.AddCookie(stateCookie)
	w := httptest.NewRecorder()

	newTestHandler(mockSvc).LoginCallback(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://localhost:3000/?login=failed", resp.Header.Get("Location"))
	assert.Nil(t, cookieByName(resp, authCookieName))
	// Az upstream hibaszöveg nem kerülhet a válaszba.
	assert.NotContains(t, w.Body.String(), assert.AnError.Error())
}

func TestHandler_Me(t *testing.T) {
	handler := newTestHandler(new(MockService))

	user := &middleware.UserClaims{
		UserID:    1,
		WCAID:     "2019ABCD01",
		Name:      "John Doe",
		AvatarURL: "https://example.com/avatar.png",
		Roles:     []types.Role{types.ROLE_ADMIN},
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserClaimsContextKey, user))
	w := httptest.NewRecorder()

	handler.Me(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))

	assert.Equal(t, float64(user.UserID), body["user_id"]) // JSON unmarshal miatt float64
	assert.Equal(t, user.WCAID, body["wca_id"])
	assert.Equal(t, user.Name, body["name"])
	assert.Equal(t, user.AvatarURL, body["avatarURL"])

	roles := body["roles"].([]any)
	assert.Contains(t, roles, string(types.ROLE_ADMIN))
}

func TestHandler_Me_Unauthenticated(t *testing.T) {
	w := httptest.NewRecorder()
	newTestHandler(new(MockService)).Me(w, httptest.NewRequest(http.MethodGet, "/me", nil))

	assert.Equal(t, http.StatusUnauthorized, w.Result().StatusCode)
}

func TestHandler_Logout(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	newTestHandler(new(MockService)).Logout(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	auth := cookieByName(resp, authCookieName)
	require.NotNil(t, auth)
	assert.Equal(t, "", auth.Value)
	assert.Equal(t, -1, auth.MaxAge)
	// A törlő sütinek egyeznie kell a kiállítottal, különben a böngésző nem
	// ugyanazt a sütit törli.
	assert.True(t, auth.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, auth.SameSite)
	assert.Equal(t, "/", auth.Path)

	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "logged out", body["message"])
}

// A Secure flag a DEPLOYMENT-ből jön. A config csomag kötelezővé teszi és
// értékkészletre ellenőrzi, így ez nem eshet vissza némán fejlesztői módra.
func TestHandler_CookiesAreSecureOutsideDevelopment(t *testing.T) {
	cfg := testConfig()
	cfg.Deployment = "production"
	handler := NewHandler(new(MockService), cfg)

	w := httptest.NewRecorder()
	handler.Logout(w, httptest.NewRequest(http.MethodPost, "/logout", nil))

	resp := w.Result()
	defer resp.Body.Close()

	auth := cookieByName(resp, authCookieName)
	require.NotNil(t, auth)
	assert.True(t, auth.Secure)
}
