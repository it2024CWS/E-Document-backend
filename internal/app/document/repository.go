package document

import (
	"context"
	"e-document-backend/internal/domain"

	"github.com/google/uuid"
)

// Repository defines the interface for document data access
type Repository interface {
	FindAll(ctx context.Context, userID string) ([]domain.Document, error)
	FindAllJoined(ctx context.Context) ([]domain.DocumentResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Document, error)
	FindByIDJoined(ctx context.Context, id uuid.UUID) (*domain.DocumentResponse, error)
	FindByFolderID(ctx context.Context, folderID uuid.UUID) ([]domain.Document, error)
	FindByFolderIDJoined(ctx context.Context, folderID uuid.UUID) ([]domain.DocumentResponse, error)
	Create(ctx context.Context, doc *domain.Document) error
	Update(ctx context.Context, id uuid.UUID, doc *domain.Document) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByDocNo(ctx context.Context, docNo string) (*domain.Document, error)
	CreateVersion(ctx context.Context, version *domain.Version) error
	GetVersionsByDocID(ctx context.Context, docID uuid.UUID) ([]domain.Version, error)
}
