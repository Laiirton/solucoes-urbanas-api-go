package models

import "time"

type AppBanner struct {
	ID         int64     `json:"id"`
	ImageURL   string    `json:"image_url"`
	Title      *string   `json:"title"`
	LinkURL    *string   `json:"link_url,omitempty"`
	OrderIndex int       `json:"order_index"`
	IsActive   bool      `json:"is_active"`
	CreatedAt  time.Time `json:"created_at"`
}

type AppConfig struct {
	LogoURL            string           `json:"logo_url"`
	FeaturedServices   []ServiceSummary `json:"featured_services"`
	FeaturedCategories []CategorySummary `json:"featured_categories"`
}

type ServiceSummary struct {
	ID       int64  `json:"id"`
	Title    string `json:"title"`
	Icon     string `json:"icon"`
	Category string `json:"category"`
}

type CategorySummary struct {
	Name string `json:"name"`
	Icon string `json:"icon"`
}

type CategoryWithServices struct {
	Name     string           `json:"name"`
	Icon     string           `json:"icon"`
	Services []ServiceSummary `json:"services"`
}

type MobileHomeResponse struct {
	LogoURL            string                 `json:"logo_url"`
	Banners            []AppBanner            `json:"banners"`
	Categories         []CategoryWithServices `json:"categories"`
	FeaturedServiceIDs []int64                `json:"featured_service_ids"`
}

type Section struct {
	Type  string      `json:"type"`
	Title string      `json:"title"`
	Data  interface{} `json:"data"`
}
