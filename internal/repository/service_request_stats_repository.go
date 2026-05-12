package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

func NewServiceRequestStatsRepository(db *pgxpool.Pool) *ServiceRequestRepository {
	return &ServiceRequestRepository{db: db}
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

func (r *ServiceRequestRepository) ListServiceRequestDetailsByService(ctx context.Context, serviceID int64, page, limit int) ([]*models.ServiceRequestDetailResponse, error) {
	offset := (page - 1) * limit
	query := `SELECT sr.id, sr.user_id, COALESCE(u.full_name, ''), sr.service_id, sr.protocol_number,
	                 sr.service_title, sr.category, sr.request_data, sr.attachments, sr.status,
	                 sr.latitude, sr.longitude, sr.geocoded_address,
	                 sr.team_id, COALESCE(t.name, ''),
	                 sr.region_id, COALESCE(rg.name, ''),
	                 sr.created_at, sr.updated_at,
	                 u.username, u.email, u.cpf, u.phone, u.birth_date, u.type, u.created_at, u.updated_at
	          FROM service_requests sr
	          LEFT JOIN users u ON sr.user_id = u.id
	          LEFT JOIN teams t ON sr.team_id = t.id
	          LEFT JOIN regions rg ON sr.region_id = rg.id
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
		var bd *time.Time
		if err := rows.Scan(
			&sr.ID, &uID, &sr.UserName, &sr.ServiceID, &sr.ProtocolNumber,
			&sr.ServiceTitle, &sr.Category, &sr.RequestData,
			&sr.Attachments, &sr.Status, &sr.Latitude, &sr.Longitude, &sr.GeocodedAddress,
			&sr.TeamID, &sr.TeamName,
			&sr.RegionID, &sr.RegionName,
			&sr.CreatedAt, &sr.UpdatedAt,
			&user.Username, &user.Email, &user.CPF, &user.Phone, &bd, &user.Type, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, err
		}
		user.BirthDate = formatBirthDate(bd)
		sr.UserID = uID

		if sr.ServiceID != nil {
			sr.Icon = models.GetServiceIcon(*sr.ServiceID)
		}

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
