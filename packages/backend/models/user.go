package models

import (
	"time"

	"github.com/google/uuid"
)

// User representa un usuario en el sistema bancario
type User struct {
	ID                   uuid.UUID  `json:"id" db:"id"`
	Email                string     `json:"email" db:"email"`
	PasswordHash         string     `json:"-" db:"password_hash"` // No incluir en JSON por seguridad
	FirstName            string     `json:"first_name" db:"first_name"`
	LastName             string     `json:"last_name" db:"last_name"`
	Phone                *string    `json:"phone,omitempty" db:"phone"`
	DateOfBirth          *time.Time `json:"date_of_birth,omitempty" db:"date_of_birth"`
	TigerBeetleAccountID *int64     `json:"tigerbeetle_account_id,omitempty" db:"tigerbeetle_account_id"`
	CreatedAt            time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at" db:"updated_at"`
	IsActive             bool       `json:"is_active" db:"is_active"`
	EmailVerified        bool       `json:"email_verified" db:"email_verified"`
}

// CreateUserRequest representa la estructura para crear un nuevo usuario
type CreateUserRequest struct {
	Email     string `json:"email" validate:"required,email"`
	Password  string `json:"password" validate:"required,min=8"`
	FirstName string `json:"first_name" validate:"required,min=2"`
	LastName  string `json:"last_name" validate:"required,min=2"`
}

// UpdateUserRequest representa la estructura para actualizar un usuario
type UpdateUserRequest struct {
	Email     *string `json:"email,omitempty" validate:"omitempty,email"`
	FirstName *string `json:"first_name,omitempty" validate:"omitempty,min=2"`
	LastName  *string `json:"last_name,omitempty" validate:"omitempty,min=2"`
}

// LoginRequest representa la estructura para el login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UserResponse representa la respuesta pública del usuario (sin datos sensibles)
type UserResponse struct {
	ID                   uuid.UUID  `json:"id"`
	Email                string     `json:"email"`
	FirstName            string     `json:"first_name"`
	LastName             string     `json:"last_name"`
	Phone                *string    `json:"phone,omitempty"`
	DateOfBirth          *time.Time `json:"date_of_birth,omitempty"`
	TigerBeetleAccountID *int64     `json:"tigerbeetle_account_id,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`
	IsActive             bool       `json:"is_active"`
	EmailVerified        bool       `json:"email_verified"`
}

// ToResponse convierte un User a UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:                   u.ID,
		Email:                u.Email,
		FirstName:            u.FirstName,
		LastName:             u.LastName,
		Phone:                u.Phone,
		DateOfBirth:          u.DateOfBirth,
		TigerBeetleAccountID: u.TigerBeetleAccountID,
		CreatedAt:            u.CreatedAt,
		UpdatedAt:            u.UpdatedAt,
		IsActive:             u.IsActive,
		EmailVerified:        u.EmailVerified,
	}
}
