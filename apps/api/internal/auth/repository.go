package auth

import (
	"database/sql"
	"encoding/json"
)

// Bekerült az Upsert az interface-be!
type Repository interface {
	GetByWCAID(wcaID string) (*User, error)
	Create(user *User) (*User, error)
	Upsert(user *User) (*User, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) GetByWCAID(wcaID string) (*User, error) {
	// Lekérjük a számlálókat is
	query := `SELECT id, wca_id, name, email, avatar_url, roles, competition_count, delegated_count, created_at, updated_at FROM users WHERE wca_id=$1`
	row := r.db.QueryRow(query, wcaID)

	user := &User{}
	var rolesJSON []byte
	var nullEmail sql.NullString  // NULL védelem
	var nullAvatar sql.NullString // avatar_url nullable

	err := row.Scan(
		&user.ID,
		&user.WCAID,
		&user.Name,
		&nullEmail,
		&nullAvatar,
		&rolesJSON,
		&user.CompetitionCount,
		&user.DelegatedCount,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if nullEmail.Valid {
		user.Email = nullEmail.String
	}
	if nullAvatar.Valid {
		user.AvatarURL = nullAvatar.String
	}

	json.Unmarshal(rolesJSON, &user.Roles)
	return user, nil
}

func (r *postgresRepository) Create(user *User) (*User, error) {
	rolesJSON, _ := json.Marshal(user.Roles)

	err := r.db.QueryRow(
		`INSERT INTO users (wca_id, name, email, avatar_url, roles)
         VALUES ($1,$2,$3,NULLIF($4,''),$5)
         RETURNING id, created_at, updated_at`,
		user.WCAID, user.Name, user.Email, user.AvatarURL, rolesJSON,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return user, nil
}

func (r *postgresRepository) Upsert(u *User) (*User, error) {
	query := `
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
        RETURNING id, wca_id, name, email, avatar_url, roles, competition_count, delegated_count, created_at, updated_at;
    `

	rolesJSON, _ := json.Marshal(u.Roles)

	row := r.db.QueryRow(query, u.WCAID, u.Name, u.Email, u.AvatarURL, rolesJSON)

	var updatedUser User
	var rolesBytes []byte
	var nullEmail sql.NullString  // NULL védelem itt is
	var nullAvatar sql.NullString // avatar_url nullable

	err := row.Scan(
		&updatedUser.ID,
		&updatedUser.WCAID,
		&updatedUser.Name,
		&nullEmail,
		&nullAvatar,
		&rolesBytes,
		&updatedUser.CompetitionCount,
		&updatedUser.DelegatedCount,
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if nullEmail.Valid {
		updatedUser.Email = nullEmail.String
	}
	if nullAvatar.Valid {
		updatedUser.AvatarURL = nullAvatar.String
	}

	json.Unmarshal(rolesBytes, &updatedUser.Roles)

	return &updatedUser, nil
}
