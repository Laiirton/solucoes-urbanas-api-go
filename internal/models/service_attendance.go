package models

import (
	"encoding/json"
	"time"
)

type ServiceAttendance struct {
	ID               int64           `json:"id"`
	ServiceRequestID int64           `json:"service_request_id"`
	AttendedBy       *int64          `json:"attended_by,omitempty"`
	AttendantName    string          `json:"attendant_name,omitempty"`
	Notes            string          `json:"notes"`
	Attachments      json.RawMessage `json:"attachments,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type CreateServiceAttendanceRequest struct {
	ServiceRequestID int64           `json:"service_request_id"`
	Notes            string          `json:"notes"`
	Attachments      json.RawMessage `json:"attachments,omitempty"`
	NewStatus        string          `json:"new_status,omitempty"` // Optional: update status of the request
}
