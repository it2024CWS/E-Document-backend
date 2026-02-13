package domain

import "time"

// Version represents a document version
type Version struct {
	ID            int       `json:"id" db:"id"`
	DocID         int       `json:"doc_id" db:"doc_id" validate:"required"`
	VersionNumber int       `json:"version_number" db:"version_number" validate:"required"`
	DocPath       string    `json:"doc_path" db:"doc_path" validate:"required"`
	CreatedAt     time.Time `json:"created_at" db:"created_at"`
}

// CreateVersionRequest represents the request body for creating a version
type CreateVersionRequest struct {
	DocID         int    `json:"doc_id" validate:"required"`
	VersionNumber int    `json:"version_number" validate:"required"`
	DocPath       string `json:"doc_path" validate:"required"`
}

// VersionResponse represents the version response
type VersionResponse struct {
	ID            int       `json:"id"`
	DocID         int       `json:"doc_id"`
	VersionNumber int       `json:"version_number"`
	DocPath       string    `json:"doc_path"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToResponse converts Version to VersionResponse
func (v *Version) ToResponse() VersionResponse {
	return VersionResponse{
		ID:            v.ID,
		DocID:         v.DocID,
		VersionNumber: v.VersionNumber,
		DocPath:       v.DocPath,
		CreatedAt:     v.CreatedAt,
	}
}
