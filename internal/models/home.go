package models

type StatDetail struct {
	Total   int `json:"total"`
	Percent int `json:"percent"`
}

type HomeStats struct {
	TotalRequests       StatDetail       `json:"total_requests"`
	PendingRequests     StatDetail       `json:"pending_requests"`
	InProgressRequests  StatDetail       `json:"in_progress_requests"`
	CompletedRequests   StatDetail       `json:"completed_requests"`
	CancelledRequests   StatDetail       `json:"cancelled_requests"`
	UrgentRequests      StatDetail       `json:"urgent_requests"`
	UnresolvedRequests  StatDetail       `json:"unresolved_requests"`
	TotalUsers          int              `json:"total_users"`
	TotalActiveServices int              `json:"total_active_services"`
	CompletedToday      int              `json:"completed_today"`
	CreatedToday        int              `json:"created_today"`
	AverageTime         float64          `json:"average_time"` // in days
	PopularServices     []PopularService `json:"popular_services"`
}

type HomeAlert struct {
	Type    string `json:"type"` // e.g., "danger", "warning", "info"
	Message string `json:"message"`
}

type CategoryStat struct {
	Category string `json:"category"`
	Percent  int    `json:"percent"`
	Count    int    `json:"count"`
}

type VolumeStat struct {
	Day   string `json:"day"`
	Count int    `json:"count"`
}

type RecentRequest struct {
	ID      int64   `json:"id"`
	Name    *string `json:"name"`
	Service string  `json:"service"`
	Address *string `json:"address"`
	Status  string  `json:"status"`
	Date    string  `json:"date"`
}

type PopularService struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Category     string `json:"category"`
	RequestCount int    `json:"request_count"`
}

type MapLocation struct {
	ID           int64   `json:"id"`
	Address      string  `json:"address"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	ServiceTitle string  `json:"service_title"`
	Status       string  `json:"status"`
	Found        bool    `json:"found"`
}

type HomeResponse struct {
	Stats           HomeStats        `json:"stats"`
	Categories      []CategoryStat   `json:"categories"`
	RecentRequests  []RecentRequest  `json:"recent_requests"`
	DelayedRequests []RecentRequest  `json:"delayed_requests"`
	NewRequests     []RecentRequest  `json:"new_requests"`
	Volume7d        []VolumeStat     `json:"volume_7d"`
	Alerts          []HomeAlert      `json:"alerts"`
	PopularServices []PopularService `json:"popular_services"`
	MapLocations    []MapLocation    `json:"map_locations,omitempty"`
}
