package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type ServiceRequestRepository struct {
	db *pgxpool.Pool
}

func NewServiceRequestRepository(db *pgxpool.Pool) *ServiceRequestRepository {
	return &ServiceRequestRepository{db: db}
}

func (r *ServiceRequestRepository) CreateServiceRequest(ctx context.Context, userID *int64, req *models.CreateServiceRequestRequest, regionID, teamID *int64, latitude, longitude *float64, geocodedAddress *string) (*models.ServiceRequest, error) {
	// Use a transaction to atomically verify service exists+active and insert request
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock the service row to prevent concurrent modifications
	var serviceCategory string
	err = tx.QueryRow(ctx,
		`SELECT category FROM services WHERE id = $1 AND is_active = TRUE FOR UPDATE`,
		req.ServiceID,
	).Scan(&serviceCategory)
	if err != nil {
		return nil, fmt.Errorf("service not found or inactive: %w", err)
	}

	// Single INSERT with subqueries in RETURNING to fetch related names
	insertQuery := `
		INSERT INTO service_requests
			(user_id, service_id, service_title, category, request_data, attachments, status, team_id, region_id, latitude, longitude, geocoded_address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id, user_id, service_id, protocol_number, service_title, category,
		          request_data, attachments, status, latitude, longitude, geocoded_address,
		          team_id, region_id,
		          COALESCE((SELECT full_name FROM users WHERE id = $1), ''),
		          COALESCE((SELECT name FROM teams WHERE id = $7), ''),
		          COALESCE((SELECT name FROM regions WHERE id = $8), ''),
		          created_at, updated_at`

sr := &models.ServiceRequest{}
	err = tx.QueryRow(ctx, insertQuery,
		userID, req.ServiceID, req.ServiceTitle, serviceCategory,
		req.RequestData, req.Attachments, teamID, regionID,
		latitude, longitude, geocodedAddress,
	).Scan(
		&sr.ID, &sr.UserID, &sr.ServiceID, &sr.ProtocolNumber,
		&sr.ServiceTitle, &sr.Category, &sr.RequestData,
		&sr.Attachments, &sr.Status, &sr.Latitude, &sr.Longitude, &sr.GeocodedAddress,
		&sr.TeamID, &sr.RegionID,
		&sr.UserName, &sr.TeamName, &sr.RegionName,
		&sr.CreatedAt, &sr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create service request: %w", err)
	}

	// Commit the transaction
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	if req.ServiceID != nil {
		sr.Icon = models.GetServiceIcon(*req.ServiceID)
	}

	// Generate unique protocol number with retry on collision
	protocolNumber, err := r.generateProtocolNumber(ctx, sr.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to set unique protocol number: %w", err)
	}
	sr.ProtocolNumber = &protocolNumber

	return sr, nil
}

func (r *ServiceRequestRepository) generateProtocolNumber(ctx context.Context, requestID int64) (string, error) {
	var protocolNumber string
	var err error

	// Retry up to 5 times in case of collision
	for i := 0; i < 5; i++ {
		err = r.db.QueryRow(ctx,
			`UPDATE service_requests
			 SET protocol_number = LPAD(CAST(FLOOR(RANDOM() * 100000000) AS TEXT), 8, '0')
			 WHERE id = $1 AND protocol_number IS NULL
			 RETURNING protocol_number`,
			requestID,
		).Scan(&protocolNumber)
		if err == nil {
			return protocolNumber, nil
		}
		// If UNIQUE violation, retry; otherwise return error
	}
	return "", fmt.Errorf("failed to generate unique protocol number after 5 attempts")
}

func (r *ServiceRequestRepository) GetServiceRequestByID(ctx context.Context, id int64) (*models.ServiceRequest, error) {
	query := `SELECT sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number,
	                 sr.service_title, sr.category, sr.request_data, sr.attachments, sr.status,
	                 sr.latitude, sr.longitude, sr.geocoded_address,
	                 sr.team_id, COALESCE(t.name, ''),
	                 sr.region_id, COALESCE(rg.name, ''),
	                 sr.created_at, sr.updated_at
	          FROM service_requests sr
	          LEFT JOIN users u ON sr.user_id = u.id
	          LEFT JOIN teams t ON sr.team_id = t.id
	          LEFT JOIN regions rg ON sr.region_id = rg.id
	          WHERE sr.id = $1`

	sr := &models.ServiceRequest{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sr.ID, &sr.UserID, &sr.UserName, &sr.ServiceID, &sr.ProtocolNumber,
		&sr.ServiceTitle, &sr.Category, &sr.RequestData,
		&sr.Attachments, &sr.Status, &sr.Latitude, &sr.Longitude, &sr.GeocodedAddress,
		&sr.TeamID, &sr.TeamName,
		&sr.RegionID, &sr.RegionName,
		&sr.CreatedAt, &sr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("service request not found: %w", err)
	}

	if sr.ServiceID != nil {
		sr.Icon = models.GetServiceIcon(*sr.ServiceID)
	}

	return sr, nil
}

func (r *ServiceRequestRepository) ListServiceRequests(ctx context.Context, search, status, category string, regionFilter, teamFilter *int64, startDate, endDate *string, page, limit int) ([]*models.ServiceRequest, error) {
	query := `SELECT sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number,
	                 sr.service_title, sr.category, sr.request_data, sr.attachments, sr.status,
	                 sr.latitude, sr.longitude, sr.geocoded_address,
	                 sr.team_id, COALESCE(t.name, ''),
	                 sr.region_id, COALESCE(rg.name, ''),
	                 sr.created_at, sr.updated_at
	          FROM service_requests sr
	          LEFT JOIN users u ON sr.user_id = u.id
	          LEFT JOIN teams t ON sr.team_id = t.id
	          LEFT JOIN regions rg ON sr.region_id = rg.id`

	var args []interface{}
	whereApplied := false

	if search != "" {
		query += ` WHERE (sr.service_title ILIKE $1 OR sr.category ILIKE $1 OR u.full_name ILIKE $1)`
		args = append(args, "%"+search+"%")
		whereApplied = true
	}

	if status != "" {
		if whereApplied {
			query += fmt.Sprintf(` AND sr.status = $%d`, len(args)+1)
		} else {
			query += ` WHERE sr.status = $1`
		}
		args = append(args, status)
		whereApplied = true
	}

	if category != "" {
		if whereApplied {
			query += fmt.Sprintf(` AND sr.category = $%d`, len(args)+1)
		} else {
			query += ` WHERE sr.category = $1`
		}
		args = append(args, category)
		whereApplied = true
	}

	if regionFilter != nil {
		if whereApplied {
			query += fmt.Sprintf(` AND sr.region_id = $%d`, len(args)+1)
		} else {
			query += ` WHERE sr.region_id = $1`
		}
		args = append(args, *regionFilter)
		whereApplied = true
	}

	if teamFilter != nil {
		if whereApplied {
			query += fmt.Sprintf(` AND sr.team_id = $%d`, len(args)+1)
		} else {
			query += ` WHERE sr.team_id = $1`
		}
		args = append(args, *teamFilter)
		whereApplied = true
	}

	if startDate != nil && *startDate != "" {
		if whereApplied {
			query += fmt.Sprintf(` AND sr.created_at >= $%d`, len(args)+1)
		} else {
			query += fmt.Sprintf(` WHERE sr.created_at >= $%d`, len(args)+1)
		}
		args = append(args, *startDate)
		whereApplied = true
	}

	if endDate != nil && *endDate != "" {
		if whereApplied {
			query += fmt.Sprintf(` AND sr.created_at <= $%d::date + interval '1 day'`, len(args)+1)
		} else {
			query += fmt.Sprintf(` WHERE sr.created_at <= $%d::date + interval '1 day'`, len(args)+1)
		}
		args = append(args, *endDate)
		whereApplied = true
	}

	query += ` ORDER BY sr.id DESC`
	if limit > 0 {
		offset := (page - 1) * limit
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	}

	return r.scanServiceRequests(ctx, query, args...)
}

func (r *ServiceRequestRepository) ListServiceRequestsByUser(ctx context.Context, userID int64, search, status string, regionFilter *int64, page, limit int) ([]*models.ServiceRequest, error) {
	query := `SELECT sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number,
	                 sr.service_title, sr.category, sr.request_data, sr.attachments, sr.status,
	                 sr.latitude, sr.longitude, sr.geocoded_address,
	                 sr.team_id, COALESCE(t.name, ''),
	                 sr.region_id, COALESCE(rg.name, ''),
	                 sr.created_at, sr.updated_at
	          FROM service_requests sr
	          LEFT JOIN users u ON sr.user_id = u.id
	          LEFT JOIN teams t ON sr.team_id = t.id
	          LEFT JOIN regions rg ON sr.region_id = rg.id
	          WHERE sr.user_id = $1`

	args := []interface{}{userID}
	if search != "" {
		query += ` AND (sr.service_title ILIKE $2 OR sr.category ILIKE $2)`
		args = append(args, "%"+search+"%")
	}

	if status != "" {
		query += fmt.Sprintf(` AND sr.status = $%d`, len(args)+1)
		args = append(args, status)
	}

	if regionFilter != nil {
		query += fmt.Sprintf(` AND sr.region_id = $%d`, len(args)+1)
		args = append(args, *regionFilter)
	}

	query += ` ORDER BY sr.id DESC`
	if limit > 0 {
		offset := (page - 1) * limit
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	}

	return r.scanServiceRequests(ctx, query, args...)
}

func (r *ServiceRequestRepository) ListServiceRequestsByTeam(ctx context.Context, teamID int64, search, status string, page, limit int) ([]*models.ServiceRequest, error) {
	// 1. Get secretary's work areas to filter by category
	var workAreaJSON []byte
	err := r.db.QueryRow(ctx, `
		SELECT work_area FROM users 
		WHERE team_id = $1 AND type = 'secretary' 
		LIMIT 1`, teamID).Scan(&workAreaJSON)

	var workAreas []string
	if err == nil && workAreaJSON != nil {
		json.Unmarshal(workAreaJSON, &workAreas)
	}

	query := `SELECT sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number,
	                 sr.service_title, sr.category, sr.request_data, sr.attachments, sr.status,
	                 sr.latitude, sr.longitude, sr.geocoded_address,
	                 sr.team_id, COALESCE(t.name, ''),
	                 sr.region_id, COALESCE(rg.name, ''),
	                 sr.created_at, sr.updated_at
	          FROM service_requests sr
	          LEFT JOIN users u ON sr.user_id = u.id
	          LEFT JOIN teams t ON sr.team_id = t.id
	          LEFT JOIN regions rg ON sr.region_id = rg.id
	          WHERE sr.team_id = $1`

	args := []interface{}{teamID}

	if len(workAreas) > 0 {
		query += fmt.Sprintf(` AND sr.category = ANY($%d)`, len(args)+1)
		args = append(args, workAreas)
	}

	if search != "" {
		query += fmt.Sprintf(` AND (sr.service_title ILIKE $%d OR sr.category ILIKE $%d)`, len(args)+1, len(args)+1)
		args = append(args, "%"+search+"%")
	}

	if status != "" {
		query += fmt.Sprintf(` AND sr.status = $%d`, len(args)+1)
		args = append(args, status)
	}

	query += ` ORDER BY sr.id DESC`
	if limit > 0 {
		offset := (page - 1) * limit
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	}

	return r.scanServiceRequests(ctx, query, args...)
}

func (r *ServiceRequestRepository) ListServiceRequestsByCategory(ctx context.Context, category string, teamID *int64, limit int) ([]*models.ServiceRequest, error) {
	query := `SELECT sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number,
	                 sr.service_title, sr.category, sr.request_data, sr.attachments, sr.status,
	                 sr.latitude, sr.longitude, sr.geocoded_address,
	                 sr.team_id, COALESCE(t.name, ''),
	                 sr.region_id, COALESCE(rg.name, ''),
	                 sr.created_at, sr.updated_at
	          FROM service_requests sr
	          LEFT JOIN users u ON sr.user_id = u.id
	          LEFT JOIN teams t ON sr.team_id = t.id
	          LEFT JOIN regions rg ON sr.region_id = rg.id
	          WHERE sr.category = $1`
	var args []interface{}
	args = append(args, category)

	if teamID != nil {
		query += fmt.Sprintf(` AND sr.team_id = $%d`, len(args)+1)
		args = append(args, *teamID)
	}

	query += ` ORDER BY sr.id DESC`

	if limit > 0 {
		query += fmt.Sprintf(` LIMIT %d`, limit)
	}

	return r.scanServiceRequests(ctx, query, args...)
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
			&sr.ID, &sr.UserID, &sr.UserName, &sr.ServiceID, &sr.ProtocolNumber,
			&sr.ServiceTitle, &sr.Category, &sr.RequestData,
			&sr.Attachments, &sr.Status, &sr.Latitude, &sr.Longitude, &sr.GeocodedAddress,
			&sr.TeamID, &sr.TeamName,
			&sr.RegionID, &sr.RegionName,
			&sr.CreatedAt, &sr.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan service request: %w", err)
		}

		if sr.ServiceID != nil {
			sr.Icon = models.GetServiceIcon(*sr.ServiceID)
		}

		list = append(list, sr)
	}
	if list == nil {
		list = []*models.ServiceRequest{}
	}
	return list, nil
}

func (r *ServiceRequestRepository) GetServiceCategory(ctx context.Context, serviceID int64) (string, error) {
	var category string
	err := r.db.QueryRow(ctx, `SELECT category FROM services WHERE id = $1 AND is_active = TRUE`, serviceID).Scan(&category)
	if err != nil {
		return "", fmt.Errorf("service not found: %w", err)
	}
	return category, nil
}

func (r *ServiceRequestRepository) UpdateServiceRequestStatus(ctx context.Context, id int64, status string) (*models.ServiceRequest, error) {
	validStatuses := map[string]bool{
		"pending": true, "in_progress": true, "completed": true, "cancelled": true,
	}
	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	query := `
		WITH updated AS (
			UPDATE service_requests
			SET status = $2, updated_at = NOW()
			WHERE id = $1
			RETURNING id, user_id, service_id, protocol_number, service_title, category,
			          request_data, attachments, status, latitude, longitude, geocoded_address,
			          team_id, region_id, created_at, updated_at
		)
		SELECT u.id, u.user_id, COALESCE(u2.full_name, ''), u.service_id, u.protocol_number,
		       u.service_title, u.category, u.request_data, u.attachments, u.status,
		       u.latitude, u.longitude, u.geocoded_address,
		       u.team_id, COALESCE(t.name, ''),
		       u.region_id, COALESCE(rg.name, ''),
		       u.created_at, u.updated_at
		FROM updated u
		LEFT JOIN users u2 ON u.user_id = u2.id
		LEFT JOIN teams t ON u.team_id = t.id
		LEFT JOIN regions rg ON u.region_id = rg.id`

	sr := &models.ServiceRequest{}
	err := r.db.QueryRow(ctx, query, id, status).Scan(
		&sr.ID, &sr.UserID, &sr.UserName, &sr.ServiceID, &sr.ProtocolNumber,
		&sr.ServiceTitle, &sr.Category, &sr.RequestData,
		&sr.Attachments, &sr.Status, &sr.Latitude, &sr.Longitude, &sr.GeocodedAddress,
		&sr.TeamID, &sr.TeamName,
		&sr.RegionID, &sr.RegionName,
		&sr.CreatedAt, &sr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update service request status: %w", err)
	}

	if sr.ServiceID != nil {
		sr.Icon = models.GetServiceIcon(*sr.ServiceID)
	}

	return sr, nil
}

func (r *ServiceRequestRepository) UpdateServiceRequest(ctx context.Context, id int64, req *models.CreateServiceRequestRequest) (*models.ServiceRequest, error) {
	query := `
		WITH updated AS (
			UPDATE service_requests
			SET request_data = $2, attachments = $3, updated_at = NOW()
			WHERE id = $1
			RETURNING id, user_id, service_id, protocol_number, service_title, category,
			          request_data, attachments, status, latitude, longitude, geocoded_address,
			          team_id, region_id, created_at, updated_at
		)
		SELECT u.id, u.user_id, COALESCE(u2.full_name, ''), u.service_id, u.protocol_number,
		       u.service_title, u.category, u.request_data, u.attachments, u.status,
		       u.latitude, u.longitude, u.geocoded_address,
		       u.team_id, COALESCE(t.name, ''),
		       u.region_id, COALESCE(rg.name, ''),
		       u.created_at, u.updated_at
		FROM updated u
		LEFT JOIN users u2 ON u.user_id = u2.id
		LEFT JOIN teams t ON u.team_id = t.id
		LEFT JOIN regions rg ON u.region_id = rg.id`

	sr := &models.ServiceRequest{}
	err := r.db.QueryRow(ctx, query, id, req.RequestData, req.Attachments).Scan(
		&sr.ID, &sr.UserID, &sr.UserName, &sr.ServiceID, &sr.ProtocolNumber,
		&sr.ServiceTitle, &sr.Category, &sr.RequestData,
		&sr.Attachments, &sr.Status, &sr.Latitude, &sr.Longitude, &sr.GeocodedAddress,
		&sr.TeamID, &sr.TeamName,
		&sr.RegionID, &sr.RegionName,
		&sr.CreatedAt, &sr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update service request: %w", err)
	}

	if sr.ServiceID != nil {
		sr.Icon = models.GetServiceIcon(*sr.ServiceID)
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

func (r *ServiceRequestRepository) SaveGeocoding(ctx context.Context, id int64, lat, lon float64, address string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE service_requests SET latitude = $1, longitude = $2, geocoded_address = $3 WHERE id = $4`,
		lat, lon, address, id,
	)
	return err
}
