package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type NewsRepository struct {
	db *pgxpool.Pool
}

func NewNewsRepository(db *pgxpool.Pool) *NewsRepository {
	return &NewsRepository{db: db}
}

func (r *NewsRepository) CreateNews(ctx context.Context, n *models.News) (*models.News, error) {
	query := `
		INSERT INTO news (title, content, image_urls, author_id)
		VALUES ($1, $2, $3, $4)
		RETURNING id, title, content, image_urls, author_id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query, n.Title, n.Content, n.ImageURLs, n.AuthorID).
		Scan(&n.ID, &n.Title, &n.Content, &n.ImageURLs, &n.AuthorID, &n.CreatedAt, &n.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return n, nil
}

func (r *NewsRepository) ListNews(ctx context.Context, search string, page, limit int) ([]*models.News, error) {
	offset := (page - 1) * limit
	query := `SELECT id, title, content, image_urls, author_id, created_at, updated_at FROM news`

	var args []interface{}
	if search != "" {
		query += ` WHERE title ILIKE $1 OR content ILIKE $1`
		args = append(args, "%"+search+"%")
	}

	query += fmt.Sprintf(` ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var newsList []*models.News
	for rows.Next() {
		var n models.News
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.ImageURLs, &n.AuthorID, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		newsList = append(newsList, &n)
	}
	return newsList, nil
}

func (r *NewsRepository) GetNews(ctx context.Context, id int64) (*models.News, error) {
	query := `SELECT id, title, content, image_urls, author_id, created_at, updated_at FROM news WHERE id = $1`
	var n models.News
	err := r.db.QueryRow(ctx, query, id).
		Scan(&n.ID, &n.Title, &n.Content, &n.ImageURLs, &n.AuthorID, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NewsRepository) UpdateNews(ctx context.Context, id int64, n *models.News) (*models.News, error) {
	query := `
		UPDATE news SET title = $1, content = $2, image_urls = $3, updated_at = NOW()
		WHERE id = $4
		RETURNING id, title, content, image_urls, author_id, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query, n.Title, n.Content, n.ImageURLs, id).
		Scan(&n.ID, &n.Title, &n.Content, &n.ImageURLs, &n.AuthorID, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return n, nil
}

func (r *NewsRepository) DeleteNews(ctx context.Context, id int64) error {
	query := `DELETE FROM news WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}
