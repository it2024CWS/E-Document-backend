package upload

import (
	"context"
	"e-document-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// Repository defines the interface for upload-related database operations
type Repository interface {
	// Transaction management
	BeginTx(ctx context.Context) (pgx.Tx, error)

	// Folder operations (within transaction)
	FindFolderByNameAndParent(ctx context.Context, tx pgx.Tx, name string, parentID *uuid.UUID, ownerID uuid.UUID) (*domain.Folder, error)
	CreateFolder(ctx context.Context, tx pgx.Tx, folder *domain.Folder) error

	// Folder operations (without transaction)
	GetFolderByID(ctx context.Context, folderID uuid.UUID) (*domain.Folder, error)

	// Document operations (within transaction)
	CreateDocument(ctx context.Context, tx pgx.Tx, doc *domain.Document) error

	// Attachment operations (within transaction)
	CreateAttachment(ctx context.Context, tx pgx.Tx, attachment *domain.DocumentAttachment) error
	GetLatestVersionByDocumentID(ctx context.Context, tx pgx.Tx, documentID uuid.UUID) (int, error)
	SetPreviousVersionsNotCurrent(ctx context.Context, tx pgx.Tx, documentID uuid.UUID) error

	// Attachment operations (without transaction)
	GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*domain.DocumentAttachment, error)
	GetAttachmentsByFolderID(ctx context.Context, folderID uuid.UUID) ([]*domain.DocumentAttachment, error)
}
