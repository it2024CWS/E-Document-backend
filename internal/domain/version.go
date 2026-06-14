package domain

import (
	"time"

	"github.com/google/uuid"
)

// Version represents a document version
type Version struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	DocDetailsID  uuid.UUID  `json:"doc_details_id" db:"doc_details_id" validate:"required"`
	FolderID      *uuid.UUID `json:"folder_id" db:"folder_id" validate:"omitempty"`
	VersionNumber int        `json:"version_number" db:"version_number" validate:"required"`
	DocPath       string     `json:"doc_path" db:"doc_path" validate:"required"`
	FileType      string     `json:"file_type" db:"file_type"`
	CreatedAt     time.Time  `json:"created_at" db:"created_at"`
	DeletedAt     *time.Time `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateVersionRequest represents the request body for creating a version
type CreateVersionRequest struct {
	DocDetailsID  uuid.UUID  `json:"doc_details_id" validate:"required"`
	FolderID      *uuid.UUID `json:"folder_id"`
	VersionNumber int        `json:"version_number" validate:"required"`
	DocPath       string     `json:"doc_path" validate:"required"`
	FileType      string     `json:"file_type"`
}

// VersionResponse represents the version response
type VersionResponse struct {
	ID            uuid.UUID  `json:"id"`
	DocDetailsID  uuid.UUID  `json:"doc_details_id"`
	FolderID      *uuid.UUID `json:"folder_id,omitempty"`
	VersionNumber int        `json:"version_number"`
	DocPath       string     `json:"doc_path"`
	FileType      string     `json:"file_type"`
	CreatedAt     time.Time  `json:"created_at"`
}

// ToResponse converts Version to VersionResponse
func (v *Version) ToResponse() VersionResponse {
	return VersionResponse{
		ID:            v.ID,
		DocDetailsID:  v.DocDetailsID,
		FolderID:      v.FolderID,
		VersionNumber: v.VersionNumber,
		DocPath:       v.DocPath,
		FileType:      v.FileType,
		CreatedAt:     v.CreatedAt,
	}
}
