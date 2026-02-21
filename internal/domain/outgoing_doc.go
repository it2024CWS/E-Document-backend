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
	UserID     *uuid.UUID `json:"user_id" db:"user_id"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
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
	UserID     *uuid.UUID `json:"user_id"`
	UserName   string     `json:"user_name,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// ToResponse converts OutgoingDoc to OutgoingDocResponse
func (o *OutgoingDoc) ToResponse() OutgoingDocResponse {
	return OutgoingDocResponse{
		ID:         o.ID,
		OutgoingNo: o.OutgoingNo,
		DocID:      o.DocID,
		UserID:     o.UserID,
		CreatedAt:  o.CreatedAt,
	}
}
