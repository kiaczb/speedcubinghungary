package auth

import (
	"speedcubinghungary/internal/types"

	"github.com/golang-jwt/jwt/v5"
)

type UserClaims struct {
	UserID    int    `json:"user_id"`
	WCAID     string `json:"wca_id"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatarURL"`
	Roles     []types.Role `json:"roles"`
	jwt.RegisteredClaims
}
