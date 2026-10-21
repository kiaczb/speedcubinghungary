package middleware

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"speedcubinghungary/internal/config"
	"speedcubinghungary/internal/httperr"
	"speedcubinghungary/internal/types"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userContextKey = contextKey("user")

var UserClaimsContextKey = userContextKey

type UserClaims struct {
	UserID         int          `json:"user_id"`
	WCAID          string       `json:"wca_id"`
	Name           string       `json:"name"`
	AvatarURL      string       `json:"avatarURL"`
	Roles          []types.Role `json:"roles"`
	SessionVersion int          `json:"session_version"`
	jwt.RegisteredClaims
}

type Session struct {
	Version int
	Roles   []types.Role
}

var ErrSessionNotFound = errors.New("session not found")

type SessionStore interface {
	GetSession(ctx context.Context, userID int) (Session, error)
}

func AuthMiddlewareWithConfig(cfg *config.Config, sessions SessionStore) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("auth")
			if err != nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			tokenString := cookie.Value

			token, err := jwt.ParseWithClaims(
				tokenString,
				&UserClaims{},
				func(token *jwt.Token) (interface{}, error) {
					if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
					}
					return []byte(cfg.JWTSecret), nil
				},
				jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
				jwt.WithExpirationRequired(),
			)

			if err != nil || !token.Valid {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(*UserClaims)
			if !ok {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			session, err := sessions.GetSession(r.Context(), claims.UserID)
			switch {
			case errors.Is(err, ErrSessionNotFound):
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			case err != nil:
				httperr.Write(w, r, http.StatusInternalServerError, err)
				return
			case session.Version != claims.SessionVersion:
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			claims.Roles = session.Roles

			ctx := context.WithValue(r.Context(), userContextKey, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetUserFromContext(r *http.Request) *UserClaims {
	user, ok := r.Context().Value(userContextKey).(*UserClaims)
	if !ok {
		return nil
	}
	return user
}

func RBACMiddleware(requiredRoles ...types.Role) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := GetUserFromContext(r)
			if user == nil || len(user.Roles) == 0 {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}

			for _, role := range user.Roles {
				for _, required := range requiredRoles {
					if role == required {
						next.ServeHTTP(w, r)
						return
					}
				}
			}

			http.Error(w, "forbidden", http.StatusForbidden)
		})
	}
}
