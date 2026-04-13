package models

import "time"

type Institution struct {
	ID        string         `json:"id" db:"id"`
	Name      string         `json:"name" db:"name"`
	Kind      *string        `json:"kind,omitempty" db:"kind"`
	CountryID *string        `json:"country_id,omitempty" db:"country_id"`
	Website   *string        `json:"website,omitempty" db:"website"`
	Tags      []string       `json:"tags" db:"tags"`
	Metadata  map[string]any `json:"metadata" db:"metadata"`
	CreatedAt time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt time.Time      `json:"updated_at" db:"updated_at"`
}

type ContactGroup struct {
	ID          string    `json:"id" db:"id"`
	Name        string    `json:"name" db:"name"`
	Description *string   `json:"description,omitempty" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

type Contact struct {
	ID            string         `json:"id" db:"id"`
	Name          string         `json:"name" db:"name"`
	Email         *string        `json:"email,omitempty" db:"email"`
	Phone         *string        `json:"phone,omitempty" db:"phone"`
	Title         *string        `json:"title,omitempty" db:"title"`
	InstitutionID *string        `json:"institution_id,omitempty" db:"institution_id"`
	GroupID       *string        `json:"group_id,omitempty" db:"group_id"`
	Relationship  *int           `json:"relationship,omitempty" db:"relationship"`
	DCGGFlag      bool           `json:"dcgg_flag" db:"dcgg_flag"`
	Verticals     []string       `json:"verticals" db:"verticals"`
	Tags          []string       `json:"tags" db:"tags"`
	Notes         *string        `json:"notes,omitempty" db:"notes"`
	Metadata      map[string]any `json:"metadata" db:"metadata"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
}
