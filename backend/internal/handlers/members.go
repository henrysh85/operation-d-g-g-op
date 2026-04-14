package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func itoaH(i int) string { return strconv.Itoa(i) }

type MembersHandler struct{ DB *pgxpool.Pool }

func NewMembersHandler(db *pgxpool.Pool) *MembersHandler { return &MembersHandler{DB: db} }

type memberRow struct {
	ID             string     `json:"id"`
	LegalName      string     `json:"legalName"`
	JurisdictionID *string    `json:"jurisdictionId,omitempty"`
	Status         string     `json:"status"`
	Tier           *string    `json:"tier,omitempty"`
	ContactID      *string    `json:"contactId,omitempty"`
	RiskScore      *int       `json:"riskScore,omitempty"`
	JoinedAt       *time.Time `json:"joinedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`
}

const memberCols = `id, legal_name, jurisdiction_id, status, tier, contact_id, risk_score, joined_at, created_at`

func scanMember(r pgx.Row) (*memberRow, error) {
	m := &memberRow{}
	if err := r.Scan(&m.ID, &m.LegalName, &m.JurisdictionID, &m.Status, &m.Tier, &m.ContactID, &m.RiskScore, &m.JoinedAt, &m.CreatedAt); err != nil {
		return nil, err
	}
	return m, nil
}

func (h *MembersHandler) List(c *gin.Context) {
	q := `SELECT ` + memberCols + ` FROM members WHERE 1=1`
	args := []any{}
	if s := c.Query("status"); s != "" {
		args = append(args, s)
		q += " AND status = $" + itoaH(len(args))
	}
	if t := c.Query("tier"); t != "" {
		args = append(args, t)
		q += " AND tier = $" + itoaH(len(args))
	}
	if search := c.Query("q"); search != "" {
		args = append(args, "%"+search+"%")
		q += " AND legal_name ILIKE $" + itoaH(len(args))
	}
	q += " ORDER BY legal_name ASC LIMIT 500"
	rows, err := h.DB.Query(c.Request.Context(), q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []*memberRow{}
	for rows.Next() {
		m, err := scanMember(rows)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, m)
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

func (h *MembersHandler) Get(c *gin.Context) {
	row := h.DB.QueryRow(c.Request.Context(), `SELECT `+memberCols+` FROM members WHERE id=$1`, c.Param("id"))
	m, err := scanMember(row)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

type memberInput struct {
	LegalName      string  `json:"legalName" binding:"required"`
	JurisdictionID *string `json:"jurisdictionId"`
	Status         string  `json:"status"`
	Tier           *string `json:"tier"`
	ContactID      *string `json:"contactId"`
	RiskScore      *int    `json:"riskScore"`
	JoinedAt       *string `json:"joinedAt"`
}

func (h *MembersHandler) Create(c *gin.Context) {
	var in memberInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.Status == "" {
		in.Status = "prospect"
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	row := h.DB.QueryRow(ctx, `
		INSERT INTO members (legal_name, jurisdiction_id, status, tier, contact_id, risk_score, joined_at)
		VALUES ($1,$2,$3,$4,$5,$6,NULLIF($7,'')::date)
		RETURNING `+memberCols,
		in.LegalName, in.JurisdictionID, in.Status, in.Tier, in.ContactID, in.RiskScore,
		strings.TrimSpace(strDeref(in.JoinedAt)))
	m, err := scanMember(row)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, m)
}

func (h *MembersHandler) Patch(c *gin.Context) {
	var in memberInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	row := h.DB.QueryRow(c.Request.Context(), `
		UPDATE members SET
		  legal_name      = COALESCE(NULLIF($1,''), legal_name),
		  jurisdiction_id = COALESCE($2, jurisdiction_id),
		  status          = COALESCE(NULLIF($3,''), status),
		  tier            = COALESCE($4, tier),
		  contact_id      = COALESCE($5, contact_id),
		  risk_score      = COALESCE($6, risk_score),
		  joined_at       = COALESCE(NULLIF($7,'')::date, joined_at),
		  updated_at      = NOW()
		WHERE id = $8
		RETURNING `+memberCols,
		in.LegalName, in.JurisdictionID, in.Status, in.Tier, in.ContactID, in.RiskScore,
		strDeref(in.JoinedAt), c.Param("id"))
	m, err := scanMember(row)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, m)
}

// Intel returns a member's risk breakdown and recent intelligence affecting
// their jurisdiction. All components are computed from real rows so the page
// is honest about what it knows (no fictional risk scores).
type intelOut struct {
	Risks []struct {
		Key   string `json:"key"`
		Label string `json:"label"`
		Value int    `json:"value"`
		Note  string `json:"note,omitempty"`
	} `json:"risks"`
	RecentRegChanges []intelRow `json:"recentRegChanges"`
	OpenConsults     []intelRow `json:"openConsults"`
	Activities       []intelRow `json:"activities"`
}
type intelRow struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	When  string `json:"when,omitempty"`
	Tag   string `json:"tag,omitempty"`
}

func (h *MembersHandler) Intel(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	var (
		jurisdictionID *string
		joinedAt       *time.Time
		tier           *int
	)
	err := h.DB.QueryRow(ctx, `
		SELECT m.jurisdiction_id, m.joined_at, co.tier
		FROM members m
		LEFT JOIN countries co ON co.id = m.jurisdiction_id
		WHERE m.id = $1`, id).Scan(&jurisdictionID, &joinedAt, &tier)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	out := intelOut{}

	jurisdictionRisk := 0
	jurisdictionNote := "No jurisdiction set."
	if tier != nil {
		jurisdictionRisk = (4 - *tier) * 25 // tier 1 → 75, tier 3 → 25
		jurisdictionNote = fmt.Sprintf("Jurisdiction tier %d.", *tier)
	}

	regChanges := 0
	if jurisdictionID != nil {
		_ = h.DB.QueryRow(ctx, `
			SELECT COUNT(*) FROM jurisdictions_status
			WHERE country_id = $1 AND updated_at >= NOW() - INTERVAL '90 days'`,
			*jurisdictionID).Scan(&regChanges)
	}

	openConsults := 0
	if jurisdictionID != nil {
		_ = h.DB.QueryRow(ctx, `
			SELECT COUNT(*) FROM consultations c
			JOIN jurisdictions_status js ON js.id = c.jurisdiction_id
			WHERE js.country_id = $1 AND c.status NOT IN ('closed','rejected','withdrawn')`,
			*jurisdictionID).Scan(&openConsults)
	}

	tenureRisk := 50
	tenureNote := "No join date."
	if joinedAt != nil {
		years := time.Since(*joinedAt).Hours() / (24 * 365)
		tenureRisk = int(60 - years*15)
		if tenureRisk < 0 {
			tenureRisk = 0
		}
		if tenureRisk > 100 {
			tenureRisk = 100
		}
		tenureNote = fmt.Sprintf("%.1f years as a member.", years)
	}

	out.Risks = []struct {
		Key   string `json:"key"`
		Label string `json:"label"`
		Value int    `json:"value"`
		Note  string `json:"note,omitempty"`
	}{
		{"jurisdiction", "Jurisdiction risk", jurisdictionRisk, jurisdictionNote},
		{"regulatory_change", "Regulatory change activity", regChanges * 10, fmt.Sprintf("%d status updates in 90 days.", regChanges)},
		{"open_consults", "Open consultations exposure", openConsults * 15, fmt.Sprintf("%d consultations open.", openConsults)},
		{"tenure", "Tenure risk", tenureRisk, tenureNote},
	}

	if jurisdictionID != nil {
		// Recent regulatory changes
		rows, err := h.DB.Query(ctx, `
			SELECT js.id, js.headline, js.updated_at::text, js.vertical
			FROM jurisdictions_status js
			WHERE js.country_id = $1
			ORDER BY js.updated_at DESC
			LIMIT 8`, *jurisdictionID)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				r := intelRow{}
				if err := rows.Scan(&r.ID, &r.Title, &r.When, &r.Tag); err == nil {
					out.RecentRegChanges = append(out.RecentRegChanges, r)
				}
			}
		}

		// Open consultations
		consRows, err := h.DB.Query(ctx, `
			SELECT c.id, c.title, COALESCE(c.deadline::text,''), c.vertical
			FROM consultations c
			JOIN jurisdictions_status js ON js.id = c.jurisdiction_id
			WHERE js.country_id = $1 AND c.status NOT IN ('closed','rejected','withdrawn')
			ORDER BY c.deadline ASC NULLS LAST
			LIMIT 8`, *jurisdictionID)
		if err == nil {
			defer consRows.Close()
			for consRows.Next() {
				r := intelRow{}
				if err := consRows.Scan(&r.ID, &r.Title, &r.When, &r.Tag); err == nil {
					out.OpenConsults = append(out.OpenConsults, r)
				}
			}
		}

		// Recent activities in jurisdiction (via region match — best effort)
		actRows, err := h.DB.Query(ctx, `
			SELECT a.id, a.title, a.occurred_on::text, COALESCE(a.vertical,'')
			FROM activities a
			WHERE a.country_id = $1
			   OR a.region_id = (SELECT region_id FROM countries WHERE id = $1)
			ORDER BY a.occurred_on DESC
			LIMIT 8`, *jurisdictionID)
		if err == nil {
			defer actRows.Close()
			for actRows.Next() {
				r := intelRow{}
				if err := actRows.Scan(&r.ID, &r.Title, &r.When, &r.Tag); err == nil {
					out.Activities = append(out.Activities, r)
				}
			}
		}
	}

	c.JSON(http.StatusOK, out)
}

func (h *MembersHandler) Delete(c *gin.Context) {
	ct, err := h.DB.Exec(c.Request.Context(), `DELETE FROM members WHERE id=$1`, c.Param("id"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ct.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}
	c.Status(http.StatusNoContent)
}

func strDeref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
