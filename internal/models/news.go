package models

import (
	"encoding/json"
	"time"
)

type News struct {
	ID          int64           `json:"id" db:"id"`
	Title       string          `json:"title" db:"title"`
	Slug        string          `json:"slug" db:"slug"`
	Summary     string          `json:"summary" db:"summary"`
	Content     json.RawMessage `json:"content" db:"content"`
	ImageURLs   []string        `json:"image_urls,omitempty" db:"image_urls"`
	Status      string          `json:"status" db:"status"` // e.g., "draft", "published"
	Category    string          `json:"category,omitempty" db:"category"`
	Tags        []string        `json:"tags,omitempty" db:"tags"`
	AuthorID    *int64          `json:"author_id,omitempty" db:"author_id"`
	PublishedAt *time.Time      `json:"published_at,omitempty" db:"published_at"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}

