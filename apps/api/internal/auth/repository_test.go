package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"speedcubinghungary/internal/middleware"
	"speedcubinghungary/internal/types"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPostgresRepository_UpsertKeepsStoredRoles(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	input := &User{
		WCAID:     "2019ABCD01",
		Name:      "John Doe",
		Email:     "john@example.com",
		AvatarURL: "https://avatars.worldcubeassociation.org/john.jpg",
		Roles:     []types.Role{types.ROLE_USER},
	}

	mock.ExpectQuery(`INSERT INTO users`).
		WithArgs(input.WCAID, input.Name, input.Email, input.AvatarURL, []byte(`["user"]`)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "roles", "session_version"}).AddRow(7, `["user","admin"]`, 3))

	user, err := NewPostgresRepository(db).Upsert(input)

	require.NoError(t, err)
	assert.Equal(t, 7, user.ID)
	assert.Equal(t, input.WCAID, user.WCAID)
	assert.Equal(t, input.AvatarURL, user.AvatarURL)
	assert.ElementsMatch(t, []types.Role{types.ROLE_USER, types.ROLE_ADMIN}, user.Roles)
	assert.Equal(t, 3, user.SessionVersion, "a belépés a már visszavont munkamenetek utáni verzióval ír alá")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_GetSession(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT session_version, roles FROM users WHERE id = \$1`).
		WithArgs(7).
		WillReturnRows(sqlmock.NewRows([]string{"session_version", "roles"}).AddRow(2, `["user","admin"]`))
	mock.ExpectQuery(`SELECT session_version, roles FROM users WHERE id = \$1`).
		WithArgs(8).
		WillReturnError(sql.ErrNoRows)

	repo := NewPostgresRepository(db)

	session, err := repo.GetSession(context.Background(), 7)
	require.NoError(t, err)
	assert.Equal(t, middleware.Session{Version: 2, Roles: []types.Role{types.ROLE_USER, types.ROLE_ADMIN}}, session)

	_, err = repo.GetSession(context.Background(), 8)
	assert.ErrorIs(t, err, middleware.ErrSessionNotFound, "törölt felhasználó tokenje ne legyen érvényes")

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_RevokeSessionsBumpsTheVersion(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec(`UPDATE users SET session_version = session_version \+ 1 WHERE id = \$1`).
		WithArgs(7).
		WillReturnResult(sqlmock.NewResult(0, 1))

	require.NoError(t, NewPostgresRepository(db).RevokeSessions(context.Background(), 7))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_SaveGrantRequiresTheMatchingAdmin(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	grant := WCAGrant{UserID: 7, Scope: "public email manage_competitions", AccessToken: []byte("a"), RefreshToken: []byte("r"), ExpiresAt: time.Now()}
	mock.ExpectExec(`(?s)INSERT INTO wca_grants.*WHERE id = \$1 AND wca_id = \$2 AND roles @> '\["admin"\]'::jsonb`).
		WithArgs(7, "2019IDEG01", grant.Scope, grant.AccessToken, grant.RefreshToken, grant.ExpiresAt).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = NewPostgresRepository(db).SaveGrant(context.Background(), grant, "2019IDEG01")
	assert.ErrorIs(t, err, ErrGrantOwnerMismatch, "idegen WCA-fiókkal vagy nem adminként nem menthető engedély")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPostgresRepository_GetGrantWithoutRowIsNoGrant(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery(`SELECT g.scope, g.access_token, g.refresh_token, g.expires_at`).
		WithArgs(7).
		WillReturnError(sql.ErrNoRows)

	_, err = NewPostgresRepository(db).GetGrant(context.Background(), 7)
	assert.ErrorIs(t, err, ErrNoGrant)
	assert.NoError(t, mock.ExpectationsWereMet())
}
