package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AuditHandler struct{ DB *pgxpool.Pool }

func NewAuditHandler(db *pgxpool.Pool) *AuditHandler { return &AuditHandler{DB: db} }

type auditRow struct {
	ID         string                 `json:"id"`
	ActorEmail *string                `json:"actorEmail,omitempty"`
	Action     string                 `json:"action"`
	Entity     string                 `json:"entity"`
	EntityID   *string                `json:"entityId,omitempty"`
	Metadata   map[string]any         `json:"metadata"`
	CreatedAt  time.Time              `json:"createdAt"`
}

func (h *AuditHandler) List(c *gin.Context) {
	q := `SELECT a.id, u.email, a.action, a.entity, a.entity_id, a.metadata, a.created_at
	      FROM audit_log a
	      LEFT JOIN users u ON u.id = a.actor_id
	      WHERE 1=1`
	args := []any{}
	if e := c.Query("entity"); e != "" {
		args = append(args, e)
		q += " AND a.entity = $" + itoaH(len(args))
	}
	if act := c.Query("actor"); act != "" {
		args = append(args, "%"+act+"%")
		q += " AND u.email ILIKE $" + itoaH(len(args))
	}
	q += " ORDER BY a.created_at DESC LIMIT 200"
	rows, err := h.DB.Query(c.Request.Context(), q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []*auditRow{}
	for rows.Next() {
		r := &auditRow{Metadata: map[string]any{}}
		if err := rows.Scan(&r.ID, &r.ActorEmail, &r.Action, &r.Entity, &r.EntityID, &r.Metadata, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, r)
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}
