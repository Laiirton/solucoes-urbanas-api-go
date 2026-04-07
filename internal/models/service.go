package models

import (
	"encoding/json"
	"time"
)

type Service struct {
	ID          int64           `json:"id"`
	Title       string          `json:"title"`
	Description *string         `json:"description,omitempty"`
	Category    string          `json:"category"`
	FormSchema  json.RawMessage `json:"form_schema"`
	IsActive    bool            `json:"is_active"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type CreateServiceRequest struct {
	Title       string          `json:"title"`
	Description *string         `json:"description,omitempty"`
	Category    string          `json:"category"`
	FormSchema  json.RawMessage `json:"form_schema"`
	IsActive    *bool           `json:"is_active,omitempty"`
}

type UpdateServiceRequest struct {
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	Category    *string         `json:"category,omitempty"`
	FormSchema  json.RawMessage `json:"form_schema,omitempty"`
	IsActive    *bool           `json:"is_active,omitempty"`
}
