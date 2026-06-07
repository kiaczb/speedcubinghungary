package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"speedcubinghungary/internal/middleware"
	"speedcubinghungary/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockService struct {
	mock.Mock
}

func (m *MockService) LoginCallback(code string) (string, error) {
	args := m.Called(code)
	return args.String(0), args.Error(1)
}

func TestHandler_LoginCallback(t *testing.T) {
	t.Setenv("FRONTEND_URL", "http://localhost:3000")
	t.Setenv("DEPLOYMENT", "development")

	mockSvc := new(MockService)
	mockSvc.On("LoginCallback", "dummy-code").Return("mock-jwt-token", nil)

	handler := NewHandler(mockSvc)

	req := httptest.NewRequest(http.MethodGet, "/login/callback?code=dummy-code", nil)
	w := httptest.NewRecorder()

	handler.LoginCallback(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusFound, resp.StatusCode)
	assert.Equal(t, "http://localhost:3000", resp.Header.Get("Location"))

	cookies := resp.Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "auth", cookies[0].Name)
	assert.Equal(t, "mock-jwt-token", cookies[0].Value)
	assert.True(t, cookies[0].HttpOnly)

	mockSvc.AssertExpectations(t)
}

func TestHandler_Me(t *testing.T) {
	handler := &Handler{}

	// Fake user létrehozása
	user := &middleware.UserClaims{
		UserID: 1,
		WCAID:  "2019ABCD01",
		Name:   "John Doe",
		Roles:  []types.Role{types.ROLE_ADMIN},
	}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	w := httptest.NewRecorder()

	// Context-be tesszük a user-t közvetlenül
	ctx := context.WithValue(req.Context(), middleware.UserClaimsContextKey, user)
	req = req.WithContext(ctx)

	handler.Me(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&body)

	assert.Equal(t, float64(user.UserID), body["user_id"]) // JSON unmarshal miatt float64
	assert.Equal(t, user.WCAID, body["wca_id"])
	assert.Equal(t, user.Name, body["name"])

	// ❌ Body["roles"] konvertálása string tömbbé
	rolesInterface := body["roles"].([]interface{})
	roles := make([]string, len(rolesInterface))
	for i, r := range rolesInterface {
		roles[i] = r.(string)
	}

	// Ellenőrzés stringként
	assert.Contains(t, roles, string(types.ROLE_ADMIN))
}


func TestHandler_Logout(t *testing.T) {
	handler := &Handler{}

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	w := httptest.NewRecorder()

	handler.Logout(w, req)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	cookies := resp.Cookies()
	assert.Len(t, cookies, 1)
	assert.Equal(t, "auth", cookies[0].Name)
	assert.Equal(t, "", cookies[0].Value)
	assert.Equal(t, -1, cookies[0].MaxAge)

	var body map[string]string
	json.NewDecoder(resp.Body).Decode(&body)
	assert.Equal(t, "logged out", body["message"])
}
