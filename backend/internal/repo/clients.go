package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
)

type ClientsRepo struct{ DB *pgxpool.Pool }

func NewClientsRepo(db *pgxpool.Pool) *ClientsRepo { return &ClientsRepo{DB: db} }

const clientCols = `id, slug, name, kind, tier, status, metrics, cross_links, verticals, tags,
onboarded_at, metadata, created_at, updated_at`

func scanClient(r pgx.Row) (*models.Client, error) {
	c := &models.Client{}
	err := r.Scan(&c.ID, &c.Slug, &c.Name, &c.Kind, &c.Tier, &c.Status,
		&c.Metrics, &c.CrossLinks, &c.Verticals, &c.Tags, &c.OnboardedAt,
		&c.Metadata, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type ClientsFilter struct {
	Vertical string
	Status   string
	Limit    int
	Offset   int
}

func (r *ClientsRepo) List(ctx context.Context, f ClientsFilter) ([]*models.Client, error) {
	q := `SELECT ` + clientCols + ` FROM clients WHERE 1=1`
	args := []any{}
	if f.Vertical != "" {
		args = append(args, f.Vertical)
		q += " AND $" + itoa(len(args)) + " = ANY(verticals)"
	}
	if f.Status != "" {
		args = append(args, f.Status)
		q += " AND status=$" + itoa(len(args))
	}
	q += " ORDER BY name ASC"
	if f.Limit > 0 {
		args = append(args, f.Limit)
		q += " LIMIT $" + itoa(len(args))
	}
	rows, err := r.DB.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Client
	for rows.Next() {
		c, err := scanClient(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ClientsRepo) Get(ctx context.Context, id string) (*models.Client, error) {
	row := r.DB.QueryRow(ctx, `SELECT `+clientCols+` FROM clients WHERE id=$1 OR slug=$1`, id)
	c, err := scanClient(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *ClientsRepo) Create(ctx context.Context, c *models.Client) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO clients (slug, name, kind, tier, status, metrics, cross_links, verticals, tags, onboarded_at, metadata)
		VALUES ($1,$2,$3,$4,COALESCE($5,'active'),COALESCE($6,'{}'::jsonb),COALESCE($7,'[]'::jsonb),$8,$9,$10,COALESCE($11,'{}'::jsonb))
		RETURNING id, created_at, updated_at`,
		c.Slug, c.Name, c.Kind, c.Tier, c.Status, c.Metrics, c.CrossLinks,
		c.Verticals, c.Tags, c.OnboardedAt, c.Metadata,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *ClientsRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.DB.Exec(ctx, `DELETE FROM clients WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
