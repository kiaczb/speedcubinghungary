package auth

import (
	"context"
	"encoding/json"
	"errors"
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

func (m *MockService) Logout(ctx context.Context, userID int) error {
	return m.Called(userID).Error(0)
}

func (m *MockService) GetSession(ctx context.Context, userID int) (middleware.Session, error) {
	args := m.Called(userID)
	return args.Get(0).(middleware.Session), args.Error(1)
}

func (m *MockService) GrantAuthorizeURL(state string) string {
	return m.Called(state).String(0)
}

func (m *MockService) CompleteGrant(ctx context.Context, userID int, code string) error {
	return m.Called(userID, code).Error(0)
}

func (m *MockService) WCAAccessToken(ctx context.Context, userID int) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

func (m *MockService) RevokeGrant(ctx context.Context, userID int) error {
	return m.Called(userID).Error(0)
}

func newTestHandler(svc Service) *Handler {
	return NewHandler(svc, testConfig())
}

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
	assert.Equal(t, http.SameSiteLaxMode, state.SameSite)
	assert.Equal(t, "/", state.Path)

	nonce, redirectPath, found := strings.Cut(state.Value, ":")
	assert.True(t, found)
	assert.Len(t, nonce, 64, "32 bájt hex-ként 64 karakter")
	assert.Equal(t, "/posts", redirectPath)

	mockSvc.AssertCalled(t, "AuthorizeURL", nonce)
}

func TestHandler_Login_RejectsOpenRedirectTargets(t *testing.T) {
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
	assert.Equal(t, "http://localhost:3000/posts", resp.Header.Get("Location"))

	auth := cookieByName(resp, authCookieName)
	require.NotNil(t, auth)
	assert.Equal(t, "mock-jwt-token", auth.Value)
	assert.True(t, auth.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, auth.SameSite)
	assert.Equal(t, 12*60*60, auth.MaxAge)

	state := cookieByName(resp, stateCookieName)
	require.NotNil(t, state)
	assert.Equal(t, -1, state.MaxAge)

	mockSvc.AssertExpectations(t)
}

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

	assert.Equal(t, float64(user.UserID), body["user_id"])
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
	assert.True(t, auth.HttpOnly)
	assert.Equal(t, http.SameSiteStrictMode, auth.SameSite)
	assert.Equal(t, "/", auth.Path)

	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, "logged out", body["message"])
}

func withUser(req *http.Request, userID int) *http.Request {
	claims := &middleware.UserClaims{UserID: userID}
	return req.WithContext(context.WithValue(req.Context(), middleware.UserClaimsContextKey, claims))
}

func TestHandler_LogoutRevokesTheSessionServerSide(t *testing.T) {
	svc := new(MockService)
	svc.On("Logout", 7).Return(nil)

	w := httptest.NewRecorder()
	newTestHandler(svc).Logout(w, withUser(httptest.NewRequest(http.MethodPost, "/logout", nil), 7))

	assert.Equal(t, http.StatusOK, w.Code)
	svc.AssertExpectations(t)
}

func TestHandler_LogoutClearsTheCookieEvenIfRevocationFails(t *testing.T) {
	svc := new(MockService)
	svc.On("Logout", 7).Return(errors.New("pq: connection refused"))

	w := httptest.NewRecorder()
	newTestHandler(svc).Logout(w, withUser(httptest.NewRequest(http.MethodPost, "/logout", nil), 7))

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode, "a kliens tudja meg, hogy a token nincs visszavonva")
	auth := cookieByName(resp, authCookieName)
	require.NotNil(t, auth)
	assert.Equal(t, -1, auth.MaxAge)
}

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
