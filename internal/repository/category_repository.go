package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type CategoryRepository struct {
	db *pgxpool.Pool
}

func NewCategoryRepository(db *pgxpool.Pool) *CategoryRepository {
	return &CategoryRepository{db: db}
}

func (r *CategoryRepository) Create(ctx context.Context, req *models.CreateCategoryRequest) (*models.Category, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	query := `
		INSERT INTO categories (name, icon, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, name, icon, is_active, created_at, updated_at`

	cat := &models.Category{}
	err := r.db.QueryRow(ctx, query, req.Name, req.Icon, isActive).Scan(
		&cat.ID, &cat.Name, &cat.Icon, &cat.IsActive, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}
	return cat, nil
}

func (r *CategoryRepository) GetByID(ctx context.Context, id int64) (*models.Category, error) {
	query := `SELECT id, name, icon, is_active, created_at, updated_at FROM categories WHERE id = $1`

	cat := &models.Category{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&cat.ID, &cat.Name, &cat.Icon, &cat.IsActive, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}
	return cat, nil
}

func (r *CategoryRepository) List(ctx context.Context, onlyActive bool) ([]*models.Category, error) {
	query := `SELECT id, name, icon, is_active, created_at, updated_at FROM categories`

	var args []interface{}
	if onlyActive {
		query += ` WHERE is_active = TRUE`
	}

	query += ` ORDER BY name ASC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	defer rows.Close()

	var categories []*models.Category
	for rows.Next() {
		cat := &models.Category{}
		if err := rows.Scan(&cat.ID, &cat.Name, &cat.Icon, &cat.IsActive, &cat.CreatedAt, &cat.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan category: %w", err)
		}
		categories = append(categories, cat)
	}
	if categories == nil {
		categories = []*models.Category{}
	}
	return categories, nil
}

func (r *CategoryRepository) Update(ctx context.Context, id int64, req *models.UpdateCategoryRequest) (*models.Category, error) {
	query := `
		UPDATE categories SET
			name = COALESCE($1, name),
			icon = COALESCE($2, icon),
			is_active = COALESCE($3, is_active),
			updated_at = NOW()
		WHERE id = $4
		RETURNING id, name, icon, is_active, created_at, updated_at`

	cat := &models.Category{}
	err := r.db.QueryRow(ctx, query, req.Name, req.Icon, req.IsActive, id).Scan(
		&cat.ID, &cat.Name, &cat.Icon, &cat.IsActive, &cat.CreatedAt, &cat.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}
	return cat, nil
}

func (r *CategoryRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete category: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("category not found")
	}
	return nil
}
