package document

import (
	"context"
	"e-document-backend/internal/domain"
)

// Repository defines the interface for document data access
type Repository interface {
	FindAll(ctx context.Context, userID string) ([]domain.Document, error)
	FindByID(ctx context.Context, id int) (*domain.Document, error)
	FindByFolderID(ctx context.Context, folderID int) ([]domain.Document, error)
	Create(ctx context.Context, doc *domain.Document) error
	Update(ctx context.Context, id int, doc *domain.Document) error
	Delete(ctx context.Context, id int) error
	FindByDocNo(ctx context.Context, docNo string) (*domain.Document, error)
	CreateVersion(ctx context.Context, version *domain.Version) error
	GetVersionsByDocID(ctx context.Context, docID int) ([]domain.Version, error)
}
