package allrounder

import (
	"database/sql"
	"time"

	"github.com/lib/pq"
)

type Repository interface {
	ListResults(competitionWCAIDs []string) (map[string][]Result, error)
	ListSyncStates(competitionWCAIDs []string) (map[string]SyncState, error)
	ReplaceResults(competitionWCAID string, results []Result, fetchedAt time.Time) error
	MarkSyncFailed(competitionWCAID string, fetchedAt time.Time, message string) error
	SaveCompetitionDates(competitionWCAID string, startDate, endDate time.Time) error
	SaveLiveCompetitionID(competitionWCAID string, liveCompetitionID int) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) ListResults(competitionWCAIDs []string) (map[string][]Result, error) {
	rows, err := r.db.Query(`
		SELECT competition_wca_id, ordinal, wca_id, name, event_id, round_type_id, pos, best
		FROM allrounder_results
		WHERE competition_wca_id = ANY($1)
		ORDER BY competition_wca_id, ordinal
	`, pq.Array(competitionWCAIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	byCompetition := make(map[string][]Result)

	for rows.Next() {
		var (
			competitionWCAID string
			result           Result
		)

		if err := rows.Scan(
			&competitionWCAID, &result.Ordinal, &result.WCAID, &result.Name,
			&result.EventID, &result.RoundTypeID, &result.Pos, &result.Best,
		); err != nil {
			return nil, err
		}

		byCompetition[competitionWCAID] = append(byCompetition[competitionWCAID], result)
	}

	return byCompetition, rows.Err()
}

func (r *PostgresRepository) ListSyncStates(competitionWCAIDs []string) (map[string]SyncState, error) {
	rows, err := r.db.Query(`
		SELECT competition_wca_id, fetched_at, result_count, COALESCE(last_error, ''),
		       start_date, end_date, live_competition_id
		FROM allrounder_syncs
		WHERE competition_wca_id = ANY($1)
	`, pq.Array(competitionWCAIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	states := make(map[string]SyncState)

	for rows.Next() {
		var (
			state             SyncState
			startDate         sql.NullTime
			endDate           sql.NullTime
			liveCompetitionID sql.NullInt64
		)
		if err := rows.Scan(
			&state.CompetitionWCAID, &state.FetchedAt, &state.ResultCount, &state.LastError,
			&startDate, &endDate, &liveCompetitionID,
		); err != nil {
			return nil, err
		}

		if startDate.Valid {
			state.StartDate = &startDate.Time
		}
		if endDate.Valid {
			state.EndDate = &endDate.Time
		}
		if liveCompetitionID.Valid {
			id := int(liveCompetitionID.Int64)
			state.LiveCompetitionID = &id
		}

		states[state.CompetitionWCAID] = state
	}

	return states, rows.Err()
}

func (r *PostgresRepository) SaveCompetitionDates(competitionWCAID string, startDate, endDate time.Time) error {
	_, err := r.db.Exec(`
		UPDATE allrounder_syncs
		SET start_date = $2, end_date = $3, updated_at = now()
		WHERE competition_wca_id = $1
	`, competitionWCAID, startDate, endDate)

	return err
}

func (r *PostgresRepository) SaveLiveCompetitionID(competitionWCAID string, liveCompetitionID int) error {
	_, err := r.db.Exec(`
		UPDATE allrounder_syncs
		SET live_competition_id = $2, updated_at = now()
		WHERE competition_wca_id = $1
	`, competitionWCAID, liveCompetitionID)

	return err
}

func (r *PostgresRepository) ReplaceResults(competitionWCAID string, results []Result, fetchedAt time.Time) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`
		INSERT INTO allrounder_syncs (competition_wca_id, fetched_at, result_count, last_error)
		VALUES ($1, $2, $3, NULL)
		ON CONFLICT (competition_wca_id) DO UPDATE SET
			fetched_at   = EXCLUDED.fetched_at,
			result_count = EXCLUDED.result_count,
			last_error   = NULL,
			updated_at   = now()
	`, competitionWCAID, fetchedAt, len(results)); err != nil {
		return err
	}

	if _, err := tx.Exec(`
		DELETE FROM allrounder_results WHERE competition_wca_id = $1
	`, competitionWCAID); err != nil {
		return err
	}

	if len(results) > 0 {
		ordinals := make([]int64, 0, len(results))
		wcaIDs := make([]string, 0, len(results))
		names := make([]string, 0, len(results))
		eventIDs := make([]string, 0, len(results))
		roundTypeIDs := make([]string, 0, len(results))
		positions := make([]int64, 0, len(results))
		bests := make([]int64, 0, len(results))

		for _, result := range results {
			ordinals = append(ordinals, int64(result.Ordinal))
			wcaIDs = append(wcaIDs, result.WCAID)
			names = append(names, result.Name)
			eventIDs = append(eventIDs, result.EventID)
			roundTypeIDs = append(roundTypeIDs, result.RoundTypeID)
			positions = append(positions, int64(result.Pos))
			bests = append(bests, int64(result.Best))
		}

		if _, err := tx.Exec(`
			INSERT INTO allrounder_results
				(competition_wca_id, ordinal, wca_id, name, event_id, round_type_id, pos, best)
			SELECT $1, ordinal, wca_id, name, event_id, round_type_id, pos, best
			FROM unnest(
				$2::int[], $3::text[], $4::text[], $5::text[], $6::text[], $7::int[], $8::int[]
			) AS t(ordinal, wca_id, name, event_id, round_type_id, pos, best)
		`,
			competitionWCAID,
			pq.Array(ordinals), pq.Array(wcaIDs), pq.Array(names),
			pq.Array(eventIDs), pq.Array(roundTypeIDs), pq.Array(positions), pq.Array(bests),
		); err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) MarkSyncFailed(competitionWCAID string, fetchedAt time.Time, message string) error {
	_, err := r.db.Exec(`
		INSERT INTO allrounder_syncs (competition_wca_id, fetched_at, result_count, last_error)
		VALUES ($1, $2, 0, $3)
		ON CONFLICT (competition_wca_id) DO UPDATE SET
			fetched_at = EXCLUDED.fetched_at,
			last_error = EXCLUDED.last_error,
			updated_at = now()
	`, competitionWCAID, fetchedAt, message)

	return err
}
