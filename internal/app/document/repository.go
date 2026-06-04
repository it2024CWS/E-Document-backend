package document

import (
	"context"
	"e-document-backend/internal/domain"
	"github.com/google/uuid"
)

type Repository interface {
	FindAllJoined(ctx context.Context) ([]domain.DocumentResponse, error)
	FindByIDJoined(ctx context.Context, id uuid.UUID) (*domain.DocumentResponse, error)
	FindByFolderIDJoined(ctx context.Context, folderID uuid.UUID) ([]domain.DocumentResponse, error)
	
	FindByID(ctx context.Context, id uuid.UUID) (*domain.DocDetails, error)
	FindByDocNo(ctx context.Context, docNo string) (*domain.DocDetails, error)
	Create(ctx context.Context, doc *domain.DocDetails) error
	Update(ctx context.Context, id uuid.UUID, doc *domain.DocDetails) error
	Delete(ctx context.Context, id uuid.UUID) error

	CreateVersion(ctx context.Context, version *domain.Version) error
	GetVersionsByDocID(ctx context.Context, docDetailsID uuid.UUID) ([]domain.Version, error)
}
