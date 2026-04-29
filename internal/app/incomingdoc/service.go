package incomingdoc

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Service defines the interface for incoming document business logic
type Service interface {
	GetAllIncomingDocs(ctx context.Context, page, limit int) ([]domain.IncomingDocResponse, int, error)
	GetIncomingDocByID(ctx context.Context, id uuid.UUID) (*domain.IncomingDocResponse, error)
	GetIncomingDocsByReceiver(ctx context.Context, receiverID uuid.UUID) ([]domain.IncomingDocResponse, error)
	GetIncomingDocsByDepartment(ctx context.Context, deptID uuid.UUID, page, limit int) ([]domain.IncomingDocResponse, int, error)
	GetIncomingDocsByStatus(ctx context.Context, status string) ([]domain.IncomingDocResponse, error)
	ReceiveDocument(ctx context.Context, req domain.ReceiveDocumentRequest) (*domain.IncomingDocResponse, error)
	ApproveDocument(ctx context.Context, id uuid.UUID, req domain.ApproveDocumentRequest) (*domain.IncomingDocResponse, error)
	CreateIncomingDocs(ctx context.Context, docID uuid.UUID, senderID *uuid.UUID, deptIDs []uuid.UUID, remark string) error
}

type service struct {
	repo Repository
}

// NewService creates a new incoming document service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// GetAllIncomingDocs retrieves all incoming documents
func (s *service) GetAllIncomingDocs(ctx context.Context, page, limit int) ([]domain.IncomingDocResponse, int, error) {
	offset := (page - 1) * limit
	docs, total, err := s.repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, util.NewDatabaseError("get all incoming documents", err)
	}

	responses := make([]domain.IncomingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, total, nil
}

// GetIncomingDocByID retrieves an incoming document by ID
func (s *service) GetIncomingDocByID(ctx context.Context, id uuid.UUID) (*domain.IncomingDocResponse, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("IncomingDoc", id.String())
	}

	response := doc.ToResponse()
	return &response, nil
}

// GetIncomingDocsByReceiver retrieves incoming documents by receiver ID
func (s *service) GetIncomingDocsByReceiver(ctx context.Context, receiverID uuid.UUID) ([]domain.IncomingDocResponse, error) {
	docs, err := s.repo.FindByReceiverID(ctx, receiverID)
	if err != nil {
		return nil, util.NewDatabaseError("get incoming documents by receiver", err)
	}

	responses := make([]domain.IncomingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, nil
}

// GetIncomingDocsByDepartment retrieves incoming documents by department ID
func (s *service) GetIncomingDocsByDepartment(ctx context.Context, deptID uuid.UUID, page, limit int) ([]domain.IncomingDocResponse, int, error) {
	offset := (page - 1) * limit
	docs, total, err := s.repo.FindByDepartmentID(ctx, deptID, limit, offset)
	if err != nil {
		return nil, 0, util.NewDatabaseError("get incoming documents by department", err)
	}

	responses := make([]domain.IncomingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, total, nil
}

// GetIncomingDocsByStatus retrieves incoming documents by status
func (s *service) GetIncomingDocsByStatus(ctx context.Context, status string) ([]domain.IncomingDocResponse, error) {
	docs, err := s.repo.FindByStatus(ctx, status)
	if err != nil {
		return nil, util.NewDatabaseError("get incoming documents by status", err)
	}

	responses := make([]domain.IncomingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, nil
}

// ReceiveDocument marks a document as received
func (s *service) ReceiveDocument(ctx context.Context, req domain.ReceiveDocumentRequest) (*domain.IncomingDocResponse, error) {
	// Validate request
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	// Find the incoming document
	doc, err := s.repo.FindByID(ctx, req.IncomingDocID)
	if err != nil {
		return nil, util.NewNotFoundError("IncomingDoc", fmt.Sprintf("%d", req.IncomingDocID))
	}

	// Update fields
	now := time.Now()
	receiverUUID, err := util.ParseUUID(req.ReceiverID)
	if err != nil {
		return nil, util.NewInvalidInputError("receiver_id", "invalid UUID format")
	}
	doc.ReceiverID = &receiverUUID
	doc.ReceivedDate = &now
	doc.Remark = req.Remark
	doc.Status = domain.IncomingStatusReceived

	// Update in database
	if err := s.repo.Update(ctx, req.IncomingDocID, doc); err != nil {
		return nil, util.NewDatabaseError("receive document", err)
	}

	// Fetch updated document
	updatedDoc, err := s.repo.FindByID(ctx, req.IncomingDocID)
	if err != nil {
		return nil, util.NewDatabaseError("get updated incoming document", err)
	}

	response := updatedDoc.ToResponse()
	return &response, nil
}

// ApproveDocument approves an incoming document
func (s *service) ApproveDocument(ctx context.Context, id uuid.UUID, req domain.ApproveDocumentRequest) (*domain.IncomingDocResponse, error) {
	// Validate request
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	// Find the incoming document
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("IncomingDoc", id.String())
	}

	// Update fields
	now := time.Now()
	approverUUID, err := util.ParseUUID(req.ApproverID)
	if err != nil {
		return nil, util.NewInvalidInputError("approver_id", "invalid UUID format")
	}
	doc.ApproverID = &approverUUID
	doc.ApproverDate = &now
	doc.Remark = req.Remark
	doc.Status = domain.IncomingDocStatus(req.Status)

	// Update in database
	if err := s.repo.Update(ctx, id, doc); err != nil {
		return nil, util.NewDatabaseError("approve document", err)
	}

	// Fetch updated document
	updatedDoc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewDatabaseError("get updated incoming document", err)
	}

	response := updatedDoc.ToResponse()
	return &response, nil
}

func (s *service) CreateIncomingDocs(ctx context.Context, docID uuid.UUID, senderID *uuid.UUID, deptIDs []uuid.UUID, remark string) error {
	for _, deptID := range deptIDs {
		deptUUID := deptID
		doc := &domain.IncomingDoc{
			IncomingNo:   util.GenerateIncomingNumber(),
			DocID:        docID,
			SenderID:     senderID,
			DepartmentID: &deptUUID,
			Status:       domain.IncomingStatusPending,
			Remark:       remark,
			CreatedAt:    time.Now(),
		}
		if err := s.repo.Create(ctx, doc); err != nil {
			return err
		}
	}
	return nil
}
