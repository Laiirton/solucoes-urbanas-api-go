package models

import "time"

type Category struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Icon      string    `json:"icon"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateCategoryRequest struct {
	Name     string `json:"name"`
	Icon     string `json:"icon,omitempty"`
	IsActive *bool  `json:"is_active,omitempty"`
}

type UpdateCategoryRequest struct {
	Name     *string `json:"name,omitempty"`
	Icon     *string `json:"icon,omitempty"`
	IsActive *bool   `json:"is_active,omitempty"`
}

type CategoryDetailResponse struct {
	*Category
	Services []*Service `json:"services"`
}
