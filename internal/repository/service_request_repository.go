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

	// Fetch the full name
	if userID != nil {
		r.db.QueryRow(ctx, `SELECT full_name FROM users WHERE id = $1`, *userID).Scan(&sr.UserName)
	}

	// Generate a unique 8-digit random protocol number
	// We use a retry loop to handle potential collisions in the unique constraint
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
	query := `SELECT sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number, sr.service_title, sr.category,
	                 sr.request_data, sr.attachments, sr.status, sr.created_at, sr.updated_at
	          FROM service_requests sr
	          LEFT JOIN users u ON sr.user_id = u.id
	          WHERE sr.id = $1`

	sr := &models.ServiceRequest{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sr.ID, &sr.UserID, &sr.UserName, &sr.ServiceID, &sr.ProtocolNumber,
		&sr.ServiceTitle, &sr.Category, &sr.RequestData,
		&sr.Attachments, &sr.Status, &sr.CreatedAt, &sr.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("service request not found: %w", err)
	}
	return sr, nil
}

func (r *ServiceRequestRepository) ListServiceRequests(ctx context.Context, search, categoryFilter string, page, limit int) ([]*models.ServiceRequest, error) {
	offset := (page - 1) * limit
	query := `SELECT sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number, sr.service_title, sr.category,
	                 sr.request_data, sr.attachments, sr.status, sr.created_at, sr.updated_at
	          FROM service_requests sr
	          LEFT JOIN users u ON sr.user_id = u.id`

	var args []interface{}
	whereApplied := false

	if search != "" {
		query += ` WHERE (CAST(sr.id AS TEXT) ILIKE $1 OR sr.service_title ILIKE $1 OR sr.category ILIKE $1 OR u.full_name ILIKE $1)`
		args = append(args, "%"+search+"%")
		whereApplied = true
	}

	if categoryFilter != "" {
		if whereApplied {
			query += fmt.Sprintf(` AND sr.category = $%d`, len(args)+1)
		} else {
			query += ` WHERE sr.category = $1`
		}
		args = append(args, categoryFilter)
	}

	query += fmt.Sprintf(` ORDER BY sr.id DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	return r.scanServiceRequests(ctx, query, args...)
}

func (r *ServiceRequestRepository) ListServiceRequestsByUser(ctx context.Context, userID int64, search, categoryFilter string, page, limit int) ([]*models.ServiceRequest, error) {
	offset := (page - 1) * limit
	query := `SELECT sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number, sr.service_title, sr.category,
	                 sr.request_data, sr.attachments, sr.status, sr.created_at, sr.updated_at
	          FROM service_requests sr
	          LEFT JOIN users u ON sr.user_id = u.id
	          WHERE sr.user_id = $1`

	args := []interface{}{userID}
	if search != "" {
		query += ` AND (CAST(sr.id AS TEXT) ILIKE $2 OR sr.service_title ILIKE $2 OR sr.category ILIKE $2)`
		args = append(args, "%"+search+"%")
	}
	
	if categoryFilter != "" {
		query += fmt.Sprintf(` AND sr.category = $%d`, len(args)+1)
		args = append(args, categoryFilter)
	}

	query += fmt.Sprintf(` ORDER BY sr.id DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
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
			&sr.ID, &sr.UserID, &sr.UserName, &sr.ServiceID, &sr.ProtocolNumber,
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

func (r *ServiceRequestRepository) CountServiceRequestsByStatusByUser(ctx context.Context, userID int64) (map[string]int, error) {
	query := `
		SELECT status, COUNT(*) 
		FROM service_requests 
		WHERE user_id = $1
		GROUP BY status`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	counts := map[string]int{
		"pending":     0,
		"in_progress": 0,
		"completed":   0,
		"cancelled":   0,
	}

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, err
		}
		counts[status] = count
	}

	return counts, nil
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

	// Fetch user name
	if sr.UserID != nil {
		r.db.QueryRow(ctx, `SELECT full_name FROM users WHERE id = $1`, *sr.UserID).Scan(&sr.UserName)
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

func (r *ServiceRequestRepository) ListServiceRequestDetailsByService(ctx context.Context, serviceID int64, page, limit int) ([]*models.ServiceRequestDetailResponse, error) {
	offset := (page - 1) * limit
	query := `SELECT sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number, sr.service_title, sr.category,
	                 sr.request_data, sr.attachments, sr.status, sr.created_at, sr.updated_at,
	                 u.username, u.email, u.cpf, u.birth_date, u.type, u.created_at, u.updated_at
	          FROM service_requests sr
	          LEFT JOIN users u ON sr.user_id = u.id
	          WHERE sr.service_id = $1
	          ORDER BY sr.created_at DESC
	          LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, serviceID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.ServiceRequestDetailResponse
	for rows.Next() {
		sr := &models.ServiceRequest{}
		user := &models.User{}
		var uID *int64
		if err := rows.Scan(
			&sr.ID, &uID, &sr.UserName, &sr.ServiceID, &sr.ProtocolNumber,
			&sr.ServiceTitle, &sr.Category, &sr.RequestData,
			&sr.Attachments, &sr.Status, &sr.CreatedAt, &sr.UpdatedAt,
			&user.Username, &user.Email, &user.CPF, &user.BirthDate, &user.Type, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			// If user is null, partial scan might fail or return zero values.
			// However, since we joined with u.id it should be fine if there is a user.
			// If u.id is null, those fields will be null/zero.
			return nil, err
		}
		sr.UserID = uID
		if uID != nil {
			user.ID = *uID
			user.FullName = &sr.UserName
			list = append(list, &models.ServiceRequestDetailResponse{
				ServiceRequest: sr,
				CreatedBy:      user,
			})
		} else {
			list = append(list, &models.ServiceRequestDetailResponse{
				ServiceRequest: sr,
			})
		}
	}
	if list == nil {
		list = []*models.ServiceRequestDetailResponse{}
	}
	return list, nil
}

func (r *ServiceRequestRepository) GetServiceStatusStats(ctx context.Context, serviceID int64) ([]models.StatusStat, error) {
	query := `SELECT status, COUNT(*) FROM service_requests WHERE service_id = $1 GROUP BY status`
	rows, err := r.db.Query(ctx, query, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stats []models.StatusStat
	for rows.Next() {
		var s models.StatusStat
		if err := rows.Scan(&s.Status, &s.Total); err != nil {
			return nil, err
		}
		stats = append(stats, s)
	}
	if stats == nil {
		stats = []models.StatusStat{}
	}
	return stats, nil
}

func (r *ServiceRequestRepository) GetAverageServiceTime(ctx context.Context, serviceID int64) (int, error) {
	queryAvg := `
		SELECT 
			COALESCE(ROUND(EXTRACT(EPOCH FROM AVG(updated_at - created_at)) / 86400)::int, 0)
		FROM service_requests
		WHERE service_id = $1 AND status = 'completed'`

	var result int
	err := r.db.QueryRow(ctx, queryAvg, serviceID).Scan(&result)
	if err != nil {
		return 0, nil
	}

	return result, nil
}

func (r *ServiceRequestRepository) GetHomeStats(ctx context.Context, isAdmin bool, userID int64, categoryFilter string) (*models.HomeResponse, error) {
	baseWhere := ""
	var args []interface{}

	if !isAdmin {
		baseWhere = "WHERE sr.user_id = $1"
		args = append(args, userID)
	} else if categoryFilter != "" {
		baseWhere = "WHERE sr.category = $1"
		args = append(args, categoryFilter)
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

	var createdToday int
	createdTodayQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM service_requests sr
		%s
		%s created_at::date = CURRENT_DATE
	`, baseWhere, map[bool]string{true: "AND", false: "WHERE"}[baseWhere != ""])
	r.db.QueryRow(ctx, createdTodayQuery, args...).Scan(&createdToday)

	var avgTime float64
	avgTimeQuery := fmt.Sprintf(`
		SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 86400), 0)
		FROM service_requests sr
		%s
		%s status = 'completed'
	`, baseWhere, map[bool]string{true: "AND", false: "WHERE"}[baseWhere != ""])
	r.db.QueryRow(ctx, avgTimeQuery, args...).Scan(&avgTime)

	// Get top 5 most requested services
	var topServices []models.PopularService
	topQuery := `
		SELECT s.id, s.title, s.category, COUNT(sr.id) as request_count
		FROM services s
		INNER JOIN service_requests sr ON s.id = sr.service_id
		WHERE sr.status != 'cancelled'
		GROUP BY s.id, s.title, s.category
		ORDER BY request_count DESC
		LIMIT 5
	`
	topRows, err := r.db.Query(ctx, topQuery)
	if err != nil {
		fmt.Printf("Warning: failed to fetch popular services: %v\n", err)
	} else {
		defer topRows.Close()
		for topRows.Next() {
			svc := models.PopularService{}
			if err := topRows.Scan(&svc.ID, &svc.Title, &svc.Category, &svc.RequestCount); err != nil {
				fmt.Printf("Warning: failed to scan popular service: %v\n", err)
				continue
			}
			topServices = append(topServices, svc)
		}
	}
	if topServices == nil {
		topServices = []models.PopularService{}
	}

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
		CreatedToday:        createdToday,
		AverageTime:         avgTime,
		PopularServices:     topServices,
	}

	// Calculate Alerts
	alerts := []models.HomeAlert{}
	
	// 1. Stagnant requests (> 3 days) - GLOBAL
	var stagnantCount int
	stagnantQuery := `
		SELECT COUNT(*) 
		FROM service_requests
		WHERE status IN ('pending', 'in_progress') AND created_at < NOW() - INTERVAL '3 days'
	`
	r.db.QueryRow(ctx, stagnantQuery).Scan(&stagnantCount)
	
	if stagnantCount > 0 {
		alerts = append(alerts, models.HomeAlert{
			Type:    "danger",
			Message: fmt.Sprintf("%d solicitações paradas há mais de 3 dias", stagnantCount),
		})
	}

	// 2. Most critical service - GLOBAL (Only if > 5 pending/urgent)
	var criticalService string
	var criticalCount int
	criticalQuery := `
		SELECT service_title, COUNT(*)
		FROM service_requests
		WHERE status IN ('pending', 'urgent')
		GROUP BY service_title
		ORDER BY COUNT(*) DESC
		LIMIT 1
	`
	r.db.QueryRow(ctx, criticalQuery).Scan(&criticalService, &criticalCount)

	if criticalService != "" && criticalCount > 5 {
		alerts = append(alerts, models.HomeAlert{
			Type:    "warning",
			Message: fmt.Sprintf("Serviço mais crítico: %s com %d pendências", criticalService, criticalCount),
		})
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

	// Ensure standard categories are present for the dashboard cards
	standardCategories := []string{
		"Limpeza Urbana", "Saúde", "Educação", "Iluminação Pública",
		"Transporte Urbano", "Segurança Pública", "Esporte e Lazer", "Cultura",
		"Tributação", "Assistência Social", "Vias Urbanas", "Arborização e Meio Ambiente",
		"Agricultura", "Vigilância Sanitária", "Animais",
	}

	for _, sc := range standardCategories {
		found := false
		for _, c := range categories {
			if c.Category == sc {
				found = true
				break
			}
		}
		if !found {
			categories = append(categories, models.CategoryStat{
				Category: sc,
				Percent:  0,
				Count:    0,
			})
		}
	}

	if categories == nil {
		categories = []models.CategoryStat{}
	}

	fetchRecent := func(query string, args ...interface{}) ([]models.RecentRequest, error) {
		rows, err := r.db.Query(ctx, query, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var list []models.RecentRequest
		for rows.Next() {
			var req models.RecentRequest
			var rawData []byte
			var createdAt time.Time
			if err := rows.Scan(&req.ID, &req.Name, &req.Service, &rawData, &req.Status, &createdAt); err != nil {
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
			list = append(list, req)
		}
		if list == nil {
			list = []models.RecentRequest{}
		}
		return list, nil
	}

	// 1. All Recent (as before)
	recentQuery := fmt.Sprintf(`
		SELECT sr.id, u.full_name, sr.service_title, sr.request_data, sr.status, sr.created_at
		FROM service_requests sr
		LEFT JOIN users u ON sr.user_id = u.id
		%s
		ORDER BY sr.created_at DESC
		LIMIT 10`, baseWhere)
	recent, _ := fetchRecent(recentQuery, args...)

	// 2. Delayed (> 3 days)
	delayedQuery := fmt.Sprintf(`
		SELECT sr.id, u.full_name, sr.service_title, sr.request_data, sr.status, sr.created_at
		FROM service_requests sr
		LEFT JOIN users u ON sr.user_id = u.id
		%s
		%s sr.status IN ('pending', 'in_progress') AND sr.created_at < NOW() - INTERVAL '3 days'
		ORDER BY sr.created_at ASC
		LIMIT 10`, baseWhere, map[bool]string{true: "AND", false: "WHERE"}[baseWhere != ""])
	delayed, _ := fetchRecent(delayedQuery, args...)

	// 3. New (last 24h)
	newQuery := fmt.Sprintf(`
		SELECT sr.id, u.full_name, sr.service_title, sr.request_data, sr.status, sr.created_at
		FROM service_requests sr
		LEFT JOIN users u ON sr.user_id = u.id
		%s
		%s sr.created_at >= NOW() - INTERVAL '24 hours'
		ORDER BY sr.created_at DESC
		LIMIT 10`, baseWhere, map[bool]string{true: "AND", false: "WHERE"}[baseWhere != ""])
	newReqs, _ := fetchRecent(newQuery, args...)

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
		Stats:            stats,
		Categories:       categories,
		RecentRequests:   recent,
		DelayedRequests:  delayed,
		NewRequests:      newReqs,
		Volume7d:         volume7d,
		Alerts:           alerts,
		PopularServices:  topServices,
	}, nil
}
