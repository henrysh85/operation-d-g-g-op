package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
)

type ActivitiesRepo struct{ DB *pgxpool.Pool }

func NewActivitiesRepo(db *pgxpool.Pool) *ActivitiesRepo { return &ActivitiesRepo{DB: db} }

const activityCols = `id, title, description, type, vertical, region_id, country_id, occurred_on,
impact, owner_id, status, highlight, metadata, created_at, updated_at`

func scanActivity(r pgx.Row) (*models.Activity, error) {
	a := &models.Activity{}
	err := r.Scan(&a.ID, &a.Title, &a.Description, &a.Type, &a.Vertical, &a.RegionID,
		&a.CountryID, &a.OccurredOn, &a.Impact, &a.OwnerID, &a.Status, &a.Highlight,
		&a.Metadata, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return a, nil
}

type ActivitiesFilter struct {
	Vertical   string
	RegionID   string
	RegionCode string
	ClientID   string
	OwnerID    string
	From, To   *time.Time
	Search     string
	Limit      int
	Offset     int
}

func (r *ActivitiesRepo) List(ctx context.Context, f ActivitiesFilter) ([]*models.Activity, error) {
	q := `SELECT ` + activityCols + ` FROM activities a WHERE 1=1`
	args := []any{}
	if f.Vertical != "" {
		args = append(args, f.Vertical)
		q += " AND a.vertical=$" + itoa(len(args))
	}
	if f.RegionID != "" {
		args = append(args, f.RegionID)
		q += " AND a.region_id=$" + itoa(len(args))
	}
	if f.RegionCode != "" {
		args = append(args, f.RegionCode)
		q += " AND a.region_id = (SELECT id FROM regions WHERE code=$" + itoa(len(args)) + ")"
	}
	if f.OwnerID != "" {
		args = append(args, f.OwnerID)
		q += " AND a.owner_id=$" + itoa(len(args))
	}
	if f.ClientID != "" {
		args = append(args, f.ClientID)
		q += " AND EXISTS (SELECT 1 FROM activity_clients ac WHERE ac.activity_id=a.id AND ac.client_id=$" + itoa(len(args)) + ")"
	}
	if f.From != nil {
		args = append(args, *f.From)
		q += " AND a.occurred_on >= $" + itoa(len(args))
	}
	if f.To != nil {
		args = append(args, *f.To)
		q += " AND a.occurred_on <= $" + itoa(len(args))
	}
	if f.Search != "" {
		args = append(args, f.Search)
		q += " AND a.search_tsv @@ plainto_tsquery('english', $" + itoa(len(args)) + ")"
	}
	q += " ORDER BY a.occurred_on DESC"
	if f.Limit > 0 {
		args = append(args, f.Limit)
		q += " LIMIT $" + itoa(len(args))
	}
	if f.Offset > 0 {
		args = append(args, f.Offset)
		q += " OFFSET $" + itoa(len(args))
	}
	rows, err := r.DB.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Activity
	for rows.Next() {
		a, err := scanActivity(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

func (r *ActivitiesRepo) Get(ctx context.Context, id string) (*models.Activity, error) {
	row := r.DB.QueryRow(ctx, `SELECT `+activityCols+` FROM activities WHERE id=$1`, id)
	a, err := scanActivity(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return a, err
}

func (r *ActivitiesRepo) Create(ctx context.Context, a *models.Activity) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO activities (title, description, type, vertical, region_id, country_id, occurred_on,
			impact, owner_id, status, highlight, metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,COALESCE($10,'done'),COALESCE($11,false),COALESCE($12,'{}'::jsonb))
		RETURNING id, created_at, updated_at`,
		a.Title, a.Description, a.Type, a.Vertical, a.RegionID, a.CountryID, a.OccurredOn,
		a.Impact, a.OwnerID, a.Status, a.Highlight, a.Metadata,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (r *ActivitiesRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.DB.Exec(ctx, `DELETE FROM activities WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ActivitiesRepo) LinkClient(ctx context.Context, activityID, clientID string) error {
	_, err := r.DB.Exec(ctx,
		`INSERT INTO activity_clients (activity_id, client_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
		activityID, clientID)
	return err
}
