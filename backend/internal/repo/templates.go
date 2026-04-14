package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
)

type TemplatesRepo struct{ DB *pgxpool.Pool }

func NewTemplatesRepo(db *pgxpool.Pool) *TemplatesRepo { return &TemplatesRepo{DB: db} }

const tmplCols = `id, slug, name, kind, body, variables, tags, created_by, created_at, updated_at`

func scanTmpl(r pgx.Row) (*models.Template, error) {
	t := &models.Template{}
	err := r.Scan(&t.ID, &t.Slug, &t.Name, &t.Kind, &t.Body, &t.Variables, &t.Tags,
		&t.CreatedBy, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func (r *TemplatesRepo) List(ctx context.Context, kind string) ([]*models.Template, error) {
	q := `SELECT ` + tmplCols + ` FROM templates WHERE 1=1`
	args := []any{}
	if kind != "" {
		args = append(args, kind)
		q += " AND kind=$1"
	}
	q += " ORDER BY name ASC"
	rows, err := r.DB.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Template
	for rows.Next() {
		t, err := scanTmpl(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

func (r *TemplatesRepo) Get(ctx context.Context, id string) (*models.Template, error) {
	row := r.DB.QueryRow(ctx, `SELECT `+tmplCols+` FROM templates WHERE id::text=$1 OR slug=$1`, id)
	t, err := scanTmpl(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return t, err
}

func (r *TemplatesRepo) Create(ctx context.Context, t *models.Template) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO templates (slug, name, kind, body, variables, tags, created_by)
		VALUES ($1,$2,$3,$4,COALESCE($5,'[]'::jsonb),$6,$7)
		RETURNING id, created_at, updated_at`,
		t.Slug, t.Name, t.Kind, t.Body, t.Variables, t.Tags, t.CreatedBy,
	).Scan(&t.ID, &t.CreatedAt, &t.UpdatedAt)
}

func (r *TemplatesRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.DB.Exec(ctx, `DELETE FROM templates WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
