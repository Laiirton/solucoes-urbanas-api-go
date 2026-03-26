package models

import "time"

type News struct {
	ID        int64     `json:"id" db:"id"`
	Title     string    `json:"title" db:"title"`
	Content   string    `json:"content" db:"content"`
	ImageURLs []string  `json:"image_urls,omitempty" db:"image_urls"`
	AuthorID  *int64    `json:"author_id,omitempty" db:"author_id"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
