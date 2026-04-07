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

type StatusStat struct {
	Status string `json:"status"`
	Total  int    `json:"total"`
}

type ServiceDetailResponse struct {
	*Service
	AverageServiceTime int                             `json:"average_service_time"`
	StatusStats        []StatusStat                    `json:"status_stats"`
	RecentRequests     []*ServiceRequestDetailResponse `json:"recent_requests"`
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
