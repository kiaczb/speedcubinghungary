package router

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"speedcubinghungary/internal/allrounder"
	"speedcubinghungary/internal/auth"
	"speedcubinghungary/internal/competition"
	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/post"
	"speedcubinghungary/internal/prediction"
	"speedcubinghungary/internal/special"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUploadsAreNotServedAsJSON(t *testing.T) {
	uploadDir := t.TempDir()
	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR")
	require.NoError(t, os.WriteFile(filepath.Join(uploadDir, "kep.png"), png, 0o644))

	cfg := &config.Config{Deployment: "development", FrontendURL: "http://localhost:3000", UploadDir: uploadDir}
	r := New(cfg, func(next http.Handler) http.Handler { return next },
		post.NewHandler(nil, uploadDir), competition.NewHandler(nil), prediction.NewHandler(nil),
		auth.NewHandler(nil, cfg), special.NewHandler(nil), allrounder.NewHandler(nil),
	)

	recorder := httptest.NewRecorder()
	r.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/uploads/kep.png", nil))

	require.Equal(t, http.StatusOK, recorder.Code)
	assert.Equal(t, "image/png", recorder.Header().Get("Content-Type"))
	assert.Equal(t, "nosniff", recorder.Header().Get("X-Content-Type-Options"))
}
