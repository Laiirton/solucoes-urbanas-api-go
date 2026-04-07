package models

import (
	"encoding/json"
	"time"
)

type ServiceRequest struct {
	ID             int64           `json:"id"`
	UserID         *int64          `json:"user_id,omitempty"`
	ServiceID      int64           `json:"service_id"`
	ProtocolNumber *string         `json:"protocol_number,omitempty"`
	ServiceTitle   string          `json:"service_title"`
	Category       string          `json:"category"`
	RequestData    json.RawMessage `json:"request_data"`
	Attachments    json.RawMessage `json:"attachments,omitempty"`
	Status         string          `json:"status"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

// CreateServiceRequestRequest — category is populated automatically from services.category
type CreateServiceRequestRequest struct {
	ServiceID    int64           `json:"service_id"`
	ServiceTitle string          `json:"service_title"`
	RequestData  json.RawMessage `json:"request_data"`
	Attachments  json.RawMessage `json:"attachments,omitempty"`
}

type UpdateServiceRequestStatusRequest struct {
	Status string `json:"status"`
}
