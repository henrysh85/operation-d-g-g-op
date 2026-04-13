package repo

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
)

type RegulatoryRepo struct{ DB *pgxpool.Pool }

func NewRegulatoryRepo(db *pgxpool.Pool) *RegulatoryRepo { return &RegulatoryRepo{DB: db} }

const jsCols = `id, country_id, vertical, status, headline, regulators, timeline, impact_matrix,
last_reviewed, created_at, updated_at`

func scanJS(r pgx.Row) (*models.JurisdictionStatus, error) {
	j := &models.JurisdictionStatus{}
	err := r.Scan(&j.ID, &j.CountryID, &j.Vertical, &j.Status, &j.Headline, &j.Regulators,
		&j.Timeline, &j.ImpactMatrix, &j.LastReviewed, &j.CreatedAt, &j.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return j, nil
}

type JSFilter struct {
	Vertical  string
	CountryID string
	Status    string
	Limit     int
}

func (r *RegulatoryRepo) ListJurisdictions(ctx context.Context, f JSFilter) ([]*models.JurisdictionStatus, error) {
	q := `SELECT ` + jsCols + ` FROM jurisdictions_status WHERE 1=1`
	args := []any{}
	if f.Vertical != "" {
		args = append(args, f.Vertical)
		q += " AND vertical=$" + itoa(len(args))
	}
	if f.CountryID != "" {
		args = append(args, f.CountryID)
		q += " AND country_id=$" + itoa(len(args))
	}
	if f.Status != "" {
		args = append(args, f.Status)
		q += " AND status=$" + itoa(len(args))
	}
	q += " ORDER BY updated_at DESC"
	if f.Limit > 0 {
		args = append(args, f.Limit)
		q += " LIMIT $" + itoa(len(args))
	}
	rows, err := r.DB.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.JurisdictionStatus
	for rows.Next() {
		j, err := scanJS(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, j)
	}
	return out, rows.Err()
}

func (r *RegulatoryRepo) GetJurisdiction(ctx context.Context, id string) (*models.JurisdictionStatus, error) {
	row := r.DB.QueryRow(ctx, `SELECT `+jsCols+` FROM jurisdictions_status WHERE id=$1`, id)
	j, err := scanJS(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return j, err
}

func (r *RegulatoryRepo) ListCountries(ctx context.Context) ([]*models.Country, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, code, name, region_id, tier, mena_sub, metadata FROM countries ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Country
	for rows.Next() {
		c := &models.Country{}
		if err := rows.Scan(&c.ID, &c.Code, &c.Name, &c.RegionID, &c.Tier, &c.MenaSub, &c.Metadata); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *RegulatoryRepo) ListRegions(ctx context.Context) ([]*models.Region, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, code, name FROM regions ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Region
	for rows.Next() {
		rg := &models.Region{}
		if err := rows.Scan(&rg.ID, &rg.Code, &rg.Name); err != nil {
			return nil, err
		}
		out = append(out, rg)
	}
	return out, rows.Err()
}
