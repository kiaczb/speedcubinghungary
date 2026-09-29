package auth

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"

	"speedcubinghungary/internal/types"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// Végponttól végpontig futó teszt a teljes WCA OAuth folyamatra, VALÓDI
// süti-kezeléssel (cookiejar) és valódi átirányítás-követéssel. A darabonkénti
// tesztek mindegyike egy-egy lépést rögzít; ez azt igazolja, hogy a lépések
// össze is állnak: a /login-ban kiállított nonce ténylegesen megfordul a
// WCA-n, és a callback ugyanazt látja viszont.
//
// Amit egy Go-teszt NEM tud ellenőrizni: a SameSite attribútumot (azt csak
// böngésző értelmezi). Azt a handler_test.go rögzíti attribútum-szinten.

// fakeWCA a WCA három végpontját utánozza: authorize, token, /me.
type fakeWCA struct {
	server *httptest.Server

	mu            sync.Mutex
	authorizeSeen url.Values // az authorize-ra érkezett query
	tokenCalls    int
	meCalls       int
}

func newFakeWCA(t *testing.T) *fakeWCA {
	t.Helper()

	wca := &fakeWCA{}
	mux := http.NewServeMux()

	// A felhasználó "jóváhagyja" a belépést: a WCA visszairányít a redirect_uri-ra
	// a code-dal és a VÁLTOZATLAN state-tel.
	mux.HandleFunc("/oauth/authorize", func(w http.ResponseWriter, r *http.Request) {
		wca.mu.Lock()
		wca.authorizeSeen = r.URL.Query()
		wca.mu.Unlock()

		back, err := url.Parse(r.URL.Query().Get("redirect_uri"))
		require.NoError(t, err)

		query := back.Query()
		query.Set("code", "wca-auth-code")
		query.Set("state", r.URL.Query().Get("state"))
		back.RawQuery = query.Encode()

		http.Redirect(w, r, back.String(), http.StatusFound)
	})

	mux.HandleFunc("/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		wca.mu.Lock()
		wca.tokenCalls++
		wca.mu.Unlock()

		require.NoError(t, r.ParseForm())
		assert.Equal(t, "wca-auth-code", r.Form.Get("code"))

		json.NewEncoder(w).Encode(map[string]string{"access_token": "wca-access-token"})
	})

	mux.HandleFunc("/api/v0/me", func(w http.ResponseWriter, r *http.Request) {
		wca.mu.Lock()
		wca.meCalls++
		wca.mu.Unlock()

		assert.Equal(t, "Bearer wca-access-token", r.Header.Get("Authorization"))

		json.NewEncoder(w).Encode(map[string]any{
			"me": map[string]any{
				"id": 42, "wca_id": "2019ABCD01", "name": "John Doe",
				"email": "john@example.com",
			},
		})
	})

	wca.server = httptest.NewServer(mux)
	t.Cleanup(wca.server.Close)

	return wca
}

// loginFlow felépíti a három szereplőt (frontend, API, WCA), és lefuttatja a
// teljes belépést. Visszaadja a végén kikötött URL-t és a kliens sütijeit.
func loginFlow(t *testing.T, startPath string) (finalURL *url.URL, cookies []*http.Cookie, wca *fakeWCA) {
	t.Helper()

	wca = newFakeWCA(t)

	// A "frontend": csak visszaigazolja, hogy a belépés ide tért vissza.
	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(frontend.Close)

	cfg := testConfig()
	cfg.Deployment = "development" // httptest http-n szolgál ki: Secure süti nem jöhet
	cfg.FrontendURL = frontend.URL

	repo := new(MockRepo)
	repo.On("Upsert", mock.AnythingOfType("*auth.User")).Return(&User{
		ID: 7, WCAID: "2019ABCD01", Name: "John Doe",
		Roles: []types.Role{types.ROLE_USER},
	}, nil)

	svc := &service{
		repo:         repo,
		cfg:          cfg,
		httpClient:   &http.Client{},
		authorizeURL: wca.server.URL + "/oauth/authorize",
		tokenURL:     wca.server.URL + "/oauth/token",
		meURL:        wca.server.URL + "/api/v0/me",
	}
	handler := NewHandler(svc, cfg)

	mux := http.NewServeMux()
	mux.HandleFunc("/login", handler.Login)
	mux.HandleFunc("/login/callback", handler.LoginCallback)
	api := httptest.NewServer(mux)
	t.Cleanup(api.Close)

	// A redirect_uri csak most, a szerver felállása után ismert.
	cfg.WCARedirectURI = api.URL + "/login/callback"

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{Jar: jar}

	resp, err := client.Get(api.URL + "/login?redirect=" + url.QueryEscape(startPath))
	require.NoError(t, err)
	defer resp.Body.Close()

	apiURL, err := url.Parse(api.URL)
	require.NoError(t, err)

	return resp.Request.URL, jar.Cookies(apiURL), wca
}

func TestLoginFlow_EndToEnd(t *testing.T) {
	final, cookies, wca := loginFlow(t, "/posts")

	// 1) A WCA a config értékeit kapta — és CSAK a szükséges scope-ot.
	wca.mu.Lock()
	authorize := wca.authorizeSeen
	tokenCalls, meCalls := wca.tokenCalls, wca.meCalls
	wca.mu.Unlock()

	assert.Equal(t, "test-client-id", authorize.Get("client_id"))
	assert.Equal(t, "public email", authorize.Get("scope"))
	assert.Len(t, authorize.Get("state"), 64, "a state a szerveroldali nonce")
	assert.Equal(t, 1, tokenCalls)
	assert.Equal(t, 1, meCalls)

	// 2) A folyamat a frontend kért útvonalán ért véget.
	assert.Equal(t, "/posts", final.Path)
	assert.Empty(t, final.Query().Get("login"), "nem a hibaágon futott végig")

	// 3) A böngésző munkamenet-sütit kapott, a state süti pedig eltűnt.
	var auth, state *http.Cookie
	for _, c := range cookies {
		switch c.Name {
		case authCookieName:
			auth = c
		case stateCookieName:
			state = c
		}
	}

	require.NotNil(t, auth, "a belépés végén auth sütinek kell lennie")
	assert.NotEmpty(t, auth.Value)
	assert.Nil(t, state, "a state sütit a callback törli")
}

// A folyamat közepén a támadó saját code-ja érkezik, a mi nonce-unk NÉLKÜL
// (login CSRF). A callback ilyenkor nem válthat be kódot és nem adhat sütit.
func TestLoginFlow_CallbackWithoutStateCookieIsRejected(t *testing.T) {
	wca := newFakeWCA(t)

	frontend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(frontend.Close)

	cfg := testConfig()
	cfg.Deployment = "development"
	cfg.FrontendURL = frontend.URL

	repo := new(MockRepo)
	svc := &service{
		repo: repo, cfg: cfg, httpClient: &http.Client{},
		authorizeURL: wca.server.URL + "/oauth/authorize",
		tokenURL:     wca.server.URL + "/oauth/token",
		meURL:        wca.server.URL + "/api/v0/me",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/login/callback", NewHandler(svc, cfg).LoginCallback)
	api := httptest.NewServer(mux)
	t.Cleanup(api.Close)

	jar, err := cookiejar.New(nil)
	require.NoError(t, err)
	client := &http.Client{Jar: jar}

	// Nincs előzetes /login, tehát nincs state süti — pont ez a támadás alakja.
	resp, err := client.Get(api.URL + "/login/callback?code=attacker-code&state=whatever")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, "failed", resp.Request.URL.Query().Get("login"))

	apiURL, _ := url.Parse(api.URL)
	for _, c := range jar.Cookies(apiURL) {
		assert.NotEqual(t, authCookieName, c.Name, "nem születhetett munkamenet")
	}

	wca.mu.Lock()
	defer wca.mu.Unlock()
	assert.Zero(t, wca.tokenCalls, "a kódot be sem szabad váltani")
	repo.AssertNotCalled(t, "Upsert", mock.Anything)
}
