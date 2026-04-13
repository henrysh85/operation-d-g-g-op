package models

import "time"

type Client struct {
	ID          string         `json:"id" db:"id"`
	Slug        string         `json:"slug" db:"slug"`
	Name        string         `json:"name" db:"name"`
	Kind        *string        `json:"kind,omitempty" db:"kind"`
	Tier        *string        `json:"tier,omitempty" db:"tier"`
	Status      string         `json:"status" db:"status"`
	Metrics     map[string]any `json:"metrics" db:"metrics"`
	CrossLinks  []any          `json:"cross_links" db:"cross_links"`
	Verticals   []string       `json:"verticals" db:"verticals"`
	Tags        []string       `json:"tags" db:"tags"`
	OnboardedAt *time.Time     `json:"onboarded_at,omitempty" db:"onboarded_at"`
	Metadata    map[string]any `json:"metadata" db:"metadata"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}
