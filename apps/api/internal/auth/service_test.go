package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"speedcubinghungary/internal/types"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// 1️⃣ Mock Repository
type MockRepo struct {
	mock.Mock
}

func (m *MockRepo) GetByWCAID(wcaID string) (*User, error) {
	args := m.Called(wcaID)
	user := args.Get(0)
	if user == nil {
		return nil, args.Error(1)
	}
	return user.(*User), args.Error(1)
}

func (m *MockRepo) Create(user *User) (*User, error) {
	args := m.Called(user)
	return args.Get(0).(*User), args.Error(1)
}

// 💥 ÚJ: Bekerült az Upsert a mockba!
func (m *MockRepo) Upsert(user *User) (*User, error) {
	args := m.Called(user)
	userRet := args.Get(0)
	if userRet == nil {
		return nil, args.Error(1)
	}
	return userRet.(*User), args.Error(1)
}

// 2️⃣ Teszt
func TestService_LoginCallback(t *testing.T) {
	// Környezeti változók
	t.Setenv("JWT_SECRET", "testsecret")
	t.Setenv("WCA_CLIENT_ID", "testid")
	t.Setenv("WCA_CLIENT_SECRET", "testsecret")

	// 3️⃣ Mock WCA token endpoint
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{
			"access_token": "mock-token",
		})
	}))
	defer tokenServer.Close()

	// 4️⃣ Mock WCA /me endpoint
	meServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		assert.Equal(t, "Bearer mock-token", auth) // ellenőrizzük header-t

		json.NewEncoder(w).Encode(map[string]interface{}{
			"me": map[string]interface{}{
				"id":     123,
				"wca_id": "2019ABCD01",
				"name":   "John Doe",
				"email":  "john@example.com",
			},
		})
	}))
	defer meServer.Close()

	// 5️⃣ Mock repository
	mockRepo := new(MockRepo)

	existingUser := &User{
		ID:    1,
		WCAID: "2019ABCD01",
		Name:  "John Doe",
		Email: "john@example.com",
		Roles: []types.Role{types.ROLE_ADMIN}, // Itt szimuláljuk, hogy admin
	}

	// 💥 ÚJ: Lecseréltük a GetByWCAID + Create párost az Upsert-re.
	// Azt mondjuk a mocknak: Ha meghívják az Upsert-et bármilyen User pointerrel,
	// akkor sikeresen add vissza a meglévő, ROLE_ADMIN joggal rendelkező userünket.
	mockRepo.On("Upsert", mock.AnythingOfType("*auth.User")).Return(existingUser, nil)

	// 6️⃣ Service létrehozása
	svc := &service{
		repo:       mockRepo,
		httpClient: &http.Client{},
		tokenURL:   tokenServer.URL,
		meURL:      meServer.URL,
	}

	// 7️⃣ Ideiglenes transport felülírás (maradhat, de a mock szerver URL-jei miatt amúgy is működne)
	origTransport := http.DefaultTransport
	http.DefaultTransport = &http.Transport{} // dummy
	defer func() { http.DefaultTransport = origTransport }()

	// 8️⃣ Teszt futtatás
	token, err := svc.LoginCallback("dummy-code")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	// 9️⃣ JWT ellenőrzés
	parsedToken, err := jwt.ParseWithClaims(token, &UserClaims{}, func(t *jwt.Token) (interface{}, error) {
		return []byte("testsecret"), nil
	})
	assert.NoError(t, err)

	claims := parsedToken.Claims.(*UserClaims)

	// Ellenőrizzük, hogy valóban az a user került-e be, amit a mock visszaadott
	assert.Equal(t, existingUser.ID, claims.UserID)
	assert.Contains(t, claims.Roles, types.ROLE_ADMIN)

	mockRepo.AssertExpectations(t)
}
