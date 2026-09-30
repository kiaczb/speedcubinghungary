package post

import (
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type Repository interface {
	GetAll() ([]Post, error)
	GetLatestPosts() ([]Post, error)
	GetByID(id string) (Post, error)
	Create(post Post) error
	Update(post Post) error
	Delete(ids []string) ([]string, error)
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func imagePathsParam(paths []string) any {
	if paths == nil {
		paths = []string{}
	}
	return pq.Array(paths)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanPost(row rowScanner) (Post, error) {
	var p Post
	if err := row.Scan(&p.ID, &p.Tag, &p.Title, &p.Content, pq.Array(&p.ImagePaths), &p.CreatedAt); err != nil {
		return Post{}, err
	}
	if p.ImagePaths == nil {
		p.ImagePaths = []string{}
	}
	return p, nil
}

func (r *PostgresRepository) listPosts(query string) ([]Post, error) {
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	posts := make([]Post, 0)
	for rows.Next() {
		p, err := scanPost(rows)
		if err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}

func (r *PostgresRepository) GetAll() ([]Post, error) {
	return r.listPosts(`
		SELECT id, tag, title, content, image_paths, created_at
		FROM posts
		ORDER BY created_at DESC
	`)
}

func (r *PostgresRepository) GetLatestPosts() ([]Post, error) {
	return r.listPosts(`
		SELECT id, tag, title, content, image_paths, created_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT 3
	`)
}

func (r *PostgresRepository) GetByID(id string) (Post, error) {
	p, err := scanPost(r.db.QueryRow(`
		SELECT id, tag, title, content, image_paths, created_at
		FROM posts
		WHERE id = $1
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Post{}, ErrPostNotFound
	}
	return p, err
}

func (r *PostgresRepository) Create(p Post) error {
	_, err := r.db.Exec(`
		INSERT INTO posts (id, tag, title, content, image_paths)
		VALUES ($1, $2, $3, $4, $5)
	`, p.ID, p.Tag, p.Title, p.Content, imagePathsParam(p.ImagePaths))

	return err
}

func (r *PostgresRepository) Delete(ids []string) ([]string, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	rows, err := r.db.Query(`
		DELETE FROM posts
		WHERE id = ANY($1)
		RETURNING image_paths
	`, pq.Array(ids))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var paths []string
	for rows.Next() {
		var postPaths []string
		if err := rows.Scan(pq.Array(&postPaths)); err != nil {
			return paths, err
		}
		paths = append(paths, postPaths...)
	}

	return paths, rows.Err()
}

func (r *PostgresRepository) Update(p Post) error {
	res, err := r.db.Exec(`
		UPDATE posts
		SET tag = $2, title = $3, content = $4, image_paths = $5, updated_at = now()
		WHERE id = $1
	`, p.ID, p.Tag, p.Title, p.Content, imagePathsParam(p.ImagePaths))
	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrPostNotFound
	}

	return nil
}
