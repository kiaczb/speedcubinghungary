package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizeRedirectPath(t *testing.T) {
	kept := []string{
		"/",
		"/posts",
		"/posts/3f1c2b4a-0000-4000-8000-000000000000",
		"/tools/special?competition=HungarianAllRounder62026",
		"/tools/special?competition=Foo&round=2",
		"/posts#section",
		"/x:y",
		"/kocka%C3%A9p%C3%ADt%C5%91", // percent-kódolt ékezetes útvonal
	}

	for _, path := range kept {
		t.Run("megtartja "+path, func(t *testing.T) {
			assert.Equal(t, path, sanitizeRedirectPath(path))
		})
	}

	// Minden más a gyökérre esik vissza. A lényeg, hogy semmi ne tudjon a saját
	// oldalunkon KÍVÜLRE mutatni (nyílt átirányítás), és semmi ne juthasson
	// vezérlőkarakterrel a Location fejlécbe (fejléc-injektálás).
	rejected := map[string]string{
		"üres":                           "",
		"abszolút http":                  "https://evil.example/x",
		"abszolút séma nélkül":           "//evil.example",
		"protokoll-relatív három perrel": "///evil.example",
		"visszaperjel":                   `/\evil.example`,
		"vegyes per/visszaperjel":        `/\/evil.example`,
		"nem per-kezdetű":                "posts",
		"javascript séma":                "javascript:alert(1)",
		"data séma":                      "data:text/html,<script>alert(1)</script>",
		"CRLF":                           "/ok\r\nX-Injected: 1",
		"sortörés":                       "/ok\nX-Injected: 1",
		"tab":                            "/ok\tvalami",
		"szóköz":                         "/két szó",
		"pontosvessző (süti-elválasztó)": "/x;y",
		"vessző":                         "/x,y",
		"idézőjel":                       `/x"y`,
		"backslash a végén":              `/posts\`,
	}

	for name, path := range rejected {
		t.Run("elutasítja "+name, func(t *testing.T) {
			assert.Equal(t, "/", sanitizeRedirectPath(path))
		})
	}
}

func TestVerifyState(t *testing.T) {
	const nonce = "a1b2c3d4e5f60718293a4b5c6d7e8f90a1b2c3d4e5f60718293a4b5c6d7e8f90"

	withCookie := func(value string) *http.Request {
		req := httptest.NewRequest(http.MethodGet, "/login/callback", nil)
		if value != "" {
			req.AddCookie(&http.Cookie{Name: stateCookieName, Value: value})
		}
		return req
	}

	t.Run("egyező nonce esetén az útvonal a sütiből jön", func(t *testing.T) {
		path, ok := verifyState(withCookie(nonce+":/posts"), nonce)

		assert.True(t, ok)
		assert.Equal(t, "/posts", path)
	})

	t.Run("útvonal nélküli süti a gyökérre esik", func(t *testing.T) {
		path, ok := verifyState(withCookie(nonce+":"), nonce)

		assert.True(t, ok)
		assert.Equal(t, "/", path)
	})

	// A sütit mi állítjuk ki (HttpOnly), de egy elrontott /login ág vagy egy
	// jövőbeli változtatás sem tud rajta keresztül nyílt átirányítást okozni.
	t.Run("a sütiben lévő útvonalat is szanitáljuk", func(t *testing.T) {
		path, ok := verifyState(withCookie(nonce+"://evil.example"), nonce)

		assert.True(t, ok)
		assert.Equal(t, "/", path)
	})

	rejected := []struct {
		name   string
		cookie string
		state  string
	}{
		{name: "nincs süti", cookie: "", state: nonce},
		{name: "nincs state", cookie: nonce + ":/posts", state: ""},
		{name: "más nonce", cookie: nonce + ":/posts", state: "ffffffff"},
		{name: "üres nonce a sütiben", cookie: ":/posts", state: ""},
		{name: "kettőspont nélküli süti", cookie: nonce, state: nonce},
		{name: "a state a nonce prefixe", cookie: nonce + ":/posts", state: nonce[:32]},
	}

	for _, tt := range rejected {
		t.Run("elutasítja: "+tt.name, func(t *testing.T) {
			_, ok := verifyState(withCookie(tt.cookie), tt.state)
			assert.False(t, ok)
		})
	}
}

// A nonce legyen kitalálhatatlan és mindig különböző.
func TestNewStateNonce(t *testing.T) {
	seen := make(map[string]bool, 100)

	for i := 0; i < 100; i++ {
		nonce, err := newStateNonce()

		assert.NoError(t, err)
		assert.Len(t, nonce, stateNonceBytes*2)
		assert.False(t, seen[nonce], "a nonce nem ismétlődhet")
		seen[nonce] = true
	}
}
