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

	CreateFolder(ctx context.Context, tx pgx.Tx, folder *domain.Folder) error
	FindFolderByNameAndParent(ctx context.Context, tx pgx.Tx, name string, parentID *uuid.UUID, userID string) (*domain.Folder, error)
	CreateDocument(ctx context.Context, tx pgx.Tx, doc *domain.Document) error
	FindDocumentByNameAndType(ctx context.Context, tx pgx.Tx, docName string, docType string, folderID *uuid.UUID) (*domain.Document, error)
	UpdateDocumentVersion(ctx context.Context, tx pgx.Tx, docID uuid.UUID, newVersionNumber int, newDocPath string) error

	// Folder operations (without transaction)
	GetFolderByID(ctx context.Context, folderID int) (*domain.Folder, error)

	// Version operations (within transaction)
	CreateVersion(ctx context.Context, tx pgx.Tx, version *domain.Version) error
	GetLatestVersionByDocumentID(ctx context.Context, tx pgx.Tx, documentID uuid.UUID) (int, error)

	// Attachment operations (within transaction)
	CreateAttachment(ctx context.Context, tx pgx.Tx, attachment *domain.DocumentAttachment) error
	SetPreviousVersionsNotCurrent(ctx context.Context, tx pgx.Tx, documentID uuid.UUID) error

	// Attachment operations (without transaction)
	GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*domain.DocumentAttachment, error)
	GetAttachmentsByFolderID(ctx context.Context, folderID int) ([]*domain.DocumentAttachment, error)
}
