package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"speedcubinghungary/internal/config"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLimiter(limit int, trustForwardedFor bool) *rateLimiter {
	return &rateLimiter{
		limit:             limit,
		window:            time.Minute,
		trustForwardedFor: trustForwardedFor,
		buckets:           make(map[string]*rateBucket),
	}
}

func TestRateLimiter_AllowsUpToTheLimitThenRefuses(t *testing.T) {
	limiter := newTestLimiter(3, false)
	now := time.Now()

	for i := 1; i <= 3; i++ {
		_, ok := limiter.allow("1.2.3.4", now)
		assert.True(t, ok, "a %d. kérés még belefér", i)
	}

	retryAfter, ok := limiter.allow("1.2.3.4", now)
	require.False(t, ok, "a limit fölötti kérést el kell utasítani")
	assert.Positive(t, retryAfter, "a Retry-After mondja meg, mennyit kell várni")
}

// A korlát IP-nként külön él: egy zajos kliens nem zárhatja ki a többieket.
func TestRateLimiter_IsScopedToTheClient(t *testing.T) {
	limiter := newTestLimiter(1, false)
	now := time.Now()

	_, ok := limiter.allow("1.2.3.4", now)
	require.True(t, ok)
	_, ok = limiter.allow("1.2.3.4", now)
	require.False(t, ok)

	_, ok = limiter.allow("5.6.7.8", now)
	assert.True(t, ok, "a másik kliens kvótája érintetlen")
}

func TestRateLimiter_ResetsAfterTheWindow(t *testing.T) {
	limiter := newTestLimiter(1, false)
	now := time.Now()

	_, ok := limiter.allow("1.2.3.4", now)
	require.True(t, ok)
	_, ok = limiter.allow("1.2.3.4", now)
	require.False(t, ok)

	_, ok = limiter.allow("1.2.3.4", now.Add(time.Minute+time.Second))
	assert.True(t, ok, "az ablak lejártával nulláról indul a számláló")
}

// A lejárt vödröket el kell dobni, különben a map minden valaha látott IP-t
// megtart — egy sok forrásból jövő szkennelés maga lenne a memóriaszivárgás.
func TestRateLimiter_ForgetsExpiredClients(t *testing.T) {
	limiter := newTestLimiter(1, false)
	now := time.Now()

	for _, ip := range []string{"1.1.1.1", "2.2.2.2", "3.3.3.3"} {
		limiter.allow(ip, now)
	}
	require.Len(t, limiter.buckets, 3)

	limiter.allow("4.4.4.4", now.Add(2*time.Minute))

	assert.Len(t, limiter.buckets, 1, "csak az él marad meg")
}

// ⚠️ Az X-Forwarded-For UTOLSÓ eleme a valódi forrás: a kliens saját fejlécéhez
// a Caddy hozzáfűzi, amit ő látott. Az elsőt venni azt jelentené, hogy egyetlen
// hamisított fejléccel bárki végtelen kvótát kap.
func TestRateLimiter_ClientKeyIgnoresSpoofedForwardedFor(t *testing.T) {
	limiter := newTestLimiter(1, true)

	request := httptest.NewRequest(http.MethodGet, "/login", nil)
	request.RemoteAddr = "10.0.0.1:34567"
	request.Header.Set("X-Forwarded-For", "9.9.9.9, 203.0.113.7")

	assert.Equal(t, "203.0.113.7", limiter.clientKey(request))
}

// Fejlesztésben nincs proxy, tehát a fejléc szabadon hamisítható — ott a
// kapcsolat tényleges forráscíme számít.
func TestRateLimiter_ClientKeyIgnoresForwardedForWithoutAProxy(t *testing.T) {
	limiter := newTestLimiter(1, false)

	request := httptest.NewRequest(http.MethodGet, "/login", nil)
	request.RemoteAddr = "10.0.0.1:34567"
	request.Header.Set("X-Forwarded-For", "9.9.9.9")

	assert.Equal(t, "10.0.0.1", limiter.clientKey(request))
}

func TestRateLimit_Answers429WithRetryAfter(t *testing.T) {
	handler := RateLimit(&config.Config{Deployment: "development"}, 1)(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}),
	)

	call := func() *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, "/posts", nil)
		request.RemoteAddr = "10.0.0.1:34567"

		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, request)
		return recorder
	}

	require.Equal(t, http.StatusOK, call().Code)

	blocked := call()
	assert.Equal(t, http.StatusTooManyRequests, blocked.Code)
	assert.NotEmpty(t, blocked.Header().Get("Retry-After"))
}
