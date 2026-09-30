package auth

import (
	"database/sql"
	"encoding/json"
)

type Repository interface {
	Upsert(user *User) (*User, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) Upsert(u *User) (*User, error) {
	rolesJSON, _ := json.Marshal(u.Roles)

	user := *u
	var roles []byte
	err := r.db.QueryRow(`
        INSERT INTO users (wca_id, name, email, avatar_url, roles)
        VALUES ($1, $2, $3, NULLIF($4, ''), $5)
        ON CONFLICT (wca_id)
        DO UPDATE SET
            name = EXCLUDED.name,
            email = EXCLUDED.email,
            avatar_url = EXCLUDED.avatar_url,
            roles = CASE
                WHEN users.roles = '[]'::jsonb OR users.roles IS NULL THEN EXCLUDED.roles
                ELSE users.roles
            END,
            updated_at = NOW()
        RETURNING id, roles`,
		u.WCAID, u.Name, u.Email, u.AvatarURL, rolesJSON,
	).Scan(&user.ID, &roles)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(roles, &user.Roles); err != nil {
		return nil, err
	}

	return &user, nil
}
