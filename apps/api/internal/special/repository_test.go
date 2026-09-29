package special

import (
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testNow = time.Date(2026, 8, 14, 10, 0, 0, 0, time.UTC)

func TestEnsureCompetition(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO special_competitions").
		WithArgs("SpeedcubingHungary2026", 11151, "Speedcubing Hungary 2026", "2026-08-14", "2026-08-15", 7).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = NewPostgresRepository(db).EnsureCompetition(Competition{
		CompetitionWCAID:  "SpeedcubingHungary2026",
		LiveCompetitionID: 11151,
		Name:              "Speedcubing Hungary 2026",
		StartDate:         "2026-08-14",
		EndDate:           "2026-08-15",
	}, 7)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSetElapsed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("INSERT INTO special_dnf_attempts").
		WithArgs(
			sqlmock.AnyArg(), "SpeedcubingHungary2026", 4973030, 2,
			3742, "bíró jegyezte", 3, testNow,
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err = NewPostgresRepository(db).SetElapsed(
		"SpeedcubingHungary2026", 4973030, 2, 3742, "bíró jegyezte", 3, testNow,
	)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeleteElapsed(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectExec("DELETE FROM special_dnf_attempts").
		WithArgs("AllRounder62026", 4973030, 2).
		WillReturnResult(sqlmock.NewResult(0, 0))

	err = NewPostgresRepository(db).DeleteElapsed("AllRounder62026", 4973030, 2)

	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestListRecordedTimes(t *testing.T) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	defer db.Close()

	mock.ExpectQuery("FROM special_dnf_attempts").
		WithArgs("SpeedcubingHungary2026").
		WillReturnRows(sqlmock.NewRows([]string{"live_result_id", "attempt_number", "elapsed_cs"}).
			AddRow(4973030, 2, 3742).
			AddRow(4973031, 1, 5000))

	recorded, err := NewPostgresRepository(db).ListRecordedTimes("SpeedcubingHungary2026")

	require.NoError(t, err)
	assert.Equal(t, []RecordedTime{
		{LiveResultID: 4973030, AttemptNumber: 2, ElapsedCs: 3742},
		{LiveResultID: 4973031, AttemptNumber: 1, ElapsedCs: 5000},
	}, recorded)
	assert.NoError(t, mock.ExpectationsWereMet())
}
