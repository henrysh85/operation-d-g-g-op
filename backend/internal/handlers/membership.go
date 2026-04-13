package handlers

import (
	"bytes"
	"net/http"
	"text/template"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/repo"
)

type MembershipHandler struct {
	DB    *pgxpool.Pool
	Tmpls *repo.TemplatesRepo
}

func NewMembershipHandler(db *pgxpool.Pool, t *repo.TemplatesRepo) *MembershipHandler {
	return &MembershipHandler{DB: db, Tmpls: t}
}

type generateReq struct {
	TemplateID    string         `json:"template_id" binding:"required"`
	ContactID     *string        `json:"contact_id"`
	InstitutionID *string        `json:"institution_id"`
	Params        map[string]any `json:"params"`
}

// Generate renders a template with text/template and records the result.
func (h *MembershipHandler) Generate(c *gin.Context) {
	var req generateReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	tpl, err := h.Tmpls.Get(c.Request.Context(), req.TemplateID)
	if err != nil {
		HandleErr(c, err)
		return
	}
	t, err := template.New(tpl.Slug).Parse(tpl.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "template parse: " + err.Error()})
		return
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, req.Params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "template exec: " + err.Error()})
		return
	}
	rendered := buf.String()

	var id string
	err = h.DB.QueryRow(c.Request.Context(), `
		INSERT INTO memberships_generated (template_id, contact_id, institution_id, rendered, params)
		VALUES ($1,$2,$3,$4,COALESCE($5,'{}'::jsonb))
		RETURNING id`,
		tpl.ID, req.ContactID, req.InstitutionID, rendered, req.Params,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id, "rendered": rendered})
}
