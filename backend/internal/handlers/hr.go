package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// HRHandler serves HR endpoints under /api/v1/hr — gated by RequireRole(admin,hr)
// + RequireHRGate so the caller has both the role and a verified HR PIN session.
type HRHandler struct{ DB *pgxpool.Pool }

func NewHRHandler(db *pgxpool.Pool) *HRHandler { return &HRHandler{DB: db} }

type holidayRow struct {
	ID         string    `json:"id"`
	PersonID   string    `json:"personId"`
	PersonName string    `json:"personName"`
	StartDate  time.Time `json:"startDate"`
	EndDate    time.Time `json:"endDate"`
	Days       float64   `json:"days"`
	Status     string    `json:"status"`
	Note       *string   `json:"note,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (h *HRHandler) ListHolidays(c *gin.Context) {
	q := `SELECT h.id, h.person_id, p.name, h.start_date, h.end_date, h.days,
	             h.status, h.note, h.created_at
	      FROM holidays h JOIN people p ON p.id = h.person_id
	      WHERE 1=1`
	args := []any{}
	if pid := c.Query("person_id"); pid != "" {
		args = append(args, pid)
		q += " AND h.person_id = $" + itoaH(len(args))
	}
	if s := c.Query("status"); s != "" {
		args = append(args, s)
		q += " AND h.status = $" + itoaH(len(args))
	}
	q += " ORDER BY h.start_date DESC LIMIT 500"
	rows, err := h.DB.Query(c.Request.Context(), q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []*holidayRow{}
	for rows.Next() {
		r := &holidayRow{}
		if err := rows.Scan(&r.ID, &r.PersonID, &r.PersonName, &r.StartDate, &r.EndDate, &r.Days, &r.Status, &r.Note, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, r)
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

type holidayInput struct {
	PersonID  string  `json:"personId"  binding:"required"`
	StartDate string  `json:"startDate" binding:"required"`
	EndDate   string  `json:"endDate"   binding:"required"`
	Days      float64 `json:"days"`
	Note      *string `json:"note"`
}

func (h *HRHandler) CreateHoliday(c *gin.Context) {
	var in holidayInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s, errA := time.Parse("2006-01-02", in.StartDate)
	e, errB := time.Parse("2006-01-02", in.EndDate)
	if errA != nil || errB != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate and endDate must be YYYY-MM-DD"})
		return
	}
	if e.Before(s) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endDate must be on or after startDate"})
		return
	}
	if in.Days <= 0 {
		in.Days = e.Sub(s).Hours()/24 + 1
	}
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()
	var id string
	err := h.DB.QueryRow(ctx, `
		INSERT INTO holidays (person_id, start_date, end_date, days, note)
		VALUES ($1, $2::date, $3::date, $4, $5)
		RETURNING id`,
		in.PersonID, in.StartDate, in.EndDate, in.Days, in.Note,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

type holidayDecision struct {
	Status string `json:"status" binding:"required,oneof=approved rejected pending"`
}

type bulkHolidayDecision struct {
	IDs    []string `json:"ids" binding:"required"`
	Status string   `json:"status" binding:"required,oneof=approved rejected pending"`
}

// BulkPatchHolidays approves/rejects multiple holiday requests in one call.
func (h *HRHandler) BulkPatchHolidays(c *gin.Context) {
	var in bulkHolidayDecision
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(in.IDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ids required"})
		return
	}
	ct, err := h.DB.Exec(c.Request.Context(),
		`UPDATE holidays SET status=$1 WHERE id = ANY($2::uuid[])`, in.Status, in.IDs)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"updated": ct.RowsAffected()})
}

func (h *HRHandler) PatchHoliday(c *gin.Context) {
	var in holidayDecision
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ct, err := h.DB.Exec(c.Request.Context(),
		`UPDATE holidays SET status=$1 WHERE id=$2`, in.Status, c.Param("id"))
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

func (h *HRHandler) DeleteHoliday(c *gin.Context) {
	ct, err := h.DB.Exec(c.Request.Context(), `DELETE FROM holidays WHERE id=$1`, c.Param("id"))
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

// HolidayBalance returns days remaining (quota − approved-days-taken-this-year)
// per person for the requested calendar year (default = current).
type holidayBalanceRow struct {
	PersonID   string  `json:"personId"`
	PersonName string  `json:"personName"`
	Quota      int     `json:"quota"`
	Taken      float64 `json:"taken"`
	Remaining  float64 `json:"remaining"`
}

func (h *HRHandler) HolidayBalances(c *gin.Context) {
	year := time.Now().Year()
	if y := c.Query("year"); y != "" {
		var parsed int
		if _, err := fmtSscan(y, &parsed); err == nil && parsed > 1900 {
			year = parsed
		}
	}
	rows, err := h.DB.Query(c.Request.Context(), `
		SELECT p.id, p.name, p.holiday_quota,
		       COALESCE(SUM(h.days) FILTER (WHERE h.status='approved'
		           AND EXTRACT(year FROM h.start_date) = $1), 0)::float AS taken
		FROM people p
		LEFT JOIN holidays h ON h.person_id = p.id
		GROUP BY p.id, p.name, p.holiday_quota
		ORDER BY p.name`, year)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []holidayBalanceRow{}
	for rows.Next() {
		r := holidayBalanceRow{}
		if err := rows.Scan(&r.PersonID, &r.PersonName, &r.Quota, &r.Taken); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		r.Remaining = float64(r.Quota) - r.Taken
		out = append(out, r)
	}
	c.JSON(http.StatusOK, gin.H{"year": year, "data": out})
}

// ---------------- Reviews ----------------

type reviewRow struct {
	ID         string    `json:"id"`
	PersonID   string    `json:"personId"`
	PersonName string    `json:"personName"`
	ReviewerID *string   `json:"reviewerId,omitempty"`
	Period     string    `json:"period"`
	Rating     *int      `json:"rating,omitempty"`
	Summary    *string   `json:"summary,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (h *HRHandler) ListReviews(c *gin.Context) {
	q := `SELECT r.id, r.person_id, p.name, r.reviewer_id, r.period, r.rating, r.summary, r.created_at
	      FROM reviews r JOIN people p ON p.id = r.person_id
	      WHERE 1=1`
	args := []any{}
	if pid := c.Query("person_id"); pid != "" {
		args = append(args, pid)
		q += " AND r.person_id = $" + itoaH(len(args))
	}
	q += " ORDER BY r.created_at DESC LIMIT 500"
	rows, err := h.DB.Query(c.Request.Context(), q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []*reviewRow{}
	for rows.Next() {
		r := &reviewRow{}
		if err := rows.Scan(&r.ID, &r.PersonID, &r.PersonName, &r.ReviewerID, &r.Period, &r.Rating, &r.Summary, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, r)
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

type reviewInput struct {
	PersonID   string  `json:"personId"  binding:"required"`
	ReviewerID *string `json:"reviewerId"`
	Period     string  `json:"period"    binding:"required"`
	Rating     *int    `json:"rating"`
	Summary    *string `json:"summary"`
}

func (h *HRHandler) CreateReview(c *gin.Context) {
	var in reviewInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var id string
	err := h.DB.QueryRow(c.Request.Context(), `
		INSERT INTO reviews (person_id, reviewer_id, period, rating, summary)
		VALUES ($1,$2,$3,$4,$5) RETURNING id`,
		in.PersonID, in.ReviewerID, in.Period, in.Rating, in.Summary,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

// ---------------- Expenses ----------------

type expenseRow struct {
	ID         string    `json:"id"`
	PersonID   string    `json:"personId"`
	PersonName string    `json:"personName"`
	Amount     float64   `json:"amount"`
	Currency   string    `json:"currency"`
	Category   *string   `json:"category,omitempty"`
	IncurredOn time.Time `json:"incurredOn"`
	Memo       *string   `json:"memo,omitempty"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"createdAt"`
}

func (h *HRHandler) ListExpenses(c *gin.Context) {
	q := `SELECT e.id, e.person_id, p.name, e.amount, e.currency, e.category,
	             e.incurred_on, e.memo, e.status, e.created_at
	      FROM expenses e JOIN people p ON p.id = e.person_id
	      WHERE 1=1`
	args := []any{}
	if pid := c.Query("person_id"); pid != "" {
		args = append(args, pid)
		q += " AND e.person_id = $" + itoaH(len(args))
	}
	if s := c.Query("status"); s != "" {
		args = append(args, s)
		q += " AND e.status = $" + itoaH(len(args))
	}
	q += " ORDER BY e.incurred_on DESC LIMIT 500"
	rows, err := h.DB.Query(c.Request.Context(), q, args...)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()
	out := []*expenseRow{}
	for rows.Next() {
		r := &expenseRow{}
		if err := rows.Scan(&r.ID, &r.PersonID, &r.PersonName, &r.Amount, &r.Currency, &r.Category, &r.IncurredOn, &r.Memo, &r.Status, &r.CreatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		out = append(out, r)
	}
	c.JSON(http.StatusOK, gin.H{"data": out})
}

type expenseInput struct {
	PersonID   string  `json:"personId"   binding:"required"`
	Amount     float64 `json:"amount"     binding:"required"`
	Currency   string  `json:"currency"`
	Category   *string `json:"category"`
	IncurredOn string  `json:"incurredOn" binding:"required"`
	Memo       *string `json:"memo"`
}

func (h *HRHandler) CreateExpense(c *gin.Context) {
	var in expenseInput
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if in.Currency == "" {
		in.Currency = "USD"
	}
	var id string
	err := h.DB.QueryRow(c.Request.Context(), `
		INSERT INTO expenses (person_id, amount, currency, category, incurred_on, memo)
		VALUES ($1,$2,$3,$4,$5::date,$6) RETURNING id`,
		in.PersonID, in.Amount, in.Currency, in.Category, in.IncurredOn, in.Memo,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"id": id})
}

type expenseDecision struct {
	Status string `json:"status" binding:"required,oneof=submitted approved rejected paid"`
}

func (h *HRHandler) PatchExpense(c *gin.Context) {
	var in expenseDecision
	if err := c.ShouldBindJSON(&in); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	ct, err := h.DB.Exec(c.Request.Context(),
		`UPDATE expenses SET status=$1 WHERE id=$2`, in.Status, c.Param("id"))
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

// minimal Sscan wrapper to avoid pulling fmt across the file's surface area.
func fmtSscan(s string, v *int) (int, error) {
	var n int
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return 0, pgx.ErrNoRows
		}
		n = n*10 + int(c-'0')
	}
	*v = n
	return 1, nil
}
