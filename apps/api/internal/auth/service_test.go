package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/middleware"
	"speedcubinghungary/internal/types"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) Upsert(user *User) (*User, error) {
	args := m.Called(user)
	userRet := args.Get(0)
	if userRet == nil {
		return nil, args.Error(1)
	}
	return userRet.(*User), args.Error(1)
}

func (m *MockRepo) GetSession(ctx context.Context, userID int) (middleware.Session, error) {
	args := m.Called(userID)
	return args.Get(0).(middleware.Session), args.Error(1)
}

func (m *MockRepo) RevokeSessions(ctx context.Context, userID int) error {
	return m.Called(userID).Error(0)
}

func (m *MockRepo) SaveGrant(ctx context.Context, grant WCAGrant, wcaID string) error {
	return m.Called(grant, wcaID).Error(0)
}

func (m *MockRepo) GetGrant(ctx context.Context, userID int) (WCAGrant, error) {
	args := m.Called(userID)
	return args.Get(0).(WCAGrant), args.Error(1)
}

func (m *MockRepo) UpdateGrant(ctx context.Context, grant WCAGrant) error {
	return m.Called(grant).Error(0)
}

func (m *MockRepo) DeleteGrant(ctx context.Context, userID int) error {
	return m.Called(userID).Error(0)
}

const testJWTSecret = "test-secret-that-is-at-least-32-chars-long"

func testConfig() *config.Config {
	return &config.Config{
		Deployment:      "development",
		FrontendURL:     "http://localhost:3000",
		JWTSecret:       testJWTSecret,
		WCAClientID:     "test-client-id",
		WCAClientSecret: "test-client-secret",
		WCARedirectURI:  "http://localhost:8080/login/callback",
	}
}

func newTestService(repo Repository, tokenURL, meURL string) *service {
	return &service{
		repo:         repo,
		cfg:          testConfig(),
		httpClient:   &http.Client{},
		authorizeURL: wcaAuthorizeURL,
		tokenURL:     tokenURL,
		meURL:        meURL,
	}
}

func TestService_LoginCallback(t *testing.T) {
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.NoError(t, r.ParseForm())
		assert.Equal(t, "test-client-id", r.Form.Get("client_id"))
		assert.Equal(t, "test-client-secret", r.Form.Get("client_secret"))
		assert.Equal(t, "dummy-code", r.Form.Get("code"))
		assert.Equal(t, "http://localhost:8080/login/callback", r.Form.Get("redirect_uri"))

		json.NewEncoder(w).Encode(map[string]string{"access_token": "mock-token"})
	}))
	defer tokenServer.Close()

	meServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "Bearer mock-token", r.Header.Get("Authorization"))

		json.NewEncoder(w).Encode(map[string]any{
			"me": map[string]any{
				"id":     123,
				"wca_id": "2019ABCD01",
				"name":   "John Doe",
				"email":  "john@example.com",
				"avatar": map[string]any{"url": "https://example.com/avatar.png"},
			},
		})
	}))
	defer meServer.Close()

	existingUser := &User{
		ID:        1,
		WCAID:     "2019ABCD01",
		Name:      "John Doe",
		Email:     "john@example.com",
		AvatarURL: "https://example.com/avatar.png",
		Roles:     []types.Role{types.ROLE_ADMIN},

		SessionVersion: 4,
	}

	mockRepo := new(MockRepo)
	mockRepo.On("Upsert", mock.AnythingOfType("*auth.User")).Return(existingUser, nil)

	svc := newTestService(mockRepo, tokenServer.URL, meServer.URL)

	token, err := svc.LoginCallback(context.Background(), "dummy-code")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	parsed, err := jwt.ParseWithClaims(token, &middleware.UserClaims{}, func(*jwt.Token) (any, error) {
		return []byte(testJWTSecret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
	assert.NoError(t, err)

	claims := parsed.Claims.(*middleware.UserClaims)
	assert.Equal(t, existingUser.ID, claims.UserID)
	assert.Equal(t, existingUser.WCAID, claims.WCAID)
	assert.Equal(t, existingUser.AvatarURL, claims.AvatarURL)
	assert.Contains(t, claims.Roles, types.ROLE_ADMIN)
	assert.NotNil(t, claims.ExpiresAt, "a token lejárat nélkül örök életű munkamenet lenne")
	assert.Equal(t, 4, claims.SessionVersion, "a verzió nélkül a kijelentkezés nem érvénytelenítené a tokent")
	assert.WithinDuration(t, time.Now().Add(12*time.Hour), claims.ExpiresAt.Time, time.Minute)

	mockRepo.AssertExpectations(t)
}

func TestService_IsTheSessionStoreOfTheAuthMiddleware(t *testing.T) {
	mockRepo := new(MockRepo)
	stored := middleware.Session{Version: 2, Roles: []types.Role{types.ROLE_ADMIN}}
	mockRepo.On("GetSession", 7).Return(stored, nil)

	var sessions middleware.SessionStore = newTestService(mockRepo, "", "")
	session, err := sessions.GetSession(context.Background(), 7)

	assert.NoError(t, err)
	assert.Equal(t, stored, session)
	mockRepo.AssertExpectations(t)
}

func TestService_LogoutRevokesTheUsersSessions(t *testing.T) {
	mockRepo := new(MockRepo)
	mockRepo.On("RevokeSessions", 7).Return(nil)

	err := newTestService(mockRepo, "", "").Logout(context.Background(), 7)

	assert.NoError(t, err)
	mockRepo.AssertExpectations(t)
}

func TestService_LoginCallback_RejectsFailedUpstream(t *testing.T) {
	tests := []struct {
		name       string
		tokenSetup func(w http.ResponseWriter)
		meSetup    func(w http.ResponseWriter)
	}{
		{
			name: "a token-beváltás 401-et ad",
			tokenSetup: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
			},
			meSetup: func(w http.ResponseWriter) { t.Error("/me-t nem lett volna szabad meghívni") },
		},
		{
			name: "a token-beváltás 200-at ad access_token nélkül",
			tokenSetup: func(w http.ResponseWriter) {
				json.NewEncoder(w).Encode(map[string]string{})
			},
			meSetup: func(w http.ResponseWriter) { t.Error("/me-t nem lett volna szabad meghívni") },
		},
		{
			name: "a /me 401-et ad",
			tokenSetup: func(w http.ResponseWriter) {
				json.NewEncoder(w).Encode(map[string]string{"access_token": "mock-token"})
			},
			meSetup: func(w http.ResponseWriter) {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "not authenticated"})
			},
		},
		{
			name: "a /me wca_id nélküli fiókot ad",
			tokenSetup: func(w http.ResponseWriter) {
				json.NewEncoder(w).Encode(map[string]string{"access_token": "mock-token"})
			},
			meSetup: func(w http.ResponseWriter) {
				json.NewEncoder(w).Encode(map[string]any{
					"me": map[string]any{"id": 1, "wca_id": "", "name": "Newcomer"},
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.tokenSetup(w)
			}))
			defer tokenServer.Close()

			meServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				tt.meSetup(w)
			}))
			defer meServer.Close()

			mockRepo := new(MockRepo)
			svc := newTestService(mockRepo, tokenServer.URL, meServer.URL)

			token, err := svc.LoginCallback(context.Background(), "dummy-code")

			assert.Error(t, err, "hibás WCA válaszra nem szabad bejelentkeztetni")
			assert.Empty(t, token)
			mockRepo.AssertNotCalled(t, "Upsert", mock.Anything)
			mockRepo.AssertExpectations(t)
		})
	}
}

func TestService_AuthorizeURL(t *testing.T) {
	svc := newTestService(new(MockRepo), wcaTokenURL, wcaMeURL)

	parsed, err := url.Parse(svc.AuthorizeURL("nonce-123"))
	assert.NoError(t, err)

	query := parsed.Query()
	assert.Equal(t, "test-client-id", query.Get("client_id"))
	assert.Equal(t, "code", query.Get("response_type"))
	assert.Equal(t, "nonce-123", query.Get("state"))
	assert.Equal(t, "http://localhost:8080/login/callback", query.Get("redirect_uri"))

	assert.Equal(t, "public email", query.Get("scope"))
	assert.NotContains(t, query.Get("scope"), "manage_competitions")
}
