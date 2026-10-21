package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	grantScopes        = "public email manage_competitions"
	grantRequiredScope = "manage_competitions"
	grantRefreshMargin = time.Minute
	grantKeyLabel      = "speedcubinghungary/wca-grant/v1:"
)

var ErrScopeNotGranted = errors.New("the WCA authorization did not include manage_competitions")

func (s *service) GrantAuthorizeURL(state string) string {
	return s.buildAuthorizeURL(state, grantScopes)
}

func (s *service) CompleteGrant(ctx context.Context, userID int, code string) error {
	tokens, err := s.exchangeCode(ctx, code)
	if err != nil {
		return err
	}
	if tokens.RefreshToken == "" {
		return fmt.Errorf("WCA token response contained no refresh_token")
	}
	if tokens.Scope != "" && !hasScope(tokens.Scope, grantRequiredScope) {
		return ErrScopeNotGranted
	}

	wca, err := s.fetchMe(ctx, tokens.AccessToken)
	if err != nil {
		return err
	}

	grant, err := s.sealGrant(userID, tokens, tokens.RefreshToken)
	if err != nil {
		return err
	}
	return s.repo.SaveGrant(ctx, grant, wca.WCAID)
}

func (s *service) WCAAccessToken(ctx context.Context, userID int) (string, error) {
	s.grantMu.Lock()
	defer s.grantMu.Unlock()

	grant, err := s.repo.GetGrant(ctx, userID)
	if err != nil {
		return "", err
	}

	aead, err := s.grantCipher()
	if err != nil {
		return "", err
	}
	accessToken, err := open(aead, grant.AccessToken, userID)
	if err != nil {
		return "", err
	}
	if time.Until(grant.ExpiresAt) > grantRefreshMargin {
		return accessToken, nil
	}

	refreshToken, err := open(aead, grant.RefreshToken, userID)
	if err != nil {
		return "", err
	}

	form := url.Values{}
	form.Set("grant_type", "refresh_token")
	form.Set("refresh_token", refreshToken)
	tokens, err := s.requestToken(ctx, form)
	if err != nil {
		var statusErr *tokenStatusError
		if errors.As(err, &statusErr) && statusErr.status >= 400 && statusErr.status < 500 {
			if deleteErr := s.repo.DeleteGrant(ctx, userID); deleteErr != nil {
				return "", deleteErr
			}
			return "", ErrNoGrant
		}
		return "", err
	}

	if tokens.RefreshToken != "" {
		refreshToken = tokens.RefreshToken
	}
	refreshed, err := s.sealGrant(userID, tokens, refreshToken)
	if err != nil {
		return "", err
	}
	if err := s.repo.UpdateGrant(ctx, refreshed); err != nil {
		return "", err
	}
	return tokens.AccessToken, nil
}

func (s *service) RevokeGrant(ctx context.Context, userID int) error {
	return s.repo.DeleteGrant(ctx, userID)
}

func (s *service) sealGrant(userID int, tokens *tokenResponse, refreshToken string) (WCAGrant, error) {
	aead, err := s.grantCipher()
	if err != nil {
		return WCAGrant{}, err
	}

	expiresIn := time.Duration(tokens.ExpiresIn) * time.Second
	if expiresIn <= 0 {
		expiresIn = time.Hour
	}

	return WCAGrant{
		UserID:       userID,
		Scope:        tokens.Scope,
		AccessToken:  seal(aead, tokens.AccessToken, userID),
		RefreshToken: seal(aead, refreshToken, userID),
		ExpiresAt:    time.Now().Add(expiresIn),
	}, nil
}

func (s *service) grantCipher() (cipher.AEAD, error) {
	key := sha256.Sum256([]byte(grantKeyLabel + s.cfg.JWTSecret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return nil, err
	}
	return cipher.NewGCM(block)
}

func seal(aead cipher.AEAD, plaintext string, userID int) []byte {
	nonce := make([]byte, aead.NonceSize())
	rand.Read(nonce)
	return aead.Seal(nonce, nonce, []byte(plaintext), []byte(strconv.Itoa(userID)))
}

func open(aead cipher.AEAD, sealed []byte, userID int) (string, error) {
	if len(sealed) < aead.NonceSize() {
		return "", fmt.Errorf("stored WCA token is malformed")
	}
	nonce, ciphertext := sealed[:aead.NonceSize()], sealed[aead.NonceSize():]
	plaintext, err := aead.Open(nil, nonce, ciphertext, []byte(strconv.Itoa(userID)))
	if err != nil {
		return "", fmt.Errorf("stored WCA token cannot be decrypted: %w", err)
	}
	return string(plaintext), nil
}

func hasScope(scopes, want string) bool {
	for _, scope := range strings.Fields(scopes) {
		if scope == want {
			return true
		}
	}
	return false
}
