package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type ServiceRequestRepository struct {
	db *pgxpool.Pool
}

func NewServiceRequestRepository(db *pgxpool.Pool) *ServiceRequestRepository {
	return &ServiceRequestRepository{db: db}
}

func (r *ServiceRequestRepository) CreateServiceRequest(ctx context.Context, userID *int64, req *models.CreateServiceRequestRequest) (*models.ServiceRequest, error) {
	// Fetch category and title from the referenced service
	var serviceCategory string
	err := r.db.QueryRow(ctx,
		`SELECT category FROM services WHERE id = $1 AND is_active = TRUE`,
		req.ServiceID,
	).Scan(&serviceCategory)
	if err != nil {
		return nil, fmt.Errorf("service not found or inactive: %w", err)
	}

	// Insert without protocol_number first to get the ID
	insertQuery := `
		INSERT INTO service_requests
			(user_id, service_id, service_title, category, request_data, attachments, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', NOW(), NOW())
		RETURNING id, user_id, service_id, protocol_number, service_title, category,
		          request_data, attachments, status, created_at, updated_at`

	sr := &models.ServiceRequest{}
	err = r.db.QueryRow(ctx, insertQuery,
		userID, req.ServiceID, req.ServiceTitle, serviceCategory,
		req.RequestData, req.Attachments,
	).Scan(
		&sr.ID, &sr.UserID, &sr.ServiceID, &sr.ProtocolNumber,
		&sr.ServiceTitle, &sr.Category, &sr.RequestData,
		&sr.Attachments, &sr.Status, &sr.CreatedAt, &sr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create service request: %w", err)
	}

	// Generate and set protocol number: SR-YYYYMMDD-{id}
	protocol := fmt.Sprintf("SR-%s-%d", time.Now().Format("20060102"), sr.ID)
	_, err = r.db.Exec(ctx,
		`UPDATE service_requests SET protocol_number = $1 WHERE id = $2`,
		protocol, sr.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to set protocol number: %w", err)
	}
	sr.ProtocolNumber = &protocol

	return sr, nil
}

func (r *ServiceRequestRepository) GetServiceRequestByID(ctx context.Context, id int64) (*models.ServiceRequest, error) {
	query := `SELECT id, user_id, service_id, protocol_number, service_title, category,
	                 request_data, attachments, status, created_at, updated_at
	          FROM service_requests WHERE id = $1`

	sr := &models.ServiceRequest{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sr.ID, &sr.UserID, &sr.ServiceID, &sr.ProtocolNumber,
		&sr.ServiceTitle, &sr.Category, &sr.RequestData,
		&sr.Attachments, &sr.Status, &sr.CreatedAt, &sr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("service request not found: %w", err)
	}
	return sr, nil
}

func (r *ServiceRequestRepository) ListServiceRequests(ctx context.Context) ([]*models.ServiceRequest, error) {
	query := `SELECT id, user_id, service_id, protocol_number, service_title, category,
	                 request_data, attachments, status, created_at, updated_at
	          FROM service_requests ORDER BY id DESC`

	return r.scanServiceRequests(ctx, query)
}

func (r *ServiceRequestRepository) ListServiceRequestsByUser(ctx context.Context, userID int64) ([]*models.ServiceRequest, error) {
	query := `SELECT id, user_id, service_id, protocol_number, service_title, category,
	                 request_data, attachments, status, created_at, updated_at
	          FROM service_requests WHERE user_id = $1 ORDER BY id DESC`

	return r.scanServiceRequests(ctx, query, userID)
}

func (r *ServiceRequestRepository) scanServiceRequests(ctx context.Context, query string, args ...interface{}) ([]*models.ServiceRequest, error) {
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list service requests: %w", err)
	}
	defer rows.Close()

	var list []*models.ServiceRequest
	for rows.Next() {
		sr := &models.ServiceRequest{}
		if err := rows.Scan(
			&sr.ID, &sr.UserID, &sr.ServiceID, &sr.ProtocolNumber,
			&sr.ServiceTitle, &sr.Category, &sr.RequestData,
			&sr.Attachments, &sr.Status, &sr.CreatedAt, &sr.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan service request: %w", err)
		}
		list = append(list, sr)
	}
	if list == nil {
		list = []*models.ServiceRequest{}
	}
	return list, nil
}

func (r *ServiceRequestRepository) UpdateServiceRequestStatus(ctx context.Context, id int64, status string) (*models.ServiceRequest, error) {
	validStatuses := map[string]bool{
		"pending": true, "in_progress": true, "completed": true, "cancelled": true,
	}
	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	query := `
		UPDATE service_requests SET status = $1, updated_at = NOW()
		WHERE id = $2
		RETURNING id, user_id, service_id, protocol_number, service_title, category,
		          request_data, attachments, status, created_at, updated_at`

	sr := &models.ServiceRequest{}
	err := r.db.QueryRow(ctx, query, status, id).Scan(
		&sr.ID, &sr.UserID, &sr.ServiceID, &sr.ProtocolNumber,
		&sr.ServiceTitle, &sr.Category, &sr.RequestData,
		&sr.Attachments, &sr.Status, &sr.CreatedAt, &sr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update service request status: %w", err)
	}
	return sr, nil
}

func (r *ServiceRequestRepository) DeleteServiceRequest(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, `DELETE FROM service_requests WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete service request: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("service request not found")
	}
	return nil
}
