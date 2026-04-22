package domain

import (
	"time"

	"github.com/google/uuid"
)

// Folder represents a folder for organizing documents
type Folder struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	FolderName     string     `json:"folder_name" db:"folder_name" validate:"required"`
	FolderPath     string     `json:"folder_path" db:"folder_path" validate:"required"`
	UserID         uuid.UUID  `json:"user_id" db:"user_id"`
	ParentFolderID *uuid.UUID `json:"parent_folder_id" db:"parent_folder_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}

// CreateFolderRequest represents the request body for creating a folder
type CreateFolderRequest struct {
	FolderName     string `json:"folder_name" validate:"required"`
	FolderPath     string `json:"folder_path" validate:"required"`
	ParentFolderID string `json:"parent_folder_id"`
}

// UpdateFolderRequest represents the request body for updating a folder
type UpdateFolderRequest struct {
	FolderName     string `json:"folder_name"`
	FolderPath     string `json:"folder_path"`
	ParentFolderID string `json:"parent_folder_id"`
}

// FolderResponse represents the folder response
type FolderResponse struct {
	ID             uuid.UUID  `json:"id"`
	FolderName     string     `json:"folder_name"`
	FolderPath     string     `json:"folder_path"`
	UserID         uuid.UUID  `json:"user_id"`
	ParentFolderID *uuid.UUID `json:"parent_folder_id"`
	CreatedAt      time.Time  `json:"created_at"`
}

// ToResponse converts Folder to FolderResponse
func (f *Folder) ToResponse() FolderResponse {
	return FolderResponse{
		ID:             f.ID,
		FolderName:     f.FolderName,
		FolderPath:     f.FolderPath,
		UserID:         f.UserID,
		ParentFolderID: f.ParentFolderID,
		CreatedAt:      f.CreatedAt,
	}
}
