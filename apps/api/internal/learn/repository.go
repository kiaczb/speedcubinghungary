package learn

import (
	"database/sql"
	"errors"
)

type Repository interface {
	ListEvents() ([]Event, error)
	GetEvent(id string) (Event, error)
	CreateEvent(e Event) error
	UpdateEvent(e Event) error
	DeleteEvent(id string) error

	ListCases(eventID string) ([]Case, error)
	ListAllCases() ([]Case, error)
	GetCase(id string) (Case, error)
	CreateCase(c Case) error
	UpdateCase(c Case) error
	DeleteCase(id string) error

	ListAlgorithms(caseID string) ([]Algorithm, error)
	ListAllAlgorithms() ([]Algorithm, error)
	GetAlgorithm(id string) (Algorithm, error)
	CreateAlgorithm(a Algorithm) error
	UpdateAlgorithm(a Algorithm) error
	DeleteAlgorithm(id string) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// --- Event -------------------------------------------------------------------

// A rendezés mindenhol (sort_order, name): az admin által megadott sorrend az
// elsődleges, azon belül a név — így az azonos (alapértelmezett 0) sorrendű
// elemek sem ugranak össze-vissza kérésről kérésre.
func (r *PostgresRepository) ListEvents() ([]Event, error) {
	rows, err := r.db.Query(`
		SELECT e.id, e.name, COALESCE(e.description, ''), e.sort_order, COUNT(c.id)
		FROM learn_events e
		LEFT JOIN learn_cases c ON c.event_id = e.id
		GROUP BY e.id, e.name, e.description, e.sort_order
		ORDER BY e.sort_order, e.name
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	events := make([]Event, 0)
	for rows.Next() {
		var e Event
		if err := rows.Scan(&e.ID, &e.Name, &e.Description, &e.Order, &e.CaseCount); err != nil {
			return nil, err
		}
		events = append(events, e)
	}

	return events, rows.Err()
}

func (r *PostgresRepository) GetEvent(id string) (Event, error) {
	var e Event
	err := r.db.QueryRow(`
		SELECT e.id, e.name, COALESCE(e.description, ''), e.sort_order,
		       (SELECT COUNT(*) FROM learn_cases c WHERE c.event_id = e.id)
		FROM learn_events e
		WHERE e.id = $1
	`, id).Scan(&e.ID, &e.Name, &e.Description, &e.Order, &e.CaseCount)

	if errors.Is(err, sql.ErrNoRows) {
		return Event{}, ErrEventNotFound
	}
	return e, err
}

func (r *PostgresRepository) CreateEvent(e Event) error {
	// NULLIF: leírás nélkül NULL kerül az oszlopba, nem üres string.
	_, err := r.db.Exec(`
		INSERT INTO learn_events (id, name, description, sort_order)
		VALUES ($1, $2, NULLIF($3, ''), $4)
	`, e.ID, e.Name, e.Description, e.Order)

	return err
}

func (r *PostgresRepository) UpdateEvent(e Event) error {
	res, err := r.db.Exec(`
		UPDATE learn_events
		SET name = $2, description = NULLIF($3, ''), sort_order = $4, updated_at = now()
		WHERE id = $1
	`, e.ID, e.Name, e.Description, e.Order)
	if err != nil {
		return err
	}

	return requireAffected(res, ErrEventNotFound)
}

// DeleteEvent: az esetek és az algoritmusok a FK ON DELETE CASCADE miatt
// magukkal mennek — nem kell kézzel takarítani.
func (r *PostgresRepository) DeleteEvent(id string) error {
	res, err := r.db.Exec(`DELETE FROM learn_events WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return requireAffected(res, ErrEventNotFound)
}

// --- Case --------------------------------------------------------------------

func (r *PostgresRepository) ListCases(eventID string) ([]Case, error) {
	return r.queryCases(`
		SELECT c.id, c.event_id, c.name, COALESCE(c.description, ''), c.sort_order,
		       COUNT(a.id)
		FROM learn_cases c
		LEFT JOIN learn_algorithms a ON a.case_id = c.id
		WHERE c.event_id = $1
		GROUP BY c.id, c.event_id, c.name, c.description, c.sort_order
		ORDER BY c.sort_order, c.name
	`, eventID)
}

// ListAllCases: az admin fájához kell, hogy a teljes hierarchia három
// lekérdezésből álljon össze, ne szintenként N+1-ből.
func (r *PostgresRepository) ListAllCases() ([]Case, error) {
	return r.queryCases(`
		SELECT c.id, c.event_id, c.name, COALESCE(c.description, ''), c.sort_order,
		       COUNT(a.id)
		FROM learn_cases c
		LEFT JOIN learn_algorithms a ON a.case_id = c.id
		GROUP BY c.id, c.event_id, c.name, c.description, c.sort_order
		ORDER BY c.sort_order, c.name
	`)
}

func (r *PostgresRepository) queryCases(query string, args ...any) ([]Case, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cases := make([]Case, 0)
	for rows.Next() {
		var c Case
		if err := rows.Scan(
			&c.ID, &c.EventID, &c.Name, &c.Description, &c.Order, &c.AlgorithmCount,
		); err != nil {
			return nil, err
		}
		cases = append(cases, c)
	}

	return cases, rows.Err()
}

func (r *PostgresRepository) GetCase(id string) (Case, error) {
	var c Case
	err := r.db.QueryRow(`
		SELECT c.id, c.event_id, c.name, COALESCE(c.description, ''), c.sort_order,
		       (SELECT COUNT(*) FROM learn_algorithms a WHERE a.case_id = c.id)
		FROM learn_cases c
		WHERE c.id = $1
	`, id).Scan(&c.ID, &c.EventID, &c.Name, &c.Description, &c.Order, &c.AlgorithmCount)

	if errors.Is(err, sql.ErrNoRows) {
		return Case{}, ErrCaseNotFound
	}
	return c, err
}

func (r *PostgresRepository) CreateCase(c Case) error {
	_, err := r.db.Exec(`
		INSERT INTO learn_cases (id, event_id, name, description, sort_order)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5)
	`, c.ID, c.EventID, c.Name, c.Description, c.Order)

	return err
}

// UpdateCase: az event_id is módosítható — így lehet egy esetet másik
// versenyszám alá mozgatni anélkül, hogy újra kellene rögzíteni.
func (r *PostgresRepository) UpdateCase(c Case) error {
	res, err := r.db.Exec(`
		UPDATE learn_cases
		SET event_id = $2, name = $3, description = NULLIF($4, ''), sort_order = $5,
		    updated_at = now()
		WHERE id = $1
	`, c.ID, c.EventID, c.Name, c.Description, c.Order)
	if err != nil {
		return err
	}

	return requireAffected(res, ErrCaseNotFound)
}

func (r *PostgresRepository) DeleteCase(id string) error {
	res, err := r.db.Exec(`DELETE FROM learn_cases WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return requireAffected(res, ErrCaseNotFound)
}

// --- Algorithm ---------------------------------------------------------------

func (r *PostgresRepository) ListAlgorithms(caseID string) ([]Algorithm, error) {
	return r.queryAlgorithms(`
		SELECT id, case_id, name, notation, COALESCE(description, ''),
		       COALESCE(video_url, ''), sort_order
		FROM learn_algorithms
		WHERE case_id = $1
		ORDER BY sort_order, name
	`, caseID)
}

func (r *PostgresRepository) ListAllAlgorithms() ([]Algorithm, error) {
	return r.queryAlgorithms(`
		SELECT id, case_id, name, notation, COALESCE(description, ''),
		       COALESCE(video_url, ''), sort_order
		FROM learn_algorithms
		ORDER BY sort_order, name
	`)
}

func (r *PostgresRepository) queryAlgorithms(query string, args ...any) ([]Algorithm, error) {
	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	algorithms := make([]Algorithm, 0)
	for rows.Next() {
		var a Algorithm
		if err := rows.Scan(
			&a.ID, &a.CaseID, &a.Name, &a.Notation, &a.Description, &a.VideoURL, &a.Order,
		); err != nil {
			return nil, err
		}
		algorithms = append(algorithms, a)
	}

	return algorithms, rows.Err()
}

func (r *PostgresRepository) GetAlgorithm(id string) (Algorithm, error) {
	var a Algorithm
	err := r.db.QueryRow(`
		SELECT id, case_id, name, notation, COALESCE(description, ''),
		       COALESCE(video_url, ''), sort_order
		FROM learn_algorithms
		WHERE id = $1
	`, id).Scan(&a.ID, &a.CaseID, &a.Name, &a.Notation, &a.Description, &a.VideoURL, &a.Order)

	if errors.Is(err, sql.ErrNoRows) {
		return Algorithm{}, ErrAlgorithmNotFound
	}
	return a, err
}

func (r *PostgresRepository) CreateAlgorithm(a Algorithm) error {
	_, err := r.db.Exec(`
		INSERT INTO learn_algorithms (id, case_id, name, notation, description, video_url, sort_order)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''), NULLIF($6, ''), $7)
	`, a.ID, a.CaseID, a.Name, a.Notation, a.Description, a.VideoURL, a.Order)

	return err
}

func (r *PostgresRepository) UpdateAlgorithm(a Algorithm) error {
	res, err := r.db.Exec(`
		UPDATE learn_algorithms
		SET case_id = $2, name = $3, notation = $4, description = NULLIF($5, ''),
		    video_url = NULLIF($6, ''), sort_order = $7, updated_at = now()
		WHERE id = $1
	`, a.ID, a.CaseID, a.Name, a.Notation, a.Description, a.VideoURL, a.Order)
	if err != nil {
		return err
	}

	return requireAffected(res, ErrAlgorithmNotFound)
}

func (r *PostgresRepository) DeleteAlgorithm(id string) error {
	res, err := r.db.Exec(`DELETE FROM learn_algorithms WHERE id = $1`, id)
	if err != nil {
		return err
	}

	return requireAffected(res, ErrAlgorithmNotFound)
}

// requireAffected: ha nincs ilyen id, 0 sor módosul — ezt 404-ként akarjuk
// jelezni, nem néma sikerként (a handler fordítja HTTP státuszra).
func requireAffected(res sql.Result, notFound error) error {
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return notFound
	}
	return nil
}
