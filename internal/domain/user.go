package domain

import (
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// UserRole represents the role type
type UserRole string

// Role Enums
const (
	RoleDirector          UserRole = "Director"
	RoleDepartmentManager UserRole = "DepartmentManager"
	RoleSectorManager     UserRole = "SectorManager"
	RoleEmployee          UserRole = "Employee"
)

// IsValid checks if the role is valid
func (r UserRole) IsValid() bool {
	switch r {
	case RoleDirector, RoleDepartmentManager, RoleSectorManager, RoleEmployee:
		return true
	}
	return false
}

// String returns the string representation of the role
func (r UserRole) String() string {
	return string(r)
}

// ValidateRole validates if a string is a valid role
func ValidateRole(role string) (UserRole, error) {
	r := UserRole(role)
	if !r.IsValid() {
		return "", errors.New("invalid role: must be Director, DepartmentManager, SectorManager, or Employee")
	}
	return r, nil
}

// User represents the user model in the system
type User struct {
	ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username     string             `json:"username" bson:"username" validate:"required"`
	Email        string             `json:"email" bson:"email" validate:"required,email"`
	Phone        string             `json:"phone" bson:"phone" validate:"required,e164"`
	FirstName    string             `json:"first_name" bson:"first_name" validate:"required"`
	LastName     string             `json:"last_name" bson:"last_name" validate:"required"`
	Password     string             `json:"password,omitempty" bson:"password" validate:"required,min=6"`
	Role         UserRole           `json:"role" bson:"role" validate:"required,oneof=Director DepartmentManager SectorManager Employee"`
	DepartmentID string             `json:"department_id" bson:"department_id"`
	SectorID     string             `json:"sector_id" bson:"sector_id"`
	CreatedAt    time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Username     string   `json:"username" validate:"required"`
	Email        string   `json:"email" validate:"required,email"`
	Password     string   `json:"password" validate:"required,min=6"`
	Role         UserRole `json:"role" validate:"required,oneof=Director DepartmentManager SectorManager Employee"`
	Phone        string   `json:"phone" validate:"required,e164"`
	FirstName    string   `json:"first_name" validate:"required"`
	LastName     string   `json:"last_name" validate:"required"`
	DepartmentID string   `json:"department_id"`
	SectorID     string   `json:"sector_id"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Username     string   `json:"username,omitempty"`
	Email        string   `json:"email,omitempty"`
	Role         UserRole `json:"role,omitempty" validate:"omitempty,oneof=Director DepartmentManager SectorManager Employee"`
	Phone        string   `json:"phone,omitempty" validate:"omitempty,e164"`
	FirstName    string   `json:"first_name,omitempty"`
	LastName     string   `json:"last_name,omitempty"`
	DepartmentID string   `json:"department_id,omitempty"`
	SectorID     string   `json:"sector_id,omitempty"`
	Password     string   `json:"password,omitempty" validate:"omitempty,min=6"`
}

// UserResponse represents the user response (without password)
type UserResponse struct {
	ID           primitive.ObjectID `json:"id"`
	Username     string             `json:"username"`
	Email        string             `json:"email"`
	Role         UserRole           `json:"role"`
	Phone        string             `json:"phone"`
	FirstName    string             `json:"first_name"`
	LastName     string             `json:"last_name"`
	DepartmentID string             `json:"department_id"`
	SectorID     string             `json:"sector_id"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

// ToResponse converts User to UserResponse (excluding password)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:           u.ID,
		Username:     u.Username,
		Email:        u.Email,
		Role:         u.Role,
		Phone:        u.Phone,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		DepartmentID: u.DepartmentID,
		SectorID:     u.SectorID,
		CreatedAt:    u.CreatedAt,
		UpdatedAt:    u.UpdatedAt,
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
	User UserResponse `json:"user"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID       string `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name"`
	Role         string `json:"role"`
	DepartmentID string `json:"department_id"`
	SectorID     string `json:"sector_id"`
	Type         string `json:"type"` // "access" or "refresh"
}
