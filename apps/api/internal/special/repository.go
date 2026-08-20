package special

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

type Repository interface {
	EnsureCompetition(c Competition, userID int) error

	ListRecordedTimes(competitionWCAID string) ([]RecordedTime, error)
	SetElapsed(competitionWCAID string, liveResultID, attemptNumber, elapsedCs int, note string, userID int, now time.Time) error
	DeleteElapsed(liveResultID, attemptNumber int) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) EnsureCompetition(c Competition, userID int) error {
	_, err := r.db.Exec(`
		INSERT INTO special_competitions
			(competition_wca_id, live_competition_id, name, start_date, end_date, added_by)
		VALUES ($1, $2, $3, NULLIF($4, '')::date, NULLIF($5, '')::date, $6)
		ON CONFLICT (competition_wca_id) DO UPDATE SET
			live_competition_id = EXCLUDED.live_competition_id,
			name                = EXCLUDED.name,
			start_date          = EXCLUDED.start_date,
			end_date            = EXCLUDED.end_date,
			updated_at          = now()
	`, c.CompetitionWCAID, c.LiveCompetitionID, c.Name, c.StartDate, c.EndDate, userID)

	return err
}

func (r *PostgresRepository) ListRecordedTimes(competitionWCAID string) ([]RecordedTime, error) {
	rows, err := r.db.Query(`
		SELECT live_result_id, attempt_number, elapsed_cs,
		       COALESCE(elapsed_note, ''), recorded_by, recorded_at
		FROM special_dnf_attempts
		WHERE competition_wca_id = $1
	`, competitionWCAID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	recorded := make([]RecordedTime, 0)

	for rows.Next() {
		var (
			t          RecordedTime
			recordedBy sql.NullInt64
			recordedAt sql.NullTime
		)

		if err := rows.Scan(
			&t.LiveResultID, &t.AttemptNumber, &t.ElapsedCs,
			&t.Note, &recordedBy, &recordedAt,
		); err != nil {
			return nil, err
		}

		if recordedBy.Valid {
			id := int(recordedBy.Int64)
			t.RecordedBy = &id
		}
		if recordedAt.Valid {
			t.RecordedAt = recordedAt.Time
		}

		recorded = append(recorded, t)
	}

	return recorded, rows.Err()
}

func (r *PostgresRepository) SetElapsed(
	competitionWCAID string,
	liveResultID, attemptNumber, elapsedCs int,
	note string,
	userID int,
	now time.Time,
) error {
	_, err := r.db.Exec(`
		INSERT INTO special_dnf_attempts
			(id, competition_wca_id, live_result_id, attempt_number,
			 elapsed_cs, elapsed_note, recorded_by, recorded_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''), $7, $8, $8, $8)
		ON CONFLICT (live_result_id, attempt_number) DO UPDATE SET
			elapsed_cs   = EXCLUDED.elapsed_cs,
			elapsed_note = EXCLUDED.elapsed_note,
			recorded_by  = EXCLUDED.recorded_by,
			recorded_at  = EXCLUDED.recorded_at,
			updated_at   = EXCLUDED.updated_at
	`, uuid.NewString(), competitionWCAID, liveResultID, attemptNumber,
		elapsedCs, note, userID, now)

	return err
}

func (r *PostgresRepository) DeleteElapsed(liveResultID, attemptNumber int) error {
	_, err := r.db.Exec(`
		DELETE FROM special_dnf_attempts
		WHERE live_result_id = $1 AND attempt_number = $2
	`, liveResultID, attemptNumber)

	return err
}
