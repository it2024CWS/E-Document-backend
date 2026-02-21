package folder

import (
	"context"
	"e-document-backend/internal/domain"

	"github.com/google/uuid"
)

// Repository defines the interface for folder data access
type Repository interface {
	FindAll(ctx context.Context) ([]domain.FolderResponse, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.FolderResponse, error)
	Create(ctx context.Context, folder *domain.Folder) error
	Update(ctx context.Context, id uuid.UUID, req domain.CreateFolderRequest) (*domain.FolderResponse, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
