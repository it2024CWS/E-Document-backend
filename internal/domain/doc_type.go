package domain

import "time"

// DocType represents a document type in the system
type DocType struct {
	ID          int       `json:"id" db:"id"`
	TypeName    string    `json:"type_name" db:"type_name" validate:"required"`
	Description string    `json:"description" db:"description"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// CreateDocTypeRequest represents the request body for creating a document type
type CreateDocTypeRequest struct {
	TypeName    string `json:"type_name" validate:"required"`
	Description string `json:"description"`
}

// UpdateDocTypeRequest represents the request body for updating a document type
type UpdateDocTypeRequest struct {
	TypeName    string `json:"type_name,omitempty"`
	Description string `json:"description,omitempty"`
}

// DocTypeResponse represents the document type response
type DocTypeResponse struct {
	ID          int       `json:"id"`
	TypeName    string    `json:"type_name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts DocType to DocTypeResponse
func (dt *DocType) ToResponse() DocTypeResponse {
	return DocTypeResponse{
		ID:          dt.ID,
		TypeName:    dt.TypeName,
		Description: dt.Description,
		CreatedAt:   dt.CreatedAt,
		UpdatedAt:   dt.UpdatedAt,
	}
}
