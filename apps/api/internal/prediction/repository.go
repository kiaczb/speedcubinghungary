package prediction

import (
	"database/sql"
	"time"

	"speedcubinghungary/internal/types"

	"github.com/lib/pq"
)

// contestRow a prediction_contests tábla nyers sora — a WCA-tól jövő
// verseny-metaadatokat (név, város, dátum) NEM tartalmazza, azokat a
// service tölti fel a competition csomagból (lásd service.go).
type contestRow struct {
	ID                 string
	CompetitionWCAID   string
	PredictionDeadline time.Time
	CreatedAt          time.Time
}

// extraQuestionRow a prediction_extra_questions tábla nyers sora.
type extraQuestionRow struct {
	ID        string
	ContestID string
	Prompt    string
	Type      types.PredictionQuestionType
	Ordinal   int
}

type Repository interface {
	GetAll() ([]contestRow, error)
	GetPodiumEvents(contestIDs []string) (map[string][]string, error)
	GetExtraQuestions(contestIDs []string) (map[string][]extraQuestionRow, error)
	Create(row contestRow, podiumEventIDs []string, questions []extraQuestionRow) error
	Delete(id string) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) GetAll() ([]contestRow, error) {
	rows, err := r.db.Query(`
		SELECT id, competition_wca_id, prediction_deadline, created_at
		FROM prediction_contests
		ORDER BY prediction_deadline DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	contests := make([]contestRow, 0)
	for rows.Next() {
		var c contestRow
		if err := rows.Scan(&c.ID, &c.CompetitionWCAID, &c.PredictionDeadline, &c.CreatedAt); err != nil {
			return nil, err
		}
		contests = append(contests, c)
	}

	return contests, rows.Err()
}

// GetPodiumEvents lekéri a megadott tippversenyek dobogó-versenyszámait,
// tippverseny id szerint csoportosítva.
func (r *PostgresRepository) GetPodiumEvents(contestIDs []string) (map[string][]string, error) {
	result := make(map[string][]string, len(contestIDs))
	if len(contestIDs) == 0 {
		return result, nil
	}

	rows, err := r.db.Query(`
		SELECT contest_id, event_id
		FROM prediction_podium_events
		WHERE contest_id = ANY($1)
	`, pq.Array(contestIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var contestID, eventID string
		if err := rows.Scan(&contestID, &eventID); err != nil {
			return nil, err
		}
		result[contestID] = append(result[contestID], eventID)
	}

	return result, rows.Err()
}

// GetExtraQuestions lekéri a megadott tippversenyek extra kérdéseit,
// tippverseny id szerint csoportosítva, megjelenítési sorrendben.
func (r *PostgresRepository) GetExtraQuestions(contestIDs []string) (map[string][]extraQuestionRow, error) {
	result := make(map[string][]extraQuestionRow, len(contestIDs))
	if len(contestIDs) == 0 {
		return result, nil
	}

	rows, err := r.db.Query(`
		SELECT id, contest_id, prompt, type, ordinal
		FROM prediction_extra_questions
		WHERE contest_id = ANY($1)
		ORDER BY contest_id, ordinal
	`, pq.Array(contestIDs))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var q extraQuestionRow
		if err := rows.Scan(&q.ID, &q.ContestID, &q.Prompt, &q.Type, &q.Ordinal); err != nil {
			return nil, err
		}
		result[q.ContestID] = append(result[q.ContestID], q)
	}

	return result, rows.Err()
}

// Create egy tippversenyt hoz létre a dobogó-versenyszámaival és az extra
// kérdéseivel együtt, egy tranzakcióban — a három tábla együtt ír le egy
// konzisztens tippversenyt, félig beírt állapot ne maradhasson.
func (r *PostgresRepository) Create(row contestRow, podiumEventIDs []string, questions []extraQuestionRow) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		INSERT INTO prediction_contests (id, competition_wca_id, prediction_deadline)
		VALUES ($1, $2, $3)
	`, row.ID, row.CompetitionWCAID, row.PredictionDeadline)
	if err != nil {
		return err
	}

	for _, eventID := range podiumEventIDs {
		if _, err := tx.Exec(`
			INSERT INTO prediction_podium_events (contest_id, event_id)
			VALUES ($1, $2)
		`, row.ID, eventID); err != nil {
			return err
		}
	}

	for _, q := range questions {
		if _, err := tx.Exec(`
			INSERT INTO prediction_extra_questions (id, contest_id, prompt, type, ordinal)
			VALUES ($1, $2, $3, $4, $5)
		`, q.ID, row.ID, q.Prompt, q.Type, q.Ordinal); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete törli a tippversenyt. A dobogó-versenyszámok és az extra kérdések
// sorai a `prediction_podium_events`/`prediction_extra_questions` FK-jának
// ON DELETE CASCADE-je miatt vele együtt törlődnek — nem kell tranzakció.
func (r *PostgresRepository) Delete(id string) error {
	res, err := r.db.Exec(`
		DELETE FROM prediction_contests
		WHERE id = $1
	`, id)
	if err != nil {
		return err
	}

	// Ha nincs ilyen id, 0 sor törlődik — ezt 404-ként akarjuk jelezni, nem
	// néma sikerként (mint a post.Update, lásd post/repository.go).
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrContestNotFound
	}

	return nil
}
