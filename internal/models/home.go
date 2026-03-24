package models

type StatDetail struct {
	Total   int `json:"total"`
	Percent int `json:"percent"`
}

type HomeStats struct {
	TotalRequests      StatDetail `json:"total_requests"`
	PendingRequests    StatDetail `json:"pending_requests"`
	InProgressRequests StatDetail `json:"in_progress_requests"`
	CompletedRequests  StatDetail `json:"completed_requests"`
	CancelledRequests  StatDetail `json:"cancelled_requests"`
	UrgentRequests     StatDetail `json:"urgent_requests"`
	UnresolvedRequests StatDetail `json:"unresolved_requests"`
}

type CategoryStat struct {
	Category string `json:"category"`
	Percent  int    `json:"percent"`
}

type RecentRequest struct {
	ID      int64   `json:"id"`
	Name    *string `json:"name"`
	Service string  `json:"service"`
	Address *string `json:"address"`
	Status  string  `json:"status"`
	Date    string  `json:"date"`
}

type HomeResponse struct {
	Stats          HomeStats       `json:"stats"`
	Categories     []CategoryStat  `json:"categories"`
	RecentRequests []RecentRequest `json:"recent_requests"`
}
