package domain

import "github.com/google/uuid"

// ReceiveDocumentRequest represents the data needed for the service to process receiving a document
type ReceiveDocumentRequest struct {
	IncomingDocID uuid.UUID `json:"incoming_doc_id"`
	ReceiverID    string    `json:"receiver_id"`
	Remark        string    `json:"remark"`
}

// ApproveDocumentRequest represents the data needed for the service to process approving a document
type ApproveDocumentRequest struct {
	ApproverID string `json:"approver_id"`
	Remark     string `json:"remark"`
	Status     string `json:"status"`
}
