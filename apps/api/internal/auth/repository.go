package auth

import (
	"database/sql"
	"encoding/json"
)

type Repository interface {
	GetByWCAID(wcaID string) (*User, error)
	Create(user *User) (*User, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &postgresRepository{db: db}
}

func (r *postgresRepository) GetByWCAID(wcaID string) (*User, error) {
	row := r.db.QueryRow("SELECT id, wca_id, name, email, roles, created_at, updated_at FROM users WHERE wca_id=$1", wcaID)
	user := &User{}
	var rolesJSON []byte
	err := row.Scan(&user.ID, &user.WCAID, &user.Name, &user.Email, &rolesJSON, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return nil, err
	}
	json.Unmarshal(rolesJSON, &user.Roles)
	return user, nil
}

func (r *postgresRepository) Create(user *User) (*User, error) {
	rolesJSON, _ := json.Marshal(user.Roles)

	err := r.db.QueryRow(
		`INSERT INTO users (wca_id, name, email, roles) 
		 VALUES ($1,$2,$3,$4) 
		 RETURNING id, created_at, updated_at`,
		user.WCAID, user.Name, user.Email, rolesJSON,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return user, nil
}
