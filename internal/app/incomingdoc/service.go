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
	CreateIncomingDocs(ctx context.Context, docDetailsID uuid.UUID, outgoingDocID uuid.UUID, creatorID *uuid.UUID, deptIDs []uuid.UUID, remark string) error
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

func (s *service) GetIncomingDocByID(ctx context.Context, id uuid.UUID) (*domain.IncomingDocResponse, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("IncomingDoc", id.String())
	}

	response := doc.ToResponse()
	return &response, nil
}

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

func (s *service) ReceiveDocument(ctx context.Context, req domain.ReceiveDocumentRequest) (*domain.IncomingDocResponse, error) {
	doc, err := s.repo.FindByID(ctx, req.IncomingDocID)
	if err != nil {
		return nil, util.NewNotFoundError("IncomingDoc", req.IncomingDocID.String())
	}

	if doc.Status != domain.IncomingStatusPending {
		return nil, util.NewValidationError(fmt.Sprintf("cannot receive a document with status '%s'", doc.Status))
	}

	now := time.Now()
	receiverID, err := uuid.Parse(req.ReceiverID)
	if err != nil {
		return nil, util.NewInvalidInputError("receiver_id", "must be a valid UUID")
	}

	// Generate incoming_no only at the moment of physical reception
	doc.IncomingNo = util.GenerateIncomingNumber()
	doc.Status = domain.IncomingStatusReceived
	doc.ReceivedDate = &now
	doc.UpdatedBy = &receiverID
	if req.Remark != "" {
		doc.Remark = req.Remark
	}

	if err := s.repo.Update(ctx, doc.ID, doc); err != nil {
		return nil, util.NewDatabaseError("receive document", err)
	}

	updated, err := s.repo.FindByID(ctx, doc.ID)
	if err != nil {
		return nil, util.NewDatabaseError("fetch updated document", err)
	}

	resp := updated.ToResponse()
	return &resp, nil
}

func (s *service) ApproveDocument(ctx context.Context, id uuid.UUID, req domain.ApproveDocumentRequest) (*domain.IncomingDocResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *service) CreateIncomingDocs(ctx context.Context, docDetailsID uuid.UUID, outgoingDocID uuid.UUID, creatorID *uuid.UUID, deptIDs []uuid.UUID, remark string) error {
	now := time.Now()
	var outgoingDocIDPtr *uuid.UUID
	if outgoingDocID != uuid.Nil {
		outgoingDocIDPtr = &outgoingDocID
	}
	for _, deptID := range deptIDs {
		deptIDCopy := deptID
		doc := &domain.IncomingDoc{
			// incoming_no is intentionally left empty — assigned by Secretary on receive.
			DocDetailsID:  docDetailsID,
			OutgoingDocID: outgoingDocIDPtr,
			CreatedBy:     creatorID,
			Status:        domain.IncomingStatusPending,
			Remark:        remark,
			IncomingDate:  &now,
			DeptID:        &deptIDCopy,
		}
		if err := s.repo.Create(ctx, doc); err != nil {
			return fmt.Errorf("failed to create incoming doc for dept %s: %w", deptID, err)
		}
	}
	return nil
}
