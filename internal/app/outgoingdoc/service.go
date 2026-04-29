package outgoingdoc

import (
	"context"
	"e-document-backend/internal/app/incomingdoc"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"

	"github.com/google/uuid"
)

// Service defines the interface for outgoing document business logic
type Service interface {
	GetAllOutgoingDocs(ctx context.Context, page, limit int) ([]domain.OutgoingDocResponse, int, error)
	GetOutgoingDocByID(ctx context.Context, id uuid.UUID) (*domain.OutgoingDocResponse, error)
	GetOutgoingDocsByUser(ctx context.Context, userID uuid.UUID) ([]domain.OutgoingDocResponse, error)
	GetOutgoingDocsByDepartment(ctx context.Context, deptID uuid.UUID, page, limit int) ([]domain.OutgoingDocResponse, int, error)
	CreateOutgoingDoc(ctx context.Context, req domain.CreateOutgoingDocRequest) (*domain.OutgoingDocResponse, error)
	CreateOutgoingDocWithParams(ctx context.Context, docID uuid.UUID, userID *uuid.UUID, deptID *uuid.UUID) error
}

type service struct {
	repo         Repository
	incomingRepo incomingdoc.Repository
}

// NewService creates a new outgoing document service
func NewService(repo Repository, incomingRepo incomingdoc.Repository) Service {
	return &service{
		repo:         repo,
		incomingRepo: incomingRepo,
	}
}

// GetAllOutgoingDocs retrieves all outgoing documents
func (s *service) GetAllOutgoingDocs(ctx context.Context, page, limit int) ([]domain.OutgoingDocResponse, int, error) {
	offset := (page - 1) * limit
	docs, total, err := s.repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, util.NewDatabaseError("get all outgoing documents", err)
	}

	responses := make([]domain.OutgoingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, total, nil
}

// GetOutgoingDocByID retrieves an outgoing document by ID
func (s *service) GetOutgoingDocByID(ctx context.Context, id uuid.UUID) (*domain.OutgoingDocResponse, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("OutgoingDoc", id.String())
	}

	response := doc.ToResponse()

	// Fetch related incoming documents
	incomings, err := s.incomingRepo.FindByDocID(ctx, doc.DocID)
	if err == nil {
		incomingResponses := make([]domain.IncomingDocResponse, len(incomings))
		for i, incoming := range incomings {
			incomingResponses[i] = incoming.ToResponse()
		}
		response.IncomingDocs = incomingResponses
	}

	return &response, nil
}

// GetOutgoingDocsByUser retrieves outgoing documents by user ID
func (s *service) GetOutgoingDocsByUser(ctx context.Context, userID uuid.UUID) ([]domain.OutgoingDocResponse, error) {
	docs, err := s.repo.FindByUserID(ctx, userID)
	if err != nil {
		return nil, util.NewDatabaseError("get outgoing documents by user", err)
	}

	responses := make([]domain.OutgoingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, nil
}

// GetOutgoingDocsByDepartment retrieves outgoing documents by department ID
func (s *service) GetOutgoingDocsByDepartment(ctx context.Context, deptID uuid.UUID, page, limit int) ([]domain.OutgoingDocResponse, int, error) {
	offset := (page - 1) * limit
	docs, total, err := s.repo.FindByDepartmentID(ctx, deptID, limit, offset)
	if err != nil {
		return nil, 0, util.NewDatabaseError("get outgoing documents by department", err)
	}

	responses := make([]domain.OutgoingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, total, nil
}

// CreateOutgoingDoc creates a new outgoing document
func (s *service) CreateOutgoingDoc(ctx context.Context, req domain.CreateOutgoingDocRequest) (*domain.OutgoingDocResponse, error) {
	// Validate request
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	// Generate outgoing number using similar LAL format
	outgoingNo := util.GenerateOutgoingNumber()

	// Create outgoing document object
	doc := &domain.OutgoingDoc{
		OutgoingNo: outgoingNo,
		DocID:      req.DocID,
		UserID:     req.UserID,
	}

	// Save to database
	if err := s.repo.Create(ctx, doc); err != nil {
		return nil, util.NewDatabaseError("create outgoing document", err)
	}

	response := doc.ToResponse()
	return &response, nil
}

func (s *service) CreateOutgoingDocWithParams(ctx context.Context, docID uuid.UUID, userID *uuid.UUID, deptID *uuid.UUID) error {
	doc := &domain.OutgoingDoc{
		OutgoingNo:   util.GenerateOutgoingNumber(),
		DocID:        docID,
		UserID:       userID,
		DepartmentID: deptID,
	}
	return s.repo.Create(ctx, doc)
}
