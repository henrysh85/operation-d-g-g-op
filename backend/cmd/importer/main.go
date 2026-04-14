package main

// Importer: loads prototype seed JSON files into Postgres (and photos into MinIO).
//
// Source: prototype/seed/*.json, produced by scripts/extract-seed.js from the v7
// HTML prototype. All upserts are idempotent — re-running only refreshes data.

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/config"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/db"
	"github.com/henrysh85/operation-d-g-g-op/backend/internal/storage"
)

var regionNames = map[string]string{
	"latam": "Latin America",
	"na":    "North America",
	"eu":    "Europe",
	"mena":  "Middle East & North Africa",
	"apac":  "Asia Pacific",
	"africa": "Africa",
}

type regRow struct {
	ID    string  `json:"id"`
	N     string  `json:"n"`
	R     string  `json:"r"`
	Tier  string  `json:"tier"`
	Lat   float64 `json:"lat"`
	Lng   float64 `json:"lng"`
	Flag  string  `json:"flag"`
	Crypto  *vertBlock `json:"crypto"`
	AI      *vertBlock `json:"ai"`
	Privacy *vertBlock `json:"privacy"`
	Market  *vertBlock `json:"market"`
	Consults []consult `json:"consults"`
}
type vertBlock struct {
	S      string              `json:"s"`
	H      string              `json:"h"`
	Reg    []string            `json:"reg"`
	TL     json.RawMessage     `json:"tl"`
	Impact json.RawMessage     `json:"impact"`
}
type consult struct {
	T string `json:"t"`
	R string `json:"r"`
	D string `json:"d"`
	V string `json:"v"`
	S string `json:"s"`
	A string `json:"a"`
}

type personRow struct {
	ID         string  `json:"id"`
	Name       string  `json:"name"`
	Title      string  `json:"title"`
	Dept       string  `json:"dept"`
	ReportsTo  *string `json:"reportsTo"`
	Status     string  `json:"status"`
	Location   string  `json:"location"`
	Joined     string  `json:"joined"`
	Salary     float64 `json:"salary"`
	HolQuota   int     `json:"holQuota"`
	Email      string  `json:"email"`
	Bio        string  `json:"bio"`
}

type activityRow struct {
	ID       string   `json:"id"`
	Date     string   `json:"date"`
	Type     string   `json:"type"`
	Title    string   `json:"title"`
	Desc     string   `json:"desc"`
	Vertical string   `json:"vertical"`
	Region   string   `json:"region"`
	Clients  []string `json:"clients"`
	Owner    string   `json:"owner"`
	Status   string   `json:"status"`
	Impact   string   `json:"impact"`
	Outputs  []string `json:"outputs"`
}

type clientProfile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Category string `json:"category"`
	Tier     string `json:"tier"`
	Color    string `json:"color"`
}

type pubRow struct {
	ID       string   `json:"id"`
	Year     string   `json:"year"`
	Month    string   `json:"month"`
	Title    string   `json:"title"`
	Authors  []string `json:"authors"`
	Outlet   string   `json:"outlet"`
	Type     string   `json:"type"`
	Abstract string   `json:"abstract"`
	URL      string   `json:"url"`
	Region   string   `json:"region"`
	Country  string   `json:"country"`
}

type stakeholderRegion struct {
	Label     string                          `json:"label"`
	Countries map[string]stakeholderCountry   `json:"countries"`
}
type stakeholderCountry struct {
	Label        string                   `json:"label"`
	Institutions []stakeholderInstitution `json:"institutions"`
}
type stakeholderInstitution struct {
	ID     string              `json:"id"`
	Label  string              `json:"label"`
	Type   string              `json:"type"`
	Groups []stakeholderGroup  `json:"groups"`
}
type stakeholderGroup struct {
	Label  string              `json:"label"`
	People []stakeholderPerson `json:"people"`
}
type stakeholderPerson struct {
	N    string   `json:"n"`
	T    string   `json:"t"`
	Rel  int      `json:"rel"`
	DCGG bool     `json:"dcgg"`
	V    []string `json:"v"`
	Tags []string `json:"tags"`
}

func impactRank(s string) int {
	switch strings.ToUpper(s) {
	case "CRITICAL":
		return 4
	case "HIGH":
		return 3
	case "MEDIUM":
		return 2
	case "LOW":
		return 1
	}
	return 0
}

func main() {
	_ = godotenv.Load()
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339})

	seedDir := flag.String("seed", "/prototype/seed", "directory containing seed JSON files")
	flag.Parse()

	if _, err := os.Stat(*seedDir); err != nil {
		// fallback: try repo-local path when run outside container
		for _, alt := range []string{"prototype/seed", "../prototype/seed"} {
			if _, e := os.Stat(alt); e == nil {
				*seedDir = alt
				break
			}
		}
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("load config")
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DBURL)
	if err != nil {
		log.Fatal().Err(err).Msg("connect postgres")
	}
	defer pool.Close()

	mc, err := storage.New(ctx, cfg)
	if err != nil {
		log.Warn().Err(err).Msg("minio connect failed — photos will be skipped")
	}

	log.Info().Str("dir", *seedDir).Msg("seeding from")

	if err := importRegionsCountries(ctx, pool, *seedDir); err != nil {
		log.Fatal().Err(err).Msg("regions/countries")
	}
	if err := importJurisdictions(ctx, pool, *seedDir); err != nil {
		log.Fatal().Err(err).Msg("jurisdictions")
	}
	peopleMap, err := importPeople(ctx, pool, *seedDir)
	if err != nil {
		log.Fatal().Err(err).Msg("people")
	}
	if mc != nil {
		if err := importPhotos(ctx, pool, mc, *seedDir); err != nil {
			log.Warn().Err(err).Msg("photos")
		}
	}
	clientMap, err := importClients(ctx, pool, *seedDir)
	if err != nil {
		log.Fatal().Err(err).Msg("clients")
	}
	if err := importActivities(ctx, pool, *seedDir, peopleMap, clientMap); err != nil {
		log.Fatal().Err(err).Msg("activities")
	}
	if err := importConsultations(ctx, pool, *seedDir, peopleMap); err != nil {
		log.Fatal().Err(err).Msg("consultations")
	}
	if err := importPublications(ctx, pool, *seedDir); err != nil {
		log.Fatal().Err(err).Msg("publications")
	}
	if err := importStakeholders(ctx, pool, *seedDir); err != nil {
		log.Warn().Err(err).Msg("stakeholders (non-fatal)")
	}

	log.Info().Msg("seed complete")
}

func readJSON(dir, name string, out any) error {
	path := filepath.Join(dir, name)
	b, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read %s: %w", path, err)
	}
	return json.Unmarshal(b, out)
}

// ---------- regions + countries ----------

func importRegionsCountries(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	var regs []regRow
	if err := readJSON(dir, "regs.json", &regs); err != nil {
		return err
	}
	seenR := map[string]bool{}
	for _, r := range regs {
		if r.R != "" {
			seenR[r.R] = true
		}
	}
	for code := range seenR {
		name := regionNames[code]
		if name == "" {
			name = strings.ToUpper(code)
		}
		if _, err := pool.Exec(ctx,
			`INSERT INTO regions (code, name) VALUES ($1,$2) ON CONFLICT (code) DO UPDATE SET name=EXCLUDED.name`,
			code, name); err != nil {
			return err
		}
	}
	for _, r := range regs {
		tier := 0
		if strings.HasPrefix(r.Tier, "t") {
			fmt.Sscanf(r.Tier[1:], "%d", &tier)
		}
		meta, _ := json.Marshal(map[string]any{"lat": r.Lat, "lng": r.Lng, "flag": r.Flag})
		if _, err := pool.Exec(ctx, `
			INSERT INTO countries (code, name, region_id, tier, metadata)
			VALUES ($1, $2, (SELECT id FROM regions WHERE code=$3), $4, $5)
			ON CONFLICT (code) DO UPDATE
			   SET name=EXCLUDED.name, region_id=EXCLUDED.region_id,
			       tier=EXCLUDED.tier, metadata=EXCLUDED.metadata`,
			r.ID, r.N, r.R, tier, meta); err != nil {
			return err
		}
	}
	log.Info().Int("regions", len(seenR)).Int("countries", len(regs)).Msg("imported")
	return nil
}

// ---------- jurisdictions_status ----------

func importJurisdictions(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	var regs []regRow
	if err := readJSON(dir, "regs.json", &regs); err != nil {
		return err
	}
	n := 0
	for _, r := range regs {
		verts := map[string]*vertBlock{
			"crypto":  r.Crypto,
			"ai":      r.AI,
			"privacy": r.Privacy,
			"market":  r.Market,
		}
		for v, b := range verts {
			if b == nil {
				continue
			}
			tl := b.TL
			if len(tl) == 0 {
				tl = []byte("[]")
			}
			impact := b.Impact
			if len(impact) == 0 {
				impact = []byte("{}")
			}
			if _, err := pool.Exec(ctx, `
				INSERT INTO jurisdictions_status (country_id, vertical, status, headline, regulators, timeline, impact_matrix, last_reviewed)
				VALUES ((SELECT id FROM countries WHERE code=$1), $2, $3, $4, $5, $6::jsonb, $7::jsonb, NOW())
				ON CONFLICT (country_id, vertical) DO UPDATE
				   SET status=EXCLUDED.status, headline=EXCLUDED.headline,
				       regulators=EXCLUDED.regulators, timeline=EXCLUDED.timeline,
				       impact_matrix=EXCLUDED.impact_matrix, last_reviewed=NOW(), updated_at=NOW()`,
				r.ID, v, b.S, b.H, b.Reg, string(tl), string(impact)); err != nil {
				return err
			}
			n++
		}
	}
	log.Info().Int("jurisdiction_rows", n).Msg("imported")
	return nil
}

// ---------- people ----------

func importPeople(ctx context.Context, pool *pgxpool.Pool, dir string) (map[string]string, error) {
	var people []personRow
	if err := readJSON(dir, "people.json", &people); err != nil {
		return nil, err
	}
	ids := make(map[string]string, len(people))
	for _, p := range people {
		var start *time.Time
		if p.Joined != "" {
			if t, err := time.Parse("2006-01-02", p.Joined); err == nil {
				start = &t
			}
		}
		status := p.Status
		if status == "" {
			status = "active"
		}
		quota := p.HolQuota
		if quota == 0 {
			quota = 25
		}
		meta, _ := json.Marshal(map[string]any{"seedId": p.ID, "bio": p.Bio})
		id, err := upsertBySeedID(ctx, pool, "people", p.ID, `
			INSERT INTO people (name, dept, title, status, location, salary, holiday_quota, start_date, metadata)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
			`UPDATE people SET name=$1, dept=$2, title=$3, status=$4, location=$5,
			                   salary=$6, holiday_quota=$7, start_date=$8,
			                   metadata=metadata || $9, updated_at=NOW()
			 WHERE id=$10`,
			p.Name, p.Dept, p.Title, status, p.Location, p.Salary, quota, start, meta)
		if err != nil {
			return nil, err
		}
		ids[p.ID] = id
	}
	// pass 2: reports_to
	for _, p := range people {
		if p.ReportsTo == nil || *p.ReportsTo == "" {
			continue
		}
		mgrID, ok := ids[*p.ReportsTo]
		if !ok {
			continue
		}
		if _, err := pool.Exec(ctx, `UPDATE people SET reports_to=$1 WHERE id=$2`, mgrID, ids[p.ID]); err != nil {
			return nil, err
		}
	}
	log.Info().Int("people", len(people)).Msg("imported")
	return ids, nil
}

// ---------- photos ----------

func importPhotos(ctx context.Context, pool *pgxpool.Pool, mc *storage.Client, dir string) error {
	var photos map[string]string
	if err := readJSON(dir, "photos.json", &photos); err != nil {
		return err
	}
	n := 0
	for seedID, dataURL := range photos {
		i := strings.Index(dataURL, ",")
		if i < 0 || !strings.HasPrefix(dataURL, "data:") {
			continue
		}
		header := dataURL[5:i]
		contentType := "image/jpeg"
		if semi := strings.Index(header, ";"); semi >= 0 {
			contentType = header[:semi]
		}
		raw, err := base64.StdEncoding.DecodeString(dataURL[i+1:])
		if err != nil {
			log.Warn().Err(err).Str("id", seedID).Msg("photo decode")
			continue
		}
		key := fmt.Sprintf("avatars/%s.jpg", seedID)
		if err := mc.PutObject(ctx, key, bytes.NewReader(raw), int64(len(raw)), contentType); err != nil {
			log.Warn().Err(err).Str("id", seedID).Msg("photo upload")
			continue
		}
		if _, err := pool.Exec(ctx,
			`UPDATE people SET photo_key=$1 WHERE metadata->>'seedId'=$2`, key, seedID); err != nil {
			log.Warn().Err(err).Str("id", seedID).Msg("photo update")
			continue
		}
		n++
	}
	log.Info().Int("photos", n).Msg("uploaded")
	return nil
}

// ---------- clients ----------

func importClients(ctx context.Context, pool *pgxpool.Pool, dir string) (map[string]string, error) {
	var raw map[string]clientProfile
	if err := readJSON(dir, "client_profiles.json", &raw); err != nil {
		return nil, err
	}
	ids := map[string]string{}
	for slug, c := range raw {
		meta, _ := json.Marshal(c)
		var id string
		if err := pool.QueryRow(ctx, `
			INSERT INTO clients (slug, name, kind, tier, metadata)
			VALUES ($1,$2,$3,$4,$5)
			ON CONFLICT (slug) DO UPDATE
			   SET name=EXCLUDED.name, kind=EXCLUDED.kind, tier=EXCLUDED.tier,
			       metadata=EXCLUDED.metadata, updated_at=NOW()
			RETURNING id`,
			slug, c.Name, c.Category, c.Tier, meta).Scan(&id); err != nil {
			return nil, err
		}
		ids[slug] = id
	}
	log.Info().Int("clients", len(ids)).Msg("imported")
	return ids, nil
}

// ---------- activities ----------

func importActivities(ctx context.Context, pool *pgxpool.Pool, dir string, peopleMap, clientMap map[string]string) error {
	var acts []activityRow
	if err := readJSON(dir, "activities.json", &acts); err != nil {
		return err
	}
	for _, a := range acts {
		var occurred *time.Time
		if a.Date != "" {
			if t, err := time.Parse("2006-01-02", a.Date); err == nil {
				occurred = &t
			}
		}
		if occurred == nil {
			continue
		}
		var ownerID *string
		if oid, ok := peopleMap[a.Owner]; ok {
			ownerID = &oid
		}
		status := a.Status
		if status == "" {
			status = "done"
		}
		highlight := strings.EqualFold(a.Impact, "HIGH") || strings.EqualFold(a.Impact, "CRITICAL")
		meta, _ := json.Marshal(map[string]any{"seedId": a.ID, "outputs": a.Outputs})
		id, err := upsertBySeedID(ctx, pool, "activities", a.ID, `
			INSERT INTO activities (title, description, type, vertical, region_id, occurred_on, impact, owner_id, status, highlight, metadata)
			VALUES ($1,$2,$3,$4,(SELECT id FROM regions WHERE code=$5),$6,$7,$8,$9,$10,$11) RETURNING id`,
			`UPDATE activities SET title=$1, description=$2, type=$3, vertical=$4,
			                        region_id=(SELECT id FROM regions WHERE code=$5),
			                        occurred_on=$6, impact=$7, owner_id=$8, status=$9,
			                        highlight=$10, metadata=metadata || $11, updated_at=NOW()
			 WHERE id=$12`,
			a.Title, a.Desc, a.Type, a.Vertical, a.Region, occurred, impactRank(a.Impact), ownerID, status, highlight, meta)
		if err != nil {
			return err
		}
		// link clients
		for _, cslug := range a.Clients {
			cid, ok := clientMap[cslug]
			if !ok {
				continue
			}
			if _, err := pool.Exec(ctx,
				`INSERT INTO activity_clients (activity_id, client_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`,
				id, cid); err != nil {
				return err
			}
		}
	}
	log.Info().Int("activities", len(acts)).Msg("imported")
	return nil
}

// ---------- consultations ----------

func importConsultations(ctx context.Context, pool *pgxpool.Pool, dir string, peopleMap map[string]string) error {
	var regs []regRow
	if err := readJSON(dir, "regs.json", &regs); err != nil {
		return err
	}
	n := 0
	for _, r := range regs {
		for _, c := range r.Consults {
			var jid string
			err := pool.QueryRow(ctx, `
				SELECT js.id FROM jurisdictions_status js
				JOIN countries co ON co.id = js.country_id
				WHERE co.code=$1 AND js.vertical=$2`, r.ID, c.V).Scan(&jid)
			if err == pgx.ErrNoRows {
				continue
			}
			if err != nil {
				return err
			}
			// find first matching assignee in peopleMap by first-name prefix
			var assignee *string
			first := strings.SplitN(strings.TrimSpace(c.A), " ", 2)[0]
			for seedID, pid := range peopleMap {
				if strings.EqualFold(seedID, strings.ToLower(first)) {
					pidCopy := pid
					assignee = &pidCopy
					break
				}
			}
			meta, _ := json.Marshal(map[string]any{"regulator": c.R, "dateLabel": c.D, "rawAssignee": c.A})
			if _, err := pool.Exec(ctx, `
				INSERT INTO consultations (jurisdiction_id, vertical, title, status, assignee_id, metadata)
				VALUES ($1,$2,$3,COALESCE(NULLIF($4,''),'open'),$5,$6)
				ON CONFLICT DO NOTHING`,
				jid, c.V, c.T, c.S, assignee, meta); err != nil {
				return err
			}
			n++
		}
	}
	log.Info().Int("consultations", n).Msg("imported")
	return nil
}

// ---------- publications ----------

func importPublications(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	var pubs []pubRow
	if err := readJSON(dir, "academic_publications.json", &pubs); err != nil {
		return err
	}
	for _, p := range pubs {
		var published *time.Time
		if p.Year != "" {
			layout := "2006"
			s := p.Year
			if p.Month != "" {
				layout = "Jan 2006"
				s = p.Month + " " + p.Year
			}
			if t, err := time.Parse(layout, s); err == nil {
				published = &t
			}
		}
		meta, _ := json.Marshal(map[string]any{"seedId": p.ID, "region": p.Region, "country": p.Country, "type": p.Type})
		if _, err := pool.Exec(ctx, `
			INSERT INTO publications (title, authors, venue, url, abstract, published_on, metadata)
			VALUES ($1,$2,$3,$4,$5,$6,$7)
			ON CONFLICT DO NOTHING`,
			p.Title, p.Authors, p.Outlet, p.URL, p.Abstract, published, meta); err != nil {
			return err
		}
	}
	log.Info().Int("publications", len(pubs)).Msg("imported")
	return nil
}

// ---------- stakeholders ----------

func importStakeholders(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	var data map[string]stakeholderRegion
	if err := readJSON(dir, "stakeholders.json", &data); err != nil {
		return err
	}
	inst, contacts := 0, 0
	for _, reg := range data {
		for cCode, country := range reg.Countries {
			for _, in := range country.Institutions {
				var instID string
				instMeta, _ := json.Marshal(map[string]any{"seedId": in.ID})
				if err := pool.QueryRow(ctx, `
					INSERT INTO institutions (name, kind, country_id, metadata)
					VALUES ($1,$2,(SELECT id FROM countries WHERE code=$3),$4)
					RETURNING id`,
					in.Label, in.Type, cCode, instMeta).Scan(&instID); err != nil {
					return err
				}
				inst++
				for _, g := range in.Groups {
					var gid string
					if err := pool.QueryRow(ctx, `
						INSERT INTO contact_groups (name, description)
						VALUES ($1,$2)
						ON CONFLICT (name) DO UPDATE SET description=EXCLUDED.description
						RETURNING id`,
						g.Label, country.Label).Scan(&gid); err != nil {
						return err
					}
					for _, per := range g.People {
						rel := per.Rel
						if rel < 1 {
							rel = 1
						}
						if rel > 5 {
							rel = 5
						}
						if _, err := pool.Exec(ctx, `
							INSERT INTO contacts (name, title, institution_id, group_id, relationship, dcgg_flag, verticals, tags)
							VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`,
							per.N, per.T, instID, gid, rel, per.DCGG, per.V, per.Tags); err != nil {
							return err
						}
						contacts++
					}
				}
			}
		}
	}
	log.Info().Int("institutions", inst).Int("contacts", contacts).Msg("imported")
	return nil
}

// ---------- helpers ----------

// upsertBySeedID inserts a row or updates it if a row with metadata->>'seedId' already exists.
// insertSQL must end with `RETURNING id`. updateSQL receives the same positional args plus
// the existing row id as a trailing parameter.
func upsertBySeedID(ctx context.Context, pool *pgxpool.Pool, table, seedID, insertSQL, updateSQL string, args ...any) (string, error) {
	var existingID string
	err := pool.QueryRow(ctx,
		fmt.Sprintf(`SELECT id FROM %s WHERE metadata->>'seedId'=$1`, table),
		seedID).Scan(&existingID)
	switch {
	case errors.Is(err, pgx.ErrNoRows):
		var id string
		if err := pool.QueryRow(ctx, insertSQL, args...).Scan(&id); err != nil {
			return "", fmt.Errorf("%s insert: %w", table, err)
		}
		return id, nil
	case err != nil:
		return "", fmt.Errorf("%s lookup: %w", table, err)
	default:
		updateArgs := append(append([]any{}, args...), existingID)
		if _, err := pool.Exec(ctx, updateSQL, updateArgs...); err != nil {
			return "", fmt.Errorf("%s update: %w", table, err)
		}
		return existingID, nil
	}
}
