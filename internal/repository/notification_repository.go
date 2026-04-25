package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PushTokenRepository struct {
	db *pgxpool.Pool
}

func NewPushTokenRepository(db *pgxpool.Pool) *PushTokenRepository {
	return &PushTokenRepository{db: db}
}

func (r *PushTokenRepository) UpsertPushToken(ctx context.Context, userID int64, token string) error {
	query := `
		INSERT INTO user_push_tokens (user_id, token, created_at, updated_at)
		VALUES ($1, $2, NOW(), NOW())
		ON CONFLICT (token)
		DO UPDATE SET user_id = EXCLUDED.user_id, updated_at = NOW()
	`
	if _, err := r.db.Exec(ctx, query, userID, token); err != nil {
		return fmt.Errorf("failed to save push token: %w", err)
	}
	return nil
}

func (r *PushTokenRepository) DeletePushToken(ctx context.Context, userID int64, token string) error {
	query := `DELETE FROM user_push_tokens WHERE user_id = $1 AND token = $2`
	result, err := r.db.Exec(ctx, query, userID, token)
	if err != nil {
		return fmt.Errorf("failed to delete push token: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("push token not found")
	}
	return nil
}

func (r *PushTokenRepository) ListTokens(ctx context.Context) ([]string, error) {
	query := `SELECT token FROM user_push_tokens ORDER BY updated_at DESC, id DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list push tokens: %w", err)
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, fmt.Errorf("failed to scan push token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}

func (r *PushTokenRepository) ListTokensByUser(ctx context.Context, userID int64) ([]string, error) {
	query := `SELECT token FROM user_push_tokens WHERE user_id = $1 ORDER BY updated_at DESC, id DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list push tokens for user: %w", err)
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, fmt.Errorf("failed to scan push token: %w", err)
		}
		tokens = append(tokens, token)
	}

	return tokens, nil
}
