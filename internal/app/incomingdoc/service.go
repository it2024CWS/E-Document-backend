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
	CreateIncomingDocs(ctx context.Context, docDetailsID uuid.UUID, creatorID *uuid.UUID, deptIDs []uuid.UUID, remark string) error
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
	// Not implemented with full ReceiverID logic yet since ReceiverID was changed to UpdatedBy
	return nil, fmt.Errorf("not implemented")
}

func (s *service) ApproveDocument(ctx context.Context, id uuid.UUID, req domain.ApproveDocumentRequest) (*domain.IncomingDocResponse, error) {
	return nil, fmt.Errorf("not implemented")
}

func (s *service) CreateIncomingDocs(ctx context.Context, docDetailsID uuid.UUID, creatorID *uuid.UUID, deptIDs []uuid.UUID, remark string) error {
	now := time.Now()
	// Create one per department for now, though dept_id is removed from table we might need to route differently
	// Just create one default for now to satisfy upload logic.
	doc := &domain.IncomingDoc{
		IncomingNo:   util.GenerateIncomingNumber(),
		DocDetailsID: docDetailsID,
		CreatedBy:    creatorID,
		Status:       domain.IncomingStatusPending,
		Remark:       remark,
		IncomingDate: &now,
	}
	if err := s.repo.Create(ctx, doc); err != nil {
		return err
	}
	return nil
}
