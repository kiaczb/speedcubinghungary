package auth

import (
	"regexp"
	"speedcubinghungary/internal/types"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestPostgresRepository_GetByWCAID(t *testing.T) {
	// 1. sqlmock setup
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db)

	// 2. Tesztadatok
	wcaID := "2019ABCD01"
	userID := 1
	name := "John Doe"
	email := "john@example.com"
	rolesJSON := `["user","admin"]`
	competitionCount := 5
	delegatedCount := 2
	createdAt := time.Now()
	updatedAt := time.Now()

	// 3. Mock elvárás: SELECT
	mock.ExpectQuery(regexp.QuoteMeta(
		"SELECT id, wca_id, name, email, roles, competition_count, delegated_count, created_at, updated_at FROM users WHERE wca_id=$1",
	)).
		WithArgs(wcaID).
		WillReturnRows(sqlmock.NewRows([]string{"id","wca_id","name","email","roles","competition_count","delegated_count","created_at","updated_at"}).
			AddRow(userID, wcaID, name, email, rolesJSON, competitionCount, delegatedCount, createdAt, updatedAt))

	// 4. Hívás
	user, err := repo.GetByWCAID(wcaID)

	// 5. Assert
	assert.NoError(t, err)
	assert.Equal(t, userID, user.ID)
	assert.Equal(t, wcaID, user.WCAID)
	assert.Equal(t, name, user.Name)
	assert.Equal(t, email, user.Email)
	assert.Equal(t, competitionCount, user.CompetitionCount)
	assert.Equal(t, delegatedCount, user.DelegatedCount)
	assert.ElementsMatch(t, []types.Role{"user","admin"}, user.Roles)

	// 6. Mock ellenőrzés
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_Create(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	repo := NewPostgresRepository(db)

	user := &User{
		WCAID: "2020XYZ01",
		Name:  "Jane Doe",
		Email: "jane@example.com",
		Roles: []types.Role{"user"},
	}

	// Mock: INSERT
	mock.ExpectQuery(regexp.QuoteMeta(
		`INSERT INTO users (wca_id, name, email, roles) 
		 VALUES ($1,$2,$3,$4) 
		 RETURNING id, created_at, updated_at`,
	)).
		WithArgs(user.WCAID, user.Name, user.Email, []byte(`["user"]`)).
		WillReturnRows(sqlmock.NewRows([]string{"id","created_at","updated_at"}).
			AddRow(2, time.Now(), time.Now()))

	createdUser, err := repo.Create(user)

	assert.NoError(t, err)
	assert.Equal(t, user.WCAID, createdUser.WCAID)
	assert.Equal(t, user.Name, createdUser.Name)
	assert.Equal(t, user.Email, createdUser.Email)

	assert.NoError(t, mock.ExpectationsWereMet())
}
