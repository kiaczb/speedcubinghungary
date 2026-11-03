package weeklyreport

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMockRepo(t *testing.T) (*PostgresRepository, sqlmock.Sqlmock) {
	db, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	return NewPostgresRepository(db), mock
}

func TestPublishReport_KeepsTheExistingPostByDefault(t *testing.T) {
	repo, mock := newMockRepo(t)
	week := time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC)
	mock.ExpectExec(`ON CONFLICT \(report_week\) DO NOTHING`).
		WithArgs("id-1", "cím", "{}", "2026-07-27").
		WillReturnResult(sqlmock.NewResult(0, 0))

	written, err := repo.PublishReport(ReportPost{ID: "id-1", Title: "cím", Content: "{}", ReportWeek: week}, false)

	require.NoError(t, err)
	assert.False(t, written)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestPublishReport_ForceOverwrites(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectExec(`ON CONFLICT \(report_week\) DO UPDATE SET`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	written, err := repo.PublishReport(ReportPost{ID: "id-1", ReportWeek: time.Now()}, true)

	require.NoError(t, err)
	assert.True(t, written)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveCounts_OverwritesInOneTransaction(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectBegin()
	prep := mock.ExpectPrepare(regexp.QuoteMeta(`competition_count = EXCLUDED.competition_count`))
	prep.ExpectExec().WithArgs("2015REKA01", "Rekorder Réka", 11, 4804).WillReturnResult(sqlmock.NewResult(0, 1))
	prep.ExpectExec().WithArgs("2019SIMO07", "Simon Ádám", 39, 2000).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.SaveCounts([]PersonCounts{
		{WCAID: "2015REKA01", Name: "Rekorder Réka", CompetitionCount: 11, SolveCount: 4804},
		{WCAID: "2019SIMO07", Name: "Simon Ádám", CompetitionCount: 39, SolveCount: 2000},
	})

	require.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveDelegatedCounts_RollsBackOnError(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectBegin()
	prep := mock.ExpectPrepare(regexp.QuoteMeta(`delegated_count = EXCLUDED.delegated_count`))
	prep.ExpectExec().WithArgs("2008BARA01", "Bence Barát", 72).WillReturnError(assert.AnError)
	mock.ExpectRollback()

	err := repo.SaveDelegatedCounts([]DelegatedCount{{WCAID: "2008BARA01", Name: "Bence Barát", Count: 72}})

	assert.ErrorIs(t, err, assert.AnError)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestSaveCounts_NothingToSave(t *testing.T) {
	repo, mock := newMockRepo(t)

	require.NoError(t, repo.SaveCounts(nil))
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestCompetitorsNearMilestones_ParameterisesEveryRange(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(regexp.QuoteMeta(`WHERE competition_count BETWEEN $1 AND $2 OR competition_count BETWEEN $3 AND $4`)).
		WithArgs(97, 100, 197, 200).
		WillReturnRows(sqlmock.NewRows([]string{"wca_id", "name", "competition_count", "delegated_count", "successful_solves_count"}).
			AddRow("2015REKA01", "Rekorder Réka", 98, 0, 4800))

	people, err := repo.CompetitorsNearMilestones([]int{100, 200}, 3)

	require.NoError(t, err)
	assert.Equal(t, []Person{{WCAID: "2015REKA01", Name: "Rekorder Réka", CompetitionCount: 98, SolveCount: 4800}}, people)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestGetRecords_KeepsMissingValuesAsNil(t *testing.T) {
	repo, mock := newMockRepo(t)
	mock.ExpectQuery(`SELECT region, event_id, single, average FROM records`).
		WillReturnRows(sqlmock.NewRows([]string{"region", "event_id", "single", "average"}).
			AddRow("hungary", "333mbf", 920315102, nil))

	table, err := repo.GetRecords()

	require.NoError(t, err)
	assert.Equal(t, 920315102, *table["hungary"]["333mbf"].Single)
	assert.Nil(t, table["hungary"]["333mbf"].Average)
}

func TestCheckSchema_FailsOnMissingColumn(t *testing.T) {
	repo, mock := newMockRepo(t)
	for _, table := range requiredTables {
		mock.ExpectQuery(`SELECT to_regclass`).WithArgs("public." + table).
			WillReturnRows(sqlmock.NewRows([]string{"to_regclass"}).AddRow(table))
	}
	mock.ExpectQuery(`information_schema.columns`).WithArgs("users", "successful_solves_count").
		WillReturnRows(sqlmock.NewRows([]string{"?column?"}))

	err := repo.CheckSchema()

	assert.ErrorContains(t, err, "users.successful_solves_count")
	assert.NoError(t, mock.ExpectationsWereMet())
}
