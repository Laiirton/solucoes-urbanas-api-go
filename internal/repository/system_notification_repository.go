package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type SystemNotificationRepository struct {
	db *pgxpool.Pool
}

func NewSystemNotificationRepository(db *pgxpool.Pool) *SystemNotificationRepository {
	return &SystemNotificationRepository{db: db}
}

func (r *SystemNotificationRepository) Create(ctx context.Context, n *models.SystemNotification) (*models.SystemNotification, error) {
	query := `
		INSERT INTO system_notifications (user_id, title, body, type, data)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, title, body, type, data, read_at, created_at
	`
	err := r.db.QueryRow(ctx, query, n.UserID, n.Title, n.Body, n.Type, n.Data).
		Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.Data, &n.ReadAt, &n.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create system notification: %w", err)
	}
	return n, nil
}

func (r *SystemNotificationRepository) List(ctx context.Context, userID *int64, notificationType string, unreadOnly bool, page, limit int) ([]*models.SystemNotification, error) {
	offset := (page - 1) * limit
	query := `SELECT id, user_id, title, body, type, data, read_at, created_at FROM system_notifications`

	var args []interface{}
	whereApplied := false

	if userID != nil {
		query += ` WHERE (user_id = $1 OR user_id IS NULL)`
		args = append(args, *userID)
		whereApplied = true
	}

	if notificationType != "" {
		if whereApplied {
			query += fmt.Sprintf(` AND type = $%d`, len(args)+1)
		} else {
			query += ` WHERE type = $1`
		}
		args = append(args, notificationType)
		whereApplied = true
	}

	if unreadOnly {
		if whereApplied {
			query += ` AND read_at IS NULL`
		} else {
			query += ` WHERE read_at IS NULL`
			whereApplied = true
		}
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list system notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*models.SystemNotification
	for rows.Next() {
		var n models.SystemNotification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.Data, &n.ReadAt, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan system notification: %w", err)
		}
		notifications = append(notifications, &n)
	}

	return notifications, nil
}

func (r *SystemNotificationRepository) GetByID(ctx context.Context, id int64) (*models.SystemNotification, error) {
	query := `SELECT id, user_id, title, body, type, data, read_at, created_at FROM system_notifications WHERE id = $1`
	var n models.SystemNotification
	err := r.db.QueryRow(ctx, query, id).
		Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.Data, &n.ReadAt, &n.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("system notification not found: %w", err)
	}
	return &n, nil
}

func (r *SystemNotificationRepository) Update(ctx context.Context, id int64, req *models.UpdateSystemNotificationRequest) (*models.SystemNotification, error) {
	query := `
		UPDATE system_notifications SET
			title = COALESCE($1, title),
			body = COALESCE($2, body),
			type = COALESCE($3, type),
			data = COALESCE($4, data),
			read_at = COALESCE($5, read_at)
		WHERE id = $6
		RETURNING id, user_id, title, body, type, data, read_at, created_at
	`

	var n models.SystemNotification
	err := r.db.QueryRow(ctx, query,
		nullableValue(req.Title),
		nullableValue(req.Body),
		nullableValue(req.Type),
		nullableValue(req.Data),
		nullableValue(req.ReadAt),
		id,
	).Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.Data, &n.ReadAt, &n.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update system notification: %w", err)
	}
	return &n, nil
}

func (r *SystemNotificationRepository) MarkAsRead(ctx context.Context, id int64) (*models.SystemNotification, error) {
	query := `
		UPDATE system_notifications SET read_at = NOW()
		WHERE id = $1
		RETURNING id, user_id, title, body, type, data, read_at, created_at
	`
	var n models.SystemNotification
	err := r.db.QueryRow(ctx, query, id).
		Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.Data, &n.ReadAt, &n.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to mark system notification as read: %w", err)
	}
	return &n, nil
}

func (r *SystemNotificationRepository) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM system_notifications WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete system notification: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("system notification not found")
	}
	return nil
}
