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
	// RateLimitWindow a számláló ablaka. Fix ablak, nem csúszó: egy percen belül
	// ennyi kérés fér bele, aztán nulláról indul.
	RateLimitWindow = time.Minute

	// LoginRateLimit a /login és a /login/callback közös kerete. Egy
	// bejelentkezés két kérés, így ez percenként 30 bejelentkezés egy IP-ről —
	// egy versenyhelyszín közös Wi-Fi-jén is elég, egy szkriptnek viszont kevés.
	LoginRateLimit = 60
)

// RateLimit IP-nkénti kérésszámlálót tesz a lánc elé.
//
// ⚠️ CSAK olyan útvonalra tedd, ahol egy IP-ről jövő legitim forgalom
// természeténél fogva kicsi. Két ok miatt nem lehet globális:
//
//   - A Server Componentek (serverJson) a web konténerből, a Docker-hálózaton
//     át hívnak minket, X-Forwarded-For nélkül — tehát az ÖSSZES látogató
//     szerveroldali kérése egyetlen kulcson, a web konténer IP-jén osztozna.
//     Betelve a serverJson csendben üres listát adna: üres hírek, posztok,
//     tabellák, hibaüzenet nélkül.
//   - A live eszközöket versenyhelyszínen használják, ahol mindenki egy közös
//     Wi-Fi IP-ről jön. A Special oldal minden élő frissítésre újrakéri a
//     riportot, nézőnként — egy per-IP korlát pont ott törne el, ahol a
//     legnagyobb szükség van rá.
//
// A WCA felé menő forgalmat ezért nem itt, hanem a kimenő oldalon korlátozzuk
// (lásd competition.WCARepository): az független attól, ki és honnan kérdez.
//
// A számláló memóriában és példányonként él — nincs Redis, egy API konténer
// fut (CLAUDE.md §12).
func RateLimit(cfg *config.Config, limit int) func(http.Handler) http.Handler {
	limiter := &rateLimiter{
		limit:  limit,
		window: RateLimitWindow,
		// Az X-Forwarded-For-t csak akkor hisszük el, ha tényleg proxy mögött
		// vagyunk. Fejlesztésben nincs Caddy, ott a fejléc szabadon hamisítható
		// lenne — és egyetlen fejléccel meg lehetne kerülni a korlátot.
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

// sweepLocked eldobja a lejárt vödröket. Enélkül a map minden valaha látott
// IP-t megtartana, és egy sok forrásból jövő szkennelés maga lenne a
// memóriaszivárgás — épp az a forgalom, ami ellen a korlát szól.
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

// clientKey a kérés forrás-IP-je.
//
// ⚠️ Az X-Forwarded-For UTOLSÓ elemét vesszük, nem az elsőt. A kliens is
// küldhet saját X-Forwarded-For fejlécet; a Caddy ahhoz HOZZÁFŰZI, amit ő
// tényleg látott, így a lánc végén a valódi forrás áll. Az első elemet venni
// azt jelentené, hogy egyetlen hamisított fejléccel bárki végtelen kvótát kap.
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
