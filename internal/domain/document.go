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

// DocDetails represents a document details record in the system
type DocDetails struct {
	ID            uuid.UUID      `json:"id" db:"id"`
	DocNo         string         `json:"doc_no" db:"doc_no" validate:"required"`
	DocName       string         `json:"doc_name" db:"doc_name" validate:"required"`
	Description   *string        `json:"description" db:"description"`
	VersionNumber int            `json:"version_number" db:"version_number"`
	Status        DocumentStatus `json:"status" db:"status"`
	DocTypeID     *uuid.UUID     `json:"doc_type_id" db:"doc_type_id"`
	UserID        *uuid.UUID     `json:"user_id" db:"user_id"`
	CreatedAt     time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at" db:"updated_at"`
	DeletedAt     *time.Time     `json:"deleted_at,omitempty" db:"deleted_at"`
}

// CreateDocumentRequest represents the request body for creating a document
type CreateDocumentRequest struct {
	DocName     string     `json:"doc_name" form:"title" validate:"required"`
	DocTypeID   *uuid.UUID `json:"doc_type_id" form:"doc_type_id"`
	Description string     `json:"description" form:"description"`
}

// UpdateDocumentRequest represents the request body for updating a document
type UpdateDocumentRequest struct {
	DocName     string     `json:"doc_name,omitempty" form:"title"`
	DocTypeID   *uuid.UUID `json:"doc_type_id,omitempty" form:"doc_type_id"`
	Description string     `json:"description,omitempty" form:"description"`
}

// DocumentResponse represents the document response
type DocumentResponse struct {
	ID          uuid.UUID      `json:"id"`
	DocNo       string         `json:"doc_no"`
	DocName     string         `json:"doc_name"`
	Description string         `json:"description"`
	Version     int            `json:"version_number"`
	Status      DocumentStatus `json:"status"`
	DocTypeID   *uuid.UUID     `json:"doc_type_id,omitempty"`
	DocTypeName string         `json:"doc_type_name,omitempty"`
	UserID      *uuid.UUID     `json:"user_id,omitempty"`
	UserName    string         `json:"user_name,omitempty"`
	FileType    string         `json:"file_type,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ToResponse converts DocDetails to DocumentResponse
func (d *DocDetails) ToResponse() DocumentResponse {
	res := DocumentResponse{
		ID:        d.ID,
		DocNo:     d.DocNo,
		DocName:   d.DocName,
		Version:   d.VersionNumber,
		Status:    d.Status,
		DocTypeID: d.DocTypeID,
		UserID:    d.UserID,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
	if d.Description != nil {
		res.Description = *d.Description
	}
	return res
}

// SendDocumentRequest represents the request body for sending a document
type SendDocumentRequest struct {
	ReceiverID   string `json:"receiver_id"`
	ReceiverType string `json:"receiver_type"`
	Remark       string `json:"remark"`
}

