package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"math/big"

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
	// Fetch category from the referenced service
	var serviceCategory string
	err := r.db.QueryRow(ctx,
		`SELECT category FROM services WHERE id = $1 AND is_active = TRUE`,
		req.ServiceID,
	).Scan(&serviceCategory)
	if err != nil {
		return nil, fmt.Errorf("service not found or inactive: %w", err)
	}

	insertQuery := `
		INSERT INTO service_requests
			(user_id, service_id, service_title, category, request_data, attachments, status, team_id, region_id, latitude, longitude, geocoded_address, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, 'pending', $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id, user_id, service_id, protocol_number, service_title, category,
		          request_data, attachments, status, latitude, longitude, geocoded_address,
		          team_id, region_id, created_at, updated_at`

	sr := &models.ServiceRequest{}
	err = r.db.QueryRow(ctx, insertQuery,
		userID, req.ServiceID, req.ServiceTitle, serviceCategory,
		req.RequestData, req.Attachments, teamID, regionID,
		latitude, longitude, geocodedAddress,
	).Scan(
		&sr.ID, &sr.UserID, &sr.ServiceID, &sr.ProtocolNumber,
		&sr.ServiceTitle, &sr.Category, &sr.RequestData,
		&sr.Attachments, &sr.Status, &sr.Latitude, &sr.Longitude, &sr.GeocodedAddress,
		&sr.TeamID, &sr.RegionID, &sr.CreatedAt, &sr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create service request: %w", err)
	}

	// Single optimized query to fetch all related data (user, team, region names)
	// This avoids N+1 queries by fetching everything in one go
	query := `
		SELECT COALESCE(u.full_name, ''), COALESCE(t.name, ''), COALESCE(rg.name, '')
		FROM service_requests sr
		LEFT JOIN users u ON sr.user_id = u.id
		LEFT JOIN teams t ON sr.team_id = t.id
		LEFT JOIN regions rg ON sr.region_id = rg.id
		WHERE sr.id = $1`
	
	err = r.db.QueryRow(ctx, query, sr.ID).Scan(&sr.UserName, &sr.TeamName, &sr.RegionName)
	if err != nil {
		// Log error but don't fail - the request was created successfully
		log.Printf("Warning: failed to fetch related data for service request %d: %v", sr.ID, err)
	}

	if req.ServiceID != nil {
		sr.Icon = models.GetServiceIcon(*req.ServiceID)
	}

	// Generate a unique 8-digit random protocol number
	var finalProtocol string
	var lastErr error
	for i := 0; i < 5; i++ {
		tempN, _ := rand.Int(rand.Reader, big.NewInt(100000000))
		p := fmt.Sprintf("%08d", tempN.Int64())

		_, lastErr = r.db.Exec(ctx,
			`UPDATE service_requests SET protocol_number = $1 WHERE id = $2`,
			p, sr.ID,
		)
		if lastErr == nil {
			finalProtocol = p
			break
		}
	}

	if finalProtocol == "" {
		return nil, fmt.Errorf("failed to set unique protocol number after retries: %w", lastErr)
	}

	sr.ProtocolNumber = &finalProtocol
	return sr, nil
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

func (r *ServiceRequestRepository) ListServiceRequests(ctx context.Context, search, status string, regionFilter, teamFilter *int64, startDate, endDate *string, page, limit int) ([]*models.ServiceRequest, error) {
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
		query += ` WHERE (CAST(sr.id AS TEXT) ILIKE $1 OR sr.service_title ILIKE $1 OR sr.category ILIKE $1 OR u.full_name ILIKE $1)`
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
		query += ` AND (CAST(sr.id AS TEXT) ILIKE $2 OR sr.service_title ILIKE $2 OR sr.category ILIKE $2)`
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
	if search != "" {
		query += ` AND (CAST(sr.id AS TEXT) ILIKE $2 OR sr.service_title ILIKE $2 OR sr.category ILIKE $2)`
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

func (r *ServiceRequestRepository) UpdateServiceRequestStatus(ctx context.Context, id int64, status string) (*models.ServiceRequest, error) {
	validStatuses := map[string]bool{
		"pending": true, "in_progress": true, "completed": true, "cancelled": true,
	}
	if !validStatuses[status] {
		return nil, fmt.Errorf("invalid status: %s", status)
	}

	// Single JOIN query to avoid N+1: fetch sr + team name + region name + user name in one query
	query := `
		SELECT sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number,
		       sr.service_title, sr.category, sr.request_data, sr.attachments, sr.status,
		       sr.latitude, sr.longitude, sr.geocoded_address,
		       sr.team_id, COALESCE(t.name, ''),
		       sr.region_id, COALESCE(rg.name, ''),
		       sr.created_at, sr.updated_at
		FROM service_requests sr
		LEFT JOIN users u ON sr.user_id = u.id
		LEFT JOIN teams t ON sr.team_id = t.id
		LEFT JOIN regions rg ON sr.region_id = rg.id
		WHERE sr.id = $1 AND sr.status = $2
		RETURNING sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number,
		          sr.service_title, sr.category, sr.request_data, sr.attachments, sr.status,
		          sr.latitude, sr.longitude, sr.geocoded_address,
		          sr.team_id, COALESCE(t.name, ''),
		          sr.region_id, COALESCE(rg.name, ''),
		          sr.created_at, sr.updated_at`

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
