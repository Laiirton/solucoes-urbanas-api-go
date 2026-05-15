package repository

import (
	"context"
	"fmt"
	"strconv"

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
	query := `SELECT sn.id, sn.user_id, sn.title, sn.body, sn.type, sn.data, sn.created_at,
		CASE
			WHEN sn.user_id IS NOT NULL THEN sn.read_at
			ELSE nr.read_at
		END AS read_at
	FROM system_notifications sn
	LEFT JOIN notification_reads nr ON nr.notification_id = sn.id AND nr.user_id = $1`

	var args []interface{}
	args = append(args, userID) // $1 - pgx handles nil *int64 as SQL NULL
	whereApplied := false

	if userID != nil {
		query += ` WHERE (sn.user_id = $1 OR sn.user_id IS NULL)`
		whereApplied = true
	} else {
		query += ` WHERE sn.user_id IS NULL`
		whereApplied = true
	}

	if notificationType != "" {
		if whereApplied {
			query += fmt.Sprintf(` AND sn.type = $%d`, len(args)+1)
		} else {
			query += ` WHERE sn.type = $1`
		}
		args = append(args, notificationType)
		whereApplied = true
	}

	if unreadOnly {
		if whereApplied {
			query += ` AND (
				(sn.user_id IS NOT NULL AND sn.read_at IS NULL)
				OR
				(sn.user_id IS NULL AND nr.read_at IS NULL)
			)`
		} else {
			query += ` WHERE (
				(sn.user_id IS NOT NULL AND sn.read_at IS NULL)
				OR
				(sn.user_id IS NULL AND nr.read_at IS NULL)
			)`
			whereApplied = true
		}
	}

	// Exclude stale notifications (references to deleted content)
	staleFilter := ` AND NOT (
		(sn.type = 'news' AND sn.data IS NOT NULL AND sn.data->>'news_id' IS NOT NULL AND NOT EXISTS (
			SELECT 1 FROM news WHERE id::text = sn.data->>'news_id'
		))
		OR
		(sn.type = 'service_request' AND sn.data IS NOT NULL AND sn.data->>'service_request_id' IS NOT NULL AND NOT EXISTS (
			SELECT 1 FROM service_requests WHERE id::text = sn.data->>'service_request_id'
		))
	)`
	query += staleFilter

	query += ` ORDER BY sn.created_at DESC`
	if limit > 0 {
		offset := (page - 1) * limit
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list system notifications: %w", err)
	}
	defer rows.Close()

	var notifications []*models.SystemNotification
	for rows.Next() {
		var n models.SystemNotification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.Data, &n.CreatedAt, &n.ReadAt); err != nil {
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

func (r *SystemNotificationRepository) MarkAsRead(ctx context.Context, id int64, userID int64) (*models.SystemNotification, error) {
	// Check if notification is user-specific or broadcast
	n, err := r.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if n.UserID != nil {
		// User-specific: update read_at directly
		query := `
			UPDATE system_notifications SET read_at = NOW()
			WHERE id = $1
			RETURNING id, user_id, title, body, type, data, read_at, created_at
		`
		err = r.db.QueryRow(ctx, query, id).
			Scan(&n.ID, &n.UserID, &n.Title, &n.Body, &n.Type, &n.Data, &n.ReadAt, &n.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to mark system notification as read: %w", err)
		}
	} else {
		// Broadcast: upsert into notification_reads for this user
		_, err = r.db.Exec(ctx,
			`INSERT INTO notification_reads (notification_id, user_id)
			 VALUES ($1, $2)
			 ON CONFLICT (notification_id, user_id) DO UPDATE SET read_at = NOW()`,
			id, userID,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to mark broadcast notification as read: %w", err)
		}
		n.ReadAt = nil // will be populated by caller if needed
	}

	return n, nil
}

func (r *SystemNotificationRepository) MarkAllAsRead(ctx context.Context, userID int64) (int64, error) {
	// 1. Mark all user-specific notifications as read
	var userCount int64
	result, err := r.db.Exec(ctx,
		`UPDATE system_notifications SET read_at = NOW() WHERE user_id = $1 AND read_at IS NULL`,
		userID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to mark user notifications as read: %w", err)
	}
	userCount = result.RowsAffected()

	// 2. Mark all broadcast notifications as read (per-user in notification_reads)
	broadcastResult, err := r.db.Exec(ctx,
		`INSERT INTO notification_reads (notification_id, user_id)
		 SELECT sn.id, $1
		 FROM system_notifications sn
		 WHERE sn.user_id IS NULL
		 AND sn.id NOT IN (
			 SELECT notification_id FROM notification_reads WHERE user_id = $1
		 )`,
		userID,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to mark broadcast notifications as read: %w", err)
	}
	broadcastCount := broadcastResult.RowsAffected()

	return userCount + broadcastCount, nil
}

func (r *SystemNotificationRepository) DeleteByTypeAndRefID(ctx context.Context, notifType string, refID int64) error {
	dataKey := "news_id"
	switch notifType {
	case "news":
		dataKey = "news_id"
	case "service_request":
		dataKey = "service_request_id"
	default:
		return fmt.Errorf("unsupported notification type for cascade delete: %s", notifType)
	}

	query := fmt.Sprintf(
		`DELETE FROM system_notifications WHERE type = $1 AND data->>'%s' = $2`,
		dataKey,
	)
	_, err := r.db.Exec(ctx, query, notifType, strconv.FormatInt(refID, 10))
	if err != nil {
		return fmt.Errorf("failed to delete system notifications by ref: %w", err)
	}
	return nil
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
