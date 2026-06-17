package domain

import (
	"time"

	"github.com/google/uuid"
)

// Rejected document sources.
const (
	RejectedSourceInbound  = "inbound"
	RejectedSourceOutbound = "outbound"
)

// RejectedDocResponse is a unified report row for a rejected document. It merges
// two sources:
//   - inbound:  an incoming_docs row rejected by a receiving department head
//     (status='rejected'); carries the receiving dept, approver, approver_date and remark.
//   - outbound: an outgoing_docs row rejected by the owner department head
//     (status='rejected'); the department is the owner's department, the rejecter is
//     updated_by and the date falls back to created_at (outgoing_docs has no
//     rejection timestamp/remark column), so remark is empty.
type RejectedDocResponse struct {
	Source       string     `json:"source"` // "inbound" | "outbound"
	ID           uuid.UUID  `json:"id"`
	DocNo        string     `json:"doc_no"`
	DocName      string     `json:"doc_name"`
	DocPath      string     `json:"doc_path,omitempty"`
	FileType     string     `json:"file_type,omitempty"`
	DeptID       *uuid.UUID `json:"dept_id,omitempty"`
	DeptName     string     `json:"dept_name,omitempty"`
	RejectedByID *uuid.UUID `json:"rejected_by_id,omitempty"`
	RejectedBy   string     `json:"rejected_by,omitempty"`
	RejectedAt   *time.Time `json:"rejected_at,omitempty"`
	Remark       string     `json:"remark,omitempty"`
}
