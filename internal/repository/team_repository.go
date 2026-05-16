package repository

import (
	"context"
	"encoding/json"
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

	workAreas, err := r.getTeamWorkAreas(ctx, id)
	if err == nil {
		team.WorkAreas = workAreas
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

	if len(teams) > 0 {
		ids := make([]int64, len(teams))
		for i, t := range teams {
			ids[i] = t.ID
		}
		areasMap, err := r.getTeamsWorkAreas(ctx, ids)
		if err == nil {
			for _, t := range teams {
				if wa, ok := areasMap[t.ID]; ok {
					t.WorkAreas = wa
				}
			}
		}
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

// FindTeamsByRegion returns all teams assigned to a region.
func (r *TeamRepository) FindTeamsByRegion(ctx context.Context, regionID int64) ([]*models.Team, error) {
	query := `
		SELECT t.id, t.name, t.region_id, COALESCE(rg.name, ''), t.description, t.created_at, t.updated_at
		FROM teams t
		LEFT JOIN regions rg ON t.region_id = rg.id
		WHERE t.region_id = $1
		ORDER BY t.name ASC`

	rows, err := r.db.Query(ctx, query, regionID)
	if err != nil {
		return nil, fmt.Errorf("failed to find teams by region: %w", err)
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

// FindTeamByRegionAndCategory finds the best team for a region that handles the given service category.
// Matches by checking which team's secretary has the category in their work_area.
func (r *TeamRepository) FindTeamByRegionAndCategory(ctx context.Context, regionID int64, serviceCategory string) (*models.Team, error) {
	catJSON, _ := json.Marshal(serviceCategory)
	query := `
		SELECT t.id, t.name, t.region_id, COALESCE(rg.name, ''), t.description, t.created_at, t.updated_at
		FROM teams t
		LEFT JOIN regions rg ON t.region_id = rg.id
		WHERE t.region_id = $1
		  AND t.id IN (
			SELECT u.team_id FROM users u
			WHERE u.team_id IS NOT NULL
			  AND u.type = 'secretary'
			  AND u.work_area::jsonb @> $2::jsonb
		  )
		LIMIT 1`

	team := &models.Team{}
	err := r.db.QueryRow(ctx, query, regionID, string(catJSON)).Scan(
		&team.ID, &team.Name, &team.RegionID, &team.RegionName, &team.Description, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("no team found for region %d and category '%s': %w", regionID, serviceCategory, err)
	}
	return team, nil
}

// GetTeamByRegion returns the first team assigned to a given region (legacy, for single-team-per-region setups).
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

	// 1. Get the user's type
	var userType *string
	err = r.db.QueryRow(ctx, `SELECT type FROM users WHERE id = $1`, userID).Scan(&userType)
	if err != nil {
		return fmt.Errorf("failed to get user type: %w", err)
	}

	// 2. If user is an attendant, get the secretary's work areas
	if userType != nil && *userType == "attendant" {
		var workAreaJSON []byte
		err = r.db.QueryRow(ctx, `
			SELECT work_area FROM users 
			WHERE team_id = $1 AND type = 'secretary' 
			LIMIT 1`, teamID).Scan(&workAreaJSON)
		
		if err == nil && workAreaJSON != nil {
			// Update the user's team_id and work_area
			result, err := r.db.Exec(ctx,
				`UPDATE users SET team_id = $1, work_area = $2, updated_at = NOW() WHERE id = $3`,
				teamID, workAreaJSON, userID,
			)
			if err != nil {
				return fmt.Errorf("failed to add member and update work area: %w", err)
			}
			if result.RowsAffected() == 0 {
				return fmt.Errorf("user not found")
			}
			return nil
		}
	}

	// Fallback to just updating team_id if not attendant or no secretary found
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

	// Consolidated query: member_count + status stats + avg resolution days in 1 round trip
	consolidatedQuery := `
		SELECT
			(SELECT COUNT(*) FROM users WHERE team_id = $1) AS member_count,
			COUNT(*) AS total_requests,
			COUNT(*) FILTER (WHERE status = 'pending') AS pending,
			COUNT(*) FILTER (WHERE status = 'in_progress') AS in_progress,
			COUNT(*) FILTER (WHERE status = 'completed') AS completed,
			COUNT(*) FILTER (WHERE status = 'cancelled') AS cancelled,
			COALESCE(ROUND(AVG(EXTRACT(EPOCH FROM (updated_at - created_at)) / 86400)::numeric, 1)
				FILTER (WHERE status = 'completed'), 0) AS avg_days
		FROM service_requests
		WHERE team_id = $1`

	err = r.db.QueryRow(ctx, consolidatedQuery, teamID).Scan(
		&stats.MemberCount,
		&stats.TotalRequests,
		&stats.PendingRequests,
		&stats.InProgressRequests,
		&stats.CompletedRequests,
		&stats.CancelledRequests,
		&stats.AvgResolutionDays,
	)
	if err != nil {
		// Fallback to individual queries
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
	}

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

func (r *TeamRepository) ListTeamsByWorkArea(ctx context.Context, categoryName string) ([]*models.Team, error) {
	catJSON, _ := json.Marshal(categoryName)
	query := `
		SELECT t.id, t.name, t.region_id, COALESCE(rg.name, ''), t.description, t.created_at, t.updated_at
		FROM teams t
		LEFT JOIN regions rg ON t.region_id = rg.id
		WHERE t.id IN (
			SELECT DISTINCT team_id FROM users 
			WHERE type = 'secretary' AND work_area::jsonb @> $1::jsonb AND team_id IS NOT NULL
		)`

	rows, err := r.db.Query(ctx, query, string(catJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to list teams by work area: %w", err)
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

func (r *TeamRepository) getTeamWorkAreas(ctx context.Context, teamID int64) ([]string, error) {
	query := `
		SELECT DISTINCT jsonb_array_elements_text(work_area::jsonb)
		FROM users
		WHERE team_id = $1 AND type = 'secretary' AND work_area IS NOT NULL`

	rows, err := r.db.Query(ctx, query, teamID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var areas []string
	for rows.Next() {
		var wa string
		if err := rows.Scan(&wa); err == nil {
			areas = append(areas, wa)
		}
	}
	if areas == nil {
		areas = []string{}
	}
	return areas, nil
}

func (r *TeamRepository) getTeamsWorkAreas(ctx context.Context, teamIDs []int64) (map[int64][]string, error) {
	query := `
		SELECT u.team_id, jsonb_array_elements_text(u.work_area::jsonb)
		FROM users u
		WHERE u.team_id = ANY($1) AND u.type = 'secretary' AND u.work_area IS NOT NULL`

	rows, err := r.db.Query(ctx, query, teamIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	areasMap := make(map[int64][]string)
	for rows.Next() {
		var teamID int64
		var wa string
		if err := rows.Scan(&teamID, &wa); err == nil {
			areasMap[teamID] = append(areasMap[teamID], wa)
		}
	}
	return areasMap, nil
}

// FindCityWideTeamByCategory finds a team with no specific region (region_id IS NULL)
// whose secretary handles the given service category. Used as fallback for small towns.
func (r *TeamRepository) FindCityWideTeamByCategory(ctx context.Context, serviceCategory string) (*models.Team, error) {
	catJSON, _ := json.Marshal(serviceCategory)
	query := `
		SELECT t.id, t.name, t.region_id, COALESCE(rg.name, ''), t.description, t.created_at, t.updated_at
		FROM teams t
		LEFT JOIN regions rg ON t.region_id = rg.id
		WHERE t.region_id IS NULL
		  AND t.id IN (
			SELECT u.team_id FROM users u
			WHERE u.team_id IS NOT NULL
			  AND u.type = 'secretary'
			  AND u.work_area::jsonb @> $1::jsonb
		  )
		LIMIT 1`

	team := &models.Team{}
	err := r.db.QueryRow(ctx, query, string(catJSON)).Scan(
		&team.ID, &team.Name, &team.RegionID, &team.RegionName, &team.Description, &team.CreatedAt, &team.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("no city-wide team found for category '%s': %w", serviceCategory, err)
	}
	return team, nil
}
