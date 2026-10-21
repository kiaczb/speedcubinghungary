package auth

import (
	"time"

	"speedcubinghungary/internal/types"
)

type User struct {
	ID        int          `db:"id"`
	WCAID     string       `db:"wca_id"`
	Name      string       `db:"name"`
	Email     string       `db:"email"`
	AvatarURL string       `db:"avatar_url"`
	Roles     []types.Role `db:"roles"`

	SessionVersion int `db:"session_version"`
}

type WCAGrant struct {
	UserID       int       `db:"user_id"`
	Scope        string    `db:"scope"`
	AccessToken  []byte    `db:"access_token"`
	RefreshToken []byte    `db:"refresh_token"`
	ExpiresAt    time.Time `db:"expires_at"`
}
