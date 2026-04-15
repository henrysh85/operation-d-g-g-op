package repo

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
)

type ConsultationsRepo struct{ DB *pgxpool.Pool }

func NewConsultationsRepo(db *pgxpool.Pool) *ConsultationsRepo { return &ConsultationsRepo{DB: db} }

const consultCols = `id, jurisdiction_id, vertical, title, deadline, status, assignee_id,
summary, metadata, created_at, updated_at`

func scanConsult(r pgx.Row) (*models.Consultation, error) {
	c := &models.Consultation{}
	err := r.Scan(&c.ID, &c.JurisdictionID, &c.Vertical, &c.Title, &c.Deadline, &c.Status,
		&c.AssigneeID, &c.Summary, &c.Metadata, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type ConsultFilter struct {
	Vertical   string
	Status     string
	AssigneeID string
	Search     string
	Before     *time.Time
	Limit      int
	Offset     int
}

func (r *ConsultationsRepo) List(ctx context.Context, f ConsultFilter) ([]*models.Consultation, error) {
	q := `SELECT ` + consultCols + ` FROM consultations WHERE 1=1`
	args := []any{}
	if f.Vertical != "" {
		args = append(args, f.Vertical)
		q += " AND vertical=$" + itoa(len(args))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		q += " AND status=$" + itoa(len(args))
	}
	if f.AssigneeID != "" {
		args = append(args, f.AssigneeID)
		q += " AND assignee_id=$" + itoa(len(args))
	}
	if f.Before != nil {
		args = append(args, *f.Before)
		q += " AND deadline <= $" + itoa(len(args))
	}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		q += " AND title ILIKE $" + itoa(len(args))
	}
	q += " ORDER BY deadline ASC NULLS LAST"
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
	var out []*models.Consultation
	for rows.Next() {
		c, err := scanConsult(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *ConsultationsRepo) Get(ctx context.Context, id string) (*models.Consultation, error) {
	row := r.DB.QueryRow(ctx, `SELECT `+consultCols+` FROM consultations WHERE id=$1`, id)
	c, err := scanConsult(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *ConsultationsRepo) Create(ctx context.Context, c *models.Consultation) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO consultations (jurisdiction_id, vertical, title, deadline, status, assignee_id, summary, metadata)
		VALUES ($1,$2,$3,$4,COALESCE($5,'open'),$6,$7,COALESCE($8,'{}'::jsonb))
		RETURNING id, created_at, updated_at`,
		c.JurisdictionID, c.Vertical, c.Title, c.Deadline, c.Status, c.AssigneeID, c.Summary, c.Metadata,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *ConsultationsRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.DB.Exec(ctx, `DELETE FROM consultations WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
