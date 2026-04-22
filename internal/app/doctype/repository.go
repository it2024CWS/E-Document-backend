package doctype

import (
	"context"
	"e-document-backend/internal/domain"
)

type Repository interface {
	FindAll(ctx context.Context) ([]domain.DocType, error)
	FindByID(ctx context.Context, id string) (*domain.DocType, error)
	Create(ctx context.Context, docType *domain.DocType) error
	Update(ctx context.Context, id string, docType *domain.DocType) error
	Delete(ctx context.Context, id string) error
}
