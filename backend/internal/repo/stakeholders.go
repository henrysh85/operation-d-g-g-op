package repo

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/models"
)

func jsonUnmarshal(b []byte, v any) error { return json.Unmarshal(b, v) }

type StakeholdersRepo struct{ DB *pgxpool.Pool }

func NewStakeholdersRepo(db *pgxpool.Pool) *StakeholdersRepo { return &StakeholdersRepo{DB: db} }

const contactCols = `id, name, email, phone, title, institution_id, group_id, relationship, dcgg_flag,
verticals, tags, notes, metadata, created_at, updated_at`

func scanContact(r pgx.Row) (*models.Contact, error) {
	c := &models.Contact{}
	err := r.Scan(&c.ID, &c.Name, &c.Email, &c.Phone, &c.Title, &c.InstitutionID, &c.GroupID,
		&c.Relationship, &c.DCGGFlag, &c.Verticals, &c.Tags, &c.Notes, &c.Metadata,
		&c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

type ContactsFilter struct {
	Vertical string
	Tag      string
	DCGGOnly bool
	Search   string
	Limit    int
}

func (r *StakeholdersRepo) ListContacts(ctx context.Context, f ContactsFilter) ([]*models.Contact, error) {
	q := `SELECT ` + contactCols + ` FROM contacts WHERE 1=1`
	args := []any{}
	if f.Vertical != "" {
		args = append(args, f.Vertical)
		q += " AND $" + itoa(len(args)) + " = ANY(verticals)"
	}
	if f.Tag != "" {
		args = append(args, f.Tag)
		q += " AND $" + itoa(len(args)) + " = ANY(tags)"
	}
	if f.DCGGOnly {
		q += " AND dcgg_flag = TRUE"
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
	rows, err := r.DB.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Contact
	for rows.Next() {
		c, err := scanContact(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *StakeholdersRepo) GetContact(ctx context.Context, id string) (*models.Contact, error) {
	row := r.DB.QueryRow(ctx, `SELECT `+contactCols+` FROM contacts WHERE id=$1`, id)
	c, err := scanContact(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	return c, err
}

func (r *StakeholdersRepo) CreateContact(ctx context.Context, c *models.Contact) error {
	return r.DB.QueryRow(ctx, `
		INSERT INTO contacts (name, email, phone, title, institution_id, group_id, relationship, dcgg_flag, verticals, tags, notes, metadata)
		VALUES ($1,$2,$3,$4,$5,$6,$7,COALESCE($8,false),$9,$10,$11,COALESCE($12,'{}'::jsonb))
		RETURNING id, created_at, updated_at`,
		c.Name, c.Email, c.Phone, c.Title, c.InstitutionID, c.GroupID, c.Relationship,
		c.DCGGFlag, c.Verticals, c.Tags, c.Notes, c.Metadata,
	).Scan(&c.ID, &c.CreatedAt, &c.UpdatedAt)
}

func (r *StakeholdersRepo) DeleteContact(ctx context.Context, id string) error {
	ct, err := r.DB.Exec(ctx, `DELETE FROM contacts WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// TreeInstitution is the shape consumed by the Stakeholders view.
type TreeInstitution struct {
	ID       string                  `json:"id"`
	Name     string                  `json:"name"`
	Type     string                  `json:"type"`
	Contacts []TreeContact           `json:"contacts"`
}
type TreeContact struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Title string `json:"title,omitempty"`
}
type TreeCountry struct {
	CountryCode  string            `json:"countryCode"`
	Institutions []TreeInstitution `json:"institutions"`
}
type TreeRegion struct {
	Region    string        `json:"region"`
	Countries []TreeCountry `json:"countries"`
}

// Tree assembles a region → country → institution → contact hierarchy in one
// pass so the Stakeholders view can render its accordion without per-node
// fetches.
func (r *StakeholdersRepo) Tree(ctx context.Context) ([]TreeRegion, error) {
	rows, err := r.DB.Query(ctx, `
		SELECT COALESCE(reg.name, 'Unassigned') AS region_name,
		       COALESCE(co.code, '?')          AS country_code,
		       i.id, i.name, COALESCE(i.kind,'') AS kind,
		       COALESCE(json_agg(json_build_object(
		           'id', ct.id, 'name', ct.name, 'title', COALESCE(ct.title,'')
		       )) FILTER (WHERE ct.id IS NOT NULL), '[]'::json) AS contacts
		FROM institutions i
		LEFT JOIN countries co ON co.id = i.country_id
		LEFT JOIN regions   reg ON reg.id = co.region_id
		LEFT JOIN contacts  ct  ON ct.institution_id = i.id
		GROUP BY reg.name, co.code, i.id, i.name, i.kind
		ORDER BY region_name, country_code, i.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type key struct{ r, c string }
	regionIdx  := map[string]int{}
	countryIdx := map[key]int{}
	out := []TreeRegion{}

	for rows.Next() {
		var (
			regionName, countryCode, instID, instName, kind string
			contactsJSON                                     []byte
		)
		if err := rows.Scan(&regionName, &countryCode, &instID, &instName, &kind, &contactsJSON); err != nil {
			return nil, err
		}
		var contacts []TreeContact
		if err := jsonUnmarshal(contactsJSON, &contacts); err != nil {
			return nil, err
		}

		ri, ok := regionIdx[regionName]
		if !ok {
			out = append(out, TreeRegion{Region: regionName})
			ri = len(out) - 1
			regionIdx[regionName] = ri
		}
		ck := key{regionName, countryCode}
		ci, ok := countryIdx[ck]
		if !ok {
			out[ri].Countries = append(out[ri].Countries, TreeCountry{CountryCode: countryCode})
			ci = len(out[ri].Countries) - 1
			countryIdx[ck] = ci
		}
		out[ri].Countries[ci].Institutions = append(out[ri].Countries[ci].Institutions, TreeInstitution{
			ID: instID, Name: instName, Type: kind, Contacts: contacts,
		})
	}
	return out, rows.Err()
}

func (r *StakeholdersRepo) ListInstitutions(ctx context.Context) ([]*models.Institution, error) {
	rows, err := r.DB.Query(ctx, `SELECT id, name, kind, country_id, website, tags, metadata, created_at, updated_at FROM institutions ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []*models.Institution
	for rows.Next() {
		i := &models.Institution{}
		if err := rows.Scan(&i.ID, &i.Name, &i.Kind, &i.CountryID, &i.Website, &i.Tags, &i.Metadata, &i.CreatedAt, &i.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, rows.Err()
}
