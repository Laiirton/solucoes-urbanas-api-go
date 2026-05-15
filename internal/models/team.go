package models

import "time"

type Team struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	RegionID    *int64    `json:"region_id,omitempty"`
	RegionName  string    `json:"region_name,omitempty"`
	Description *string   `json:"description,omitempty"`
	WorkAreas   []string  `json:"work_areas,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CreateTeamRequest struct {
	Name        string  `json:"name"`
	RegionID    int64   `json:"region_id"`
	Description *string `json:"description,omitempty"`
}

type UpdateTeamRequest struct {
	Name        *string `json:"name,omitempty"`
	RegionID    *int64  `json:"region_id,omitempty"`
	Description *string `json:"description,omitempty"`
}

type TeamStats struct {
	Team               Team              `json:"team"`
	MemberCount        int               `json:"member_count"`
	TotalRequests      int               `json:"total_requests"`
	PendingRequests    int               `json:"pending_requests"`
	InProgressRequests int               `json:"in_progress_requests"`
	CompletedRequests  int               `json:"completed_requests"`
	CancelledRequests  int               `json:"cancelled_requests"`
	AvgResolutionDays  float64           `json:"avg_resolution_days"`
	CompletionRate     float64           `json:"completion_rate"`
	RecentRequests     []*ServiceRequest `json:"recent_requests,omitempty"`
}

type TeamMember struct {
	ID              int64     `json:"id"`
	Username        string    `json:"username"`
	Email           string    `json:"email"`
	FullName        *string   `json:"full_name,omitempty"`
	Type            *string   `json:"type,omitempty"`
	ProfileImageURL *string   `json:"profile_image_url,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type MyTeamResponse struct {
	Team       Team    `json:"team"`
	Secretary  *User   `json:"secretary,omitempty"`
	Attendants []*User `json:"attendants"`
}
