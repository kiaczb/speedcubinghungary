package post

import (
	"database/sql"

	"github.com/lib/pq"
)

type Repository interface {
	GetAll() ([]Post, error)
	GetLatestPosts() ([]Post, error)
	GetByID(id string) (Post, error)
	Create(post Post) error
	Update(post Post) error
	Delete(ids []string) error
}


type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// imagePathsParam: az image_paths oszlop NOT NULL, a pq.Array viszont nil
// szeletből NULL-t csinálna — kép nélküli poszthoz üres tömb kell.
func imagePathsParam(paths []string) any {
	if paths == nil {
		paths = []string{}
	}
	return pq.Array(paths)
}

func (r *PostgresRepository) GetAll() ([]Post, error) {
	rows, err := r.db.Query(`
		SELECT id, tag, title, content, image_paths, created_at
		FROM posts
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts = make([]Post, 0)
	for rows.Next() {
		var p Post
		if err := rows.Scan(&p.ID, &p.Tag, &p.Title, &p.Content, pq.Array(&p.ImagePaths), &p.CreatedAt); err != nil {
			return nil, err
		}
		if p.ImagePaths == nil {
			p.ImagePaths = []string{}
		}
		posts = append(posts, p)
	}

	return posts, rows.Err()
}
func (r *PostgresRepository) GetLatestPosts() ([]Post, error){
	rows, err := r.db.Query(`
		SELECT id, tag, title, content, image_paths, created_at
		FROM posts
		ORDER BY created_at DESC
		LIMIT 3
	`)
	if err != nil {
	  return nil, err
	}
	defer rows.Close()

	var posts = make([]Post, 0)

	for rows.Next(){
		var p Post
		err := rows.Scan(&p.ID, &p.Tag, &p.Title, &p.Content, pq.Array(&p.ImagePaths), &p.CreatedAt)
		if err != nil{
			return nil, err
		}
		if p.ImagePaths == nil {
			p.ImagePaths = []string{}
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}
func (r *PostgresRepository) GetByID(id string) (Post, error) {
	var p Post
	err := r.db.QueryRow(`
		SELECT id, tag, title, content, image_paths, created_at
		FROM posts
		WHERE id = $1
	`, id).Scan(&p.ID, &p.Tag, &p.Title, &p.Content, pq.Array(&p.ImagePaths), &p.CreatedAt)
	if p.ImagePaths == nil {
		p.ImagePaths = []string{}
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

func (r *PostgresRepository) Delete(ids []string) error {
    if len(ids) == 0 {
        return nil
    }

    _, err := r.db.Exec(`
        DELETE FROM posts
        WHERE id = ANY($1)
    `, pq.Array(ids))

    return err
}

func (r *PostgresRepository) Update(p Post) error {
	// A report_week-et szándékosan nem írjuk felül: az a heti riport idempotencia-
	// kulcsa, a szerkesztés csak a megjelenített mezőket érinti.
	res, err := r.db.Exec(`
		UPDATE posts
		SET tag = $2, title = $3, content = $4, image_paths = $5, updated_at = now()
		WHERE id = $1
	`, p.ID, p.Tag, p.Title, p.Content, imagePathsParam(p.ImagePaths))
	if err != nil {
		return err
	}

	// Ha nincs ilyen id, 0 sor módosul — ezt 404-ként akarjuk jelezni, nem néma
	// sikerként (a handler az ErrPostNotFound-ot fordítja HTTP státuszra).
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrPostNotFound
	}

	return nil
}
