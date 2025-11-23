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
	RoleAdmin  UserRole = "admin"
	RoleEditor UserRole = "editor"
	RoleViewer UserRole = "viewer"
)

// IsValid checks if the role is valid
func (r UserRole) IsValid() bool {
	switch r {
	case RoleAdmin, RoleEditor, RoleViewer:
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
		return "", errors.New("invalid role: must be admin, editor, or viewer")
	}
	return r, nil
}

// User represents the user model in the system
type User struct {
	ID        primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Username  string             `json:"username" bson:"username" validate:"required"`
	Email     string             `json:"email" bson:"email" validate:"required,email"`
	Password  string             `json:"password,omitempty" bson:"password" validate:"required,min=6"`
	Role      UserRole           `json:"role" bson:"role" validate:"required,oneof=admin editor viewer"`
	CreatedAt time.Time          `json:"created_at" bson:"created_at"`
	UpdatedAt time.Time          `json:"updated_at" bson:"updated_at"`
}

// CreateUserRequest represents the request body for creating a user
type CreateUserRequest struct {
	Username string   `json:"username" validate:"required"`
	Email    string   `json:"email" validate:"required,email"`
	Password string   `json:"password" validate:"required,min=6"`
	Role     UserRole `json:"role" validate:"required,oneof=admin editor viewer"`
}

// UpdateUserRequest represents the request body for updating a user
type UpdateUserRequest struct {
	Username string   `json:"username,omitempty"`
	Email    string   `json:"email,omitempty"`
	Role     UserRole `json:"role,omitempty" validate:"omitempty,oneof=admin editor viewer"`
}


// UserResponse represents the user response (without password)
type UserResponse struct {
	ID        primitive.ObjectID `json:"id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	Role      UserRole           `json:"role"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}

// ToResponse converts User to UserResponse (excluding password)
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		Username:  u.Username,
		Email:     u.Email,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
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
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Type     string `json:"type"` // "access" or "refresh"
}
