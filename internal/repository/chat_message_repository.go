package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type ChatMessageRepository struct {
	db *pgxpool.Pool
}

func NewChatMessageRepository(db *pgxpool.Pool) *ChatMessageRepository {
	return &ChatMessageRepository{db: db}
}

func (r *ChatMessageRepository) Create(ctx context.Context, senderID int64, senderName string, req *models.CreateChatMessageRequest) (*models.ChatMessage, error) {
	query := `
		INSERT INTO chat_messages (service_request_id, sender_id, sender_name, content, attachments, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, service_request_id, sender_id, sender_name, content, attachments, created_at, updated_at`

	msg := &models.ChatMessage{}
	err := r.db.QueryRow(ctx, query,
		req.ServiceRequestID, senderID, senderName, req.Content, req.Attachments,
	).Scan(
		&msg.ID, &msg.ServiceRequestID, &msg.SenderID, &msg.SenderName,
		&msg.Content, &msg.Attachments, &msg.CreatedAt, &msg.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create chat message: %w", err)
	}

	return msg, nil
}

func (r *ChatMessageRepository) ListByRequestID(ctx context.Context, requestID int64) ([]*models.ChatMessage, error) {
	query := `
		SELECT id, service_request_id, sender_id, sender_name, content, attachments, created_at, updated_at
		FROM chat_messages
		WHERE service_request_id = $1
		ORDER BY created_at ASC`

	rows, err := r.db.Query(ctx, query, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*models.ChatMessage
	for rows.Next() {
		msg := &models.ChatMessage{}
		if err := rows.Scan(
			&msg.ID, &msg.ServiceRequestID, &msg.SenderID, &msg.SenderName,
			&msg.Content, &msg.Attachments, &msg.CreatedAt, &msg.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, msg)
	}
	if list == nil {
		list = []*models.ChatMessage{}
	}
	return list, nil
}
