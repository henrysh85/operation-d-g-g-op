package models

import "time"

type Publication struct {
	ID          string         `json:"id" db:"id"`
	Title       string         `json:"title" db:"title"`
	Authors     []string       `json:"authors" db:"authors"`
	Venue       *string        `json:"venue,omitempty" db:"venue"`
	URL         *string        `json:"url,omitempty" db:"url"`
	Abstract    *string        `json:"abstract,omitempty" db:"abstract"`
	PublishedOn *time.Time     `json:"published_on,omitempty" db:"published_on"`
	Verticals   []string       `json:"verticals" db:"verticals"`
	Tags        []string       `json:"tags" db:"tags"`
	Metadata    map[string]any `json:"metadata" db:"metadata"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}
