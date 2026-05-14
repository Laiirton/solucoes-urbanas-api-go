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

type CategoryDashboardResponse struct {
	Category           CategoryInfo               `json:"category"`
	KPIs               CategoryKPIs               `json:"kpis"`
	StatusDistribution CategoryStatusDistribution `json:"status_distribution"`
	Services           []CategoryServiceDetail    `json:"services"`
	Teams              []CategoryTeamDetail       `json:"teams"`
	RecentRequests     []CategoryRecentRequest    `json:"recent_requests"`
}

type CategoryInfo struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type CategoryKPIs struct {
	TotalServices         int     `json:"total_services"`
	TotalTeams            int     `json:"total_teams"`
	TotalRequests         int     `json:"total_requests"`
	AverageRating         float64 `json:"average_rating"`
	AverageResolutionDays float64 `json:"average_resolution_days"`
}

type CategoryStatusDistribution struct {
	Pending    int `json:"pending"`
	InProgress int `json:"in_progress"`
	Completed  int `json:"completed"`
	Cancelled  int `json:"cancelled"`
}

type CategoryServiceDetail struct {
	ID                int64   `json:"id"`
	Title             string  `json:"title"`
	IsActive          bool    `json:"is_active"`
	TotalRequests     int     `json:"total_requests"`
	CompletedRequests int     `json:"completed_requests"`
	AverageRating     float64 `json:"average_rating"`
}

type CategoryTeamDetail struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

type CategoryRecentRequest struct {
	ID           int64     `json:"id"`
	ServiceTitle string    `json:"service_title"`
	Status       string    `json:"status"`
	Address      *string   `json:"address,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}
