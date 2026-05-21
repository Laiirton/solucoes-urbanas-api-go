package models

import (
	"encoding/json"
	"time"
)

type ServiceRequest struct {
	ID              int64           `json:"id"`
	UserID          *int64          `json:"user_id,omitempty"`
	UserName        string          `json:"user_name,omitempty"`
	ServiceID       *int64          `json:"service_id,omitempty"`
	ProtocolNumber  *string         `json:"protocol_number,omitempty"`
	ServiceTitle    string          `json:"service_title"`
	Category        string          `json:"category"`
	Icon            string          `json:"icon,omitempty"`
	RequestData     json.RawMessage `json:"request_data"`
	Attachments     json.RawMessage `json:"attachments,omitempty"`
	Status          string          `json:"status"`
	TeamID          *int64          `json:"team_id,omitempty"`
	TeamName        string          `json:"team_name,omitempty"`
	RegionID        *int64          `json:"region_id,omitempty"`
	RegionName      string          `json:"region_name,omitempty"`
	Latitude        *float64        `json:"latitude,omitempty"`
	Longitude       *float64        `json:"longitude,omitempty"`
	GeocodedAddress *string         `json:"geocoded_address,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type ServiceRequestDetailResponse struct {
	*ServiceRequest
	CreatedBy    *User                `json:"created_by,omitempty"`
	Rating       *ServiceRating       `json:"rating,omitempty"`
	Attendances  []*ServiceAttendance `json:"attendances,omitempty"`
	ChatMessages []*ChatMessage       `json:"chat_messages,omitempty"`
	UserRequests int                  `json:"user_requests_count"`
}

// CreateServiceRequestRequest — category is populated automatically from services.category
type CreateServiceRequestRequest struct {
	ServiceID    *int64          `json:"service_id,omitempty"`
	ServiceTitle string          `json:"service_title"`
	RequestData  json.RawMessage `json:"request_data"`
	Attachments  json.RawMessage `json:"attachments,omitempty"`
}

type UpdateServiceRequestStatusRequest struct {
	Status string `json:"status"`
}
