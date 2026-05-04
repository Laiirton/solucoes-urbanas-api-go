package models

import (
	"fmt"
	"net/mail"
	"time"
)

type User struct {
	ID              int64      `json:"id"`
	Username        string     `json:"username"`
	Password        string     `json:"-"` // never expose password in JSON
	Email           string     `json:"email"`
	FullName        *string    `json:"full_name,omitempty"`
	CPF             *string    `json:"cpf,omitempty"`
	BirthDate       *time.Time `json:"birth_date,omitempty"`
	Type            *string    `json:"type,omitempty"`
	TeamID          *int64     `json:"team_id,omitempty"`
	Team            *Team      `json:"team,omitempty"`
	WorkArea        *string    `json:"work_area,omitempty"`
	ProfileImageURL *string    `json:"profile_image_url,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type CreateUserRequest struct {
	Username        string  `json:"username"`
	Password        string  `json:"password"`
	Email           string  `json:"email"`
	FullName        *string `json:"full_name,omitempty"`
	CPF             *string `json:"cpf,omitempty"`
	BirthDate       *string `json:"birth_date,omitempty"`
	Type            *string `json:"type,omitempty"`
	TeamID          *int64  `json:"team_id,omitempty"`
	WorkArea        *string `json:"work_area,omitempty"`
	ProfileImageURL *string `json:"profile_image_url,omitempty"`
}

func (r *CreateUserRequest) Validate() error {
	if r.Username == "" {
		return fmt.Errorf("username is required")
	}
	if r.Email == "" {
		return fmt.Errorf("email is required")
	}
	if r.Password == "" {
		return fmt.Errorf("password is required")
	}

	if r.CPF == nil || *r.CPF == "" {
		return fmt.Errorf("cpf is required")
	}

	if r.BirthDate == nil || *r.BirthDate == "" {
		return fmt.Errorf("birth_date is required")
	}

	if _, err := mail.ParseAddress(r.Email); err != nil {
		return fmt.Errorf("invalid email format")
	}

	if _, err := time.Parse("02/01/2006", *r.BirthDate); err != nil {
		return fmt.Errorf("invalid birth_date format, expected DD/MM/YYYY")
	}

	return nil
}

type UpdateUserRequest struct {
	Username        *string `json:"username,omitempty"`
	Email           *string `json:"email,omitempty"`
	FullName        *string `json:"full_name,omitempty"`
	CPF             *string `json:"cpf,omitempty"`
	BirthDate       *string `json:"birth_date,omitempty"`
	Type            *string `json:"type,omitempty"`
	TeamID          *int64  `json:"team_id,omitempty"`
	WorkArea        *string `json:"work_area,omitempty"`
	ProfileImageURL *string `json:"profile_image_url,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type UserDetailResponse struct {
	User           User              `json:"user"`
	TotalRequests  int               `json:"total_requests"`
	Requests       []*ServiceRequest `json:"requests"`
	RequestSummary map[string]int    `json:"request_summary"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
