package auth

import (
	"speedcubinghungary/internal/types"
	"time"
)

type User struct {
	ID    int    `db:"id"`
	WCAID string `db:"wca_id"`
	Name  string `db:"name"`
	Email string `db:"email"`
	//AvatarURL string	`db:"avatar_url"`
	Roles            []types.Role `db:"roles"` // JSONB vagy TEXT[]
	CompetitionCount int          `db:"competition_count"`
	DelegatedCount   int          `db:"delegated_count"`
	CreatedAt        time.Time    `db:"created_at"`
	UpdatedAt        time.Time    `db:"updated_at"`
}
