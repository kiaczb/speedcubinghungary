package middleware

import (
	"math"
	"net"
	"net/http"
	"speedcubinghungary/internal/config"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	RateLimitWindow = time.Minute

	LoginRateLimit = 60
)

func RateLimit(cfg *config.Config, limit int) func(http.Handler) http.Handler {
	limiter := &rateLimiter{
		limit:             limit,
		window:            RateLimitWindow,
		trustForwardedFor: !cfg.IsDevelopment(),
		buckets:           make(map[string]*rateBucket),
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if retryAfter, ok := limiter.allow(limiter.clientKey(r), time.Now()); !ok {
				w.Header().Set("Retry-After", strconv.Itoa(retryAfter))
				http.Error(w, "too many requests", http.StatusTooManyRequests)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

type rateLimiter struct {
	limit             int
	window            time.Duration
	trustForwardedFor bool

	mu        sync.Mutex
	buckets   map[string]*rateBucket
	nextSweep time.Time
}

type rateBucket struct {
	count   int
	resetAt time.Time
}

func (l *rateLimiter) allow(key string, now time.Time) (retryAfterSeconds int, ok bool) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.sweepLocked(now)

	bucket, known := l.buckets[key]
	if !known || now.After(bucket.resetAt) {
		bucket = &rateBucket{resetAt: now.Add(l.window)}
		l.buckets[key] = bucket
	}

	bucket.count++
	if bucket.count > l.limit {
		retryAfter := int(math.Ceil(bucket.resetAt.Sub(now).Seconds()))
		if retryAfter < 1 {
			retryAfter = 1
		}
		return retryAfter, false
	}

	return 0, true
}

func (l *rateLimiter) sweepLocked(now time.Time) {
	if now.Before(l.nextSweep) {
		return
	}
	l.nextSweep = now.Add(l.window)

	for key, bucket := range l.buckets {
		if now.After(bucket.resetAt) {
			delete(l.buckets, key)
		}
	}
}

func (l *rateLimiter) clientKey(r *http.Request) string {
	if l.trustForwardedFor {
		if forwarded := r.Header.Get("X-Forwarded-For"); forwarded != "" {
			parts := strings.Split(forwarded, ",")
			if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
				return ip
			}
		}
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}

	return host
}
