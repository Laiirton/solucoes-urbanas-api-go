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
		INSERT INTO teams (name, region_id, description, created_at, updated_at)
		VALUES ($1, $2, $3, NOW(), NOW())
		RETURNING id, name, region_id, description, created_at, updated_at`

	team := &models.Team{}
	err := r.db.QueryRow(ctx, query, req.Name, req.RegionID, req.Description).Scan(
		&team.ID, &team.Name, &team.RegionID, &team.Description, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create team: %w", err)
	}

	// Fetch region name
	if team.RegionID != nil {
		r.db.QueryRow(ctx, `SELECT name FROM regions WHERE id = $1`, *team.RegionID).Scan(&team.RegionName)
	}

	return team, nil
}

func (r *TeamRepository) GetTeamByID(ctx context.Context, id int64) (*models.Team, error) {
	query := `
		SELECT t.id, t.name, t.region_id, COALESCE(rg.name, ''), t.description, t.created_at, t.updated_at
		FROM teams t
		LEFT JOIN regions rg ON t.region_id = rg.id
		WHERE t.id = $1`

	team := &models.Team{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&team.ID, &team.Name, &team.RegionID, &team.RegionName, &team.Description, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("team not found: %w", err)
	}

	return team, nil
}

func (r *TeamRepository) ListTeams(ctx context.Context, search string, page, limit int) ([]*models.Team, error) {
	query := `
		SELECT t.id, t.name, t.region_id, COALESCE(rg.name, ''), t.description, t.created_at, t.updated_at
		FROM teams t
		LEFT JOIN regions rg ON t.region_id = rg.id`

	var args []interface{}
	if search != "" {
		query += ` WHERE t.name ILIKE $1 OR rg.name ILIKE $1`
		args = append(args, "%"+search+"%")
	}

	query += ` ORDER BY t.id ASC`
	if limit > 0 {
		offset := (page - 1) * limit
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list teams: %w", err)
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		if err := rows.Scan(
			&team.ID, &team.Name, &team.RegionID, &team.RegionName, &team.Description, &team.CreatedAt, &team.UpdatedAt,
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
			region_id = COALESCE($2, region_id),
			description = COALESCE($3, description),
			updated_at = NOW()
		WHERE id = $4
		RETURNING id, name, region_id, description, created_at, updated_at`

	team := &models.Team{}
	err := r.db.QueryRow(ctx, query, req.Name, req.RegionID, req.Description, id).Scan(
		&team.ID, &team.Name, &team.RegionID, &team.Description, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update team: %w", err)
	}

	// Fetch region name
	if team.RegionID != nil {
		r.db.QueryRow(ctx, `SELECT name FROM regions WHERE id = $1`, *team.RegionID).Scan(&team.RegionName)
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

// GetTeamByRegion returns the team responsible for a given region.
func (r *TeamRepository) GetTeamByRegion(ctx context.Context, regionID int64) (*models.Team, error) {
	query := `
		SELECT t.id, t.name, t.region_id, COALESCE(rg.name, ''), t.description, t.created_at, t.updated_at
		FROM teams t
		LEFT JOIN regions rg ON t.region_id = rg.id
		WHERE t.region_id = $1
		LIMIT 1`

	team := &models.Team{}
	err := r.db.QueryRow(ctx, query, regionID).Scan(
		&team.ID, &team.Name, &team.RegionID, &team.RegionName, &team.Description, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("no team found for region %d: %w", regionID, err)
	}
	return team, nil
}

func (r *TeamRepository) ListMembers(ctx context.Context, teamID int64) ([]*models.User, error) {
	query := `SELECT id, username, email, full_name, type, profile_image_url, created_at, updated_at
	          FROM users WHERE team_id = $1 ORDER BY username ASC`

	rows, err := r.db.Query(ctx, query, teamID)
	if err != nil {
		return nil, fmt.Errorf("failed to list team members: %w", err)
	}
	defer rows.Close()

	var members []*models.User
	for rows.Next() {
		m := &models.User{}
		if err := rows.Scan(
			&m.ID, &m.Username, &m.Email, &m.FullName, &m.Type,
			&m.ProfileImageURL, &m.CreatedAt, &m.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan member: %w", err)
		}
		members = append(members, m)
	}
	if members == nil {
		members = []*models.User{}
	}
	return members, nil
}

func (r *TeamRepository) AddMember(ctx context.Context, teamID, userID int64) error {
	var exists bool
	err := r.db.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM teams WHERE id = $1)`, teamID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify team: %w", err)
	}
	if !exists {
		return fmt.Errorf("team not found")
	}

	result, err := r.db.Exec(ctx,
		`UPDATE users SET team_id = $1, updated_at = NOW() WHERE id = $2`,
		teamID, userID,
	)
	if err != nil {
		return fmt.Errorf("failed to add member: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}

func (r *TeamRepository) RemoveMember(ctx context.Context, teamID, userID int64) error {
	result, err := r.db.Exec(ctx,
		`UPDATE users SET team_id = NULL, updated_at = NOW() WHERE id = $1 AND team_id = $2`,
		userID, teamID,
	)
	if err != nil {
		return fmt.Errorf("failed to remove member: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found or not a member of this team")
	}
	return nil
}

func (r *TeamRepository) GetTeamStats(ctx context.Context, teamID int64) (*models.TeamStats, error) {
	team, err := r.GetTeamByID(ctx, teamID)
	if err != nil {
		return nil, err
	}

	stats := &models.TeamStats{Team: *team}

	r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users WHERE team_id = $1`, teamID).Scan(&stats.MemberCount)

	rows, err := r.db.Query(ctx, `
		SELECT status, COUNT(*) FROM service_requests
		WHERE team_id = $1 GROUP BY status`, teamID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var status string
			var count int
			if err := rows.Scan(&status, &count); err != nil {
				continue
			}
			stats.TotalRequests += count
			switch status {
			case "pending":
				stats.PendingRequests = count
			case "in_progress":
				stats.InProgressRequests = count
			case "completed":
				stats.CompletedRequests = count
			case "cancelled":
				stats.CancelledRequests = count
			}
		}
	}

	r.db.QueryRow(ctx, `
		SELECT COALESCE(ROUND(AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 86400)::numeric, 1), 0)
		FROM service_requests
		WHERE team_id = $1 AND status = 'completed'`, teamID,
	).Scan(&stats.AvgResolutionDays)

	if stats.TotalRequests > 0 {
		stats.CompletionRate = float64(stats.CompletedRequests) / float64(stats.TotalRequests) * 100
	}

	recentRows, err := r.db.Query(ctx, `
		SELECT sr.id, sr.user_id, sr.service_id, sr.protocol_number, sr.service_title, sr.category,
		       sr.request_data, sr.attachments, sr.status, sr.latitude, sr.longitude, sr.geocoded_address,
		       sr.team_id, sr.region_id, COALESCE(rg.name, ''), sr.created_at, sr.updated_at
		FROM service_requests sr
		LEFT JOIN regions rg ON sr.region_id = rg.id
		WHERE sr.team_id = $1
		ORDER BY sr.created_at DESC LIMIT 5`, teamID)
	if err == nil {
		defer recentRows.Close()
		for recentRows.Next() {
			sr := &models.ServiceRequest{}
			if err := recentRows.Scan(
				&sr.ID, &sr.UserID, &sr.ServiceID, &sr.ProtocolNumber,
				&sr.ServiceTitle, &sr.Category, &sr.RequestData,
				&sr.Attachments, &sr.Status, &sr.Latitude,
				&sr.Longitude, &sr.GeocodedAddress, &sr.TeamID,
				&sr.RegionID, &sr.RegionName,
				&sr.CreatedAt, &sr.UpdatedAt,
			); err == nil {
				stats.RecentRequests = append(stats.RecentRequests, sr)
			}
		}
	}
	if stats.RecentRequests == nil {
		stats.RecentRequests = []*models.ServiceRequest{}
	}

	return stats, nil
}
