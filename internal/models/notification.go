package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type RegisterPushTokenRequest struct {
	Token string `json:"token"`
}

func (r *RegisterPushTokenRequest) Validate() error {
	r.Token = strings.TrimSpace(r.Token)
	if r.Token == "" {
		return fmt.Errorf("token is required")
	}

	if !strings.HasPrefix(r.Token, "ExponentPushToken[") && !strings.HasPrefix(r.Token, "ExpoPushToken[") {
		return fmt.Errorf("invalid push token format")
	}

	return nil
}

type SystemNotification struct {
	ID        int64           `json:"id"`
	UserID    *int64          `json:"user_id,omitempty"`
	Title     string          `json:"title"`
	Body      string          `json:"body"`
	Type      string          `json:"type"`
	Data      json.RawMessage `json:"data,omitempty"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

type CreateSystemNotificationRequest struct {
	UserID *int64          `json:"user_id,omitempty"`
	Title  string          `json:"title"`
	Body   string          `json:"body"`
	Type   string          `json:"type,omitempty"`
	Data   json.RawMessage `json:"data,omitempty"`
}

func (r *CreateSystemNotificationRequest) Validate() error {
	r.Title = strings.TrimSpace(r.Title)
	if r.Title == "" {
		return fmt.Errorf("title is required")
	}

	r.Body = strings.TrimSpace(r.Body)
	if r.Body == "" {
		return fmt.Errorf("body is required")
	}

	if r.Type == "" {
		r.Type = "general"
	}

	return nil
}

type UpdateSystemNotificationRequest struct {
	Title  *string          `json:"title,omitempty"`
	Body   *string          `json:"body,omitempty"`
	Type   *string          `json:"type,omitempty"`
	Data   *json.RawMessage `json:"data,omitempty"`
	ReadAt *time.Time       `json:"read_at,omitempty"`
}
