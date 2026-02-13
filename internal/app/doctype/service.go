package doctype

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"
)

type Service interface {
	GetAllDocTypes(ctx context.Context) ([]domain.DocTypeResponse, error)
	GetDocTypeByID(ctx context.Context, id int) (*domain.DocTypeResponse, error)
	CreateDocType(ctx context.Context, req domain.CreateDocTypeRequest) (*domain.DocTypeResponse, error)
	UpdateDocType(ctx context.Context, id int, req domain.UpdateDocTypeRequest) (*domain.DocTypeResponse, error)
	DeleteDocType(ctx context.Context, id int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAllDocTypes(ctx context.Context) ([]domain.DocTypeResponse, error) {
	docTypes, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, util.NewDatabaseError("get all doc types", err)
	}

	responses := make([]domain.DocTypeResponse, len(docTypes))
	for i, docType := range docTypes {
		responses[i] = docType.ToResponse()
	}

	return responses, nil
}

func (s *service) GetDocTypeByID(ctx context.Context, id int) (*domain.DocTypeResponse, error) {
	docType, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("DocType", fmt.Sprintf("%d", id))
	}

	response := docType.ToResponse()
	return &response, nil
}

func (s *service) CreateDocType(ctx context.Context, req domain.CreateDocTypeRequest) (*domain.DocTypeResponse, error) {
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	docType := &domain.DocType{TypeName: req.TypeName}

	if err := s.repo.Create(ctx, docType); err != nil {
		return nil, util.NewDatabaseError("create doc type", err)
	}

	response := docType.ToResponse()
	return &response, nil
}

func (s *service) UpdateDocType(ctx context.Context, id int, req domain.UpdateDocTypeRequest) (*domain.DocTypeResponse, error) {
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	existingDocType, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("DocType", fmt.Sprintf("%d", id))
	}

	if req.TypeName != "" {
		existingDocType.TypeName = req.TypeName
	}

	if err := s.repo.Update(ctx, id, existingDocType); err != nil {
		return nil, util.NewDatabaseError("update doc type", err)
	}

	updatedDocType, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewDatabaseError("get updated doc type", err)
	}

	response := updatedDocType.ToResponse()
	return &response, nil
}

func (s *service) DeleteDocType(ctx context.Context, id int) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return util.NewNotFoundError("DocType", fmt.Sprintf("%d", id))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return util.NewDatabaseError("delete doc type", err)
	}

	return nil
}
