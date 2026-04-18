package models

import "time"

type Team struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	ServiceCategory string    `json:"service_category"` // The area of service this team takes care of
	Description     *string   `json:"description,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type CreateTeamRequest struct {
	Name            string  `json:"name"`
	ServiceCategory string  `json:"service_category"`
	Description     *string `json:"description,omitempty"`
}

type UpdateTeamRequest struct {
	Name            *string `json:"name,omitempty"`
	ServiceCategory *string `json:"service_category,omitempty"`
	Description     *string `json:"description,omitempty"`
}
