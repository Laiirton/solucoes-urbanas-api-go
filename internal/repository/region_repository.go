package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type RegionRepository struct {
	db *pgxpool.Pool
}

func NewRegionRepository(db *pgxpool.Pool) *RegionRepository {
	return &RegionRepository{db: db}
}

func (r *RegionRepository) Create(ctx context.Context, req *models.CreateRegionRequest) (*models.Region, error) {
	region := &models.Region{}
	err := r.db.QueryRow(ctx,
		`INSERT INTO regions (name, neighborhoods, created_at, updated_at)
		 VALUES ($1, $2, NOW(), NOW())
		 RETURNING id, name, neighborhoods, created_at, updated_at`,
		req.Name, req.Neighborhoods,
	).Scan(&region.ID, &region.Name, &region.Neighborhoods, &region.CreatedAt, &region.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create region: %w", err)
	}
	return region, nil
}

func (r *RegionRepository) GetByID(ctx context.Context, id int64) (*models.Region, error) {
	region := &models.Region{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, neighborhoods, created_at, updated_at
		 FROM regions WHERE id = $1`, id,
	).Scan(&region.ID, &region.Name, &region.Neighborhoods, &region.CreatedAt, &region.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("region not found: %w", err)
	}
	return region, nil
}

func (r *RegionRepository) List(ctx context.Context, search string, page, limit int) ([]*models.Region, error) {
	query := `SELECT id, name, neighborhoods, created_at, updated_at FROM regions`
	var args []interface{}

	if search != "" {
		query += ` WHERE name ILIKE $1`
		args = append(args, "%"+search+"%")
	}

	query += ` ORDER BY name ASC`
	if limit > 0 {
		offset := (page - 1) * limit
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list regions: %w", err)
	}
	defer rows.Close()

	var regions []*models.Region
	for rows.Next() {
		region := &models.Region{}
		if err := rows.Scan(&region.ID, &region.Name, &region.Neighborhoods, &region.CreatedAt, &region.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan region: %w", err)
		}
		regions = append(regions, region)
	}
	if regions == nil {
		regions = []*models.Region{}
	}
	return regions, nil
}

func (r *RegionRepository) Update(ctx context.Context, id int64, req *models.UpdateRegionRequest) (*models.Region, error) {
	region := &models.Region{}
	err := r.db.QueryRow(ctx,
		`UPDATE regions SET
			name = COALESCE($1, name),
			neighborhoods = COALESCE($2, neighborhoods),
			updated_at = NOW()
		 WHERE id = $3
		 RETURNING id, name, neighborhoods, created_at, updated_at`,
		req.Name, req.Neighborhoods, id,
	).Scan(&region.ID, &region.Name, &region.Neighborhoods, &region.CreatedAt, &region.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to update region: %w", err)
	}
	return region, nil
}

func (r *RegionRepository) Delete(ctx context.Context, id int64) error {
	result, err := r.db.Exec(ctx, `DELETE FROM regions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete region: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("region not found")
	}
	return nil
}

// FindByNeighborhood checks each region's neighborhoods JSON array for a match.
// Uses PostgreSQL JSONB containment (@>) to check if the given bairro exists in the array.
func (r *RegionRepository) FindByNeighborhood(ctx context.Context, bairro string) (*models.Region, error) {
	bairroJSON, _ := json.Marshal(bairro)
	region := &models.Region{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, neighborhoods, created_at, updated_at
		 FROM regions
		 WHERE neighborhoods @> $1::jsonb
		 LIMIT 1`,
		string(bairroJSON),
	).Scan(&region.ID, &region.Name, &region.Neighborhoods, &region.CreatedAt, &region.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("no region found for neighborhood '%s': %w", bairro, err)
	}
	return region, nil
}

// FindByNeighborhoodCaseInsensitive busca região por nome do bairro ignorando acentos e caixa.
// Usa a extensão unaccent do PostgreSQL para comparação flexível.
func (r *RegionRepository) FindByNeighborhoodCaseInsensitive(ctx context.Context, bairro string) (*models.Region, error) {
	region := &models.Region{}
	err := r.db.QueryRow(ctx,
		`SELECT id, name, neighborhoods, created_at, updated_at
		 FROM regions
		 WHERE EXISTS (
			 SELECT 1 FROM jsonb_array_elements_text(neighborhoods) AS nb
			 WHERE LOWER(TRANSLATE(nb, 'áàâãäéèêëíìîïóòôõöúùûüçñ', 'aaaaaeeeeiiiiooooouuuucn')) = LOWER(TRANSLATE($1, 'áàâãäéèêëíìîïóòôõöúùûüçñ', 'aaaaaeeeeiiiiooooouuuucn'))
		 )
		 LIMIT 1`,
		bairro,
	).Scan(&region.ID, &region.Name, &region.Neighborhoods, &region.CreatedAt, &region.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("no region found for neighborhood '%s': %w", bairro, err)
	}
	return region, nil
}

// ListAll returns all regions (used for matching region from address text).
func (r *RegionRepository) ListAll(ctx context.Context) ([]*models.Region, error) {
	return r.List(ctx, "", 1, 1000)
}
