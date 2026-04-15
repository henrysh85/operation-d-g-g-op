package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/henrysh85/operation-d-g-g-op/backend/internal/auth"
)

type InboxHandler struct{ DB *pgxpool.Pool }

func NewInboxHandler(db *pgxpool.Pool) *InboxHandler { return &InboxHandler{DB: db} }

type inboxItem struct {
	Kind   string    `json:"kind"`   // holiday_pending | expense_pending | consultation_deadline | holiday_decided
	ID     string    `json:"id"`
	Title  string    `json:"title"`
	Detail string    `json:"detail,omitempty"`
	When   time.Time `json:"when"`
	Link   string    `json:"link,omitempty"`
}

// Tasks aggregates the things the signed-in user most likely cares about:
// approvals they can action (if they hold admin/hr/lead), upcoming deadlines
// on consultations assigned to them, and recent decisions on their own
// holidays.
func (h *InboxHandler) Tasks(c *gin.Context) {
	claims, ok := auth.ClaimsFrom(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	hasRole := func(r string) bool {
		for _, x := range claims.Roles {
			if x == r {
				return true
			}
		}
		return false
	}
	canApproveHR := hasRole("admin") || hasRole("hr")
	canApproveExpenses := canApproveHR || hasRole("lead")

	out := []inboxItem{}
	ctx := c.Request.Context()

	if canApproveHR {
		rows, err := h.DB.Query(ctx, `
			SELECT h.id, p.name, h.start_date, h.end_date, h.days, h.created_at
			FROM holidays h JOIN people p ON p.id = h.person_id
			WHERE h.status = 'pending' ORDER BY h.created_at DESC LIMIT 25`)
		if err == nil {
			for rows.Next() {
				var id, name string
				var s, e time.Time
				var days float64
				var when time.Time
				if err := rows.Scan(&id, &name, &s, &e, &days, &when); err == nil {
					out = append(out, inboxItem{
						Kind:   "holiday_pending",
						ID:     id,
						Title:  name + " — holiday request",
						Detail: s.Format("2006-01-02") + " → " + e.Format("2006-01-02") + " · " + fmtDays(days),
						When:   when,
						Link:   "/people",
					})
				}
			}
			rows.Close()
		}
	}

	if canApproveExpenses {
		rows, err := h.DB.Query(ctx, `
			SELECT e.id, p.name, e.amount, e.currency, e.memo, e.created_at
			FROM expenses e JOIN people p ON p.id = e.person_id
			WHERE e.status = 'submitted' ORDER BY e.created_at DESC LIMIT 25`)
		if err == nil {
			for rows.Next() {
				var id, name, currency string
				var memo *string
				var amount float64
				var when time.Time
				if err := rows.Scan(&id, &name, &amount, &currency, &memo, &when); err == nil {
					d := memo
					note := ""
					if d != nil {
						note = " · " + *d
					}
					out = append(out, inboxItem{
						Kind:   "expense_pending",
						ID:     id,
						Title:  name + " — expense " + currency + " " + fmtMoney(amount),
						Detail: "submitted" + note,
						When:   when,
						Link:   "/people",
					})
				}
			}
			rows.Close()
		}
	}

	// Consultations assigned to this user — use the people row with same email
	// because consultations.assignee_id references people, not users.
	rows, err := h.DB.Query(ctx, `
		SELECT c.id, c.title, c.deadline, c.vertical
		FROM consultations c
		JOIN people p ON p.id = c.assignee_id
		WHERE LOWER(p.email) = LOWER($1)
		  AND c.status NOT IN ('closed','rejected','withdrawn')
		  AND c.deadline IS NOT NULL
		  AND c.deadline <= CURRENT_DATE + INTERVAL '60 days'
		ORDER BY c.deadline ASC
		LIMIT 25`, claims.Email)
	if err == nil {
		for rows.Next() {
			var id, title, vertical string
			var deadline time.Time
			if err := rows.Scan(&id, &title, &deadline, &vertical); err == nil {
				out = append(out, inboxItem{
					Kind:   "consultation_deadline",
					ID:     id,
					Title:  title,
					Detail: "deadline " + deadline.Format("2006-01-02") + " · " + vertical,
					When:   deadline,
					Link:   "/consultations",
				})
			}
		}
		rows.Close()
	}

	// Recent decisions on the user's own holidays.
	rows2, err := h.DB.Query(ctx, `
		SELECT h.id, h.start_date, h.end_date, h.status, h.created_at
		FROM holidays h
		JOIN people p ON p.id = h.person_id
		WHERE LOWER(p.email) = LOWER($1)
		  AND h.status IN ('approved','rejected')
		  AND h.created_at >= NOW() - INTERVAL '30 days'
		ORDER BY h.created_at DESC LIMIT 10`, claims.Email)
	if err == nil {
		for rows2.Next() {
			var id, status string
			var s, e, when time.Time
			if err := rows2.Scan(&id, &s, &e, &status, &when); err == nil {
				out = append(out, inboxItem{
					Kind:   "holiday_decided",
					ID:     id,
					Title:  "Your holiday was " + status,
					Detail: s.Format("2006-01-02") + " → " + e.Format("2006-01-02"),
					When:   when,
					Link:   "/people",
				})
			}
		}
		rows2.Close()
	}

	counts := map[string]int{}
	for _, it := range out {
		counts[it.Kind]++
	}
	c.JSON(http.StatusOK, gin.H{"data": out, "counts": counts})
}

func fmtDays(d float64) string {
	// Avoid pulling fmt import for a single call.
	if d == float64(int64(d)) {
		return itoaH(int(d)) + " days"
	}
	// crude
	return itoaH(int(d)) + "+ days"
}
func fmtMoney(amount float64) string {
	cents := int64(amount * 100)
	whole := cents / 100
	frac := cents % 100
	if frac < 0 {
		frac = -frac
	}
	return itoaH(int(whole)) + "." + pad2(int(frac))
}
func pad2(n int) string {
	if n < 10 {
		return "0" + itoaH(n)
	}
	return itoaH(n)
}
