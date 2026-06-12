package domain

import (
	"time"

	"github.com/google/uuid"
)

// OutgoingDoc represents an outgoing document
type OutgoingDoc struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	OutgoingNo   string     `json:"outgoing_no" db:"outgoing_no" validate:"required"`
	DocDetailsID uuid.UUID  `json:"doc_details_id" db:"doc_details_id" validate:"required"`
	FolderID     *uuid.UUID `json:"folder_id" db:"folder_id"`
	CreatedBy    *uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedBy    *uuid.UUID `json:"updated_by" db:"updated_by"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`

	// Joined fields
	DocNo       string `json:"doc_no,omitempty" db:"doc_no"`
	DocName     string `json:"doc_name,omitempty" db:"doc_name"`
	DocPath     string `json:"doc_path,omitempty" db:"doc_path"`
	CreatorName string `json:"creator_name,omitempty" db:"creator_name"`
	UpdaterName string `json:"updater_name,omitempty" db:"updater_name"`
}

// CreateOutgoingDocRequest represents the request body for creating an outgoing document
type CreateOutgoingDocRequest struct {
	DocDetailsID uuid.UUID  `json:"doc_details_id" validate:"required"`
	FolderID     *uuid.UUID `json:"folder_id"`
	CreatedBy    *uuid.UUID `json:"created_by"`
	UpdatedBy    *uuid.UUID `json:"updated_by"`
}

// RecipientInfo holds per-department tracking for an outgoing document
type RecipientInfo struct {
	DepartmentID   uuid.UUID         `json:"department_id"`
	DepartmentName string            `json:"department_name"`
	IncomingNo     string            `json:"incoming_no"`
	Status         IncomingDocStatus `json:"status"`
	ReceivedDate   *time.Time        `json:"received_date"`
	ApproverDate   *time.Time        `json:"approver_date"`
}

// StatusCounts holds the count of incoming docs grouped by status
type StatusCounts struct {
	Total    int `json:"total"`
	Pending  int `json:"pending"`
	Received int `json:"received"`
	Approved int `json:"approved"`
	Rejected int `json:"rejected"`
}

// OutgoingDocResponse represents the outgoing document response
type OutgoingDocResponse struct {
	ID           uuid.UUID      `json:"id"`
	OutgoingNo   string         `json:"outgoing_no"`
	DocDetailsID uuid.UUID      `json:"doc_details_id"`
	DocNo        string         `json:"doc_no,omitempty"`
	DocName      string         `json:"doc_name,omitempty"`
	DocPath      string         `json:"doc_path,omitempty"`
	FolderID     *uuid.UUID     `json:"folder_id"`
	CreatedBy    *uuid.UUID     `json:"created_by"`
	CreatorName  string         `json:"creator_name,omitempty"`
	UpdatedBy    *uuid.UUID     `json:"updated_by"`
	UpdaterName  string         `json:"updater_name,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	Recipients   []RecipientInfo `json:"recipients"`
	StatusCounts StatusCounts   `json:"status_counts"`
}

// ToResponse converts OutgoingDoc to OutgoingDocResponse (without recipients; service fills them in)
func (o *OutgoingDoc) ToResponse() OutgoingDocResponse {
	return OutgoingDocResponse{
		ID:           o.ID,
		OutgoingNo:   o.OutgoingNo,
		DocDetailsID: o.DocDetailsID,
		FolderID:     o.FolderID,
		CreatedBy:    o.CreatedBy,
		UpdatedBy:    o.UpdatedBy,
		CreatedAt:    o.CreatedAt,
		DocNo:        o.DocNo,
		DocName:      o.DocName,
		DocPath:      o.DocPath,
		CreatorName:  o.CreatorName,
		UpdaterName:  o.UpdaterName,
		Recipients:   []RecipientInfo{},
	}
}
