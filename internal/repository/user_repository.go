package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/laiirton/solucoes-urbanas-api/internal/models"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) CreateUser(ctx context.Context, req *models.CreateUserRequest, hashedPassword string) (*models.User, error) {
	query := `
		INSERT INTO users (username, password, email, full_name, cpf, birth_date, type, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
		RETURNING id, username, email, full_name, cpf, birth_date, type, created_at, updated_at`

	var birthDate *time.Time
	if req.BirthDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			return nil, fmt.Errorf("invalid birth_date format, expected YYYY-MM-DD: %w", err)
		}
		birthDate = &parsed
	}

	user := &models.User{}
	err := r.db.QueryRow(ctx, query,
		req.Username, hashedPassword, req.Email,
		req.FullName, req.CPF, birthDate, req.Type,
	).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.FullName, &user.CPF, &user.BirthDate,
		&user.Type, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, username, password, email, full_name, cpf, birth_date, type, created_at, updated_at
              FROM users WHERE email = $1`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Username, &user.Password, &user.Email,
		&user.FullName, &user.CPF, &user.BirthDate,
		&user.Type, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id int64) (*models.User, error) {
	query := `SELECT id, username, email, full_name, cpf, birth_date, type, created_at, updated_at
              FROM users WHERE id = $1`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.FullName, &user.CPF, &user.BirthDate,
		&user.Type, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	return user, nil
}

func (r *UserRepository) ListUsers(ctx context.Context) ([]*models.User, error) {
	query := `SELECT id, username, email, full_name, cpf, birth_date, type, created_at, updated_at
              FROM users ORDER BY id ASC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(
			&user.ID, &user.Username, &user.Email,
			&user.FullName, &user.CPF, &user.BirthDate,
			&user.Type, &user.CreatedAt, &user.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) UpdateUser(ctx context.Context, id int64, req *models.UpdateUserRequest) (*models.User, error) {
	var birthDate *time.Time
	if req.BirthDate != nil {
		parsed, err := time.Parse("2006-01-02", *req.BirthDate)
		if err != nil {
			return nil, fmt.Errorf("invalid birth_date format, expected YYYY-MM-DD: %w", err)
		}
		birthDate = &parsed
	}

	query := `
		UPDATE users SET
			username   = COALESCE($1, username),
			full_name  = COALESCE($2, full_name),
			cpf        = COALESCE($3, cpf),
			birth_date = COALESCE($4, birth_date),
			type       = COALESCE($5, type),
			updated_at = NOW()
		WHERE id = $6
		RETURNING id, username, email, full_name, cpf, birth_date, type, created_at, updated_at`

	user := &models.User{}
	err := r.db.QueryRow(ctx, query,
		req.Username, req.FullName, req.CPF, birthDate, req.Type, id,
	).Scan(
		&user.ID, &user.Username, &user.Email,
		&user.FullName, &user.CPF, &user.BirthDate,
		&user.Type, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
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
