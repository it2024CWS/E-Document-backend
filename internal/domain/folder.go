package domain

import (
	"time"

	"github.com/google/uuid"
)

// DocumentType represents the type of document
type DocumentType string

const (
	DocumentTypeGeneral DocumentType = "General"
	DocumentTypeBarcode DocumentType = "Barcode"
)

// IsValid checks if the document type is valid
func (dt DocumentType) IsValid() bool {
	switch dt {
	case DocumentTypeGeneral, DocumentTypeBarcode:
		return true
	}
	return false
}

// DocumentStatus represents the status of a document
type DocumentStatus string

const (
	DocumentStatusDraft    DocumentStatus = "Draft"
	DocumentStatusPending  DocumentStatus = "Pending"
	DocumentStatusApproved DocumentStatus = "Approved"
	DocumentStatusRejected DocumentStatus = "Rejected"
)

// IsValid checks if the document status is valid
func (ds DocumentStatus) IsValid() bool {
	switch ds {
	case DocumentStatusDraft, DocumentStatusPending, DocumentStatusApproved, DocumentStatusRejected:
		return true
	}
	return false
}

// Folder represents a folder in the hierarchical structure
type Folder struct {
	ID             uuid.UUID  `json:"id" db:"id"`
	Name           string     `json:"name" db:"name"`
	Path           string     `json:"path" db:"path"`
	IsRootFolder   bool       `json:"is_root_folder" db:"is_root_folder"`
	ParentFolderID *uuid.UUID `json:"parent_folder_id,omitempty" db:"parent_folder_id"`
	OwnerID        uuid.UUID  `json:"owner_id" db:"owner_id"`
	CreatedAt      time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at" db:"updated_at"`
}

// Document represents a document in the system
type Document struct {
	ID                  uuid.UUID      `json:"id" db:"id"`
	Title               string         `json:"title" db:"title"`
	Description         string         `json:"description,omitempty" db:"description"`
	Type                DocumentType   `json:"type" db:"type"`
	CategoryID          *uuid.UUID     `json:"category_id,omitempty" db:"category_id"`
	FolderID            *uuid.UUID     `json:"folder_id,omitempty" db:"folder_id"`
	Barcode             *string        `json:"barcode,omitempty" db:"barcode"`
	RegistrantID        *uuid.UUID     `json:"registrant_id,omitempty" db:"registrant_id"`
	CurrentDepartmentID *uuid.UUID     `json:"current_department_id,omitempty" db:"current_department_id"`
	Status              DocumentStatus `json:"status" db:"status"`
	CreatedAt           time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at" db:"updated_at"`
}

// DocumentAttachment represents a file attachment to a document
type DocumentAttachment struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	DocumentID uuid.UUID  `json:"document_id" db:"document_id"`
	FileName   string     `json:"file_name" db:"file_name"`
	FilePath   string     `json:"file_path" db:"file_path"`
	FileSize   int64      `json:"file_size" db:"file_size"`
	FileType   string     `json:"file_type,omitempty" db:"file_type"`
	Version    int        `json:"version" db:"version"`
	IsCurrent  bool       `json:"is_current" db:"is_current"`
	UploadedBy *uuid.UUID `json:"uploaded_by,omitempty" db:"uploaded_by"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
}

// FolderResponse represents the folder response
type FolderResponse struct {
	ID             uuid.UUID  `json:"id"`
	Name           string     `json:"name"`
	Path           string     `json:"path"`
	IsRootFolder   bool       `json:"is_root_folder"`
	ParentFolderID *uuid.UUID `json:"parent_folder_id,omitempty"`
	OwnerID        uuid.UUID  `json:"owner_id"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// DocumentResponse represents the document response
type DocumentResponse struct {
	ID                  uuid.UUID      `json:"id"`
	Title               string         `json:"title"`
	Description         string         `json:"description,omitempty"`
	Type                DocumentType   `json:"type"`
	CategoryID          *uuid.UUID     `json:"category_id,omitempty"`
	FolderID            *uuid.UUID     `json:"folder_id,omitempty"`
	Barcode             *string        `json:"barcode,omitempty"`
	RegistrantID        *uuid.UUID     `json:"registrant_id,omitempty"`
	CurrentDepartmentID *uuid.UUID     `json:"current_department_id,omitempty"`
	Status              DocumentStatus `json:"status"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

// DocumentAttachmentResponse represents the attachment response
type DocumentAttachmentResponse struct {
	ID         uuid.UUID  `json:"id"`
	DocumentID uuid.UUID  `json:"document_id"`
	FileName   string     `json:"file_name"`
	FilePath   string     `json:"file_path"`
	FileSize   int64      `json:"file_size"`
	FileType   string     `json:"file_type,omitempty"`
	Version    int        `json:"version"`
	IsCurrent  bool       `json:"is_current"`
	UploadedBy *uuid.UUID `json:"uploaded_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ToResponse converts Folder to FolderResponse
func (f *Folder) ToResponse() FolderResponse {
	return FolderResponse{
		ID:             f.ID,
		Name:           f.Name,
		Path:           f.Path,
		IsRootFolder:   f.IsRootFolder,
		ParentFolderID: f.ParentFolderID,
		OwnerID:        f.OwnerID,
		CreatedAt:      f.CreatedAt,
		UpdatedAt:      f.UpdatedAt,
	}
}

// ToResponse converts Document to DocumentResponse
func (d *Document) ToResponse() DocumentResponse {
	return DocumentResponse{
		ID:                  d.ID,
		Title:               d.Title,
		Description:         d.Description,
		Type:                d.Type,
		CategoryID:          d.CategoryID,
		FolderID:            d.FolderID,
		Barcode:             d.Barcode,
		RegistrantID:        d.RegistrantID,
		CurrentDepartmentID: d.CurrentDepartmentID,
		Status:              d.Status,
		CreatedAt:           d.CreatedAt,
		UpdatedAt:           d.UpdatedAt,
	}
}

// ToResponse converts DocumentAttachment to DocumentAttachmentResponse
func (a *DocumentAttachment) ToResponse() DocumentAttachmentResponse {
	return DocumentAttachmentResponse{
		ID:         a.ID,
		DocumentID: a.DocumentID,
		FileName:   a.FileName,
		FilePath:   a.FilePath,
		FileSize:   a.FileSize,
		FileType:   a.FileType,
		Version:    a.Version,
		IsCurrent:  a.IsCurrent,
		UploadedBy: a.UploadedBy,
		CreatedAt:  a.CreatedAt,
	}
}
