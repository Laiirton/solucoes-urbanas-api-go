package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

func (r *ServiceRequestRepository) GetHomeStats(ctx context.Context, isAdmin bool, userID int64, regionFilter, teamFilter *int64, startDate, endDate *string) (*models.HomeResponse, error) {
	baseWhere, args := buildBaseWhere(isAdmin, userID, regionFilter, teamFilter, startDate, endDate)

	statusCounts, total := r.computeStatusStats(ctx, baseWhere, args)
	pct := calcPercentFunc(total)

	stats, err := r.computeAggregatedMetrics(ctx, isAdmin, baseWhere, args)
	if err != nil {
		return nil, err
	}
	stats.TotalRequests = models.StatDetail{Total: total, Percent: 100}
	stats.PendingRequests = models.StatDetail{Total: statusCounts["pending"], Percent: pct(statusCounts["pending"], total)}
	stats.InProgressRequests = models.StatDetail{Total: statusCounts["in_progress"], Percent: pct(statusCounts["in_progress"], total)}
	stats.CompletedRequests = models.StatDetail{Total: statusCounts["completed"], Percent: pct(statusCounts["completed"], total)}
	stats.CancelledRequests = models.StatDetail{Total: statusCounts["cancelled"], Percent: pct(statusCounts["cancelled"], total)}
	// Note: "urgent" is not a valid status in the DB enum. Kept for model compatibility.
	stats.UrgentRequests = models.StatDetail{Total: 0, Percent: 0}
	unresolved := statusCounts["pending"] + statusCounts["in_progress"]
	stats.UnresolvedRequests = models.StatDetail{Total: unresolved, Percent: pct(unresolved, total)}

	popularServices := r.computePopularServices(ctx, baseWhere, args)
	topRatedServices := r.computeTopRatedServices(ctx, baseWhere, args)
	alerts := r.computeAlerts(ctx, baseWhere, args)
	categories := r.computeCategories(ctx, baseWhere, args, total)
	recent, delayed, newReqs := r.computeRecentRequests(ctx, baseWhere, args)
	volume7d := r.computeVolume7d(ctx, baseWhere, args)

	return &models.HomeResponse{
		Stats:            *stats,
		Categories:       categories,
		RecentRequests:   recent,
		DelayedRequests:  delayed,
		NewRequests:      newReqs,
		Volume7d:         volume7d,
		Alerts:           alerts,
		PopularServices:  popularServices,
		TopRatedServices: topRatedServices,
	}, nil
}

func buildBaseWhere(isAdmin bool, userID int64, regionFilter, teamFilter *int64, startDate, endDate *string) (string, []interface{}) {
	baseWhere := ""
	var args []interface{}

	if !isAdmin {
		baseWhere = "WHERE sr.user_id = $1"
		args = append(args, userID)
	} else if teamFilter != nil {
		baseWhere = "WHERE sr.team_id = $1"
		args = append(args, *teamFilter)
	} else if regionFilter != nil {
		baseWhere = "WHERE sr.region_id = $1"
		args = append(args, *regionFilter)
	}

	if startDate != nil && *startDate != "" {
		if baseWhere != "" {
			baseWhere += fmt.Sprintf(" AND sr.created_at >= $%d", len(args)+1)
		} else {
			baseWhere = fmt.Sprintf("WHERE sr.created_at >= $%d", len(args)+1)
		}
		args = append(args, *startDate)
	}
	if endDate != nil && *endDate != "" {
		if baseWhere != "" {
			baseWhere += fmt.Sprintf(" AND sr.created_at <= $%d::date + interval '1 day'", len(args)+1)
		} else {
			baseWhere = fmt.Sprintf("WHERE sr.created_at <= $%d::date + interval '1 day'", len(args)+1)
		}
		args = append(args, *endDate)
	}

	return baseWhere, args
}

func calcPercentFunc(total int) func(int, int) int {
	return func(val, tot int) int {
		if tot > 0 {
			return int((float64(val) / float64(tot)) * 100)
		}
		return 0
	}
}

func (r *ServiceRequestRepository) computeStatusStats(ctx context.Context, baseWhere string, args []interface{}) (map[string]int, int) {
	counts := make(map[string]int)
	total := 0

	statsQuery := fmt.Sprintf(`
		SELECT status, COUNT(*)
		FROM service_requests sr
		%s
		GROUP BY status`, baseWhere)

	rows, err := r.db.Query(ctx, statsQuery, args...)
	if err != nil {
		return counts, 0
	}
	defer rows.Close()

	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			continue
		}
		counts[status] = count
		total += count
	}

	return counts, total
}

func (r *ServiceRequestRepository) computeAggregatedMetrics(ctx context.Context, isAdmin bool, baseWhere string, args []interface{}) (*models.HomeStats, error) {
	var totalUsers int
	if isAdmin {
		r.db.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&totalUsers)
	}

	var activeServices int
	if isAdmin {
		r.db.QueryRow(ctx, "SELECT COUNT(*) FROM services WHERE is_active = TRUE").Scan(&activeServices)
	}

	// Consolidated query: 3 metrics in 1 query
	consolidatedQuery := fmt.Sprintf(`
		SELECT
			COUNT(*) FILTER (WHERE status = 'completed' AND updated_at::date = CURRENT_DATE) AS completed_today,
			COUNT(*) FILTER (WHERE created_at::date = CURRENT_DATE) AS created_today,
			COALESCE(AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 86400) FILTER (WHERE status = 'completed'), 0) AS avg_time
		FROM service_requests sr
		%s`,
		baseWhere)

	var completedToday int
	var createdToday int
	var avgTime float64
	err := r.db.QueryRow(ctx, consolidatedQuery, args...).Scan(&completedToday, &createdToday, &avgTime)
	if err != nil {
		// Fallback to individual queries
		completedToday = 0
		createdToday = 0
		avgTime = 0
	}

	return &models.HomeStats{
		TotalUsers:          totalUsers,
		TotalActiveServices: activeServices,
		CompletedToday:      completedToday,
		CreatedToday:        createdToday,
		AverageTime:         avgTime,
	}, nil
}

func (r *ServiceRequestRepository) computePopularServices(ctx context.Context, baseWhere string, args []interface{}) []models.PopularService {
	var result []models.PopularService
	var qArgs []interface{}
	filter := ""
	if baseWhere != "" {
		cond := strings.TrimPrefix(baseWhere, "WHERE ")
		filter = "AND " + cond
		qArgs = append(qArgs, args...)
	}
	query := fmt.Sprintf(`
		SELECT s.id, s.title, s.category, COUNT(sr.id) as request_count
		FROM services s
		INNER JOIN service_requests sr ON s.id = sr.service_id
		WHERE 1=1 %s AND sr.status != 'cancelled'
		GROUP BY s.id, s.title, s.category
		ORDER BY request_count DESC
		LIMIT 5`, filter)

	rows, err := r.db.Query(ctx, query, qArgs...)
	if err != nil {
		fmt.Printf("Warning: failed to fetch popular services: %v\n", err)
		return []models.PopularService{}
	}
	defer rows.Close()

	for rows.Next() {
		svc := models.PopularService{}
		if err := rows.Scan(&svc.ID, &svc.Title, &svc.Category, &svc.RequestCount); err != nil {
			fmt.Printf("Warning: failed to scan popular service: %v\n", err)
			continue
		}
		result = append(result, svc)
	}
	if result == nil {
		result = []models.PopularService{}
	}
	return result
}

func (r *ServiceRequestRepository) computeTopRatedServices(ctx context.Context, baseWhere string, args []interface{}) []models.TopRatedService {
	var result []models.TopRatedService
	var qArgs []interface{}
	filter := ""
	if baseWhere != "" {
		cond := strings.TrimPrefix(baseWhere, "WHERE ")
		filter = "AND " + cond
		qArgs = append(qArgs, args...)
	}
	query := fmt.Sprintf(`
		SELECT s.id, s.title, s.category, AVG(r.stars) as avg_stars, COUNT(r.id) as rating_count
		FROM services s
		INNER JOIN service_ratings r ON s.id = r.service_id
		INNER JOIN service_requests sr ON r.service_request_id = sr.id
		WHERE 1=1 %s
		GROUP BY s.id, s.title, s.category
		ORDER BY avg_stars DESC, rating_count DESC
		LIMIT 5`, filter)

	rows, err := r.db.Query(ctx, query, qArgs...)
	if err != nil {
		fmt.Printf("Warning: failed to fetch top rated services: %v\n", err)
		return []models.TopRatedService{}
	}
	defer rows.Close()

	for rows.Next() {
		svc := models.TopRatedService{}
		if err := rows.Scan(&svc.ID, &svc.Title, &svc.Category, &svc.AverageStars, &svc.RatingCount); err != nil {
			fmt.Printf("Warning: failed to scan top rated service: %v\n", err)
			continue
		}
		result = append(result, svc)
	}
	if result == nil {
		result = []models.TopRatedService{}
	}
	return result
}

func (r *ServiceRequestRepository) computeAlerts(ctx context.Context, baseWhere string, args []interface{}) []models.HomeAlert {
	alerts := []models.HomeAlert{}

	// 1. Stagnant requests (> 3 days)
	var stagnantCount int
	var stagnantArgs []interface{}
	stagnantPrefix := map[bool]string{true: baseWhere + " AND ", false: "WHERE "}[baseWhere != ""]
	if baseWhere != "" {
		stagnantArgs = append(stagnantArgs, args...)
	}
	stagnantQuery := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM service_requests sr
		%sstatus IN ('pending', 'in_progress') AND created_at < NOW() - INTERVAL '3 days'
	`, stagnantPrefix)
	r.db.QueryRow(ctx, stagnantQuery, stagnantArgs...).Scan(&stagnantCount)

	if stagnantCount > 0 {
		alerts = append(alerts, models.HomeAlert{
			Type:    "danger",
			Message: fmt.Sprintf("%d solicitações paradas há mais de 3 dias", stagnantCount),
		})
	}

	// 2. Most critical service (Only if > 5 pending)
	var criticalService string
	var criticalCount int
	var criticalArgs []interface{}
	criticalPrefix := map[bool]string{true: baseWhere + " AND ", false: "WHERE "}[baseWhere != ""]
	if baseWhere != "" {
		criticalArgs = append(criticalArgs, args...)
	}
	criticalQuery := fmt.Sprintf(`
		SELECT service_title, COUNT(*)
		FROM service_requests sr
		%sstatus = 'pending'
		GROUP BY service_title
		ORDER BY COUNT(*) DESC
		LIMIT 1
	`, criticalPrefix)
	r.db.QueryRow(ctx, criticalQuery, criticalArgs...).Scan(&criticalService, &criticalCount)

	if criticalService != "" && criticalCount > 5 {
		alerts = append(alerts, models.HomeAlert{
			Type:    "warning",
			Message: fmt.Sprintf("Serviço mais crítico: %s com %d pendências", criticalService, criticalCount),
		})
	}

	return alerts
}

func (r *ServiceRequestRepository) computeCategories(ctx context.Context, baseWhere string, args []interface{}, total int) []models.CategoryStat {
	pct := calcPercentFunc(total)

	catQuery := fmt.Sprintf(`
		SELECT category, COUNT(*)
		FROM service_requests sr
		%s
		GROUP BY category
		ORDER BY COUNT(*) DESC`, baseWhere)

	catRows, err := r.db.Query(ctx, catQuery, args...)
	if err != nil {
		return []models.CategoryStat{}
	}
	defer catRows.Close()

	var categories []models.CategoryStat
	for catRows.Next() {
		var cat string
		var count int
		if err := catRows.Scan(&cat, &count); err != nil {
			continue
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
	return categories
}

func (r *ServiceRequestRepository) computeRecentRequests(ctx context.Context, baseWhere string, args []interface{}) (recent, delayed, newReqs []models.RecentRequest) {
	fetchRecent := func(query string, args ...interface{}) []models.RecentRequest {
		rows, err := r.db.Query(ctx, query, args...)
		if err != nil {
			return []models.RecentRequest{}
		}
		defer rows.Close()

		var list []models.RecentRequest
		for rows.Next() {
			var req models.RecentRequest
			var rawData []byte
			var createdAt time.Time
			if err := rows.Scan(&req.ID, &req.Name, &req.Service, &rawData, &req.Status, &createdAt); err != nil {
				continue
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
		return list
	}

	recentQuery := fmt.Sprintf(`
		SELECT sr.id, u.full_name, sr.service_title, sr.request_data, sr.status, sr.created_at
		FROM service_requests sr
		LEFT JOIN users u ON sr.user_id = u.id
		%s
		ORDER BY sr.created_at DESC
		LIMIT 10`, baseWhere)
	recent = fetchRecent(recentQuery, args...)

	delayedQuery := fmt.Sprintf(`
		SELECT sr.id, u.full_name, sr.service_title, sr.request_data, sr.status, sr.created_at
		FROM service_requests sr
		LEFT JOIN users u ON sr.user_id = u.id
		%s
		%s sr.status IN ('pending', 'in_progress') AND sr.created_at < NOW() - INTERVAL '3 days'
		ORDER BY sr.created_at ASC
		LIMIT 10`, baseWhere, map[bool]string{true: "AND", false: "WHERE"}[baseWhere != ""])
	delayed = fetchRecent(delayedQuery, args...)

	newQuery := fmt.Sprintf(`
		SELECT sr.id, u.full_name, sr.service_title, sr.request_data, sr.status, sr.created_at
		FROM service_requests sr
		LEFT JOIN users u ON sr.user_id = u.id
		%s
		%s sr.created_at >= NOW() - INTERVAL '24 hours'
		ORDER BY sr.created_at DESC
		LIMIT 10`, baseWhere, map[bool]string{true: "AND", false: "WHERE"}[baseWhere != ""])
	newReqs = fetchRecent(newQuery, args...)

	return
}

func (r *ServiceRequestRepository) computeVolume7d(ctx context.Context, baseWhere string, args []interface{}) []models.VolumeStat {
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
	return volume7d
}

func (r *ServiceRequestRepository) ListMapLocations(ctx context.Context, regionFilter, teamFilter *int64, startDate, endDate *string, limit int) ([]models.MapLocation, error) {
	query := `SELECT sr.id, sr.latitude, sr.longitude, sr.geocoded_address, sr.service_title, sr.status, COALESCE(sr.service_id, 0)
	          FROM service_requests sr
	          LEFT JOIN services s ON sr.service_id = s.id`

	var args []interface{}
	whereApplied := false

	if regionFilter != nil {
		query += ` WHERE sr.region_id = $1`
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
	}

	query += ` ORDER BY sr.id DESC`
	if limit > 0 {
		query += fmt.Sprintf(` LIMIT $%d`, len(args)+1)
		args = append(args, limit)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list map locations: %w", err)
	}
	defer rows.Close()

	var locations []models.MapLocation
	for rows.Next() {
		var loc models.MapLocation
		var lat, lon *float64
		var geoAddr *string
		var serviceID int64
		if err := rows.Scan(&loc.ID, &lat, &lon, &geoAddr, &loc.ServiceTitle, &loc.Status, &serviceID); err != nil {
			continue
		}
		// Map service_id to icon using the predefined mapping
		loc.Icon = models.GetServiceIcon(serviceID)
		if lat != nil && lon != nil {
			loc.Latitude = *lat
			loc.Longitude = *lon
			loc.Found = true
			if geoAddr != nil {
				loc.Address = *geoAddr
			}
		}
		locations = append(locations, loc)
	}
	if locations == nil {
		locations = []models.MapLocation{}
	}
	return locations, nil
}
