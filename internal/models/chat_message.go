package models

import (
	"encoding/json"
	"time"
)

type ChatMessage struct {
	ID               int64           `json:"id"`
	ServiceRequestID int64           `json:"service_request_id"`
	SenderID         *int64          `json:"sender_id,omitempty"`
	SenderName       string          `json:"sender_name"`
	Content          string          `json:"content"`
	Attachments      json.RawMessage `json:"attachments,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type CreateChatMessageRequest struct {
	ServiceRequestID int64           `json:"service_request_id"`
	Content          string          `json:"content"`
	Attachments      json.RawMessage `json:"attachments,omitempty"`
}
