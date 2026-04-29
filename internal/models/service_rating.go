package models

import "time"

type ServiceRating struct {
	ID               int64     `json:"id"`
	ServiceRequestID int64     `json:"service_request_id"`
	ServiceID        int64     `json:"service_id"`
	UserID           int64     `json:"user_id"`
	Stars            int        `json:"stars"`
	Comment          string     `json:"comment,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type CreateServiceRatingRequest struct {
	ServiceRequestID int64  `json:"service_request_id"`
	Stars            int    `json:"stars"`
	Comment          string `json:"comment,omitempty"`
}

type ServiceRatingResponse struct {
	ID               int64     `json:"id"`
	Stars            int        `json:"stars"`
	Comment          string     `json:"comment,omitempty"`
	UserName         string     `json:"user_name,omitempty"`
	UserProfileImage string     `json:"user_profile_image,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

type ServiceRatingStats struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}
