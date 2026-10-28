package certificate

import (
	"database/sql"
	"errors"
)

type Repository interface {
	ListTemplates() ([]Template, error)
	GetTemplate(id string) (Template, error)
	CreateTemplate(id string, input TemplateInput, userID int) error
	UpdateTemplate(id string, input TemplateInput, version int, userID int) error
	DeleteTemplate(id string) error
	ListAssets() ([]Asset, error)
	GetAssetData(id string) (AssetData, error)
	CreateAsset(asset Asset, data []byte, userID int) error
	DeleteAsset(id string) error
}

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) *PostgresRepository {
	return &PostgresRepository{db: db}
}

const templateColumns = `
	t.id, t.name, t.description, t.kind, t.document, t.version,
	COALESCE(u.name, ''), t.created_at, t.updated_at`

type scanner interface {
	Scan(dest ...any) error
}

func scanTemplate(row scanner) (Template, error) {
	var t Template
	var document []byte
	err := row.Scan(&t.ID, &t.Name, &t.Description, &t.Kind, &document, &t.Version, &t.UpdatedByName, &t.CreatedAt, &t.UpdatedAt)
	t.Document = document
	return t, err
}

func (r *PostgresRepository) ListTemplates() ([]Template, error) {
	rows, err := r.db.Query(`
		SELECT` + templateColumns + `
		FROM certificate_templates t
		LEFT JOIN users u ON u.id = t.updated_by
		ORDER BY t.updated_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	templates := make([]Template, 0)
	for rows.Next() {
		t, err := scanTemplate(rows)
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}
	return templates, rows.Err()
}

func (r *PostgresRepository) GetTemplate(id string) (Template, error) {
	t, err := scanTemplate(r.db.QueryRow(`
		SELECT`+templateColumns+`
		FROM certificate_templates t
		LEFT JOIN users u ON u.id = t.updated_by
		WHERE t.id = $1
	`, id))
	if errors.Is(err, sql.ErrNoRows) {
		return Template{}, ErrTemplateNotFound
	}
	return t, err
}

func (r *PostgresRepository) CreateTemplate(id string, input TemplateInput, userID int) error {
	_, err := r.db.Exec(`
		INSERT INTO certificate_templates (id, name, description, kind, document, created_by, updated_by)
		VALUES ($1, $2, $3, $4, $5, $6, $6)
	`, id, input.Name, input.Description, input.Kind, []byte(input.Document), userID)
	return err
}

func (r *PostgresRepository) UpdateTemplate(id string, input TemplateInput, version int, userID int) error {
	res, err := r.db.Exec(`
		UPDATE certificate_templates
		SET name = $2, description = $3, document = $4, version = version + 1,
		    updated_by = $5, updated_at = NOW()
		WHERE id = $1 AND version = $6
	`, id, input.Name, input.Description, []byte(input.Document), userID, version)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil || n > 0 {
		return err
	}

	var exists bool
	if err := r.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM certificate_templates WHERE id = $1)`, id).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return ErrVersionConflict
	}
	return ErrTemplateNotFound
}

func (r *PostgresRepository) DeleteTemplate(id string) error {
	res, err := r.db.Exec(`DELETE FROM certificate_templates WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrTemplateNotFound
	}
	return nil
}

func (r *PostgresRepository) ListAssets() ([]Asset, error) {
	rows, err := r.db.Query(`
		SELECT id, name, mime, width, height, size, created_at
		FROM certificate_assets
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	assets := make([]Asset, 0)
	for rows.Next() {
		var a Asset
		if err := rows.Scan(&a.ID, &a.Name, &a.Mime, &a.Width, &a.Height, &a.Size, &a.CreatedAt); err != nil {
			return nil, err
		}
		assets = append(assets, a)
	}
	return assets, rows.Err()
}

func (r *PostgresRepository) GetAssetData(id string) (AssetData, error) {
	var asset AssetData
	err := r.db.QueryRow(`SELECT mime, data FROM certificate_assets WHERE id = $1`, id).Scan(&asset.Mime, &asset.Data)
	if errors.Is(err, sql.ErrNoRows) {
		return AssetData{}, ErrAssetNotFound
	}
	return asset, err
}

func (r *PostgresRepository) CreateAsset(asset Asset, data []byte, userID int) error {
	_, err := r.db.Exec(`
		INSERT INTO certificate_assets (id, name, mime, width, height, size, data, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, asset.ID, asset.Name, asset.Mime, asset.Width, asset.Height, asset.Size, data, userID)
	return err
}

func (r *PostgresRepository) DeleteAsset(id string) error {
	res, err := r.db.Exec(`
		DELETE FROM certificate_assets a
		WHERE a.id = $1::uuid
		  AND NOT EXISTS (
		    SELECT 1 FROM certificate_templates t
		    WHERE jsonb_path_exists(t.document, '$.elements[*] ? (@.assetId == $id)', jsonb_build_object('id', $1::text))
		  )
	`, id)
	if err != nil {
		return err
	}
	if n, err := res.RowsAffected(); err != nil || n > 0 {
		return err
	}

	var exists bool
	if err := r.db.QueryRow(`SELECT EXISTS (SELECT 1 FROM certificate_assets WHERE id = $1)`, id).Scan(&exists); err != nil {
		return err
	}
	if exists {
		return ErrAssetInUse
	}
	return ErrAssetNotFound
}
