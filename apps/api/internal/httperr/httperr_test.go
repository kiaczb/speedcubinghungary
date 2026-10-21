package httperr

import (
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestWrite_HidesServerErrorsFromTheClient(t *testing.T) {
	var logged strings.Builder
	log.SetOutput(&logged)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	leaky := errors.New(`pq: column "secret_token" of relation "users" does not exist`)

	for _, status := range []int{
		http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
	} {
		recorder := httptest.NewRecorder()
		Write(recorder, httptest.NewRequest(http.MethodGet, "/posts/42", nil), status, leaky)

		require.Equal(t, status, recorder.Code)
		assert.NotContains(t, recorder.Body.String(), "secret_token",
			"a nyers hiba szövege nem mehet ki a kliensnek")
		assert.NotEmpty(t, strings.TrimSpace(recorder.Body.String()),
			"a kliens azért kapjon valamilyen üzenetet")
	}

	assert.Contains(t, logged.String(), "secret_token",
		"a részletes hibát a szerver naplójában KELL látni, különben vakon nyomozunk")
	assert.Contains(t, logged.String(), "/posts/42", "a naplóban ott a kérés útvonala is")
}

func TestWrite_PassesClientErrorsThrough(t *testing.T) {
	recorder := httptest.NewRecorder()

	Write(recorder, httptest.NewRequest(http.MethodGet, "/posts/abc", nil),
		http.StatusNotFound, errors.New("post not found"))

	assert.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Contains(t, recorder.Body.String(), "post not found")
}
