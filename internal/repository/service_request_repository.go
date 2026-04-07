package repository

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"math/big"
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

	// Generate random protocol number: SR-YYYYMMDD + random 4 digits (e.g. SR-202603242048)
	randomNum, _ := rand.Int(rand.Reader, big.NewInt(10000))
	protocol := fmt.Sprintf("SR-%s%04d", time.Now().Format("20060102"), randomNum.Int64())

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

func (r *ServiceRequestRepository) ListServiceRequests(ctx context.Context, search string, page, limit int) ([]*models.ServiceRequest, error) {
	offset := (page - 1) * limit
	query := `SELECT sr.id, sr.user_id, sr.service_id, sr.protocol_number, sr.service_title, sr.category,
	                 sr.request_data, sr.attachments, sr.status, sr.created_at, sr.updated_at
	          FROM service_requests sr
	          LEFT JOIN users u ON sr.user_id = u.id`

	var args []interface{}
	if search != "" {
		query += ` WHERE (CAST(sr.id AS TEXT) ILIKE $1 OR sr.service_title ILIKE $1 OR sr.category ILIKE $1 OR u.full_name ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	query += fmt.Sprintf(` ORDER BY sr.id DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	return r.scanServiceRequests(ctx, query, args...)
}

func (r *ServiceRequestRepository) ListServiceRequestsByUser(ctx context.Context, userID int64, search string, page, limit int) ([]*models.ServiceRequest, error) {
	offset := (page - 1) * limit
	query := `SELECT id, user_id, service_id, protocol_number, service_title, category,
	                 request_data, attachments, status, created_at, updated_at
	          FROM service_requests WHERE user_id = $1`

	args := []interface{}{userID}
	if search != "" {
		query += ` AND (CAST(id AS TEXT) ILIKE $2 OR service_title ILIKE $2 OR category ILIKE $2)`
		args = append(args, "%"+search+"%")
	}

	query += fmt.Sprintf(` ORDER BY id DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

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

func (r *ServiceRequestRepository) CountServiceRequestsByUser(ctx context.Context, userID int64) (int, error) {
	var count int
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM service_requests WHERE user_id = $1`, userID).Scan(&count)
	return count, err
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

func (r *ServiceRequestRepository) GetHomeStats(ctx context.Context, isAdmin bool, userID int64) (*models.HomeResponse, error) {
	baseWhere := ""
	var args []interface{}

	if !isAdmin {
		baseWhere = "WHERE sr.user_id = $1"
		args = append(args, userID)
	}

	statsQuery := fmt.Sprintf(`
		SELECT status, COUNT(*) 
		FROM service_requests sr
		%s
		GROUP BY status`, baseWhere)

	rows, err := r.db.Query(ctx, statsQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %w", err)
	}
	defer rows.Close()

	counts := make(map[string]int)
	total := 0
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
		total += count
	}

	pct := func(val, tot int) int {
		if tot > 0 {
			return int((float64(val) / float64(tot)) * 100)
		}
		return 0
	}

	unresolved := counts["pending"] + counts["in_progress"] + counts["urgent"]

	var totalUsers int
	if isAdmin {
		r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	}

	var activeServices int
	if isAdmin {
		r.db.QueryRow(ctx, "SELECT COUNT(*) FROM services WHERE is_active = TRUE").Scan(&activeServices)
	}

	var completedToday int
	completedTodayQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM service_requests sr
		%s
		%s status = 'completed' AND updated_at::date = CURRENT_DATE
	`, baseWhere, map[bool]string{true: "AND", false: "WHERE"}[baseWhere != ""])
	r.db.QueryRow(ctx, completedTodayQuery, args...).Scan(&completedToday)

	stats := models.HomeStats{
		TotalRequests:       models.StatDetail{Total: total, Percent: 100},
		PendingRequests:     models.StatDetail{Total: counts["pending"], Percent: pct(counts["pending"], total)},
		InProgressRequests:  models.StatDetail{Total: counts["in_progress"], Percent: pct(counts["in_progress"], total)},
		CompletedRequests:   models.StatDetail{Total: counts["completed"], Percent: pct(counts["completed"], total)},
		CancelledRequests:   models.StatDetail{Total: counts["cancelled"], Percent: pct(counts["cancelled"], total)},
		UrgentRequests:      models.StatDetail{Total: counts["urgent"], Percent: pct(counts["urgent"], total)},
		UnresolvedRequests:  models.StatDetail{Total: unresolved, Percent: pct(unresolved, total)},
		TotalUsers:          totalUsers,
		TotalActiveServices: activeServices,
		CompletedToday:      completedToday,
	}

	catQuery := fmt.Sprintf(`
		SELECT category, COUNT(*) 
		FROM service_requests sr
		%s
		GROUP BY category
		ORDER BY COUNT(*) DESC`, baseWhere)

	catRows, err := r.db.Query(ctx, catQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	defer catRows.Close()

	var categories []models.CategoryStat
	for catRows.Next() {
		var cat string
		var count int
		if err := catRows.Scan(&cat, &count); err != nil {
			return nil, err
		}
		categories = append(categories, models.CategoryStat{
			Category: cat,
			Percent:  pct(count, total),
			Count:    count,
		})
	}
	if categories == nil {
		categories = []models.CategoryStat{}
	}

	recentQuery := fmt.Sprintf(`
		SELECT sr.id, u.full_name, sr.service_title, sr.request_data, sr.status, sr.created_at
		FROM service_requests sr
		LEFT JOIN users u ON sr.user_id = u.id
		%s
		ORDER BY sr.created_at DESC
		LIMIT 10`, baseWhere)

	recentRows, err := r.db.Query(ctx, recentQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent requests: %w", err)
	}
	defer recentRows.Close()

	var recent []models.RecentRequest
	for recentRows.Next() {
		var req models.RecentRequest
		var rawData []byte
		var createdAt time.Time
		if err := recentRows.Scan(&req.ID, &req.Name, &req.Service, &rawData, &req.Status, &createdAt); err != nil {
			return nil, err
		}
		req.Date = createdAt.Format("2006-01-02")

		var data map[string]interface{}
		if err := json.Unmarshal(rawData, &data); err == nil {
			if addr, ok := data["address"].(string); ok {
				req.Address = &addr
			} else if end, ok := data["endereco"].(string); ok {
				req.Address = &end
			}
		}

		recent = append(recent, req)
	}
	if recent == nil {
		recent = []models.RecentRequest{}
	}

	// Volume for the last 7 days
	var volume7d []models.VolumeStat
	volQuery := fmt.Sprintf(`
		SELECT date_trunc('day', created_at) as day, COUNT(*) 
		FROM service_requests sr
		%s
		%s created_at >= CURRENT_DATE - INTERVAL '7 days' 
		GROUP BY day 
		ORDER BY day ASC
	`, baseWhere, map[bool]string{true: "AND", false: "WHERE"}[baseWhere != ""])
	vRows, err := r.db.Query(ctx, volQuery, args...)
	if err == nil {
		defer vRows.Close()
		for vRows.Next() {
			var day time.Time
			var count int
			if err := vRows.Scan(&day, &count); err == nil {
				volume7d = append(volume7d, models.VolumeStat{
					Day:   day.Format("2006-01-02"),
					Count: count,
				})
			}
		}
	}
	if volume7d == nil {
		volume7d = []models.VolumeStat{}
	}

	return &models.HomeResponse{
		Stats:          stats,
		Categories:     categories,
		RecentRequests: recent,
		Volume7d:       volume7d,
	}, nil
}
