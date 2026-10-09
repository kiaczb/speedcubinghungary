package learn

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Repository interface {
	ResolvePath(path []string) ([]Node, error)
	GetChildren(parentID *string) ([]Node, error)
	GetAlgorithms(nodeID string) ([]Algorithm, error)
	GetAllNodes() ([]Node, error)
	GetNode(id string) (Node, error)
	CreateNode(id string, in NodeInput) error
	UpdateNode(id string, in NodeInput) error
	DeleteNode(id string) error
	ReorderNodes(parentID *string, ids []string) error
	CreateAlgorithm(id, nodeID string, in AlgorithmInput) error
	UpdateAlgorithm(id string, in AlgorithmInput) (string, error)
	DeleteAlgorithm(id string) error
	ReorderAlgorithms(nodeID string, ids []string) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

const nodeColumns = `n.id, n.parent_id, n.slug, n.title, n.description, n.icon, n.image_url,
	(SELECT count(*) FROM learn_nodes c WHERE c.parent_id = n.id),
	(SELECT count(*) FROM learn_algorithms a WHERE a.node_id = n.id)`

type rowScanner interface {
	Scan(dest ...any) error
}

func scanNode(row rowScanner) (Node, error) {
	var (
		n        Node
		parentID sql.NullString
	)
	if err := row.Scan(&n.ID, &parentID, &n.Slug, &n.Title, &n.Description, &n.Icon, &n.ImageURL, &n.ChildCount, &n.AlgorithmCount); err != nil {
		return Node{}, err
	}
	if parentID.Valid {
		n.ParentID = &parentID.String
	}
	return n, nil
}

func scanNodes(rows *sql.Rows) ([]Node, error) {
	defer rows.Close()

	nodes := make([]Node, 0)
	for rows.Next() {
		n, err := scanNode(rows)
		if err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, rows.Err()
}

func nullableID(id *string) any {
	if id == nil {
		return nil
	}
	return *id
}

func (r *PostgresRepository) ResolvePath(path []string) ([]Node, error) {
	if len(path) == 0 {
		return []Node{}, nil
	}

	rows, err := r.db.Query(`
		WITH RECURSIVE walk AS (
			SELECT n.id, 1 AS depth
			FROM learn_nodes n
			WHERE n.parent_id IS NULL AND n.slug = ($1::text[])[1]
			UNION ALL
			SELECT c.id, w.depth + 1
			FROM walk w
			JOIN learn_nodes c ON c.parent_id = w.id
			WHERE c.slug = ($1::text[])[w.depth + 1]
		)
		SELECT `+nodeColumns+`
		FROM walk w
		JOIN learn_nodes n ON n.id = w.id
		ORDER BY w.depth
	`, pq.Array(path))
	if err != nil {
		return nil, err
	}
	return scanNodes(rows)
}

func (r *PostgresRepository) GetChildren(parentID *string) ([]Node, error) {
	rows, err := r.db.Query(`
		SELECT `+nodeColumns+`
		FROM learn_nodes n
		WHERE n.parent_id IS NOT DISTINCT FROM $1::uuid
		ORDER BY n.sort_order, n.title
	`, nullableID(parentID))
	if err != nil {
		return nil, err
	}
	return scanNodes(rows)
}

func (r *PostgresRepository) GetAllNodes() ([]Node, error) {
	rows, err := r.db.Query(`
		SELECT ` + nodeColumns + `
		FROM learn_nodes n
		ORDER BY n.sort_order, n.title
	`)
	if err != nil {
		return nil, err
	}
	return scanNodes(rows)
}

func (r *PostgresRepository) GetNode(id string) (Node, error) {
	n, err := scanNode(r.db.QueryRow(`
		SELECT `+nodeColumns+`
		FROM learn_nodes n
		WHERE n.id = $1
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Node{}, ErrNodeNotFound
	}
	if err != nil {
		return Node{}, err
	}

	rows, err := r.db.Query(`
		WITH RECURSIVE up AS (
			SELECT id, parent_id, slug, 0 AS height
			FROM learn_nodes
			WHERE id = $1
			UNION ALL
			SELECT p.id, p.parent_id, p.slug, up.height + 1
			FROM learn_nodes p
			JOIN up ON p.id = up.parent_id
		)
		SELECT slug FROM up ORDER BY height DESC
	`, id)
	if err != nil {
		return Node{}, err
	}
	defer rows.Close()

	n.Path = make([]string, 0)
	for rows.Next() {
		var slug string
		if err := rows.Scan(&slug); err != nil {
			return Node{}, err
		}
		n.Path = append(n.Path, slug)
	}
	return n, rows.Err()
}

func (r *PostgresRepository) GetAlgorithms(nodeID string) ([]Algorithm, error) {
	rows, err := r.db.Query(`
		SELECT id, node_id, notation, name, notes
		FROM learn_algorithms
		WHERE node_id = $1
		ORDER BY sort_order, created_at
	`, nodeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	algorithms := make([]Algorithm, 0)
	for rows.Next() {
		var a Algorithm
		if err := rows.Scan(&a.ID, &a.NodeID, &a.Notation, &a.Name, &a.Notes); err != nil {
			return nil, err
		}
		algorithms = append(algorithms, a)
	}
	return algorithms, rows.Err()
}

func (r *PostgresRepository) inTreeTx(fn func(tx *sql.Tx) error) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`SELECT pg_advisory_xact_lock(hashtext('learn_nodes'))`); err != nil {
		return err
	}
	if err := fn(tx); err != nil {
		return err
	}
	return tx.Commit()
}

func checkParent(tx *sql.Tx, parentID string, subtreeHeight int) error {
	var depth, algorithms int
	err := tx.QueryRow(`
		WITH RECURSIVE up AS (
			SELECT id, parent_id FROM learn_nodes WHERE id = $1
			UNION ALL
			SELECT p.id, p.parent_id FROM learn_nodes p JOIN up ON p.id = up.parent_id
		)
		SELECT (SELECT count(*) FROM up),
		       (SELECT count(*) FROM learn_algorithms WHERE node_id = $1)
	`, parentID).Scan(&depth, &algorithms)
	if err != nil {
		return err
	}

	switch {
	case depth == 0:
		return ErrParentNotFound
	case algorithms > 0:
		return ErrParentHasAlgorithms
	case depth+subtreeHeight > MaxDepth:
		return ErrTooDeep
	}
	return nil
}

func checkSlugFree(tx *sql.Tx, parentID *string, slug, nodeID string) error {
	var taken bool
	err := tx.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM learn_nodes
			WHERE parent_id IS NOT DISTINCT FROM $1::uuid AND slug = $2 AND id <> $3::uuid
		)
	`, nullableID(parentID), slug, nodeID).Scan(&taken)
	if err != nil {
		return err
	}
	if taken {
		return ErrSlugTaken
	}
	return nil
}

func (r *PostgresRepository) CreateNode(id string, in NodeInput) error {
	return r.inTreeTx(func(tx *sql.Tx) error {
		if in.ParentID != nil {
			if err := checkParent(tx, *in.ParentID, 1); err != nil {
				return err
			}
		}
		if err := checkSlugFree(tx, in.ParentID, in.Slug, id); err != nil {
			return err
		}

		_, err := tx.Exec(`
			INSERT INTO learn_nodes (id, parent_id, slug, title, description, icon, image_url, sort_order)
			VALUES ($1, $2, $3, $4, $5, $6, $7, (
				SELECT COALESCE(MAX(sort_order) + 1, 0)
				FROM learn_nodes
				WHERE parent_id IS NOT DISTINCT FROM $2::uuid
			))
		`, id, nullableID(in.ParentID), in.Slug, in.Title, in.Description, in.Icon, in.ImageURL)
		return err
	})
}

func sameParent(current sql.NullString, next *string) bool {
	if !current.Valid || next == nil {
		return !current.Valid && next == nil
	}
	return current.String == *next
}

func (r *PostgresRepository) UpdateNode(id string, in NodeInput) error {
	return r.inTreeTx(func(tx *sql.Tx) error {
		var current sql.NullString
		err := tx.QueryRow(`SELECT parent_id FROM learn_nodes WHERE id = $1`, id).Scan(&current)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrNodeNotFound
		}
		if err != nil {
			return err
		}

		moving := !sameParent(current, in.ParentID)
		if moving && in.ParentID != nil {
			var (
				height      int
				intoOwnTree bool
			)
			err := tx.QueryRow(`
				WITH RECURSIVE down AS (
					SELECT id, 1 AS height FROM learn_nodes WHERE id = $1
					UNION ALL
					SELECT c.id, d.height + 1 FROM learn_nodes c JOIN down d ON c.parent_id = d.id
				)
				SELECT max(height), bool_or(id = $2::uuid) FROM down
			`, id, *in.ParentID).Scan(&height, &intoOwnTree)
			if err != nil {
				return err
			}
			if intoOwnTree {
				return ErrMoveIntoOwnSubtree
			}
			if err := checkParent(tx, *in.ParentID, height); err != nil {
				return err
			}
		}

		if err := checkSlugFree(tx, in.ParentID, in.Slug, id); err != nil {
			return err
		}

		if !moving {
			_, err = tx.Exec(`
				UPDATE learn_nodes
				SET slug = $2, title = $3, description = $4, icon = $5, image_url = $6, updated_at = now()
				WHERE id = $1
			`, id, in.Slug, in.Title, in.Description, in.Icon, in.ImageURL)
			return err
		}

		_, err = tx.Exec(`
			UPDATE learn_nodes
			SET parent_id = $2, slug = $3, title = $4, description = $5, icon = $6, image_url = $7, updated_at = now(),
			    sort_order = (
			        SELECT COALESCE(MAX(sort_order) + 1, 0)
			        FROM learn_nodes
			        WHERE parent_id IS NOT DISTINCT FROM $2::uuid
			    )
			WHERE id = $1
		`, id, nullableID(in.ParentID), in.Slug, in.Title, in.Description, in.Icon, in.ImageURL)
		return err
	})
}

func (r *PostgresRepository) DeleteNode(id string) error {
	return r.inTreeTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(`DELETE FROM learn_nodes WHERE id = $1`, id)
		if err != nil {
			return err
		}
		return requireAffected(res, ErrNodeNotFound)
	})
}

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

func sameSet(rows *sql.Rows, ids []string) (bool, error) {
	defer rows.Close()

	wanted := make(map[string]bool, len(ids))
	for _, id := range ids {
		wanted[id] = true
	}

	count := 0
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return false, err
		}
		if !wanted[id] {
			return false, rows.Err()
		}
		count++
	}
	return count == len(ids), rows.Err()
}

func (r *PostgresRepository) ReorderNodes(parentID *string, ids []string) error {
	return r.inTreeTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(`
			SELECT id FROM learn_nodes WHERE parent_id IS NOT DISTINCT FROM $1::uuid
		`, nullableID(parentID))
		if err != nil {
			return err
		}
		matches, err := sameSet(rows, ids)
		if err != nil {
			return err
		}
		if !matches {
			return ErrOrderMismatch
		}

		_, err = tx.Exec(`
			UPDATE learn_nodes n
			SET sort_order = o.position - 1
			FROM unnest($1::uuid[]) WITH ORDINALITY AS o(id, position)
			WHERE n.id = o.id
		`, pq.Array(ids))
		return err
	})
}

func (r *PostgresRepository) CreateAlgorithm(id, nodeID string, in AlgorithmInput) error {
	return r.inTreeTx(func(tx *sql.Tx) error {
		var (
			exists   bool
			children int
		)
		err := tx.QueryRow(`
			SELECT EXISTS (SELECT 1 FROM learn_nodes WHERE id = $1),
			       (SELECT count(*) FROM learn_nodes WHERE parent_id = $1)
		`, nodeID).Scan(&exists, &children)
		if err != nil {
			return err
		}
		if !exists {
			return ErrNodeNotFound
		}
		if children > 0 {
			return ErrNodeHasChildren
		}

		_, err = tx.Exec(`
			INSERT INTO learn_algorithms (id, node_id, notation, name, notes, sort_order)
			VALUES ($1, $2, $3, $4, $5, (
				SELECT COALESCE(MAX(sort_order) + 1, 0)
				FROM learn_algorithms
				WHERE node_id = $2
			))
		`, id, nodeID, in.Notation, in.Name, in.Notes)
		return err
	})
}

func (r *PostgresRepository) UpdateAlgorithm(id string, in AlgorithmInput) (string, error) {
	var nodeID string
	err := r.inTreeTx(func(tx *sql.Tx) error {
		err := tx.QueryRow(`
			UPDATE learn_algorithms
			SET notation = $2, name = $3, notes = $4, updated_at = now()
			WHERE id = $1
			RETURNING node_id
		`, id, in.Notation, in.Name, in.Notes).Scan(&nodeID)
		if errors.Is(err, sql.ErrNoRows) {
			return ErrAlgorithmNotFound
		}
		return err
	})
	return nodeID, err
}

func (r *PostgresRepository) DeleteAlgorithm(id string) error {
	return r.inTreeTx(func(tx *sql.Tx) error {
		res, err := tx.Exec(`DELETE FROM learn_algorithms WHERE id = $1`, id)
		if err != nil {
			return err
		}
		return requireAffected(res, ErrAlgorithmNotFound)
	})
}

func (r *PostgresRepository) ReorderAlgorithms(nodeID string, ids []string) error {
	return r.inTreeTx(func(tx *sql.Tx) error {
		rows, err := tx.Query(`SELECT id FROM learn_algorithms WHERE node_id = $1`, nodeID)
		if err != nil {
			return err
		}
		matches, err := sameSet(rows, ids)
		if err != nil {
			return err
		}
		if !matches {
			return ErrOrderMismatch
		}

		_, err = tx.Exec(`
			UPDATE learn_algorithms a
			SET sort_order = o.position - 1
			FROM unnest($1::uuid[]) WITH ORDINALITY AS o(id, position)
			WHERE a.id = o.id
		`, pq.Array(ids))
		return err
	})
}
