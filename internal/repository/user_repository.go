package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func formatBirthDate(bd *time.Time) *string {
	if bd == nil {
		return nil
	}
	s := bd.Format("02/01/2006")
	return &s
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, req *models.CreateUserRequest, hashedPassword string) (*models.User, error) {
	var workAreaJSON []byte
	if req.WorkArea != nil {
		var err error
		workAreaJSON, err = json.Marshal(req.WorkArea)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal work_area: %w", err)
		}
	}

	query := `
		INSERT INTO users (username, password, email, full_name, cpf, phone, birth_date, type, team_id, work_area, profile_image_url, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
		RETURNING id, username, email, full_name, cpf, phone, birth_date, type, team_id, work_area, profile_image_url, created_at, updated_at`

	var birthDate *time.Time
	if req.BirthDate != nil {
		parsed, err := time.Parse("02/01/2006", *req.BirthDate)
		if err != nil {
			return nil, fmt.Errorf("invalid birth_date format, expected DD/MM/YYYY: %w", err)
		}
		birthDate = &parsed
	}

	user := &models.User{}
	var workAreaResult []byte
	var bd time.Time
	err := r.db.QueryRow(ctx, query,
		req.Username, hashedPassword, req.Email,
		req.FullName, req.CPF, req.Phone, birthDate, req.Type,
		req.TeamID, workAreaJSON, req.ProfileImageURL,
	).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.FullName, &user.CPF, &user.Phone, &bd,
		&user.Type, &user.TeamID, &workAreaResult, &user.ProfileImageURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	user.BirthDate = formatBirthDate(&bd)

	if workAreaResult != nil {
		if err := json.Unmarshal(workAreaResult, &user.WorkArea); err != nil {
			return nil, fmt.Errorf("failed to unmarshal work_area: %w", err)
		}
	}

	return user, nil
}

func (r *UserRepository) GetUserByUsernameOrEmail(ctx context.Context, identifier string) (*models.User, error) {
	query := `SELECT id, username, password, email, full_name, cpf, phone, birth_date, type, team_id, work_area, profile_image_url, created_at, updated_at
 FROM users WHERE username = $1 OR email = $1`

	user := &models.User{}
	var workAreaResult []byte
	var bd *time.Time
	err := r.db.QueryRow(ctx, query, identifier).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.FullName, &user.CPF, &user.Phone, &bd,
		&user.Type, &user.TeamID, &workAreaResult, &user.ProfileImageURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.BirthDate = formatBirthDate(bd)

	if workAreaResult != nil {
		if err := json.Unmarshal(workAreaResult, &user.WorkArea); err != nil {
			return nil, fmt.Errorf("failed to unmarshal work_area: %w", err)
		}
	}

	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.full_name, u.cpf, u.phone, u.birth_date, u.type, u.team_id, u.work_area, u.profile_image_url, u.created_at, u.updated_at,
		 t.id, t.name, t.region_id, COALESCE(rg.name, ''), t.description, t.created_at, t.updated_at
		FROM users u
		LEFT JOIN teams t ON u.team_id = t.id
		LEFT JOIN regions rg ON t.region_id = rg.id
		WHERE u.id = $1`

	user := &models.User{}
	var tID *int64
	var tName *string
	var tRegionID *int64
	var tRegionName string
	var tDesc *string
	var tCreatedAt, tUpdatedAt *time.Time
	var workAreaResult []byte
	var bd *time.Time

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.FullName, &user.CPF, &user.Phone, &bd,
		&user.Type, &user.TeamID, &workAreaResult, &user.ProfileImageURL,
		&user.CreatedAt, &user.UpdatedAt,
		&tID, &tName, &tRegionID, &tRegionName, &tDesc, &tCreatedAt, &tUpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	user.BirthDate = formatBirthDate(bd)

	if workAreaResult != nil {
		if err := json.Unmarshal(workAreaResult, &user.WorkArea); err != nil {
			return nil, fmt.Errorf("failed to unmarshal work_area: %w", err)
		}
	}

	if tID != nil {
		user.Team = &models.Team{
			ID:          *tID,
			Name:        *tName,
			RegionID:    tRegionID,
			RegionName:  tRegionName,
			Description: tDesc,
			CreatedAt:   *tCreatedAt,
			UpdatedAt:   *tUpdatedAt,
		}
	}

	return user, nil
}

func (r *UserRepository) ListUsers(ctx context.Context, search, userType string, page, limit int) ([]*models.User, error) {
	query := `SELECT u.id, u.username, u.email, u.full_name, u.cpf, u.phone, u.birth_date, u.type, u.team_id, u.work_area, u.profile_image_url, u.created_at, u.updated_at,
		 t.id, t.name, t.description, t.region_id, COALESCE(rg.name, ''), t.created_at, t.updated_at
 FROM users u
 LEFT JOIN teams t ON u.team_id = t.id
 LEFT JOIN regions rg ON t.region_id = rg.id`

	var args []interface{}
	whereApplied := false

	if search != "" {
		query += ` WHERE (u.username ILIKE $1 OR u.full_name ILIKE $1 OR u.email ILIKE $1 OR u.type ILIKE $1 OR u.cpf ILIKE $1)`
		args = append(args, "%"+search+"%")
		whereApplied = true
	}

	if userType != "" {
		if whereApplied {
			query += fmt.Sprintf(` AND u.type ILIKE $%d`, len(args)+1)
		} else {
			query += ` WHERE u.type ILIKE $1`
		}
		args = append(args, userType)
	}

	query += ` ORDER BY u.id ASC`
	if limit > 0 {
		offset := (page - 1) * limit
		query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, len(args)+1, len(args)+2)
		args = append(args, limit, offset)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		var workAreaResult []byte
		var tID *int64
		var tName *string
		var tDesc *string
		var tRegionID *int64
		var tRegionName string
		var tCreatedAt *time.Time
		var tUpdatedAt *time.Time
		var bd *time.Time
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email,
			&user.FullName, &user.CPF, &user.Phone, &bd,
			&user.Type, &user.TeamID, &workAreaResult, &user.ProfileImageURL,
			&user.CreatedAt, &user.UpdatedAt,
			&tID, &tName, &tDesc, &tRegionID, &tRegionName, &tCreatedAt, &tUpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}

		user.BirthDate = formatBirthDate(bd)

		if workAreaResult != nil {
			if err := json.Unmarshal(workAreaResult, &user.WorkArea); err != nil {
				return nil, fmt.Errorf("failed to unmarshal work_area: %w", err)
			}
		}

		if tID != nil {
			user.Team = &models.Team{
				ID:          *tID,
				Name:        *tName,
				Description: tDesc,
				RegionID:    tRegionID,
				RegionName:  tRegionName,
				CreatedAt:   *tCreatedAt,
				UpdatedAt:   *tUpdatedAt,
			}
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error) {
	var birthDate *time.Time
	if req.BirthDate != nil && *req.BirthDate != "" {
		// Try DD/MM/YYYY first (standard in Brazil)
		parsed, err := time.Parse("02/01/2006", *req.BirthDate)
		if err != nil {
			// Fallback to YYYY-MM-DD
			parsed2, err2 := time.Parse("2006-01-02", *req.BirthDate)
			if err2 != nil {
				return nil, fmt.Errorf("invalid birth_date format, expected DD/MM/YYYY or YYYY-MM-DD: %w", err)
			}
			birthDate = &parsed2
		} else {
			birthDate = &parsed
		}
	}

	var workAreaJSON []byte
	if req.WorkArea != nil {
		var err error
		workAreaJSON, err = json.Marshal(req.WorkArea)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal work_area: %w", err)
		}
	}

	query := `
		UPDATE users SET
			username = COALESCE($1, username),
			email = COALESCE($2, email),
			full_name = COALESCE($3, full_name),
			cpf = COALESCE($4, cpf),
			phone = COALESCE($5, phone),
			birth_date = COALESCE($6, birth_date),
			type = COALESCE($7, type),
			team_id = COALESCE($8, team_id),
			work_area = COALESCE($9, work_area),
			profile_image_url = COALESCE($10, profile_image_url),
			updated_at = NOW()
		WHERE id = $11
		RETURNING id, username, email, full_name, cpf, phone, birth_date, type, team_id, work_area, profile_image_url, created_at, updated_at`

	user := &models.User{}
	var workAreaResult []byte
	var bd *time.Time
	err := r.db.QueryRow(ctx, query,
		req.Username, req.Email, req.FullName, req.CPF, req.Phone, birthDate, req.Type, req.TeamID, workAreaJSON, req.ProfileImageURL, id,
	).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.FullName, &user.CPF, &user.Phone, &bd,
		&user.Type, &user.TeamID, &workAreaResult, &user.ProfileImageURL,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	user.BirthDate = formatBirthDate(bd)

	if workAreaResult != nil {
		if err := json.Unmarshal(workAreaResult, &user.WorkArea); err != nil {
			return nil, fmt.Errorf("failed to unmarshal work_area: %w", err)
		}
	}

	return user, nil
}

func (r *UserRepository) DeleteUser(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("user not found")
	}
	return nil
}
