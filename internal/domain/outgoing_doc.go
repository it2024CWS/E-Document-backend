package domain

import (
	"time"

	"github.com/google/uuid"
)

// OutgoingDoc represents an outgoing document
type OutgoingDoc struct {
	ID         uuid.UUID  `json:"id" db:"id"`
	OutgoingNo string     `json:"outgoing_no" db:"outgoing_no" validate:"required"`
	DocID      uuid.UUID  `json:"doc_id" db:"doc_id" validate:"required"`
	UserID       *uuid.UUID `json:"user_id" db:"user_id"`
	DepartmentID *uuid.UUID `json:"dept_id" db:"dept_id"`
	CreatedAt    time.Time  `json:"created_at" db:"created_at"`

	// Joined fields
	DocNo      string `json:"doc_no,omitempty" db:"doc_no"`
	DocName    string `json:"doc_name,omitempty" db:"doc_name"`
	DocPath    string `json:"doc_path,omitempty" db:"doc_path"`
	Type       string `json:"type,omitempty" db:"type"`
	UserName   string `json:"user_name,omitempty" db:"user_name"`
}

// CreateOutgoingDocRequest represents the request body for creating an outgoing document
type CreateOutgoingDocRequest struct {
	DocID  uuid.UUID  `json:"doc_id" validate:"required"`
	UserID *uuid.UUID `json:"user_id"`
}

// OutgoingDocResponse represents the outgoing document response
type OutgoingDocResponse struct {
	ID         uuid.UUID  `json:"id"`
	OutgoingNo string     `json:"outgoing_no"`
	DocID      uuid.UUID  `json:"doc_id"`
	DocNo      string     `json:"doc_no,omitempty"`
	DocName    string     `json:"doc_name,omitempty"`
	DocPath    string     `json:"doc_path,omitempty"`
	Type       string     `json:"type,omitempty"`
	UserID       *uuid.UUID `json:"user_id"`
	UserName     string     `json:"user_name,omitempty"`
	DepartmentID *uuid.UUID           `json:"dept_id"`
	CreatedAt    time.Time           `json:"created_at"`
	IncomingDocs []IncomingDocResponse `json:"incoming_docs,omitempty"`
}

// ToResponse converts OutgoingDoc to OutgoingDocResponse
func (o *OutgoingDoc) ToResponse() OutgoingDocResponse {
	return OutgoingDocResponse{
		ID:         o.ID,
		OutgoingNo: o.OutgoingNo,
		DocID:      o.DocID,
		UserID:       o.UserID,
		DepartmentID: o.DepartmentID,
		CreatedAt:    o.CreatedAt,
		DocNo:      o.DocNo,
		DocName:    o.DocName,
		DocPath:    o.DocPath,
		Type:       o.Type,
		UserName:   o.UserName,
	}
}
