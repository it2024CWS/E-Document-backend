package domain

import (
	"time"

	"github.com/google/uuid"
) // User represents the user model in the system
type User struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Username       string     `json:"username" db:"username" validate:"required"`
	Email          string     `json:"email" db:"email" validate:"required,email"`
	Phone          string     `json:"phone" db:"phone"`
	Firstname      string     `json:"firstname" db:"firstname"`
	Lastname       string     `json:"lastname" db:"lastname"`
	Nickname       string     `json:"nickname" db:"nickname"`
	Password       string     `json:"password,omitempty" db:"password" validate:"required,min=6"`
	RoleID         *uuid.UUID `json:"role_id" db:"role_id"`
	DepartmentID   *uuid.UUID `json:"department_id" db:"department_id"`
	SectorID       *uuid.UUID `json:"sector_id" db:"sector_id"`
	IsActive       bool       `json:"is_active" db:"is_active"`
	ProfilePicture string     `json:"profile_picture,omitempty" db:"profile_picture"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`

	// Joined fields
	RoleName       string `json:"role_name,omitempty" db:"role_name"`
	DepartmentName string `json:"department_name,omitempty" db:"department_name"`
	SectorName     string `json:"sector_name,omitempty" db:"sector_name"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Username     string     `json:"username" validate:"required"`
	Email        string     `json:"email" validate:"required,email"`
	Password     string     `json:"password" validate:"required,min=6"`
	RoleID       *uuid.UUID `json:"role_id" validate:"required"`
	Phone        string     `json:"phone"`
	Firstname    string     `json:"firstname"`
	Lastname     string     `json:"lastname"`
	Nickname     string     `json:"nickname"`
	DepartmentID *uuid.UUID `json:"department_id"`
	SectorID     *uuid.UUID `json:"sector_id"`
	IsActive     *bool      `json:"is_active"` // Defaults to true in database
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Username     string     `json:"username,omitempty"`
	Email        string     `json:"email,omitempty"`
	RoleID       *uuid.UUID `json:"role_id,omitempty"`
	Phone        string     `json:"phone,omitempty"`
	Firstname    string     `json:"firstname,omitempty"`
	Lastname     string     `json:"lastname,omitempty"`
	Nickname     string     `json:"nickname,omitempty"`
	DepartmentID *uuid.UUID `json:"department_id,omitempty"`
	SectorID     *uuid.UUID `json:"sector_id,omitempty"`
	IsActive     *bool      `json:"is_active,omitempty"`
	Password     string     `json:"password,omitempty" validate:"omitempty,min=6"`
}

// UserResponse represents the user response (without password)
type UserResponse struct {
	ID             uuid.UUID  `json:"id"`
	Username       string     `json:"username"`
	Email          string     `json:"email"`
	RoleID         *uuid.UUID `json:"role_id"`
	RoleName       string     `json:"role_name,omitempty"` // Joined from user_roles table
	Phone          string     `json:"phone"`
	Firstname      string     `json:"firstname"`
	Lastname       string     `json:"lastname"`
	Nickname       string     `json:"nickname"`
	ProfilePicture string     `json:"profile_picture,omitempty"`
	DepartmentID   *uuid.UUID `json:"department_id"`
	DepartmentName string     `json:"department_name,omitempty"` // Joined from departments table
	SectorID       *uuid.UUID `json:"sector_id"`
	SectorName     string     `json:"sector_name,omitempty"` // Joined from sectors table
	IsActive       bool       `json:"is_active"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ToResponse converts User to UserResponse (excluding password)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:             u.ID,
		Username:       u.Username,
		Email:          u.Email,
		RoleID:         u.RoleID,
		RoleName:       u.RoleName,
		Phone:          u.Phone,
		Firstname:      u.Firstname,
		Lastname:       u.Lastname,
		Nickname:       u.Nickname,
		ProfilePicture: u.ProfilePicture,
		DepartmentID:   u.DepartmentID,
		DepartmentName: u.DepartmentName,
		SectorID:       u.SectorID,
		SectorName:     u.SectorName,
		IsActive:       u.IsActive,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
}

// Auth-related structs

// LoginRequest represents the request body for user login
type LoginRequest struct {
	UsernameOrEmail string `json:"usernameOrEmail" validate:"required"`
	Password        string `json:"password" validate:"required"`
}

// RefreshTokenRequest represents the request body for refreshing token
type RefreshTokenRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

// AuthResponse represents the authentication response
type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"accessToken,omitempty"`  // For mobile apps
	RefreshToken string       `json:"refreshToken,omitempty"` // For mobile apps
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID       string     `json:"user_id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	Phone        string     `json:"phone"`
	Firstname    string     `json:"firstname"`
	Lastname     string     `json:"lastname"`
	Nickname     string     `json:"nickname"`
	RoleID       *uuid.UUID `json:"role_id"`
	RoleName     string     `json:"role_name"`
	DepartmentID *uuid.UUID `json:"department_id"`
	SectorID     *uuid.UUID `json:"sector_id"`
	Type         string     `json:"type"` // "access" or "refresh"
}
