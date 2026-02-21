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
	DocID        uuid.UUID         `json:"doc_id" db:"doc_id" validate:"required"`
	SenderID     *uuid.UUID        `json:"sender_id" db:"sender_id"`
	ReceiverID   *uuid.UUID        `json:"receiver_id" db:"receiver_id"`
	ApproverID   *uuid.UUID        `json:"approver_id" db:"approver_id"`
	ReceivedDate *time.Time        `json:"received_date" db:"received_date"`
	ApproverDate *time.Time        `json:"approver_date" db:"approver_date"`
	Remark       string            `json:"remark" db:"remark"`
	Status       IncomingDocStatus `json:"status" db:"status"`
	CreatedAt    time.Time         `json:"created_at" db:"created_at"`
}

// CreateIncomingDocRequest represents the request body for creating an incoming document
type CreateIncomingDocRequest struct {
	DocID      uuid.UUID  `json:"doc_id" validate:"required"`
	SenderID   *uuid.UUID `json:"sender_id"`
	ReceiverID *uuid.UUID `json:"receiver_id"`
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
	DocID        uuid.UUID         `json:"doc_id"`
	DocNo        string            `json:"doc_no,omitempty"`
	DocName      string            `json:"doc_name,omitempty"`
	SenderID     *uuid.UUID        `json:"sender_id"`
	SenderName   string            `json:"sender_name,omitempty"`
	ReceiverID   *uuid.UUID        `json:"receiver_id"`
	ReceiverName string            `json:"receiver_name,omitempty"`
	ApproverID   *uuid.UUID        `json:"approver_id"`
	ApproverName string            `json:"approver_name,omitempty"`
	ReceivedDate *time.Time        `json:"received_date"`
	ApproverDate *time.Time        `json:"approver_date"`
	Remark       string            `json:"remark"`
	Status       IncomingDocStatus `json:"status"`
	CreatedAt    time.Time         `json:"created_at"`
}

// ToResponse converts IncomingDoc to IncomingDocResponse
func (i *IncomingDoc) ToResponse() IncomingDocResponse {
	return IncomingDocResponse{
		ID:           i.ID,
		IncomingNo:   i.IncomingNo,
		DocID:        i.DocID,
		SenderID:     i.SenderID,
		ReceiverID:   i.ReceiverID,
		ApproverID:   i.ApproverID,
		ReceivedDate: i.ReceivedDate,
		ApproverDate: i.ApproverDate,
		Remark:       i.Remark,
		Status:       i.Status,
		CreatedAt:    i.CreatedAt,
	}
}
