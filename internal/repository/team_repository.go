package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type TeamRepository struct {
	db *pgxpool.Pool
}

func NewTeamRepository(db *pgxpool.Pool) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, req *models.CreateTeamRequest) (*models.Team, error) {
	query := `
		INSERT INTO teams (name, service_category, description, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, name, service_category, description, created_at, updated_at`

	team := &models.Team{}
	err := r.db.QueryRow(ctx, query, req.Name, req.ServiceCategory, req.Description).Scan(
		&team.ID, &team.Name, &team.ServiceCategory, &team.Description, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	return team, nil
}

func (r *TeamRepository) GetTeamByID(ctx context.Context, id int64) (*models.Team, error) {
	query := `
		SELECT id, name, service_category, description, created_at, updated_at
		FROM teams WHERE id = $1`

	team := &models.Team{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.ServiceCategory, &team.Description, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	return team, nil
}

func (r *TeamRepository) ListTeams(ctx context.Context, search string, page, limit int) ([]*models.Team, error) {
	offset := (page - 1) * limit
	query := `
		SELECT id, name, service_category, description, created_at, updated_at
		FROM teams`

	var args []interface{}
	if search != "" {
		query += ` WHERE name ILIKE $1 OR service_category ILIKE $1`
		args = append(args, "%"+search+"%")
	}

	query += fmt.Sprintf(` ORDER BY id ASC LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		if err := rows.Scan(
			&team.ID, &team.Name, &team.ServiceCategory, &team.Description, &team.CreatedAt, &team.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}
		teams = append(teams, team)
	}

	if teams == nil {
		teams = []*models.Team{}
	}

	return teams, nil
}

func (r *TeamRepository) UpdateTeam(ctx context.Context, id int64, req *models.UpdateTeamRequest) (*models.Team, error) {
	query := `
		UPDATE teams SET
			name = COALESCE($1, name),
			service_category = COALESCE($2, service_category),
			description = COALESCE($3, description),
			updated_at = NOW()
		WHERE id = $4
		RETURNING id, name, service_category, description, created_at, updated_at`

	team := &models.Team{}
	err := r.db.QueryRow(ctx, query, req.Name, req.ServiceCategory, req.Description, id).Scan(
		&team.ID, &team.Name, &team.ServiceCategory, &team.Description, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	return team, nil
}

func (r *TeamRepository) DeleteTeam(ctx context.Context, id int64) error {
	query := `DELETE FROM teams WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("team not found")
	}
	return nil
}
