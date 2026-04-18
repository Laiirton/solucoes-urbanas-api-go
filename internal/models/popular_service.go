package models

type PopularService struct {
	ID           int64  `json:"id"`
	Title        string `json:"title"`
	Category     string `json:"category"`
	RequestCount int    `json:"request_count"`
}
