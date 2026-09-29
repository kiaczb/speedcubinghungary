package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/middleware"
	"speedcubinghungary/internal/types"

	"github.com/golang-jwt/jwt/v5"
)

// A WCA OAuth végpontjai. Nem env-változók: a WCA egyetlen ismert példány, a
// tesztek pedig a service mezőit írják felül (lásd service_test.go).
const (
	wcaAuthorizeURL = "https://www.worldcubeassociation.org/oauth/authorize"
	wcaTokenURL     = "https://www.worldcubeassociation.org/oauth/token"
	wcaMeURL        = "https://www.worldcubeassociation.org/api/v0/me"

	// wcaScopes: CSAK a bejelentkezéshez szükséges jogok — a WCA access tokent
	// egyetlen /me hívásra használjuk, majd eldobjuk. Bővebb scope-ot (pl.
	// `manage_competitions`) csak akkor kérj, ha van is ilyen hívás.
	wcaScopes = "public email"

	// tokenLifetime a saját JWT-nk élettartama; a cookie MaxAge-e ehhez igazodik.
	tokenLifetime = time.Hour

	upstreamTimeout = 10 * time.Second
	maxUpstreamBody = 1 << 20 // 1 MiB
)

type Service interface {
	// AuthorizeURL a WCA bejelentkező oldalának URL-je a megadott CSRF nonce-szal.
	AuthorizeURL(state string) string
	// LoginCallback beváltja a WCA authorization code-ot, felveszi/frissíti a
	// felhasználót, és visszaadja a saját JWT-nket.
	LoginCallback(ctx context.Context, code string) (string, error)
}

type service struct {
	repo         Repository
	cfg          *config.Config
	httpClient   *http.Client
	authorizeURL string
	tokenURL     string
	meURL        string
}

func NewService(repo Repository, cfg *config.Config) Service {
	return &service{
		repo:         repo,
		cfg:          cfg,
		httpClient:   &http.Client{Timeout: upstreamTimeout},
		authorizeURL: wcaAuthorizeURL,
		tokenURL:     wcaTokenURL,
		meURL:        wcaMeURL,
	}
}

func (s *service) AuthorizeURL(state string) string {
	query := url.Values{}
	query.Set("client_id", s.cfg.WCAClientID)
	// Egyeznie kell a token-beváltáskor küldöttel, különben a WCA elutasítja.
	query.Set("redirect_uri", s.cfg.WCARedirectURI)
	query.Set("response_type", "code")
	query.Set("scope", wcaScopes)
	query.Set("state", state)

	return s.authorizeURL + "?" + query.Encode()
}

// wcaUser a WCA /api/v0/me válaszának általunk használt része.
type wcaUser struct {
	ID     int    `json:"id"`
	WCAID  string `json:"wca_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Avatar struct {
		URL string `json:"url"`
	} `json:"avatar"`
}

func (s *service) LoginCallback(ctx context.Context, code string) (string, error) {
	accessToken, err := s.exchangeCode(ctx, code)
	if err != nil {
		return "", err
	}

	wca, err := s.fetchMe(ctx, accessToken)
	if err != nil {
		return "", err
	}

	user, err := s.repo.Upsert(&User{
		WCAID:     wca.WCAID,
		Name:      wca.Name,
		Email:     wca.Email,
		AvatarURL: wca.Avatar.URL,
		Roles:     []types.Role{types.ROLE_USER}, // Csak új usernél számít (lásd repository.Upsert)
	})
	if err != nil {
		return "", err
	}

	return s.signToken(user)
}

// exchangeCode az authorization code-ot WCA access tokenre váltja.
func (s *service) exchangeCode(ctx context.Context, code string) (string, error) {
	form := url.Values{}
	form.Set("client_id", s.cfg.WCAClientID)
	form.Set("client_secret", s.cfg.WCAClientSecret)
	form.Set("grant_type", "authorization_code")
	form.Set("code", code)
	form.Set("redirect_uri", s.cfg.WCARedirectURI)

	req, err := http.NewRequestWithContext(
		ctx, http.MethodPost, s.tokenURL, strings.NewReader(form.Encode()),
	)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("WCA token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	// ⚠️ EZ A LÉPÉS NEM HAGYHATÓ KI. Enélkül egy 4xx válasz törzse is simán
	// dekódolódik (csak minden mező üres marad), a folyamat végigfut, és a végén
	// ÉRVÉNYES munkamenet-sütit állítanánk ki egy olyan kérésre, ami sosem járt a
	// WCA-nál — azaz bárki bejelentkezhetne egy kitalált `code`-dal.
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("WCA token exchange returned status %d", resp.StatusCode)
	}

	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	if err := decodeJSON(resp.Body, &tokenResp); err != nil {
		return "", fmt.Errorf("WCA token response is not valid JSON: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("WCA token response contained no access_token")
	}

	return tokenResp.AccessToken, nil
}

// fetchMe lekéri a bejelentkezett WCA felhasználó adatait.
func (s *service) fetchMe(ctx context.Context, accessToken string) (*wcaUser, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.meURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("WCA /me request failed: %w", err)
	}
	defer resp.Body.Close()

	// Ugyanaz az indok, mint a token-beváltásnál: egy 401-es JSON törzs is
	// "sikeresen" dekódolódna, csak üres felhasználót adna.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WCA /me returned status %d", resp.StatusCode)
	}

	var body struct {
		Me wcaUser `json:"me"`
	}
	if err := decodeJSON(resp.Body, &body); err != nil {
		return nil, fmt.Errorf("WCA /me response is not valid JSON: %w", err)
	}

	// A felhasználó azonosítója a wca_id — ez a users tábla UNIQUE kulcsa és a
	// jogosultságok hordozója. Üres értékkel NEM hozunk létre/keresünk sort:
	// abból egyetlen, mindenki által osztott "szellemfelhasználó" lenne.
	if strings.TrimSpace(body.Me.WCAID) == "" {
		// Van WCA-fiók wca_id nélkül (aki még sosem versenyzett), ezért ez nem
		// feltétlenül támadás — de bejelentkezni sem tud nálunk.
		return nil, fmt.Errorf("the WCA account has no WCA ID")
	}

	return &body.Me, nil
}

// signToken kiállítja a saját JWT-nket a felhasználóra.
func (s *service) signToken(user *User) (string, error) {
	now := time.Now()

	claims := middleware.UserClaims{
		UserID:    user.ID,
		WCAID:     user.WCAID,
		Name:      user.Name,
		AvatarURL: user.AvatarURL,
		Roles:     user.Roles,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifetime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.cfg.JWTSecret))
}

// decodeJSON korlátozott méretű JSON-t olvas be.
func decodeJSON(body io.Reader, out any) error {
	return json.NewDecoder(io.LimitReader(body, maxUpstreamBody)).Decode(out)
}
