package domain

import (
	"time"

	"github.com/google/uuid"
)

// OutgoingDoc represents an outgoing document
type OutgoingDoc struct {
	ID           uuid.UUID  `json:"id" db:"id"`
	DocDetailsID uuid.UUID  `json:"doc_details_id" db:"doc_details_id" validate:"required"`
	FolderID     *uuid.UUID `json:"folder_id" db:"folder_id"`
	CreatedBy    *uuid.UUID `json:"created_by" db:"created_by"`
	UpdatedBy    *uuid.UUID `json:"updated_by" db:"updated_by"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`
	// Status is the owner department head's approval gate:
	// pending|approved|rejected (see Outgoing*Status constants).
	Status string `json:"status" db:"status"`

	// Joined fields
	DocNo       string `json:"doc_no,omitempty" db:"doc_no"`
	DocName     string `json:"doc_name,omitempty" db:"doc_name"`
	DocPath     string `json:"doc_path,omitempty" db:"doc_path"`
	FileType    string `json:"file_type,omitempty" db:"file_type"`
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

// OutgoingDocRoute represents one ordered step in an outgoing document's
// department flow. sequence_order 1 = owner department, then selected
// departments in order, then Director Office (if requested).
type OutgoingDocRoute struct {
	ID            uuid.UUID  `json:"id" db:"id"`
	OutgoingDocID uuid.UUID  `json:"outgoing_doc_id" db:"outgoing_doc_id"`
	DeptID        uuid.UUID  `json:"dept_id" db:"dept_id"`
	SequenceOrder int        `json:"sequence_order" db:"sequence_order"`
	IncomingDocID *uuid.UUID `json:"incoming_doc_id" db:"incoming_doc_id"`
}

// RouteStep is the lightweight view used to advance the flow.
type RouteStep struct {
	ID            uuid.UUID
	OutgoingDocID uuid.UUID
	DeptID        uuid.UUID
	SequenceOrder int
	IncomingDocID *uuid.UUID
}

// Outgoing document status — the owner department head's approval gate.
const (
	OutgoingStatusPending  = "pending"
	OutgoingStatusApproved = "approved"
	OutgoingStatusRejected = "rejected"
)

// Flow status values for an outgoing document
const (
	FlowStatusInProgress = "in_progress"
	FlowStatusCompleted  = "completed"
	FlowStatusRejected   = "rejected"
	// FlowStatusPendingApproval means the owner department head has not yet
	// approved the outgoing document, so no recipient incoming docs exist.
	FlowStatusPendingApproval = "pending_approval"
	// StatusWaiting is a synthetic recipient status for a route step whose
	// incoming document has not been created yet (waiting for the previous step).
	StatusWaiting = "waiting"
)

// RecipientInfo holds per-department tracking for an outgoing document.
// Built from outgoing_doc_routes LEFT JOIN incoming_docs, ordered by sequence_order.
type RecipientInfo struct {
	DepartmentID   uuid.UUID  `json:"department_id"`
	DepartmentName string     `json:"department_name"`
	Status         string     `json:"status"` // pending|received|approved|rejected|waiting
	SequenceOrder  int        `json:"sequence_order"`
	IsCurrent      bool       `json:"is_current"`
	ReceivedDate   *time.Time `json:"received_date"`
	ApproverDate   *time.Time `json:"approver_date"`
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
	DocDetailsID uuid.UUID      `json:"doc_details_id"`
	DocNo        string         `json:"doc_no,omitempty"`
	DocName      string         `json:"doc_name,omitempty"`
	DocPath      string         `json:"doc_path,omitempty"`
	FileType     string         `json:"file_type,omitempty"`
	FolderID     *uuid.UUID     `json:"folder_id"`
	CreatedBy    *uuid.UUID     `json:"created_by"`
	CreatorName  string         `json:"creator_name,omitempty"`
	UpdatedBy    *uuid.UUID     `json:"updated_by"`
	UpdaterName  string         `json:"updater_name,omitempty"`
	CreatedAt    time.Time      `json:"created_at"`
	Status       string         `json:"status"` // owner approval gate: pending|approved|rejected
	Recipients   []RecipientInfo `json:"recipients"`
	StatusCounts StatusCounts   `json:"status_counts"`

	// Flow tracking — where the document currently sits and its overall state.
	CurrentDepartment string `json:"current_department"`
	CurrentStatus     string `json:"current_status"` // pending|received|rejected|"" (completed)
	FlowStatus        string `json:"flow_status"`    // in_progress|completed|rejected
}

// ToResponse converts OutgoingDoc to OutgoingDocResponse (without recipients; service fills them in)
func (o *OutgoingDoc) ToResponse() OutgoingDocResponse {
	return OutgoingDocResponse{
		ID:           o.ID,
		DocDetailsID: o.DocDetailsID,
		FolderID:     o.FolderID,
		CreatedBy:    o.CreatedBy,
		UpdatedBy:    o.UpdatedBy,
		CreatedAt:    o.CreatedAt,
		DocNo:        o.DocNo,
		DocName:      o.DocName,
		DocPath:      o.DocPath,
		FileType:     o.FileType,
		CreatorName:  o.CreatorName,
		UpdaterName:  o.UpdaterName,
		Status:       o.Status,
		Recipients:   []RecipientInfo{},
	}
}
