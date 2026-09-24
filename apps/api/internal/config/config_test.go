package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setRequiredEnv beállítja a minimális, érvényes környezetet.
func setRequiredEnv(t *testing.T) {
	t.Helper()

	t.Setenv("DEPLOYMENT", "production")
	t.Setenv("FRONTEND_URL", "https://example.com")
	t.Setenv("DATABASE_URL", "postgres://user:pass@db:5432/app?sslmode=disable")
	t.Setenv("JWT_SECRET", "a-secret-that-is-at-least-32-characters")
	t.Setenv("WCA_CLIENT_ID", "client-id")
	t.Setenv("WCA_CLIENT_SECRET", "client-secret")
	t.Setenv("WCA_REDIRECT_URI", "https://example.com/api/login/callback")
}

func TestLoad_Valid(t *testing.T) {
	setRequiredEnv(t)

	cfg, err := Load()
	require.NoError(t, err)

	assert.Equal(t, "production", cfg.Deployment)
	assert.True(t, cfg.IsProduction())
	assert.False(t, cfg.IsDevelopment())
	// A nem kötelező változóknak marad alapértelmezésük.
	assert.Equal(t, "/data/uploads", cfg.UploadDir)
	assert.Equal(t, "https://live.worldcubeassociation.org", cfg.WCALiveURL)
}

// ⚠️ EZ A KETTŐ VÉDI A COOKIE-BIZTONSÁGOT.
//
// Korábban a DEPLOYMENT-nek "development", a FRONTEND_URL-nek
// "http://localhost:3000" alapértelmezése volt, ezért a "kötelező" ellenőrzésük
// SOSEM csapott le. Egy prodban lemaradt DEPLOYMENT így nem leállást okozott,
// hanem némán fejlesztői módot — vagyis `Secure` NÉLKÜLI auth sütit —, a
// hiányzó FRONTEND_URL pedig localhostra engedte a CORS-t és oda irányított
// bejelentkezés után.
func TestLoad_FailsWithoutDeployment(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("DEPLOYMENT", "")

	_, err := Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DEPLOYMENT")
}

func TestLoad_FailsWithoutFrontendURL(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("FRONTEND_URL", "")

	_, err := Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "FRONTEND_URL")
}

// Egy elgépelt DEPLOYMENT nem üres, tehát a "kötelező" ellenőrzésen átcsúszna,
// az IsProduction() viszont hamisat adna rá — az elgépelés némán viselkedés-
// változás lenne (pl. Secure süti helyett nem-Secure).
func TestLoad_FailsOnUnknownDeployment(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("DEPLOYMENT", "produciton")

	_, err := Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DEPLOYMENT")
}

func TestLoad_FailsOnShortJWTSecret(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("JWT_SECRET", "too-short")

	_, err := Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "JWT_SECRET")
}

func TestLoad_ListsEveryMissingRequiredVariable(t *testing.T) {
	setRequiredEnv(t)
	t.Setenv("DATABASE_URL", "")
	t.Setenv("WCA_CLIENT_ID", "")

	_, err := Load()

	require.Error(t, err)
	assert.Contains(t, err.Error(), "DATABASE_URL")
	assert.Contains(t, err.Error(), "WCA_CLIENT_ID")
}
