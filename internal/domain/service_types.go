package domain

// ReceiveDocumentRequest represents the data needed for the service to process receiving a document
type ReceiveDocumentRequest struct {
	IncomingDocID int
	ReceiverID    string
	Remark        string
}

// ApproveDocumentRequest represents the data needed for the service to process approving a document
type ApproveDocumentRequest struct {
	ApproverID string
	Remark     string
	Status     string
}
