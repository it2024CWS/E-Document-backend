package outgoingdoc

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"
)

// Service defines the interface for outgoing document business logic
type Service interface {
	GetAllOutgoingDocs(ctx context.Context) ([]domain.OutgoingDocResponse, error)
	GetOutgoingDocByID(ctx context.Context, id int) (*domain.OutgoingDocResponse, error)
	GetOutgoingDocsByUser(ctx context.Context, userID string) ([]domain.OutgoingDocResponse, error)
	CreateOutgoingDoc(ctx context.Context, req domain.CreateOutgoingDocRequest) (*domain.OutgoingDocResponse, error)
}

type service struct {
	repo Repository
}

// NewService creates a new outgoing document service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// GetAllOutgoingDocs retrieves all outgoing documents
func (s *service) GetAllOutgoingDocs(ctx context.Context) ([]domain.OutgoingDocResponse, error) {
	docs, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, util.NewDatabaseError("get all outgoing documents", err)
	}

	responses := make([]domain.OutgoingDocResponse, len(docs))
	for i, doc := range docs {
		responses[i] = doc.ToResponse()
	}

	return responses, nil
}

// GetOutgoingDocByID retrieves an outgoing document by ID
func (s *service) GetOutgoingDocByID(ctx context.Context, id int) (*domain.OutgoingDocResponse, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("OutgoingDoc", fmt.Sprintf("%d", id))
	}

	response := doc.ToResponse()
	return &response, nil
}

// GetOutgoingDocsByUser retrieves outgoing documents by user ID
func (s *service) GetOutgoingDocsByUser(ctx context.Context, userID string) ([]domain.OutgoingDocResponse, error) {
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

// CreateOutgoingDoc creates a new outgoing document
func (s *service) CreateOutgoingDoc(ctx context.Context, req domain.CreateOutgoingDocRequest) (*domain.OutgoingDocResponse, error) {
	// Validate request
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	// Generate outgoing number using similar LAL format
	outgoingNo := util.GenerateLALDocNumber()

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
