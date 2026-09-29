package auth

import "speedcubinghungary/internal/types"

type User struct {
	ID        int          `db:"id"`
	WCAID     string       `db:"wca_id"`
	Name      string       `db:"name"`
	Email     string       `db:"email"`
	AvatarURL string       `db:"avatar_url"`
	Roles     []types.Role `db:"roles"`
}
