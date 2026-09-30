package auth

import (
	"testing"

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
		WillReturnRows(sqlmock.NewRows([]string{"id", "roles"}).AddRow(7, `["user","admin"]`))

	user, err := NewPostgresRepository(db).Upsert(input)

	require.NoError(t, err)
	assert.Equal(t, 7, user.ID)
	assert.Equal(t, input.WCAID, user.WCAID)
	assert.Equal(t, input.AvatarURL, user.AvatarURL)
	assert.ElementsMatch(t, []types.Role{types.ROLE_USER, types.ROLE_ADMIN}, user.Roles)
	assert.NoError(t, mock.ExpectationsWereMet())
}
