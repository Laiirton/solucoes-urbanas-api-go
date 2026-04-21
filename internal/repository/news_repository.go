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
		INSERT INTO news (title, slug, summary, content, image_urls, status, category, tags, author_id, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, title, slug, summary, content, image_urls, status, category, tags, author_id, published_at, created_at, updated_at
	`
	err := r.db.QueryRow(ctx, query,
		n.Title, n.Slug, n.Summary, n.Content, n.ImageURLs, n.Status, n.Category, n.Tags, n.AuthorID, n.PublishedAt).
		Scan(&n.ID, &n.Title, &n.Slug, &n.Summary, &n.Content, &n.ImageURLs, &n.Status, &n.Category, &n.Tags, &n.AuthorID, &n.PublishedAt, &n.CreatedAt, &n.UpdatedAt)

	if err != nil {
		return nil, err
	}
	return n, nil
}

func (r *NewsRepository) ListNews(ctx context.Context, search string, status string, page, limit int) ([]*models.News, error) {
	offset := (page - 1) * limit
	query := `SELECT id, title, slug, summary, content, image_urls, status, category, tags, author_id, published_at, created_at, updated_at FROM news`

	var args []interface{}
	where := ""
	if search != "" {
		where = ` WHERE (title ILIKE $1 OR summary ILIKE $1)`
		args = append(args, "%"+search+"%")
	}

	if status != "" {
		if where == "" {
			where = " WHERE status = $1"
		} else {
			where += " AND status = $2"
		}
		args = append(args, status)
	}

	query += where
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
		if err := rows.Scan(&n.ID, &n.Title, &n.Slug, &n.Summary, &n.Content, &n.ImageURLs, &n.Status, &n.Category, &n.Tags, &n.AuthorID, &n.PublishedAt, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		newsList = append(newsList, &n)
	}
	return newsList, nil
}

func (r *NewsRepository) GetNews(ctx context.Context, id int64) (*models.News, error) {
	query := `SELECT id, title, slug, summary, content, image_urls, status, category, tags, author_id, published_at, created_at, updated_at FROM news WHERE id = $1`
	var n models.News
	err := r.db.QueryRow(ctx, query, id).
		Scan(&n.ID, &n.Title, &n.Slug, &n.Summary, &n.Content, &n.ImageURLs, &n.Status, &n.Category, &n.Tags, &n.AuthorID, &n.PublishedAt, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NewsRepository) GetNewsBySlug(ctx context.Context, slug string) (*models.News, error) {
	query := `SELECT id, title, slug, summary, content, image_urls, status, category, tags, author_id, published_at, created_at, updated_at FROM news WHERE slug = $1`
	var n models.News
	err := r.db.QueryRow(ctx, query, slug).
		Scan(&n.ID, &n.Title, &n.Slug, &n.Summary, &n.Content, &n.ImageURLs, &n.Status, &n.Category, &n.Tags, &n.AuthorID, &n.PublishedAt, &n.CreatedAt, &n.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &n, nil
}

func (r *NewsRepository) UpdateNews(ctx context.Context, id int64, n *models.UpdateNewsRequest) (*models.News, error) {
	query := `
		UPDATE news SET 
			title = COALESCE($1, title),
			slug = COALESCE($2, slug),
			summary = COALESCE($3, summary),
			content = COALESCE($4, content),
			image_urls = COALESCE($5, image_urls), 
			status = COALESCE($6, status),
			category = COALESCE($7, category),
			tags = COALESCE($8, tags),
			published_at = COALESCE($9, published_at),
			updated_at = NOW()
		WHERE id = $10
		RETURNING id, title, slug, summary, content, image_urls, status, category, tags, author_id, published_at, created_at, updated_at
	`

	news := &models.News{}
	err := r.db.QueryRow(ctx, query,
		nullableValue(n.Title),
		nullableValue(n.Slug),
		nullableValue(n.Summary),
		nullableValue(n.Content),
		nullableValue(n.ImageURLs),
		nullableValue(n.Status),
		nullableValue(n.Category),
		nullableValue(n.Tags),
		nullableValue(n.PublishedAt),
		id,
	).
		Scan(&news.ID, &news.Title, &news.Slug, &news.Summary, &news.Content, &news.ImageURLs, &news.Status, &news.Category, &news.Tags, &news.AuthorID, &news.PublishedAt, &news.CreatedAt, &news.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return news, nil
}

func (r *NewsRepository) DeleteNews(ctx context.Context, id int64) error {
	query := `DELETE FROM news WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func nullableValue[T any](value *T) any {
	if value == nil {
		return nil
	}
	return *value
}
