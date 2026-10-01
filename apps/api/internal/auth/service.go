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

const (
	wcaAuthorizeURL = "https://www.worldcubeassociation.org/oauth/authorize"
	wcaTokenURL     = "https://www.worldcubeassociation.org/oauth/token"
	wcaMeURL        = "https://www.worldcubeassociation.org/api/v0/me"

	wcaScopes = "public email"

	tokenLifetime = 12 * time.Hour

	upstreamTimeout = 10 * time.Second
	maxUpstreamBody = 1 << 20
)

type Service interface {
	AuthorizeURL(state string) string
	LoginCallback(ctx context.Context, code string) (string, error)
	Logout(ctx context.Context, userID int) error
	GetSession(ctx context.Context, userID int) (middleware.Session, error)
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
	query.Set("redirect_uri", s.cfg.WCARedirectURI)
	query.Set("response_type", "code")
	query.Set("scope", wcaScopes)
	query.Set("state", state)

	return s.authorizeURL + "?" + query.Encode()
}

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
		Roles:     []types.Role{types.ROLE_USER},
	})
	if err != nil {
		return "", err
	}

	return s.signToken(user)
}

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

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("WCA /me returned status %d", resp.StatusCode)
	}

	var body struct {
		Me wcaUser `json:"me"`
	}
	if err := decodeJSON(resp.Body, &body); err != nil {
		return nil, fmt.Errorf("WCA /me response is not valid JSON: %w", err)
	}

	if strings.TrimSpace(body.Me.WCAID) == "" {
		return nil, fmt.Errorf("the WCA account has no WCA ID")
	}

	return &body.Me, nil
}

func (s *service) signToken(user *User) (string, error) {
	now := time.Now()

	claims := middleware.UserClaims{
		UserID:         user.ID,
		WCAID:          user.WCAID,
		Name:           user.Name,
		AvatarURL:      user.AvatarURL,
		Roles:          user.Roles,
		SessionVersion: user.SessionVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(tokenLifetime)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.cfg.JWTSecret))
}

func decodeJSON(body io.Reader, out any) error {
	return json.NewDecoder(io.LimitReader(body, maxUpstreamBody)).Decode(out)
}

func (s *service) Logout(ctx context.Context, userID int) error {
	return s.repo.RevokeSessions(ctx, userID)
}

func (s *service) GetSession(ctx context.Context, userID int) (middleware.Session, error) {
	return s.repo.GetSession(ctx, userID)
}
