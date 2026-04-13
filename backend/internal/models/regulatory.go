package models

import "time"

type Region struct {
	ID   string `json:"id" db:"id"`
	Code string `json:"code" db:"code"`
	Name string `json:"name" db:"name"`
}

type Country struct {
	ID       string         `json:"id" db:"id"`
	Code     string         `json:"code" db:"code"`
	Name     string         `json:"name" db:"name"`
	RegionID *string        `json:"region_id,omitempty" db:"region_id"`
	Tier     *int           `json:"tier,omitempty" db:"tier"`
	MenaSub  *string        `json:"mena_sub,omitempty" db:"mena_sub"`
	Metadata map[string]any `json:"metadata" db:"metadata"`
}

type JurisdictionStatus struct {
	ID            string         `json:"id" db:"id"`
	CountryID     string         `json:"country_id" db:"country_id"`
	Vertical      string         `json:"vertical" db:"vertical"`
	Status        string         `json:"status" db:"status"`
	Headline      *string        `json:"headline,omitempty" db:"headline"`
	Regulators    []string       `json:"regulators" db:"regulators"`
	Timeline      []any          `json:"timeline" db:"timeline"`
	ImpactMatrix  map[string]any `json:"impact_matrix" db:"impact_matrix"`
	LastReviewed  *time.Time     `json:"last_reviewed,omitempty" db:"last_reviewed"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}
