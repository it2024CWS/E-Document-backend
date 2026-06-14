package domain

import (
	"time"

	"github.com/google/uuid"
)

// IncomingDocStatus represents the status of an incoming document
type IncomingDocStatus string

const (
	IncomingStatusPending  IncomingDocStatus = "pending"
	IncomingStatusReceived IncomingDocStatus = "received"
	IncomingStatusApproved IncomingDocStatus = "approved"
	IncomingStatusRejected IncomingDocStatus = "rejected"
)

// IncomingDoc represents an incoming document
type IncomingDoc struct {
	ID           uuid.UUID         `json:"id" db:"id"`
	IncomingNo   string            `json:"incoming_no" db:"incoming_no" validate:"required"`
	IncomingDate *time.Time        `json:"incoming_date" db:"incoming_date"`
	ReceivedDate *time.Time        `json:"received_date" db:"received_date"`
	Status       IncomingDocStatus `json:"status" db:"status"`
	DocDetailsID uuid.UUID         `json:"doc_details_id" db:"doc_details_id" validate:"required"`
	FolderID     *uuid.UUID        `json:"folder_id" db:"folder_id"`
	CreatedBy    *uuid.UUID        `json:"created_by" db:"created_by"`
	UpdatedBy    *uuid.UUID        `json:"updated_by" db:"updated_by"`
	ApproverID   *uuid.UUID        `json:"approver_id" db:"approver_id"`
	ApproverDate *time.Time        `json:"approver_date" db:"approver_date"`
	Remark       string            `json:"remark" db:"remark"`
	UpdatedAt    time.Time         `json:"updated_at" db:"updated_at"`

	DeptID        *uuid.UUID `json:"dept_id,omitempty" db:"dept_id"`
	OutgoingDocID *uuid.UUID `json:"outgoing_doc_id,omitempty" db:"outgoing_doc_id"`

	// Joined fields
	DocNo        string `json:"doc_no,omitempty" db:"doc_no"`
	DocName      string `json:"doc_name,omitempty" db:"doc_name"`
	DocPath      string `json:"doc_path,omitempty" db:"doc_path"`
	FileType     string `json:"file_type,omitempty" db:"file_type"`
	CreatorName  string `json:"creator_name,omitempty" db:"creator_name"`
	UpdaterName  string `json:"updater_name,omitempty" db:"updater_name"`
	ApproverName string `json:"approver_name,omitempty" db:"approver_name"`
	DeptName     string `json:"dept_name,omitempty" db:"dept_name"`
}

// CreateIncomingDocRequest represents the request body for creating an incoming document
type CreateIncomingDocRequest struct {
	DocDetailsID uuid.UUID  `json:"doc_details_id" validate:"required"`
	FolderID     *uuid.UUID `json:"folder_id"`
	CreatedBy    *uuid.UUID `json:"created_by"`
	UpdatedBy    *uuid.UUID `json:"updated_by"`
}

// ReceiveIncomingDocRequest represents the request for receiving a document (by secretary)
type ReceiveIncomingDocRequest struct {
	Remark string `json:"remark"`
}

// ApproveIncomingDocRequest represents the request for approving a document (by department head)
type ApproveIncomingDocRequest struct {
	Approved bool   `json:"approved"` // true = approved, false = rejected
	Remark   string `json:"remark"`
}

// IncomingDocResponse represents the incoming document response
type IncomingDocResponse struct {
	ID           uuid.UUID         `json:"id"`
	IncomingNo   string            `json:"incoming_no"`
	IncomingDate *time.Time        `json:"incoming_date"`
	ReceivedDate *time.Time        `json:"received_date"`
	Status       IncomingDocStatus `json:"status"`
	DocDetailsID uuid.UUID         `json:"doc_details_id"`
	DocNo        string            `json:"doc_no,omitempty"`
	DocName      string            `json:"doc_name,omitempty"`
	DocPath      string            `json:"doc_path,omitempty"`
	FileType     string            `json:"file_type,omitempty"`
	FolderID     *uuid.UUID        `json:"folder_id"`
	CreatedBy    *uuid.UUID        `json:"created_by"`
	CreatorName  string            `json:"creator_name,omitempty"`
	UpdatedBy    *uuid.UUID        `json:"updated_by"`
	UpdaterName  string            `json:"updater_name,omitempty"`
	ApproverID   *uuid.UUID        `json:"approver_id"`
	ApproverName string            `json:"approver_name,omitempty"`
	ApproverDate *time.Time        `json:"approver_date"`
	Remark       string            `json:"remark"`
	UpdatedAt    time.Time         `json:"updated_at"`
	DeptID       *uuid.UUID        `json:"dept_id,omitempty"`
	DeptName     string            `json:"dept_name,omitempty"`
}

// ToResponse converts IncomingDoc to IncomingDocResponse
func (i *IncomingDoc) ToResponse() IncomingDocResponse {
	return IncomingDocResponse{
		ID:           i.ID,
		IncomingNo:   i.IncomingNo,
		IncomingDate: i.IncomingDate,
		ReceivedDate: i.ReceivedDate,
		Status:       i.Status,
		DocDetailsID: i.DocDetailsID,
		FolderID:     i.FolderID,
		CreatedBy:    i.CreatedBy,
		UpdatedBy:    i.UpdatedBy,
		ApproverID:   i.ApproverID,
		ApproverDate: i.ApproverDate,
		Remark:       i.Remark,
		UpdatedAt:    i.UpdatedAt,
		DocNo:        i.DocNo,
		DocName:      i.DocName,
		DocPath:      i.DocPath,
		FileType:     i.FileType,
		CreatorName:  i.CreatorName,
		UpdaterName:  i.UpdaterName,
		ApproverName: i.ApproverName,
		DeptID:       i.DeptID,
		DeptName:     i.DeptName,
	}
}
