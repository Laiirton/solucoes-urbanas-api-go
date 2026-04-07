package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type ServiceRepository struct {
	db *pgxpool.Pool
}

func NewServiceRepository(db *pgxpool.Pool) *ServiceRepository {
	return &ServiceRepository{db: db}
}

func (r *ServiceRepository) CreateService(ctx context.Context, req *models.CreateServiceRequest) (*models.Service, error) {
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	formSchema := req.FormSchema
	if len(formSchema) == 0 {
		formSchema = []byte("[]")
	}

	query := `
		INSERT INTO services (title, description, category, form_schema, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, title, description, category, form_schema, is_active, created_at, updated_at`

	svc := &models.Service{}
	err := r.db.QueryRow(ctx, query,
		req.Title, req.Description, req.Category, formSchema, isActive,
	).Scan(
		&svc.ID, &svc.Title, &svc.Description, &svc.Category, &svc.FormSchema,
		&svc.IsActive, &svc.CreatedAt, &svc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create service: %w", err)
	}
	return svc, nil
}

func (r *ServiceRepository) GetServiceByID(ctx context.Context, id int64) (*models.Service, error) {
	query := `SELECT id, title, description, category, form_schema, is_active, created_at, updated_at
              FROM services WHERE id = $1`

	svc := &models.Service{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&svc.ID, &svc.Title, &svc.Description, &svc.Category, &svc.FormSchema,
		&svc.IsActive, &svc.CreatedAt, &svc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("service not found: %w", err)
	}
	return svc, nil
}

func (r *ServiceRepository) ListServices(ctx context.Context, onlyActive bool) ([]*models.Service, error) {
	query := `SELECT id, title, description, category, form_schema, is_active, created_at, updated_at
              FROM services`
	if onlyActive {
		query += ` WHERE is_active = TRUE`
	}
	query += ` ORDER BY id ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list services: %w", err)
	}
	defer rows.Close()

	var services []*models.Service
	for rows.Next() {
		svc := &models.Service{}
		if err := rows.Scan(
			&svc.ID, &svc.Title, &svc.Description, &svc.Category, &svc.FormSchema,
			&svc.IsActive, &svc.CreatedAt, &svc.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan service: %w", err)
		}
		services = append(services, svc)
	}
	if services == nil {
		services = []*models.Service{}
	}
	return services, nil
}

func (r *ServiceRepository) UpdateService(ctx context.Context, id int64, req *models.UpdateServiceRequest) (*models.Service, error) {
	query := `
		UPDATE services SET
			title       = COALESCE($1, title),
			description = COALESCE($2, description),
			category    = COALESCE($3, category),
			form_schema = COALESCE($4, form_schema),
			is_active   = COALESCE($5, is_active),
			updated_at  = NOW()
		WHERE id = $6
		RETURNING id, title, description, category, form_schema, is_active, created_at, updated_at`

	svc := &models.Service{}
	err := r.db.QueryRow(ctx, query,
		req.Title, req.Description, req.Category, req.FormSchema, req.IsActive, id,
	).Scan(
		&svc.ID, &svc.Title, &svc.Description, &svc.Category, &svc.FormSchema,
		&svc.IsActive, &svc.CreatedAt, &svc.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update service: %w", err)
	}
	return svc, nil
}

func (r *ServiceRepository) DeleteService(ctx context.Context, id int64) error {
	query := `DELETE FROM services WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete service: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("service not found")
	}
	return nil
}
