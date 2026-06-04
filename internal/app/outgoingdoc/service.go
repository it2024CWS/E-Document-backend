package outgoingdoc

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"github.com/google/uuid"
)

type Service interface {
	GetAllOutgoingDocs(ctx context.Context, page, limit int) ([]domain.OutgoingDocResponse, int, error)
	GetOutgoingDocByID(ctx context.Context, id uuid.UUID) (*domain.OutgoingDocResponse, error)
	GetOutgoingDocsByDepartment(ctx context.Context, deptID uuid.UUID, page, limit int) ([]domain.OutgoingDocResponse, int, error)
	CreateOutgoingDocWithParams(ctx context.Context, docDetailsID uuid.UUID, creatorID *uuid.UUID, deptID *uuid.UUID) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

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

func (s *service) GetOutgoingDocByID(ctx context.Context, id uuid.UUID) (*domain.OutgoingDocResponse, error) {
	doc, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("OutgoingDoc", id.String())
	}

	response := doc.ToResponse()
	return &response, nil
}

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

func (s *service) CreateOutgoingDocWithParams(ctx context.Context, docDetailsID uuid.UUID, creatorID *uuid.UUID, deptID *uuid.UUID) error {
	doc := &domain.OutgoingDoc{
		OutgoingNo:   util.GenerateOutgoingNumber(),
		DocDetailsID: docDetailsID,
		CreatedBy:    creatorID,
	}

	if err := s.repo.Create(ctx, doc); err != nil {
		return util.NewDatabaseError("create outgoing document", err)
	}

	return nil
}
