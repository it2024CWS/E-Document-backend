package upload

import (
	"context"
	"e-document-backend/internal/domain"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL-backed upload repository
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// FindFolderByNameAndParent finds an existing folder by name, parent ID, and owner
func (r *postgresRepository) FindFolderByNameAndParent(ctx context.Context, tx pgx.Tx, name string, parentID *uuid.UUID, userID string) (*domain.Folder, error) {
	var query string
	var args []interface{}
	query = `
		SELECT id, folder_name, folder_path, user_id, parent_folder_id, created_at
		FROM folders
		WHERE folder_name = $1
		  AND user_id = $2`
	if parentID == nil {
		query += ` AND parent_folder_id IS NULL`
		args = []interface{}{name, userID}
	} else {
		query += ` AND parent_folder_id = $3`
		args = []interface{}{name, userID, parentID}
	}
	query += ` LIMIT 1`

	var folder domain.Folder
	err := tx.QueryRow(ctx, query, args...).Scan(
		&folder.ID,
		&folder.FolderName,
		&folder.FolderPath,
		&folder.UserID,
		&folder.ParentFolderID,
		&folder.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // not found → caller will create
		}
		return nil, fmt.Errorf("failed to find folder: %w", err)
	}
	return &folder, nil
}

// CreateFolder inserts a new folder and populates folder.ID
func (r *postgresRepository) CreateFolder(ctx context.Context, tx pgx.Tx, folder *domain.Folder) error {
	query := `
		INSERT INTO folders (folder_name, folder_path, user_id, parent_folder_id, created_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`
	folder.CreatedAt = time.Now().UTC()
	return tx.QueryRow(ctx, query,
		folder.FolderName,
		folder.FolderPath,
		folder.UserID,
		folder.ParentFolderID,
		folder.CreatedAt,
	).Scan(&folder.ID)
}

// GetFolderByID retrieves a folder without a transaction
func (r *postgresRepository) GetFolderByID(ctx context.Context, folderID int) (*domain.Folder, error) {
	query := `
		SELECT id, folder_name, folder_path, user_id, parent_folder_id, created_at
		FROM folders
		WHERE id = $1
	`
	var folder domain.Folder
	err := r.pool.QueryRow(ctx, query, folderID).Scan(
		&folder.ID,
		&folder.FolderName,
		&folder.FolderPath,
		&folder.UserID,
		&folder.ParentFolderID,
		&folder.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}
	return &folder, nil
}

// CreateDocument inserts a new document and populates doc.ID
func (r *postgresRepository) CreateDocument(ctx context.Context, tx pgx.Tx, doc *domain.Document) error {
	query := `
		INSERT INTO docs (doc_no, doc_name, doc_path, type, folder_id, registrant_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING id, created_at, updated_at
	`
	now := time.Now().UTC()
	doc.CreatedAt = now
	doc.UpdatedAt = now

	return tx.QueryRow(ctx, query,
		doc.DocNo,
		doc.DocName,
		doc.DocPath,
		doc.Type,
		doc.FolderID,
		doc.RegistrantID,
		doc.Status,
		doc.CreatedAt,
		doc.UpdatedAt,
	).Scan(&doc.ID, &doc.CreatedAt, &doc.UpdatedAt)
}

// CreateVersion inserts a new document version
func (r *postgresRepository) CreateVersion(ctx context.Context, tx pgx.Tx, version *domain.Version) error {
	query := `
		INSERT INTO versions (doc_id, version_number, doc_path, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at
	`
	version.CreatedAt = time.Now().UTC()
	return tx.QueryRow(ctx, query,
		version.DocID,
		version.VersionNumber,
		version.DocPath,
		version.CreatedAt,
	).Scan(&version.ID, &version.CreatedAt)
}

// CreateAttachment inserts a new document attachment
func (r *postgresRepository) CreateAttachment(ctx context.Context, tx pgx.Tx, attachment *domain.DocumentAttachment) error {
	query := `
		INSERT INTO document_attachments (id, document_id, file_name, file_path, file_size, file_type, version, is_current, uploaded_by, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`
	attachment.CreatedAt = time.Now().UTC()
	return tx.QueryRow(ctx, query,
		attachment.ID,
		attachment.DocumentID,
		attachment.FileName,
		attachment.FilePath,
		attachment.FileSize,
		attachment.FileType,
		attachment.Version,
		attachment.IsCurrent,
		attachment.UploadedBy,
		attachment.CreatedAt,
	).Scan(&attachment.ID, &attachment.CreatedAt)
}

// GetLatestVersionByDocumentID returns the highest version number for the given document
func (r *postgresRepository) GetLatestVersionByDocumentID(ctx context.Context, tx pgx.Tx, documentID int) (int, error) {
	query := `SELECT COALESCE(MAX(version), 0) FROM document_attachments WHERE document_id = $1`
	var version int
	if err := tx.QueryRow(ctx, query, documentID).Scan(&version); err != nil {
		return 0, fmt.Errorf("failed to get latest version: %w", err)
	}
	return version, nil
}

// SetPreviousVersionsNotCurrent marks all existing current attachments of a document as not current
func (r *postgresRepository) SetPreviousVersionsNotCurrent(ctx context.Context, tx pgx.Tx, documentID int) error {
	query := `UPDATE document_attachments SET is_current = false WHERE document_id = $1 AND is_current = true`
	if _, err := tx.Exec(ctx, query, documentID); err != nil {
		return fmt.Errorf("failed to update previous versions: %w", err)
	}
	return nil
}

// GetAttachmentByID retrieves an attachment by its UUID (no transaction)
func (r *postgresRepository) GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*domain.DocumentAttachment, error) {
	query := `
		SELECT id, document_id, file_name, file_path, file_size, file_type, version, is_current, uploaded_by, created_at
		FROM document_attachments
		WHERE id = $1
	`
	var a domain.DocumentAttachment
	err := r.pool.QueryRow(ctx, query, attachmentID).Scan(
		&a.ID, &a.DocumentID, &a.FileName, &a.FilePath,
		&a.FileSize, &a.FileType, &a.Version, &a.IsCurrent, &a.UploadedBy, &a.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("attachment not found")
		}
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}
	return &a, nil
}

// GetAttachmentsByFolderID retrieves all current attachments in a folder (recursively)
func (r *postgresRepository) GetAttachmentsByFolderID(ctx context.Context, folderID int) ([]*domain.DocumentAttachment, error) {
	query := `
		WITH RECURSIVE folder_tree AS (
			SELECT id FROM folders WHERE id = $1
			UNION ALL
			SELECT f.id FROM folders f
			INNER JOIN folder_tree ft ON f.parent_folder_id = ft.id
		)
		SELECT DISTINCT
			da.id, da.document_id, da.file_name, da.file_path, da.file_size, da.file_type,
			da.version, da.is_current, da.uploaded_by, da.created_at
		FROM document_attachments da
		INNER JOIN documents d ON d.id = da.document_id
		INNER JOIN folder_tree ft ON d.folder_id = ft.id
		WHERE da.is_current = true
		ORDER BY da.created_at DESC
	`
	rows, err := r.pool.Query(ctx, query, folderID)
	if err != nil {
		return nil, fmt.Errorf("failed to get attachments by folder: %w", err)
	}
	defer rows.Close()

	var attachments []*domain.DocumentAttachment
	for rows.Next() {
		var a domain.DocumentAttachment
		if err := rows.Scan(
			&a.ID, &a.DocumentID, &a.FileName, &a.FilePath,
			&a.FileSize, &a.FileType, &a.Version, &a.IsCurrent, &a.UploadedBy, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}
		attachments = append(attachments, &a)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attachments: %w", err)
	}
	return attachments, nil
}
