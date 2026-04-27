package repository

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type AppConfigRepository struct {
	db *pgxpool.Pool
}

func NewAppConfigRepository(db *pgxpool.Pool) *AppConfigRepository {
	return &AppConfigRepository{db: db}
}

func (r *AppConfigRepository) GetBanners(ctx context.Context) ([]models.AppBanner, error) {
	query := `SELECT id, image_url, title, link_url, order_index, is_active, created_at 
	          FROM app_banners WHERE is_active = TRUE ORDER BY order_index ASC`
	
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var banners []models.AppBanner
	for rows.Next() {
		var b models.AppBanner
		if err := rows.Scan(&b.ID, &b.ImageURL, &b.Title, &b.LinkURL, &b.OrderIndex, &b.IsActive, &b.CreatedAt); err != nil {
			return nil, err
		}
		banners = append(banners, b)
	}
	return banners, nil
}

func (r *AppConfigRepository) GetBannerByID(ctx context.Context, id int64) (*models.AppBanner, error) {
	query := `SELECT id, image_url, title, link_url, order_index, is_active, created_at FROM app_banners WHERE id = $1`
	var b models.AppBanner
	err := r.db.QueryRow(ctx, query, id).Scan(&b.ID, &b.ImageURL, &b.Title, &b.LinkURL, &b.OrderIndex, &b.IsActive, &b.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *AppConfigRepository) GetSetting(ctx context.Context, key string, target interface{}) error {
	var val []byte
	err := r.db.QueryRow(ctx, "SELECT value FROM app_settings WHERE key = $1", key).Scan(&val)
	if err != nil {
		return err
	}
	return json.Unmarshal(val, target)
}

func (r *AppConfigRepository) GetFeaturedServices(ctx context.Context) ([]models.ServiceSummary, error) {
	var ids []int64
	if err := r.GetSetting(ctx, "featured_services", &ids); err != nil {
		return nil, err
	}

	if len(ids) == 0 {
		return []models.ServiceSummary{}, nil
	}

	query := `SELECT id, title, category FROM services WHERE id = ANY($1) AND is_active = TRUE`
	rows, err := r.db.Query(ctx, query, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []models.ServiceSummary
	for rows.Next() {
		var s models.ServiceSummary
		if err := rows.Scan(&s.ID, &s.Title, &s.Category); err != nil {
			return nil, err
		}
		s.Icon = models.GetServiceIcon(s.ID)
		services = append(services, s)
	}
	return services, nil
}

func (r *AppConfigRepository) GetFeaturedCategories(ctx context.Context) ([]models.CategorySummary, error) {
	var categories []string
	if err := r.GetSetting(ctx, "featured_categories", &categories); err != nil {
		return nil, err
	}

	var result []models.CategorySummary
	for _, cat := range categories {
		result = append(result, models.CategorySummary{
			Name: cat,
			Icon: models.GetCategoryIcon(cat),
		})
	}
	return result, nil
}

func (r *AppConfigRepository) UpdateSetting(ctx context.Context, key string, value interface{}) error {
	val, err := json.Marshal(value)
	if err != nil {
		return err
	}
	_, err = r.db.Exec(ctx, "INSERT INTO app_settings (key, value) VALUES ($1, $2) ON CONFLICT (key) DO UPDATE SET value = $2, updated_at = NOW()", key, val)
	return err
}

func (r *AppConfigRepository) CreateBanner(ctx context.Context, banner *models.AppBanner) error {
	query := `INSERT INTO app_banners (image_url, title, link_url, order_index, is_active) 
	          VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`
	return r.db.QueryRow(ctx, query, banner.ImageURL, banner.Title, banner.LinkURL, banner.OrderIndex, banner.IsActive).
		Scan(&banner.ID, &banner.CreatedAt)
}

func (r *AppConfigRepository) UpdateBanner(ctx context.Context, banner *models.AppBanner) error {
	query := `UPDATE app_banners SET image_url = $1, title = $2, link_url = $3, order_index = $4, is_active = $5 
	          WHERE id = $6`
	_, err := r.db.Exec(ctx, query, banner.ImageURL, banner.Title, banner.LinkURL, banner.OrderIndex, banner.IsActive, banner.ID)
	return err
}

func (r *AppConfigRepository) DeleteBanner(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, "DELETE FROM app_banners WHERE id = $1", id)
	return err
}
