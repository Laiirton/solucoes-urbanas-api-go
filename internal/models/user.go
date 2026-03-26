package models

import "time"

type User struct {
	ID        int64      `json:"id"`
	Username  string     `json:"username"`
	Password  string     `json:"-"` // never expose password in JSON
	Email     string     `json:"email"`
	FullName  *string    `json:"full_name,omitempty"`
	CPF       *string    `json:"cpf,omitempty"`
	BirthDate *time.Time `json:"birth_date,omitempty"`
	Type      *string    `json:"type,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreateUserRequest struct {
	Username  string  `json:"username"`
	Password  string  `json:"password"`
	Email     string  `json:"email"`
	FullName  *string `json:"full_name,omitempty"`
	CPF       *string `json:"cpf,omitempty"`
	BirthDate *string `json:"birth_date,omitempty"` // parsed as string "YYYY-MM-DD"
	Type      *string `json:"type,omitempty"`
}

type UpdateUserRequest struct {
	Username  *string `json:"username,omitempty"`
	FullName  *string `json:"full_name,omitempty"`
	CPF       *string `json:"cpf,omitempty"`
	BirthDate *string `json:"birth_date,omitempty"`
	Type      *string `json:"type,omitempty"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type MessageResponse struct {
	Message string `json:"message"`
}
