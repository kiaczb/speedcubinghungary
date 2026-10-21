package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	"speedcubinghungary/internal/middleware"
)

var ErrNoGrant = errors.New("no WCA grant stored for this user")

var ErrGrantOwnerMismatch = errors.New("the WCA account does not belong to this admin")

type Repository interface {
	Upsert(user *User) (*User, error)
	GetSession(ctx context.Context, userID int) (middleware.Session, error)
	RevokeSessions(ctx context.Context, userID int) error
	SaveGrant(ctx context.Context, grant WCAGrant, wcaID string) error
	GetGrant(ctx context.Context, userID int) (WCAGrant, error)
	UpdateGrant(ctx context.Context, grant WCAGrant) error
	DeleteGrant(ctx context.Context, userID int) error
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

func (r *postgresRepository) SaveGrant(ctx context.Context, grant WCAGrant, wcaID string) error {
	result, err := r.db.ExecContext(ctx, `
        INSERT INTO wca_grants (user_id, scope, access_token, refresh_token, expires_at)
        SELECT id, $3, $4, $5, $6 FROM users
        WHERE id = $1 AND wca_id = $2 AND roles @> '["admin"]'::jsonb
        ON CONFLICT (user_id) DO UPDATE SET
            scope = EXCLUDED.scope,
            access_token = EXCLUDED.access_token,
            refresh_token = EXCLUDED.refresh_token,
            expires_at = EXCLUDED.expires_at,
            granted_at = NOW(),
            updated_at = NOW()`,
		grant.UserID, wcaID, grant.Scope, grant.AccessToken, grant.RefreshToken, grant.ExpiresAt,
	)
	if err != nil {
		return err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if affected == 0 {
		return ErrGrantOwnerMismatch
	}
	return nil
}

func (r *postgresRepository) GetGrant(ctx context.Context, userID int) (WCAGrant, error) {
	grant := WCAGrant{UserID: userID}
	err := r.db.QueryRowContext(ctx, `
        SELECT g.scope, g.access_token, g.refresh_token, g.expires_at
        FROM wca_grants g
        JOIN users u ON u.id = g.user_id
        WHERE g.user_id = $1 AND u.roles @> '["admin"]'::jsonb`, userID,
	).Scan(&grant.Scope, &grant.AccessToken, &grant.RefreshToken, &grant.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return WCAGrant{}, ErrNoGrant
	}
	return grant, err
}

func (r *postgresRepository) UpdateGrant(ctx context.Context, grant WCAGrant) error {
	_, err := r.db.ExecContext(ctx, `
        UPDATE wca_grants
        SET access_token = $2, refresh_token = $3, expires_at = $4, updated_at = NOW()
        WHERE user_id = $1`,
		grant.UserID, grant.AccessToken, grant.RefreshToken, grant.ExpiresAt,
	)
	return err
}

func (r *postgresRepository) DeleteGrant(ctx context.Context, userID int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM wca_grants WHERE user_id = $1`, userID)
	return err
}
