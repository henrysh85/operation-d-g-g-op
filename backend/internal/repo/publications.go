package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
)

type PublicationsRepo struct{ DB *pgxpool.Pool }

func NewPublicationsRepo(db *pgxpool.Pool) *PublicationsRepo { return &PublicationsRepo{DB: db} }

const pubCols = `id, title, authors, venue, url, abstract, published_on, verticals, tags, metadata, created_at, updated_at`

func scanPub(r pgx.Row) (*models.Publication, error) {
	p := &models.Publication{}
	err := r.Scan(&p.ID, &p.Title, &p.Authors, &p.Venue, &p.URL, &p.Abstract,
		&p.PublishedOn, &p.Verticals, &p.Tags, &p.Metadata, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

func (r *PublicationsRepo) List(ctx context.Context, vertical string, limit int) ([]*models.Publication, error) {
	q := `SELECT ` + pubCols + ` FROM publications WHERE 1=1`
	args := []any{}
	if vertical != "" {
		args = append(args, vertical)
		q += " AND $" + itoa(len(args)) + " = ANY(verticals)"
	}
	q += " ORDER BY published_on DESC NULLS LAST"
	if limit > 0 {
		args = append(args, limit)
		q += " LIMIT $" + itoa(len(args))
	}
	rows, err := r.DB.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Publication
	for rows.Next() {
		p, err := scanPub(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PublicationsRepo) Get(ctx context.Context, id string) (*models.Publication, error) {
	row := r.DB.QueryRow(ctx, `SELECT `+pubCols+` FROM publications WHERE id=$1`, id)
	p, err := scanPub(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *PublicationsRepo) Create(ctx context.Context, p *models.Publication) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO publications (title, authors, venue, url, abstract, published_on, verticals, tags, metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,COALESCE($9,'{}'::jsonb))
		RETURNING id, created_at, updated_at`,
		p.Title, p.Authors, p.Venue, p.URL, p.Abstract, p.PublishedOn, p.Verticals, p.Tags, p.Metadata,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PublicationsRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.DB.Exec(ctx, `DELETE FROM publications WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
