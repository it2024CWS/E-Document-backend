package upload

import (
	"context"
	"e-document-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Repository defines the data access methods for the upload service
type Repository interface {
	BeginTx(ctx context.Context) (pgx.Tx, error)

	FindFolderByNameAndParent(ctx context.Context, tx pgx.Tx, name string, parentID *uuid.UUID, userID string) (*domain.Folder, error)
	CreateFolder(ctx context.Context, tx pgx.Tx, folder *domain.Folder) error
	GetFolderByID(ctx context.Context, folderID int) (*domain.Folder, error)

	FindDocumentByNameAndFolder(ctx context.Context, tx pgx.Tx, docName string, folderID *uuid.UUID) (*domain.DocDetails, error)
	CreateDocument(ctx context.Context, tx pgx.Tx, doc *domain.DocDetails) error
	
	CreateVersion(ctx context.Context, tx pgx.Tx, version *domain.Version) error
	GetLatestVersionByDocumentID(ctx context.Context, tx pgx.Tx, documentID uuid.UUID) (int, error)
	UpdateDocumentVersion(ctx context.Context, tx pgx.Tx, docID uuid.UUID, newVersionNumber int) error
	
	GetVersionWithDoc(ctx context.Context, versionID uuid.UUID) (*domain.Version, *domain.DocDetails, error)
	GetVersionsByFolderID(ctx context.Context, folderID int) ([]*domain.Version, error)
}
