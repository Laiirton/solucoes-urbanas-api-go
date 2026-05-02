package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type ServiceAttendanceRepository struct {
	db *pgxpool.Pool
}

func NewServiceAttendanceRepository(db *pgxpool.Pool) *ServiceAttendanceRepository {
	return &ServiceAttendanceRepository{db: db}
}

func (r *ServiceAttendanceRepository) Create(ctx context.Context, attendantID int64, req *models.CreateServiceAttendanceRequest) (*models.ServiceAttendance, error) {
	// Start a transaction if we need to update the status too
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO service_attendances (service_request_id, attended_by, notes, attachments, created_at, updated_at)
		VALUES ($1, $2, $3, $4, NOW(), NOW())
		RETURNING id, service_request_id, attended_by, notes, attachments, created_at, updated_at`

	attendance := &models.ServiceAttendance{}
	err = tx.QueryRow(ctx, query,
		req.ServiceRequestID, attendantID, req.Notes, req.Attachments,
	).Scan(
		&attendance.ID, &attendance.ServiceRequestID, &attendance.AttendedBy,
		&attendance.Notes, &attendance.Attachments, &attendance.CreatedAt, &attendance.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create service attendance: %w", err)
	}

	// Update status of the service request if provided
	if req.NewStatus != "" {
		_, err = tx.Exec(ctx, `UPDATE service_requests SET status = $1, updated_at = NOW() WHERE id = $2`, req.NewStatus, req.ServiceRequestID)
		if err != nil {
			return nil, fmt.Errorf("failed to update service request status: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	// Fetch attendant name
	tx.QueryRow(ctx, `SELECT full_name FROM users WHERE id = $1`, attendantID).Scan(&attendance.AttendantName)

	return attendance, nil
}

func (r *ServiceAttendanceRepository) ListByRequestID(ctx context.Context, requestID int64) ([]*models.ServiceAttendance, error) {
	query := `
		SELECT sa.id, sa.service_request_id, sa.attended_by, COALESCE(u.full_name, ''), sa.notes, sa.attachments, sa.created_at, sa.updated_at
		FROM service_attendances sa
		LEFT JOIN users u ON sa.attended_by = u.id
		WHERE sa.service_request_id = $1
		ORDER BY sa.created_at DESC`

	rows, err := r.db.Query(ctx, query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.ServiceAttendance
	for rows.Next() {
		sa := &models.ServiceAttendance{}
		if err := rows.Scan(
			&sa.ID, &sa.ServiceRequestID, &sa.AttendedBy, &sa.AttendantName,
			&sa.Notes, &sa.Attachments, &sa.CreatedAt, &sa.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, sa)
	}
	if list == nil {
		list = []*models.ServiceAttendance{}
	}
	return list, nil
}

func (r *ServiceAttendanceRepository) GetByID(ctx context.Context, id int64) (*models.ServiceAttendance, error) {
	query := `
		SELECT sa.id, sa.service_request_id, sa.attended_by, COALESCE(u.full_name, ''), sa.notes, sa.attachments, sa.created_at, sa.updated_at
		FROM service_attendances sa
		LEFT JOIN users u ON sa.attended_by = u.id
		WHERE sa.id = $1`

	sa := &models.ServiceAttendance{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&sa.ID, &sa.ServiceRequestID, &sa.AttendedBy, &sa.AttendantName,
		&sa.Notes, &sa.Attachments, &sa.CreatedAt, &sa.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("service attendance not found: %w", err)
	}

	return sa, nil
}
