package domain

import (
	"time"

	"github.com/google/uuid"
)

// Department represents a department in the system
type Department struct {
	ID          uuid.UUID `json:"id" db:"id"`
	DeptName    string    `json:"dept_name" db:"dept_name" validate:"required"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateDepartmentRequest represents the request body for creating a department
type CreateDepartmentRequest struct {
	DeptName    string `json:"dept_name" validate:"required"`
	Description string `json:"description"`
}

// UpdateDepartmentRequest represents the request body for updating a department
type UpdateDepartmentRequest struct {
	DeptName    string `json:"dept_name,omitempty"`
	Description string `json:"description,omitempty"`
}

// DepartmentResponse represents the department response
type DepartmentResponse struct {
	ID          uuid.UUID `json:"id"`
	DeptName    string    `json:"dept_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts Department to DepartmentResponse
func (d *Department) ToResponse() DepartmentResponse {
	return DepartmentResponse{
		ID:          d.ID,
		DeptName:    d.DeptName,
		Description: d.Description,
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}
