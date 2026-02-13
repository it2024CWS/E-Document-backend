package domain

import "time"

// Sector represents a sector within a department
type Sector struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name" validate:"required"`
	DeptID    int       `json:"dept_id" db:"dept_id" validate:"required"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// CreateSectorRequest represents the request body for creating a sector
type CreateSectorRequest struct {
	Name   string `json:"name" validate:"required"`
	DeptID int    `json:"dept_id" validate:"required"`
}

// UpdateSectorRequest represents the request body for updating a sector
type UpdateSectorRequest struct {
	Name   string `json:"name,omitempty"`
	DeptID int    `json:"dept_id,omitempty"`
}

// SectorResponse represents the sector response
type SectorResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	DeptID    int       `json:"dept_id"`
	DeptName  string    `json:"dept_name,omitempty"` // Joined from departments table
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts Sector to SectorResponse
func (s *Sector) ToResponse() SectorResponse {
	return SectorResponse{
		ID:        s.ID,
		Name:      s.Name,
		DeptID:    s.DeptID,
		CreatedAt: s.CreatedAt,
		UpdatedAt: s.UpdatedAt,
	}
}
