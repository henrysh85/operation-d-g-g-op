package models

import "time"

type Consultation struct {
	ID             string         `json:"id" db:"id"`
	JurisdictionID string         `json:"jurisdiction_id" db:"jurisdiction_id"`
	Vertical       string         `json:"vertical" db:"vertical"`
	Title          string         `json:"title" db:"title"`
	Deadline       *time.Time     `json:"deadline,omitempty" db:"deadline"`
	Status         string         `json:"status" db:"status"`
	AssigneeID     *string        `json:"assignee_id,omitempty" db:"assignee_id"`
	Summary        *string        `json:"summary,omitempty" db:"summary"`
	Metadata       map[string]any `json:"metadata" db:"metadata"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}
