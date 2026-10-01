package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"speedcubinghungary/internal/middleware"
)

type Repository interface {
	Upsert(user *User) (*User, error)
	GetSession(ctx context.Context, userID int) (middleware.Session, error)
	RevokeSessions(ctx context.Context, userID int) error
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
        RETURNING id, roles, session_version`,
		u.WCAID, u.Name, u.Email, u.AvatarURL, rolesJSON,
	).Scan(&user.ID, &roles, &user.SessionVersion)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(roles, &user.Roles); err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *postgresRepository) GetSession(ctx context.Context, userID int) (middleware.Session, error) {
	var session middleware.Session
	var roles []byte
	err := r.db.QueryRowContext(ctx,
		`SELECT session_version, roles FROM users WHERE id = $1`, userID,
	).Scan(&session.Version, &roles)
	if errors.Is(err, sql.ErrNoRows) {
		return middleware.Session{}, middleware.ErrSessionNotFound
	}
	if err != nil {
		return middleware.Session{}, err
	}

	if err := json.Unmarshal(roles, &session.Roles); err != nil {
		return middleware.Session{}, err
	}
	return session, nil
}

func (r *postgresRepository) RevokeSessions(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET session_version = session_version + 1 WHERE id = $1`, userID,
	)
	return err
}
