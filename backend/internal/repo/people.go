package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
)

type PeopleRepo struct{ DB *pgxpool.Pool }

func NewPeopleRepo(db *pgxpool.Pool) *PeopleRepo { return &PeopleRepo{DB: db} }

const peopleCols = `id, name, email, dept, title, reports_to, status, location, country_code,
salary, currency, holiday_quota, photo_key, start_date, metadata, created_at, updated_at`

func scanPerson(r pgx.Row) (*models.Person, error) {
	p := &models.Person{}
	err := r.Scan(&p.ID, &p.Name, &p.Email, &p.Dept, &p.Title, &p.ReportsTo, &p.Status,
		&p.Location, &p.CountryCode, &p.Salary, &p.Currency, &p.HolidayQuota, &p.PhotoKey,
		&p.StartDate, &p.Metadata, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return p, nil
}

type PeopleFilter struct {
	Dept   string
	Status string
	Search string
	Limit  int
	Offset int
}

func (r *PeopleRepo) List(ctx context.Context, f PeopleFilter) ([]*models.Person, error) {
	q := `SELECT ` + peopleCols + ` FROM people WHERE 1=1`
	args := []any{}
	if f.Dept != "" {
		args = append(args, f.Dept)
		q += " AND dept = $" + itoa(len(args))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		q += " AND status = $" + itoa(len(args))
	}
	if f.Search != "" {
		args = append(args, "%"+f.Search+"%")
		q += " AND name ILIKE $" + itoa(len(args))
	}
	q += " ORDER BY name ASC"
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
	var out []*models.Person
	for rows.Next() {
		p, err := scanPerson(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *PeopleRepo) Get(ctx context.Context, id string) (*models.Person, error) {
	row := r.DB.QueryRow(ctx, `SELECT `+peopleCols+` FROM people WHERE id=$1`, id)
	p, err := scanPerson(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return p, err
}

func (r *PeopleRepo) Create(ctx context.Context, p *models.Person) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO people (name, email, dept, title, reports_to, status, location, country_code,
			salary, currency, holiday_quota, photo_key, start_date, metadata)
		VALUES ($1,$2,$3,$4,$5,COALESCE($6,'active'),$7,$8,$9,COALESCE($10,'USD'),COALESCE($11,25),$12,$13,COALESCE($14,'{}'::jsonb))
		RETURNING id, created_at, updated_at`,
		p.Name, p.Email, p.Dept, p.Title, p.ReportsTo, p.Status, p.Location, p.CountryCode,
		p.Salary, p.Currency, p.HolidayQuota, p.PhotoKey, p.StartDate, p.Metadata,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *PeopleRepo) Patch(ctx context.Context, id string, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	q := "UPDATE people SET updated_at=NOW()"
	args := []any{}
	for k, v := range fields {
		args = append(args, v)
		q += ", " + k + "=$" + itoa(len(args))
	}
	args = append(args, id)
	q += " WHERE id=$" + itoa(len(args))
	ct, err := r.DB.Exec(ctx, q, args...)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PeopleRepo) Delete(ctx context.Context, id string) error {
	ct, err := r.DB.Exec(ctx, `DELETE FROM people WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
