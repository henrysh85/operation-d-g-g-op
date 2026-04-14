package handlers

import (
	"context"
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
