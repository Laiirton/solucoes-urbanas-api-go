package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type ServiceRatingRepository struct {
	db *pgxpool.Pool
}

func NewServiceRatingRepository(db *pgxpool.Pool) *ServiceRatingRepository {
	return &ServiceRatingRepository{db: db}
}

func (r *ServiceRatingRepository) Create(ctx context.Context, rating *models.ServiceRating) error {
	query := `
		INSERT INTO service_ratings (service_request_id, service_id, user_id, stars, comment, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at`

	err := r.db.QueryRow(ctx, query,
		rating.ServiceRequestID, rating.ServiceID, rating.UserID, rating.Stars, rating.Comment,
	).Scan(&rating.ID, &rating.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create rating: %w", err)
	}
	return nil
}

func (r *ServiceRatingRepository) GetByRequestID(ctx context.Context, requestID int64) (*models.ServiceRating, error) {
	query := `SELECT id, service_request_id, service_id, user_id, stars, comment, created_at
              FROM service_ratings WHERE service_request_id = $1`

	rating := &models.ServiceRating{}
	err := r.db.QueryRow(ctx, query, requestID).Scan(
		&rating.ID, &rating.ServiceRequestID, &rating.ServiceID, &rating.UserID, &rating.Stars, &rating.Comment, &rating.CreatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("rating not found: %w", err)
	}
	return rating, nil
}

func (r *ServiceRatingRepository) GetStatsByServiceID(ctx context.Context, serviceID int64) (*models.ServiceRatingStats, error) {
	query := `SELECT COALESCE(AVG(stars), 0), COUNT(*) FROM service_ratings WHERE service_id = $1`

	stats := &models.ServiceRatingStats{}
	err := r.db.QueryRow(ctx, query, serviceID).Scan(&stats.Average, &stats.Count)
	if err != nil {
		return nil, fmt.Errorf("failed to get rating stats: %w", err)
	}
	return stats, nil
}

func (r *ServiceRatingRepository) ListByServiceID(ctx context.Context, serviceID int64, limit, offset int) ([]*models.ServiceRatingResponse, error) {
	query := `
		SELECT r.id, r.stars, r.comment, u.full_name, COALESCE(u.profile_image_url, ''), r.created_at
		FROM service_ratings r
		JOIN users u ON r.user_id = u.id
		WHERE r.service_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, serviceID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list ratings: %w", err)
	}
	defer rows.Close()

	var ratings []*models.ServiceRatingResponse
	for rows.Next() {
		res := &models.ServiceRatingResponse{}
		if err := rows.Scan(&res.ID, &res.Stars, &res.Comment, &res.UserName, &res.UserProfileImage, &res.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan rating: %w", err)
		}
		ratings = append(ratings, res)
	}
	if ratings == nil {
		ratings = []*models.ServiceRatingResponse{}
	}
	return ratings, nil
}
