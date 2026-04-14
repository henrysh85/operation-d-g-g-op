package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DashboardHandler struct{ DB *pgxpool.Pool }

func NewDashboardHandler(db *pgxpool.Pool) *DashboardHandler { return &DashboardHandler{DB: db} }

// Summary aggregates top-level counts for the landing dashboard.
func (h *DashboardHandler) Summary(c *gin.Context) {
	ctx := c.Request.Context()
	counts := map[string]int{}
	deltas := map[string]int{}
	// Each KPI: value query + previous-period query for the delta.
	queries := []struct{ key, val, prev string }{
		{"people",
			`SELECT COUNT(*) FROM people WHERE status='active'`,
			`SELECT COUNT(*) FROM people WHERE status='active' AND created_at < NOW() - INTERVAL '30 days'`},
		{"activities",
			`SELECT COUNT(*) FROM activities WHERE occurred_on >= NOW() - INTERVAL '90 days'`,
			`SELECT COUNT(*) FROM activities WHERE occurred_on BETWEEN NOW() - INTERVAL '180 days' AND NOW() - INTERVAL '90 days'`},
		{"open_consultations",
			`SELECT COUNT(*) FROM consultations WHERE status NOT IN ('closed','rejected','withdrawn')`,
			`SELECT COUNT(*) FROM consultations WHERE status NOT IN ('closed','rejected','withdrawn') AND created_at < NOW() - INTERVAL '30 days'`},
		{"clients",
			`SELECT COUNT(*) FROM clients WHERE status='active'`,
			`SELECT COUNT(*) FROM clients WHERE status='active' AND created_at < NOW() - INTERVAL '30 days'`},
		{"contacts",
			`SELECT COUNT(*) FROM contacts`,
			`SELECT COUNT(*) FROM contacts WHERE created_at < NOW() - INTERVAL '30 days'`},
		{"jurisdictions",
			`SELECT COUNT(*) FROM jurisdictions_status`,
			`SELECT COUNT(*) FROM jurisdictions_status WHERE created_at < NOW() - INTERVAL '30 days'`},
	}
	for _, q := range queries {
		var v, p int
		if err := h.DB.QueryRow(ctx, q.val).Scan(&v); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "at": q.key})
			return
		}
		_ = h.DB.QueryRow(ctx, q.prev).Scan(&p)
		counts[q.key] = v
		deltas[q.key] = v - p
	}

	rows, err := h.DB.Query(ctx, `SELECT id, title, occurred_on, vertical FROM activities
		WHERE highlight = TRUE ORDER BY occurred_on DESC LIMIT 10`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	highlights := []gin.H{}
	for rows.Next() {
		var id, title, vertical string
		var occurredOn interface{}
		if err := rows.Scan(&id, &title, &occurredOn, &vertical); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		highlights = append(highlights, gin.H{
			"id": id, "title": title, "occurred_on": occurredOn, "vertical": vertical,
		})
	}
	deadlineRows, err := h.DB.Query(ctx, `
		SELECT c.id, c.title, c.deadline,
		       COALESCE(c.metadata->>'regulator', '') AS regulator,
		       c.vertical
		FROM consultations c
		WHERE c.status NOT IN ('closed','rejected','withdrawn')
		  AND c.deadline IS NOT NULL
		  AND c.deadline >= CURRENT_DATE
		ORDER BY c.deadline ASC
		LIMIT 10`)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer deadlineRows.Close()
	deadlines := []gin.H{}
	for deadlineRows.Next() {
		var id, title, regulator, vertical string
		var deadline interface{}
		if err := deadlineRows.Scan(&id, &title, &deadline, &regulator, &vertical); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		deadlines = append(deadlines, gin.H{
			"id": id, "title": title, "deadline": deadline, "regulator": regulator, "vertical": vertical,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"counts": counts, "deltas": deltas,
		"highlights": highlights, "deadlines": deadlines,
	})
}
