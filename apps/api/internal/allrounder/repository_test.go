package allrounder

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testFetchedAt = time.Date(2026, 8, 20, 9, 0, 0, 0, time.UTC)

func TestListResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("FROM allrounder_results").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"competition_wca_id", "ordinal", "wca_id", "name",
			"event_id", "round_type_id", "pos", "best",
		}).
			AddRow("C1", 0, "2020AAAA01", "Anna", "333", "f", 1, 700).
			AddRow("C1", 1, "2020BBBB01", "Bence", "333", "f", 2, 800).
			AddRow("C2", 0, "2020AAAA01", "Anna", "222", "f", 3, 300))

	byCompetition, err := NewPostgresRepository(db).ListResults([]string{"C1", "C2"})

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	require.Len(t, byCompetition["C1"], 2)
	require.Len(t, byCompetition["C2"], 1)
	assert.Equal(t, 0, byCompetition["C1"][0].Ordinal)
	assert.Equal(t, "Bence", byCompetition["C1"][1].Name)
}

func TestListSyncStates(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("FROM allrounder_syncs").
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{
			"competition_wca_id", "fetched_at", "result_count", "last_error",
			"start_date", "end_date", "live_competition_id",
		}).
			AddRow("C1", testFetchedAt, 298, "", testFetchedAt, testFetchedAt, 11151).
			AddRow("C2", testFetchedAt, 0, "boom", nil, nil, nil))

	states, err := NewPostgresRepository(db).ListSyncStates([]string{"C1", "C2"})

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
	assert.Equal(t, 298, states["C1"].ResultCount)
	assert.Equal(t, "boom", states["C2"].LastError)

	require.NotNil(t, states["C1"].LiveCompetitionID)
	assert.Equal(t, 11151, *states["C1"].LiveCompetitionID)
	require.NotNil(t, states["C1"].StartDate)

	assert.Nil(t, states["C2"].StartDate)
	assert.Nil(t, states["C2"].LiveCompetitionID)
}

func TestReplaceResults(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO allrounder_syncs").
		WithArgs("C1", testFetchedAt, 2).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM allrounder_results").
		WithArgs("C1").
		WillReturnResult(sqlmock.NewResult(0, 5))
	mock.ExpectExec("INSERT INTO allrounder_results").
		WithArgs(
			"C1",
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
			sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(),
		).
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	err = NewPostgresRepository(db).ReplaceResults("C1", results(
		Result{WCAID: "2020AAAA01", Name: "Anna", EventID: "333", RoundTypeID: "f", Pos: 1, Best: 700},
		Result{WCAID: "2020BBBB01", Name: "Bence", EventID: "333", RoundTypeID: "f", Pos: 2, Best: 800},
	), testFetchedAt)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestReplaceResults_EmptySkipsInsert(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO allrounder_syncs").
		WithArgs("C1", testFetchedAt, 0).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM allrounder_results").
		WithArgs("C1").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = NewPostgresRepository(db).ReplaceResults("C1", nil, testFetchedAt)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestMarkSyncFailed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO allrounder_syncs").
		WithArgs("C1", testFetchedAt, "WCA API returned status 503").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err = NewPostgresRepository(db).MarkSyncFailed("C1", testFetchedAt, "WCA API returned status 503")

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}
