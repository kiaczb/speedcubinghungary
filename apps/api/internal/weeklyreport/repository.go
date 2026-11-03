package weeklyreport

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

type Repository interface {
	CheckSchema() error
	GetRecords() (RecordTable, error)
	UpsertRecord(region, eventID string, values RecordValues) error
	TrackedPeople() ([]Person, error)
	SaveCounts(counts []PersonCounts) error
	SaveDelegatedCounts(counts []DelegatedCount) error
	CompetitorsNearMilestones(milestones []int, difference int) ([]Person, error)
	DelegatesNearMilestones(milestones []int, difference int) ([]Person, error)
	SolversNearMilestones(step, difference int) ([]Person, error)
	PublishReport(post ReportPost, replace bool) (bool, error)
}

var requiredTables = []string{"users", "records", "posts"}

var requiredColumns = [][2]string{{"users", "successful_solves_count"}}

const personColumns = `wca_id, COALESCE(name, ''), COALESCE(competition_count, 0),
	COALESCE(delegated_count, 0), successful_solves_count`

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) CheckSchema() error {
	for _, table := range requiredTables {
		var exists sql.NullString
		if err := r.db.QueryRow(`SELECT to_regclass($1)::text`, "public."+table).Scan(&exists); err != nil {
			return err
		}
		if !exists.Valid {
			return fmt.Errorf("required table %q is missing, run the API migrations first", table)
		}
	}
	for _, column := range requiredColumns {
		var one int
		err := r.db.QueryRow(`
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public' AND table_name = $1 AND column_name = $2`,
			column[0], column[1]).Scan(&one)
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("required column %s.%s is missing, run the API migrations first", column[0], column[1])
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) GetRecords() (RecordTable, error) {
	rows, err := r.db.Query(`SELECT region, event_id, single, average FROM records`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	table := RecordTable{}
	for rows.Next() {
		var region, eventID string
		var single, average sql.NullInt64
		if err := rows.Scan(&region, &eventID, &single, &average); err != nil {
			return nil, err
		}
		if table[region] == nil {
			table[region] = map[string]RecordValues{}
		}
		table[region][eventID] = RecordValues{Single: nullableInt(single), Average: nullableInt(average)}
	}
	return table, rows.Err()
}

func nullableInt(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	n := int(v.Int64)
	return &n
}

func (r *PostgresRepository) UpsertRecord(region, eventID string, values RecordValues) error {
	_, err := r.db.Exec(`
		INSERT INTO records (region, event_id, single, average)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (region, event_id) DO UPDATE SET
			single = EXCLUDED.single,
			average = EXCLUDED.average`,
		region, eventID, values.Single, values.Average)
	return err
}

func (r *PostgresRepository) TrackedPeople() ([]Person, error) {
	return r.queryPeople(`
		SELECT ` + personColumns + ` FROM users
		WHERE competition_count > 0 OR delegated_count > 0
		ORDER BY wca_id`)
}

func (r *PostgresRepository) SaveCounts(counts []PersonCounts) error {
	return r.inTransaction(`
		INSERT INTO users (wca_id, name, competition_count, successful_solves_count)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (wca_id) DO UPDATE SET
			name = EXCLUDED.name,
			competition_count = EXCLUDED.competition_count,
			successful_solves_count = EXCLUDED.successful_solves_count,
			updated_at = NOW()`,
		len(counts), func(i int) []any {
			c := counts[i]
			return []any{c.WCAID, c.Name, c.CompetitionCount, c.SolveCount}
		})
}

func (r *PostgresRepository) SaveDelegatedCounts(counts []DelegatedCount) error {
	return r.inTransaction(`
		INSERT INTO users (wca_id, name, delegated_count)
		VALUES ($1, $2, $3)
		ON CONFLICT (wca_id) DO UPDATE SET
			name = EXCLUDED.name,
			delegated_count = EXCLUDED.delegated_count,
			updated_at = NOW()`,
		len(counts), func(i int) []any {
			c := counts[i]
			return []any{c.WCAID, c.Name, c.Count}
		})
}

func (r *PostgresRepository) inTransaction(query string, n int, args func(i int) []any) error {
	if n == 0 {
		return nil
	}
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for i := range n {
		if _, err := stmt.Exec(args(i)...); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (r *PostgresRepository) CompetitorsNearMilestones(milestones []int, difference int) ([]Person, error) {
	return r.peopleNearMilestones("competition_count", milestones, difference)
}

func (r *PostgresRepository) DelegatesNearMilestones(milestones []int, difference int) ([]Person, error) {
	return r.peopleNearMilestones("delegated_count", milestones, difference)
}

func (r *PostgresRepository) peopleNearMilestones(column string, milestones []int, difference int) ([]Person, error) {
	if len(milestones) == 0 {
		return []Person{}, nil
	}
	conditions := make([]string, 0, len(milestones))
	args := make([]any, 0, 2*len(milestones))
	for i, milestone := range milestones {
		conditions = append(conditions, fmt.Sprintf("%s BETWEEN $%d AND $%d", column, 2*i+1, 2*i+2))
		args = append(args, milestone-difference, milestone)
	}
	return r.queryPeople(
		`SELECT `+personColumns+` FROM users WHERE `+strings.Join(conditions, " OR ")+` ORDER BY name`,
		args...)
}

func (r *PostgresRepository) SolversNearMilestones(step, difference int) ([]Person, error) {
	return r.queryPeople(`
		SELECT `+personColumns+` FROM users
		WHERE successful_solves_count > 0
		  AND (
			(MOD(successful_solves_count, $1) <= $2 AND successful_solves_count >= $1)
			OR MOD(successful_solves_count, $1) >= ($1 - $2)
		  )
		ORDER BY name`,
		step, difference)
}

func (r *PostgresRepository) queryPeople(query string, args ...any) ([]Person, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	people := make([]Person, 0)
	for rows.Next() {
		var p Person
		if err := rows.Scan(&p.WCAID, &p.Name, &p.CompetitionCount, &p.DelegatedCount, &p.SolveCount); err != nil {
			return nil, err
		}
		people = append(people, p)
	}
	return people, rows.Err()
}

func (r *PostgresRepository) PublishReport(post ReportPost, replace bool) (bool, error) {
	conflict := `DO NOTHING`
	if replace {
		conflict = `DO UPDATE SET
			title = EXCLUDED.title,
			content = EXCLUDED.content,
			updated_at = NOW()`
	}

	result, err := r.db.Exec(`
		INSERT INTO posts (id, tag, title, content, report_week)
		VALUES ($1, 'report', $2, $3, $4)
		ON CONFLICT (report_week) `+conflict,
		post.ID, post.Title, post.Content, post.ReportWeek.Format(time.DateOnly))
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected == 1, nil
}
