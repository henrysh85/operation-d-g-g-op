package models

import "time"

type Template struct {
	ID        string    `json:"id" db:"id"`
	Slug      string    `json:"slug" db:"slug"`
	Name      string    `json:"name" db:"name"`
	Kind      string    `json:"kind" db:"kind"`
	Body      string    `json:"body" db:"body"`
	Variables []any     `json:"variables" db:"variables"`
	Tags      []string  `json:"tags" db:"tags"`
	CreatedBy *string   `json:"created_by,omitempty" db:"created_by"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

type MembershipGenerated struct {
	ID            string         `json:"id" db:"id"`
	TemplateID    *string        `json:"template_id,omitempty" db:"template_id"`
	ContactID     *string        `json:"contact_id,omitempty" db:"contact_id"`
	InstitutionID *string        `json:"institution_id,omitempty" db:"institution_id"`
	Rendered      string         `json:"rendered" db:"rendered"`
	Params        map[string]any `json:"params" db:"params"`
	MinioKey      *string        `json:"minio_key,omitempty" db:"minio_key"`
	GeneratedBy   *string        `json:"generated_by,omitempty" db:"generated_by"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
}
