package models

import (
	"encoding/json"
	"time"
)

type Service struct {
	ID          int64           `json:"id"`
	Title       string          `json:"title"`
	Description *string         `json:"description,omitempty"`
	Category    string          `json:"category,omitempty"`
	CategoryID  *int64          `json:"category_id,omitempty"`
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
	RatingStats        *ServiceRatingStats             `json:"rating_stats,omitempty"`
	StatusStats        []StatusStat                    `json:"status_stats"`
	RecentRequests     []*ServiceRequestDetailResponse `json:"recent_requests"`
	RecentRatings      []*ServiceRatingResponse        `json:"recent_ratings,omitempty"`
}

type CreateServiceRequest struct {
	Title       string          `json:"title"`
	Description *string         `json:"description,omitempty"`
	Category    string          `json:"category,omitempty"`
	CategoryID  *int64          `json:"category_id,omitempty"`
	FormSchema  json.RawMessage `json:"form_schema"`
	IsActive    *bool           `json:"is_active,omitempty"`
}

type CategoryGroupResponse struct {
	ID       int64                `json:"id"`
	Order    int                  `json:"order"`
	Category string               `json:"category"`
	Name     string               `json:"name"`
	Icon     string               `json:"icon"`
	Link     string               `json:"link"`
	Services []ServiceItemResponse `json:"services"`
}

type ServiceItemResponse struct {
	IDService int64           `json:"id_service"`
	Title     string          `json:"title"`
	NewLink   string          `json:"newLink"`
	Form      json.RawMessage `json:"form"`
	IsActive  bool            `json:"is_active"`
}

type UpdateServiceRequest struct {
	Title       *string         `json:"title,omitempty"`
	Description *string         `json:"description,omitempty"`
	Category    *string         `json:"category,omitempty"`
	CategoryID  *int64          `json:"category_id,omitempty"`
	FormSchema  json.RawMessage `json:"form_schema,omitempty"`
	IsActive    *bool           `json:"is_active,omitempty"`
}
