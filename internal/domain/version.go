package domain

import (
	"time"

	"github.com/google/uuid"
)

// Version represents a document version
type Version struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	DocID         uuid.UUID  `json:"doc_id" db:"doc_id" validate:"required"`
	FolderID      *uuid.UUID `json:"folder_id" db:"folder_id" validate:"omitempty"` // Optional (nullable)
	VersionNumber int        `json:"version_number" db:"version_number" validate:"required"`
	DocPath       string     `json:"doc_path" db:"doc_path" validate:"required"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
}

// CreateVersionRequest represents the request body for creating a version
type CreateVersionRequest struct {
	DocID         uuid.UUID `json:"doc_id" validate:"required"`
	VersionNumber int       `json:"version_number" validate:"required"`
	DocPath       string    `json:"doc_path" validate:"required"`
}

// VersionResponse represents the version response
type VersionResponse struct {
	ID            uuid.UUID  `json:"id"`
	DocID         uuid.UUID  `json:"doc_id"`
	FolderID      *uuid.UUID `json:"folder_id,omitempty"`
	VersionNumber int        `json:"version_number"`
	DocPath       string     `json:"doc_path"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ToResponse converts Version to VersionResponse
func (v *Version) ToResponse() VersionResponse {
	return VersionResponse{
		ID:            v.ID,
		DocID:         v.DocID,
		FolderID:      v.FolderID,
		VersionNumber: v.VersionNumber,
		DocPath:       v.DocPath,
		CreatedAt:     v.CreatedAt,
	}
}
