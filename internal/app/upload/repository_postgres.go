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

// postgresRepository implements the Repository interface for PostgreSQL
type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL upload repository
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{
		pool: pool,
	}
}

// BeginTx starts a new database transaction
func (r *postgresRepository) BeginTx(ctx context.Context) (pgx.Tx, error) {
	return r.pool.Begin(ctx)
}

// FindFolderByNameAndParent finds a folder by name, parent, and owner
func (r *postgresRepository) FindFolderByNameAndParent(ctx context.Context, tx pgx.Tx, name string, parentID *uuid.UUID, ownerID uuid.UUID) (*domain.Folder, error) {
	var query string
	var args []interface{}

	if parentID == nil {
		query = `
			SELECT id, name, path, is_root_folder, parent_folder_id, owner_id, created_at, updated_at
			FROM folders
			WHERE name = $1 AND parent_folder_id IS NULL AND owner_id = $2
		`
		args = []interface{}{name, ownerID}
	} else {
		query = `
			SELECT id, name, path, is_root_folder, parent_folder_id, owner_id, created_at, updated_at
			FROM folders
			WHERE name = $1 AND parent_folder_id = $2 AND owner_id = $3
		`
		args = []interface{}{name, *parentID, ownerID}
	}

	var folder domain.Folder
	err := tx.QueryRow(ctx, query, args...).Scan(
		&folder.ID,
		&folder.Name,
		&folder.Path,
		&folder.IsRootFolder,
		&folder.ParentFolderID,
		&folder.OwnerID,
		&folder.CreatedAt,
		&folder.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil // Folder not found, not an error
		}
		return nil, fmt.Errorf("failed to find folder: %w", err)
	}

	return &folder, nil
}

// CreateFolder creates a new folder in the database
func (r *postgresRepository) CreateFolder(ctx context.Context, tx pgx.Tx, folder *domain.Folder) error {
	query := `
		INSERT INTO folders (id, name, path, is_root_folder, parent_folder_id, owner_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`

	folder.ID = uuid.New()
	folder.CreatedAt = time.Now()
	folder.UpdatedAt = time.Now()

	err := tx.QueryRow(ctx, query,
		folder.ID,
		folder.Name,
		folder.Path,
		folder.IsRootFolder,
		folder.ParentFolderID,
		folder.OwnerID,
		folder.CreatedAt,
		folder.UpdatedAt,
	).Scan(&folder.ID, &folder.CreatedAt, &folder.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create folder: %w", err)
	}

	return nil
}

// GetFolderByID retrieves a folder by its ID (without transaction)
func (r *postgresRepository) GetFolderByID(ctx context.Context, folderID uuid.UUID) (*domain.Folder, error) {
	query := `
		SELECT id, name, path, is_root_folder, parent_folder_id, owner_id, created_at, updated_at
		FROM folders
		WHERE id = $1
	`

	var folder domain.Folder
	err := r.pool.QueryRow(ctx, query, folderID).Scan(
		&folder.ID,
		&folder.Name,
		&folder.Path,
		&folder.IsRootFolder,
		&folder.ParentFolderID,
		&folder.OwnerID,
		&folder.CreatedAt,
		&folder.UpdatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("folder not found")
		}
		return nil, fmt.Errorf("failed to get folder: %w", err)
	}

	return &folder, nil
}

// CreateDocument creates a new document in the database
func (r *postgresRepository) CreateDocument(ctx context.Context, tx pgx.Tx, doc *domain.Document) error {
	query := `
		INSERT INTO documents (
			id, title, description, type, category_id, folder_id, barcode,
			registrant_id, current_department_id, status, created_at, updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING id, created_at, updated_at
	`

	doc.ID = uuid.New()
	doc.CreatedAt = time.Now()
	doc.UpdatedAt = time.Now()

	err := tx.QueryRow(ctx, query,
		doc.ID,
		doc.Title,
		doc.Description,
		doc.Type,
		doc.CategoryID,
		doc.FolderID,
		doc.Barcode,
		doc.RegistrantID,
		doc.CurrentDepartmentID,
		doc.Status,
		doc.CreatedAt,
		doc.UpdatedAt,
	).Scan(&doc.ID, &doc.CreatedAt, &doc.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create document: %w", err)
	}

	return nil
}

// CreateAttachment creates a new document attachment in the database
func (r *postgresRepository) CreateAttachment(ctx context.Context, tx pgx.Tx, attachment *domain.DocumentAttachment) error {
	query := `
		INSERT INTO document_attachments (
			id, document_id, file_name, file_path, file_size, file_type,
			version, is_current, uploaded_by, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at
	`

	attachment.ID = uuid.New()
	attachment.CreatedAt = time.Now()

	err := tx.QueryRow(ctx, query,
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

	if err != nil {
		return fmt.Errorf("failed to create attachment: %w", err)
	}

	return nil
}

// GetLatestVersionByDocumentID gets the latest version number for a document's attachments
func (r *postgresRepository) GetLatestVersionByDocumentID(ctx context.Context, tx pgx.Tx, documentID uuid.UUID) (int, error) {
	query := `
		SELECT COALESCE(MAX(version), 0)
		FROM document_attachments
		WHERE document_id = $1
	`

	var version int
	err := tx.QueryRow(ctx, query, documentID).Scan(&version)
	if err != nil {
		return 0, fmt.Errorf("failed to get latest version: %w", err)
	}

	return version, nil
}

// SetPreviousVersionsNotCurrent marks all previous versions as not current
func (r *postgresRepository) SetPreviousVersionsNotCurrent(ctx context.Context, tx pgx.Tx, documentID uuid.UUID) error {
	query := `
		UPDATE document_attachments
		SET is_current = false
		WHERE document_id = $1 AND is_current = true
	`

	_, err := tx.Exec(ctx, query, documentID)
	if err != nil {
		return fmt.Errorf("failed to update previous versions: %w", err)
	}

	return nil
}

// GetAttachmentByID retrieves an attachment by its ID (without transaction)
func (r *postgresRepository) GetAttachmentByID(ctx context.Context, attachmentID uuid.UUID) (*domain.DocumentAttachment, error) {
	query := `
		SELECT id, document_id, file_name, file_path, file_size, file_type,
		       version, is_current, uploaded_by, created_at
		FROM document_attachments
		WHERE id = $1
	`

	var attachment domain.DocumentAttachment
	err := r.pool.QueryRow(ctx, query, attachmentID).Scan(
		&attachment.ID,
		&attachment.DocumentID,
		&attachment.FileName,
		&attachment.FilePath,
		&attachment.FileSize,
		&attachment.FileType,
		&attachment.Version,
		&attachment.IsCurrent,
		&attachment.UploadedBy,
		&attachment.CreatedAt,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("attachment not found")
		}
		return nil, fmt.Errorf("failed to get attachment: %w", err)
	}

	return &attachment, nil
}

// GetAttachmentsByFolderID retrieves all attachments in a folder (recursively including subfolders)
func (r *postgresRepository) GetAttachmentsByFolderID(ctx context.Context, folderID uuid.UUID) ([]*domain.DocumentAttachment, error) {
	query := `
		WITH RECURSIVE folder_tree AS (
			-- Base case: the specified folder
			SELECT id FROM folders WHERE id = $1
			UNION ALL
			-- Recursive case: all subfolders
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
		var attachment domain.DocumentAttachment
		err := rows.Scan(
			&attachment.ID,
			&attachment.DocumentID,
			&attachment.FileName,
			&attachment.FilePath,
			&attachment.FileSize,
			&attachment.FileType,
			&attachment.Version,
			&attachment.IsCurrent,
			&attachment.UploadedBy,
			&attachment.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan attachment: %w", err)
		}
		attachments = append(attachments, &attachment)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating attachments: %w", err)
	}

	return attachments, nil
}
