package models

import "time"

type Person struct {
	ID           string         `json:"id" db:"id"`
	Name         string         `json:"name" db:"name"`
	Email        *string        `json:"email,omitempty" db:"email"`
	Dept         *string        `json:"dept,omitempty" db:"dept"`
	Title        *string        `json:"title,omitempty" db:"title"`
	ReportsTo    *string        `json:"reports_to,omitempty" db:"reports_to"`
	Status       string         `json:"status" db:"status"`
	Location     *string        `json:"location,omitempty" db:"location"`
	CountryCode  *string        `json:"country_code,omitempty" db:"country_code"`
	Salary       *float64       `json:"salary,omitempty" db:"salary"`
	Currency     *string        `json:"currency,omitempty" db:"currency"`
	HolidayQuota int            `json:"holiday_quota" db:"holiday_quota"`
	PhotoKey     *string        `json:"photo_key,omitempty" db:"photo_key"`
	StartDate    *time.Time     `json:"start_date,omitempty" db:"start_date"`
	Metadata     map[string]any `json:"metadata" db:"metadata"`
	CreatedAt    time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at" db:"updated_at"`
}

type Holiday struct {
	ID        string    `json:"id" db:"id"`
	PersonID  string    `json:"person_id" db:"person_id"`
	StartDate time.Time `json:"start_date" db:"start_date"`
	EndDate   time.Time `json:"end_date" db:"end_date"`
	Days      float64   `json:"days" db:"days"`
	Status    string    `json:"status" db:"status"`
	Note      *string   `json:"note,omitempty" db:"note"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Review struct {
	ID         string         `json:"id" db:"id"`
	PersonID   string         `json:"person_id" db:"person_id"`
	ReviewerID *string        `json:"reviewer_id,omitempty" db:"reviewer_id"`
	Period     string         `json:"period" db:"period"`
	Rating     *int           `json:"rating,omitempty" db:"rating"`
	Summary    *string        `json:"summary,omitempty" db:"summary"`
	Details    map[string]any `json:"details" db:"details"`
	CreatedAt  time.Time      `json:"created_at" db:"created_at"`
}

type Expense struct {
	ID         string    `json:"id" db:"id"`
	PersonID   string    `json:"person_id" db:"person_id"`
	Amount     float64   `json:"amount" db:"amount"`
	Currency   string    `json:"currency" db:"currency"`
	Category   *string   `json:"category,omitempty" db:"category"`
	IncurredOn time.Time `json:"incurred_on" db:"incurred_on"`
	Memo       *string   `json:"memo,omitempty" db:"memo"`
	ReceiptKey *string   `json:"receipt_key,omitempty" db:"receipt_key"`
	Status     string    `json:"status" db:"status"`
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
