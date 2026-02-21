package domain

import (
	"time"

	"github.com/google/uuid"
)

// DocumentStatus represents the status of a document
type DocumentStatus string

const (
	StatusNone            DocumentStatus = "none"
	StatusPending         DocumentStatus = "pending"
	StatusWaitingApproval DocumentStatus = "waiting_approval"
	StatusApproved        DocumentStatus = "approved"
)

// IsValid checks if the document status is valid
func (s DocumentStatus) IsValid() bool {
	switch s {
	case StatusNone, StatusPending, StatusWaitingApproval, StatusApproved:
		return true
	}
	return false
}

// Document represents a document in the system
type Document struct {
	ID             uuid.UUID      `json:"id" db:"id"`
	DocNo          string         `json:"doc_no" db:"doc_no" validate:"required"`
	DocName        string         `json:"doc_name" db:"doc_name" validate:"required"`
	DocPath        string         `json:"doc_path" db:"doc_path"`
	Type           string         `json:"type" db:"type"` // file extension (doc, pdf, excel)
	DocTypeID      *uuid.UUID     `json:"doc_type_id" db:"doc_type_id"`
	FolderID       *uuid.UUID     `json:"folder_id" db:"folder_id"`
	RegistrantID   *uuid.UUID     `json:"registrant_id" db:"registrant_id"`
	Status         DocumentStatus `json:"status" db:"status"`
	VersionNumber  int            `json:"version_number" db:"version_number"`
	Description    *string        `json:"description" db:"description"`
	SendToDirector bool           `json:"send_to_director" db:"send_to_director"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

// CreateDocumentRequest represents the request body for creating a document
type CreateDocumentRequest struct {
	DocName        string      `json:"doc_name" form:"title" validate:"required"` // Added form tag
	DocTypeID      *uuid.UUID  `json:"doc_type_id" form:"doc_type_id"`
	FolderID       *uuid.UUID  `json:"folder_id" form:"folder_id"`
	Description    string      `json:"description" form:"description"`
	SendToDirector bool        `json:"send_to_director" form:"send_to_director"`
	DepartmentIDs  []uuid.UUID `json:"department_ids" form:"department_ids"`
	DocPath        string      `json:"doc_path"` // Not from form directly, set by handler
}

// UpdateDocumentRequest represents the request body for updating a document
type UpdateDocumentRequest struct {
	DocName        string      `json:"doc_name,omitempty" form:"title"`
	DocTypeID      *uuid.UUID  `json:"doc_type_id,omitempty" form:"doc_type_id"`
	FolderID       *uuid.UUID  `json:"folder_id,omitempty" form:"folder_id"`
	Description    string      `json:"description,omitempty" form:"description"`
	SendToDirector *bool       `json:"send_to_director,omitempty" form:"send_to_director"`
	DepartmentIDs  []uuid.UUID `json:"department_ids,omitempty" form:"department_ids"`
	DocPath        string      `json:"doc_path,omitempty"` // Not from form directly, set by handler
}

// SendDocumentRequest represents the request body for sending a document
type SendDocumentRequest struct {
	ReceiverID   string `json:"receiver_id"`   // Can be User UUID or Department ID (if implementing logic for that)
	ReceiverType string `json:"receiver_type"` // "user" or "department"
	Remark       string `json:"remark"`
}

// DepartmentApprovalStatus represents approval status for a department
type DepartmentApprovalStatus struct {
	DepartmentID   uuid.UUID  `json:"department_id"`
	DepartmentName string     `json:"department_name"`
	Received       bool       `json:"received"`
	ReceivedDate   *time.Time `json:"received_date,omitempty"`
	Approved       bool       `json:"approved"`
	ApprovedDate   *time.Time `json:"approved_date,omitempty"`
}

// DocumentResponse represents the document response
type DocumentResponse struct {
	ID              uuid.UUID                  `json:"id"`
	DocNo           string                     `json:"doc_no"`
	DocName         string                     `json:"doc_name"`
	DocPath         string                     `json:"doc_path"`
	Type            string                     `json:"type"`
	DocTypeID       *uuid.UUID                 `json:"doc_type_id"`
	DocTypeName     string                     `json:"doc_type_name,omitempty"`
	FolderID        *uuid.UUID                 `json:"folder_id"`
	FolderName      string                     `json:"folder_name,omitempty"`
	RegistrantID    *uuid.UUID                 `json:"registrant_id,omitempty"`
	RegistrantName  string                     `json:"registrant_name,omitempty"`
	RegistrantEmail string                     `json:"registrant_email,omitempty"`
	DepartmentName  string                     `json:"department_name,omitempty"`
	SectorName      string                     `json:"sector_name,omitempty"`
	Status          DocumentStatus             `json:"status"`
	VersionNumber   int                        `json:"version_number"`
	Description     string                     `json:"description"`
	SendToDirector  bool                       `json:"send_to_director"`
	ApprovalStatus  []DepartmentApprovalStatus `json:"approval_status,omitempty"`
	CreatedAt       time.Time                  `json:"created_at"`
	UpdatedAt       time.Time                  `json:"updated_at"`
}

// ToResponse converts Document to DocumentResponse
func (d *Document) ToResponse() DocumentResponse {
	res := DocumentResponse{
		ID:             d.ID,
		DocNo:          d.DocNo,
		DocName:        d.DocName,
		DocPath:        d.DocPath,
		Type:           d.Type,
		DocTypeID:      d.DocTypeID,
		FolderID:       d.FolderID,
		RegistrantID:   d.RegistrantID,
		Status:         d.Status,
		VersionNumber:  d.VersionNumber,
		SendToDirector: d.SendToDirector,
		CreatedAt:      d.CreatedAt,
		UpdatedAt:      d.UpdatedAt,
	}
	if d.Description != nil {
		res.Description = *d.Description
	}
	return res
}

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
	FolderName     string     `json:"folder_name" validate:"required"`
	FolderPath     string     `json:"folder_path" validate:"required"`
	ParentFolderID *uuid.UUID `json:"parent_folder_id"`
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

// DocumentAttachment represents a file attachment associated with a document.
// Used by the tusd upload pipeline.
type DocumentAttachment struct {
	ID         uuid.UUID `json:"id"          db:"id"`
	DocumentID uuid.UUID `json:"document_id" db:"document_id"`
	FileName   string    `json:"file_name"   db:"file_name"`
	FilePath   string    `json:"file_path"   db:"file_path"`
	FileSize   int64     `json:"file_size"   db:"file_size"`
	FileType   string    `json:"file_type"   db:"file_type"`
	Version    int       `json:"version"     db:"version"`
	IsCurrent  bool      `json:"is_current"  db:"is_current"`
	UploadedBy uuid.UUID `json:"uploaded_by" db:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"  db:"created_at"`
}

// DocumentAttachmentResponse is the public representation of a DocumentAttachment.
type DocumentAttachmentResponse struct {
	ID         uuid.UUID `json:"id"`
	DocumentID uuid.UUID `json:"document_id"`
	FileName   string    `json:"file_name"`
	FilePath   string    `json:"file_path"`
	FileSize   int64     `json:"file_size"`
	FileType   string    `json:"file_type"`
	Version    int       `json:"version"`
	IsCurrent  bool      `json:"is_current"`
	UploadedBy uuid.UUID `json:"uploaded_by"`
	CreatedAt  time.Time `json:"created_at"`
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
