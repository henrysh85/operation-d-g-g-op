package models

import "time"

type Activity struct {
	ID          string         `json:"id" db:"id"`
	Title       string         `json:"title" db:"title"`
	Description *string        `json:"description,omitempty" db:"description"`
	Type        string         `json:"type" db:"type"`
	Vertical    *string        `json:"vertical,omitempty" db:"vertical"`
	RegionID    *string        `json:"region_id,omitempty" db:"region_id"`
	CountryID   *string        `json:"country_id,omitempty" db:"country_id"`
	OccurredOn  time.Time      `json:"occurred_on" db:"occurred_on"`
	Impact      *int           `json:"impact,omitempty" db:"impact"`
	OwnerID     *string        `json:"owner_id,omitempty" db:"owner_id"`
	Status      string         `json:"status" db:"status"`
	Highlight   bool           `json:"highlight" db:"highlight"`
	Metadata    map[string]any `json:"metadata" db:"metadata"`
	CreatedAt   time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at" db:"updated_at"`
}

type ActivityOutput struct {
	ID          string    `json:"id" db:"id"`
	ActivityID  string    `json:"activity_id" db:"activity_id"`
	Label       string    `json:"label" db:"label"`
	MinioKey    string    `json:"minio_key" db:"minio_key"`
	ContentType *string   `json:"content_type,omitempty" db:"content_type"`
	SizeBytes   *int64    `json:"size_bytes,omitempty" db:"size_bytes"`
	UploadedBy  *string   `json:"uploaded_by,omitempty" db:"uploaded_by"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
