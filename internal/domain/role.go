package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserRole represents a user role in the system
type UserRole struct {
	ID          uuid.UUID `json:"id" db:"id"`
	RoleName    string    `json:"role_name" db:"role_name" validate:"required"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateRoleRequest represents the request body for creating a role
type CreateRoleRequest struct {
	RoleName    string `json:"role_name" validate:"required"`
	Description string `json:"description"`
}

// UpdateRoleRequest represents the request body for updating a role
type UpdateRoleRequest struct {
	RoleName    string `json:"role_name,omitempty"`
	Description string `json:"description,omitempty"`
}

// RoleResponse represents the role response
type RoleResponse struct {
	ID          uuid.UUID `json:"id"`
	RoleName    string    `json:"role_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts UserRole to RoleResponse
func (r *UserRole) ToResponse() RoleResponse {
	return RoleResponse{
		ID:          r.ID,
		RoleName:    r.RoleName,
		Description: r.Description,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
