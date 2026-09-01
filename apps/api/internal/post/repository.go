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

func (r *PostgresRepository) GetAll() ([]Post, error) {
	rows, err := r.db.Query(`
		SELECT id, tag, title, content, image_path, created_at
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
		var imagePath sql.NullString
		if err := rows.Scan(&p.ID, &p.Tag, &p.Title, &p.Content, &imagePath, &p.CreatedAt); err != nil {
			return nil, err
		}
		p.ImagePath = imagePath.String
		posts = append(posts, p)
	}

	return posts, rows.Err()
}
func (r *PostgresRepository) GetLatestPosts() ([]Post, error){
	rows, err := r.db.Query(`
		SELECT id, tag, title, content, image_path, created_at
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
		var imagePath sql.NullString
		err := rows.Scan(&p.ID, &p.Tag, &p.Title, &p.Content, &imagePath, &p.CreatedAt)
		if err != nil{
			return nil, err
		}
		p.ImagePath = imagePath.String
		posts = append(posts, p)
	}
	return posts, rows.Err()
}
func (r *PostgresRepository) GetByID(id string) (Post, error) {
	var p Post
	var imagePath sql.NullString
	err := r.db.QueryRow(`
		SELECT id, tag, title, content, image_path, created_at
		FROM posts
		WHERE id = $1
	`, id).Scan(&p.ID, &p.Tag, &p.Title, &p.Content, &imagePath, &p.CreatedAt)
	p.ImagePath = imagePath.String

	return p, err
}

func (r *PostgresRepository) Create(p Post) error {
	// NULLIF: kép nélküli posztnál NULL kerül az oszlopba, nem üres string.
	_, err := r.db.Exec(`
		INSERT INTO posts (id, tag, title, content, image_path)
		VALUES ($1, $2, $3, $4, NULLIF($5, ''))
	`, p.ID, p.Tag, p.Title, p.Content, p.ImagePath)

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
		SET tag = $2, title = $3, content = $4, image_path = NULLIF($5, ''), updated_at = now()
		WHERE id = $1
	`, p.ID, p.Tag, p.Title, p.Content, p.ImagePath)
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