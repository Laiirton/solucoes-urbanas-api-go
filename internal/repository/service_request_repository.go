package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type ServiceRequestRepository struct {
	db *pgxpool.Pool
}

func NewServiceRequestRepository(db *pgxpool.Pool) *ServiceRequestRepository {
	return &ServiceRequestRepository{db: db}
}

const srFields = `sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number, sr.service_title, sr.category, sr.request_data, sr.attachments, sr.status, sr.created_at, sr.updated_at`

func (r *ServiceRequestRepository) scan(row pgx.Row, sr *models.ServiceRequest) error {
	err := row.Scan(&sr.ID, &sr.UserID, &sr.UserName, &sr.ServiceID, &sr.ProtocolNumber, &sr.ServiceTitle, &sr.Category, &sr.RequestData, &sr.Attachments, &sr.Status, &sr.CreatedAt, &sr.UpdatedAt)
	if err == nil && sr.ServiceID != nil {
		sr.Icon = models.GetServiceIcon(*sr.ServiceID)
	}
	return err
}

func (r *ServiceRequestRepository) CreateServiceRequest(ctx context.Context, userID *int64, req *models.CreateServiceRequestRequest) (*models.ServiceRequest, error) {
	// Consolidate creation and protocol generation in one transaction using CTE
	query := `
		WITH ins AS (
			INSERT INTO service_requests (user_id, service_id, service_title, category, request_data, attachments, status, created_at, updated_at)
			SELECT $1, $2, $3, category, $4, $5, 'pending', NOW(), NOW()
			FROM services WHERE id = $2 AND is_active = TRUE
			RETURNING id
		), 
		upd AS (
			UPDATE service_requests SET protocol_number = LPAD(ins.id::TEXT, 8, '0')
			FROM ins WHERE service_requests.id = ins.id
			RETURNING id
		)
		SELECT ` + srFields + ` FROM service_requests sr
		LEFT JOIN users u ON sr.user_id = u.id
		WHERE sr.id = (SELECT id FROM upd)`

	sr := &models.ServiceRequest{}
	if err := r.scan(r.db.QueryRow(ctx, query, userID, req.ServiceID, req.ServiceTitle, req.RequestData, req.Attachments), sr); err != nil {
		return nil, fmt.Errorf("failed to create service request: %w", err)
	}
	return sr, nil
}

func (r *ServiceRequestRepository) GetServiceRequestByID(ctx context.Context, id int64) (*models.ServiceRequest, error) {
	sr := &models.ServiceRequest{}
	if err := r.scan(r.db.QueryRow(ctx, "SELECT "+srFields+" FROM service_requests sr LEFT JOIN users u ON sr.user_id = u.id WHERE sr.id = $1", id), sr); err != nil {
		return nil, err
	}
	return sr, nil
}

func (r *ServiceRequestRepository) ListServiceRequests(ctx context.Context, search, cat string, page, limit int) ([]*models.ServiceRequest, error) {
	query, args := r.buildListQuery(search, cat, 0, page, limit)
	return r.list(ctx, query, args...)
}

func (r *ServiceRequestRepository) ListServiceRequestsByUser(ctx context.Context, userID int64, search, cat string, page, limit int) ([]*models.ServiceRequest, error) {
	query, args := r.buildListQuery(search, cat, userID, page, limit)
	return r.list(ctx, query, args...)
}

func (r *ServiceRequestRepository) buildListQuery(search, cat string, userID int64, page, limit int) (string, []any) {
	query := "SELECT " + srFields + " FROM service_requests sr LEFT JOIN users u ON sr.user_id = u.id WHERE 1=1"
	args := []any{}
	if userID > 0 {
		query += " AND sr.user_id = $1"; args = append(args, userID)
	}
	if search != "" {
		query += fmt.Sprintf(" AND (CAST(sr.id AS TEXT) ILIKE $%d OR sr.service_title ILIKE $%d OR sr.category ILIKE $%d OR u.full_name ILIKE $%d)", len(args)+1, len(args)+1, len(args)+1, len(args)+1)
		args = append(args, "%"+search+"%")
	}
	if cat != "" {
		query += fmt.Sprintf(" AND sr.category = $%d", len(args)+1); args = append(args, cat)
	}
	query += fmt.Sprintf(" ORDER BY sr.id DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, (page-1)*limit)
	return query, args
}

func (r *ServiceRequestRepository) list(ctx context.Context, query string, args ...any) ([]*models.ServiceRequest, error) {
	rows, _ := r.db.Query(ctx, query, args...)
	defer rows.Close()
	var list []*models.ServiceRequest
	for rows.Next() {
		sr := &models.ServiceRequest{}
		if err := r.scan(rows, sr); err == nil {
			list = append(list, sr)
		}
	}
	if list == nil { list = []*models.ServiceRequest{} }
	return list, nil
}

func (r *ServiceRequestRepository) CountServiceRequestsByUser(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM service_requests WHERE user_id = $1", userID).Scan(&n)
	return n, err
}

func (r *ServiceRequestRepository) UpdateServiceRequestStatus(ctx context.Context, id int64, status string) (*models.ServiceRequest, error) {
	query := "UPDATE service_requests SET status = $1, updated_at = NOW() WHERE id = $2 RETURNING " + srFields
	sr := &models.ServiceRequest{}
	if err := r.scan(r.db.QueryRow(ctx, query, status, id), sr); err != nil {
		return nil, err
	}
	return sr, nil
}

func (r *ServiceRequestRepository) DeleteServiceRequest(ctx context.Context, id int64) error {
	res, err := r.db.Exec(ctx, "DELETE FROM service_requests WHERE id = $1", id)
	if err != nil || res.RowsAffected() == 0 {
		return fmt.Errorf("not found or error")
	}
	return nil
}

func (r *ServiceRequestRepository) GetServiceStatusStats(ctx context.Context, serviceID int64) ([]models.StatusStat, error) {
	rows, _ := r.db.Query(ctx, "SELECT status, COUNT(*) FROM service_requests WHERE service_id = $1 GROUP BY status", serviceID)
	defer rows.Close()
	var stats []models.StatusStat
	for rows.Next() {
		var s models.StatusStat
		rows.Scan(&s.Status, &s.Total)
		stats = append(stats, s)
	}
	return stats, nil
}

func (r *ServiceRequestRepository) GetAverageServiceTime(ctx context.Context, serviceID int64) (int, error) {
	var avg int
	r.db.QueryRow(ctx, "SELECT COALESCE(ROUND(EXTRACT(EPOCH FROM AVG(updated_at - created_at)) / 86400)::int, 0) FROM service_requests WHERE service_id = $1 AND status = 'completed'", serviceID).Scan(&avg)
	return avg, nil
}

func (r *ServiceRequestRepository) ListServiceRequestDetailsByService(ctx context.Context, serviceID int64, page, limit int) ([]*models.ServiceRequestDetailResponse, error) {
	query := "SELECT " + srFields + ", u.username, u.email, u.cpf, u.birth_date, u.type, u.created_at, u.updated_at FROM service_requests sr LEFT JOIN users u ON sr.user_id = u.id WHERE sr.service_id = $1 ORDER BY sr.created_at DESC LIMIT $2 OFFSET $3"
	rows, _ := r.db.Query(ctx, query, serviceID, limit, (page-1)*limit)
	defer rows.Close()
	var list []*models.ServiceRequestDetailResponse
	for rows.Next() {
		sr := &models.ServiceRequest{}
		user := &models.User{}
		if err := rows.Scan(&sr.ID, &sr.UserID, &sr.UserName, &sr.ServiceID, &sr.ProtocolNumber, &sr.ServiceTitle, &sr.Category, &sr.RequestData, &sr.Attachments, &sr.Status, &sr.CreatedAt, &sr.UpdatedAt, &user.Username, &user.Email, &user.CPF, &user.BirthDate, &user.Type, &user.CreatedAt, &user.UpdatedAt); err == nil {
			if sr.UserID != nil {
				user.ID = *sr.UserID
				user.FullName = &sr.UserName
				list = append(list, &models.ServiceRequestDetailResponse{ServiceRequest: sr, CreatedBy: user})
			} else {
				list = append(list, &models.ServiceRequestDetailResponse{ServiceRequest: sr})
			}
		}
	}
	return list, nil
}

func (r *ServiceRequestRepository) CountServiceRequestsByStatusByUser(ctx context.Context, userID int64) (map[string]int, error) {
	counts := map[string]int{"pending": 0, "in_progress": 0, "completed": 0, "cancelled": 0}
	rows, _ := r.db.Query(ctx, "SELECT status, COUNT(*) FROM service_requests WHERE user_id = $1 GROUP BY status", userID)
	defer rows.Close()
	for rows.Next() {
		var s string; var c int
		if err := rows.Scan(&s, &c); err == nil { counts[s] = c }
	}
	return counts, nil
}

func (r *ServiceRequestRepository) GetHomeStats(ctx context.Context, isAdmin bool, userID int64, catFilter string) (*models.HomeResponse, error) {
	// The "Holy Grail" Query: Single round-trip dashboard aggregator
	query := `
	WITH base AS (
		SELECT sr.id, sr.status, sr.category, sr.service_title, sr.request_data, sr.created_at, sr.updated_at, u.full_name as user_name
		FROM service_requests sr
		LEFT JOIN users u ON sr.user_id = u.id
		WHERE ($1::BOOLEAN OR sr.user_id = $2) AND ($3 = '' OR sr.category = $3)
	),
	counts AS (
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'pending') as pending,
			COUNT(*) FILTER (WHERE status = 'in_progress') as in_progress,
			COUNT(*) FILTER (WHERE status = 'completed') as completed,
			COUNT(*) FILTER (WHERE status = 'cancelled') as cancelled,
			COUNT(*) FILTER (WHERE status = 'urgent') as urgent,
			COUNT(*) FILTER (WHERE status IN ('pending', 'in_progress', 'urgent')) as unresolved,
			COUNT(*) FILTER (WHERE status = 'completed' AND updated_at::date = CURRENT_DATE) as completed_today,
			COUNT(*) FILTER (WHERE created_at::date = CURRENT_DATE) as created_today,
			COALESCE(AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 86400) FILTER (WHERE status = 'completed'), 0) as avg_time
		FROM base
	),
	cat_stats AS (
		SELECT json_agg(j) FROM (
			SELECT category, COUNT(*) as count, ROUND(COUNT(*)::numeric / NULLIF((SELECT total FROM counts), 0) * 100) as percent
			FROM base GROUP BY category ORDER BY count DESC
		) j
	),
	recent AS (
		SELECT json_agg(j) FROM (SELECT id, user_name as name, service_title as service, request_data, status, created_at FROM base ORDER BY created_at DESC LIMIT 10) j
	),
	delayed AS (
		SELECT json_agg(j) FROM (SELECT id, user_name as name, service_title as service, request_data, status, created_at FROM base WHERE status IN ('pending', 'in_progress') AND created_at < NOW() - INTERVAL '3 days' ORDER BY created_at ASC LIMIT 10) j
	),
	new_reqs AS (
		SELECT json_agg(j) FROM (SELECT id, user_name as name, service_title as service, request_data, status, created_at FROM base WHERE created_at >= NOW() - INTERVAL '24 hours' ORDER BY created_at DESC LIMIT 10) j
	),
	volume AS (
		SELECT json_agg(j) FROM (SELECT TO_CHAR(created_at, 'YYYY-MM-DD') as day, COUNT(*) as count FROM base WHERE created_at >= CURRENT_DATE - INTERVAL '7 days' GROUP BY day ORDER BY day ASC) j
	),
	popular AS (
		SELECT json_agg(j) FROM (
			SELECT s.id, s.title, s.category, COUNT(sr.id) as request_count
			FROM services s INNER JOIN service_requests sr ON s.id = sr.service_id
			WHERE sr.status != 'cancelled' GROUP BY s.id, s.title, s.category ORDER BY request_count DESC LIMIT 5
		) j
	),
	alerts AS (
		SELECT json_agg(j) FROM (
			SELECT 'danger' as type, COUNT(*)::text || ' solicitações paradas há mais de 3 dias' as message FROM base WHERE status IN ('pending', 'in_progress') AND created_at < NOW() - INTERVAL '3 days' HAVING COUNT(*) > 0
			UNION ALL
			SELECT 'warning' as type, 'Serviço crítico: ' || service_title || ' com ' || COUNT(*)::text || ' pendências' as message FROM base WHERE status IN ('pending', 'urgent') GROUP BY service_title HAVING COUNT(*) > 5 ORDER BY COUNT(*) DESC LIMIT 1
		) j
	)
	SELECT 
		(SELECT total FROM counts), (SELECT pending FROM counts), (SELECT in_progress FROM counts), (SELECT completed FROM counts),
		(SELECT cancelled FROM counts), (SELECT urgent FROM counts), (SELECT unresolved FROM counts), (SELECT completed_today FROM counts),
		(SELECT created_today FROM counts), (SELECT avg_time FROM counts),
		(SELECT COUNT(*) FROM users), (SELECT COUNT(*) FROM services WHERE is_active = TRUE),
		COALESCE((SELECT * FROM cat_stats), '[]'), COALESCE((SELECT * FROM recent), '[]'), COALESCE((SELECT * FROM delayed), '[]'),
		COALESCE((SELECT * FROM new_reqs), '[]'), COALESCE((SELECT * FROM volume), '[]'), COALESCE((SELECT * FROM popular), '[]'),
		COALESCE((SELECT * FROM alerts), '[]')`

	var total, p, ip, c, can, u, ur, ct, cr, users, svcs int
	var avg float64
	var cats, rec, del, newR, vol, pop, alrt []byte

	err := r.db.QueryRow(ctx, query, isAdmin, userID, catFilter).Scan(
		&total, &p, &ip, &c, &can, &u, &ur, &ct, &cr, &avg, &users, &svcs,
		&cats, &rec, &del, &newR, &vol, &pop, &alrt,
	)
	if err != nil { return nil, err }

	pct := func(v int) int { if total > 0 { return int(float64(v)/float64(total)*100) }; return 0 }
	
	resp := &models.HomeResponse{
		Stats: models.HomeStats{
			TotalRequests: models.StatDetail{Total: total, Percent: 100},
			PendingRequests: models.StatDetail{Total: p, Percent: pct(p)},
			InProgressRequests: models.StatDetail{Total: ip, Percent: pct(ip)},
			CompletedRequests: models.StatDetail{Total: c, Percent: pct(c)},
			CancelledRequests: models.StatDetail{Total: can, Percent: pct(can)},
			UrgentRequests: models.StatDetail{Total: u, Percent: pct(u)},
			UnresolvedRequests: models.StatDetail{Total: ur, Percent: pct(ur)},
			TotalUsers: users, TotalActiveServices: svcs, CompletedToday: ct, CreatedToday: cr, AverageTime: avg,
		},
	}
	
	json.Unmarshal(cats, &resp.Categories)
	json.Unmarshal(pop, &resp.PopularServices)
	json.Unmarshal(alrt, &resp.Alerts)
	json.Unmarshal(vol, &resp.Volume7d)
	
	mapRecent := func(data []byte) []models.RecentRequest {
		var raw []struct {
			ID int64 `json:"id"`; Name *string `json:"name"`; Service string `json:"service"`
			Data json.RawMessage `json:"request_data"`; Status string `json:"status"`; CreatedAt time.Time `json:"created_at"`
		}
		json.Unmarshal(data, &raw)
		res := make([]models.RecentRequest, len(raw))
		for i, x := range raw {
			res[i] = models.RecentRequest{ID: x.ID, Name: x.Name, Service: x.Service, Status: x.Status, Date: x.CreatedAt.Format("2006-01-02")}
			var m map[string]any; json.Unmarshal(x.Data, &m)
			if addr, ok := m["address"].(string); ok { res[i].Address = &addr } else if end, ok := m["endereco"].(string); ok { res[i].Address = &end }
		}
		return res
	}
	resp.RecentRequests = mapRecent(rec)
	resp.DelayedRequests = mapRecent(del)
	resp.NewRequests = mapRecent(newR)
	
	return resp, nil
}
