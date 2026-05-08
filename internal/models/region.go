package models

import (
	"encoding/json"
	"time"
)

type Region struct {
	ID            int64           `json:"id"`
	Name          string          `json:"name"`
	Neighborhoods json.RawMessage `json:"neighborhoods"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type CreateRegionRequest struct {
	Name          string          `json:"name"`
	Neighborhoods json.RawMessage `json:"neighborhoods"`
}

type UpdateRegionRequest struct {
	Name          *string          `json:"name,omitempty"`
	Neighborhoods *json.RawMessage `json:"neighborhoods,omitempty"`
}
