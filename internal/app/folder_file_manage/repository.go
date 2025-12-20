package folder_file_manage

import (
	"context"
	"e-document-backend/internal/domain"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the interface for storage-related database operations
type Repository interface {
	// Folder operations
	GetFolderByID(ctx context.Context, folderID uuid.UUID) (*domain.Folder, error)
	GetRootFolders(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*domain.Folder, int, error)
	GetSubfolders(ctx context.Context, parentFolderID uuid.UUID, limit, offset int) ([]*domain.Folder, int, error)
	GetFolderContents(ctx context.Context, folderID uuid.UUID) (*FolderContents, error)

	// Document operations
	GetDocumentByID(ctx context.Context, documentID uuid.UUID) (*DocumentWithAttachment, error)
	GetDocumentsByFolderID(ctx context.Context, folderID uuid.UUID, limit, offset int) ([]*DocumentWithAttachment, int, error)
	GetAllDocuments(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*DocumentWithAttachment, int, error)

	// Recent files
	GetRecentFiles(ctx context.Context, ownerID uuid.UUID, limit int) ([]*RecentFile, error)
}

// FolderContents represents the contents of a folder (subfolders + documents)
type FolderContents struct {
	Folder     *domain.Folder            `json:"folder"`
	Subfolders []*domain.Folder          `json:"subfolders"`
	Documents  []*DocumentWithAttachment `json:"documents"`
}

// DocumentWithAttachment represents a document with its current attachment
type DocumentWithAttachment struct {
	*domain.Document
	Attachment *domain.DocumentAttachment `json:"attachment,omitempty"`
}

// RecentFile represents a recently modified file
type RecentFile struct {
	DocumentID   uuid.UUID  `json:"document_id"`
	Title        string     `json:"title"`
	FolderID     *uuid.UUID `json:"folder_id"`
	FolderName   *string    `json:"folder_name,omitempty"`
	FolderPath   *string    `json:"folder_path,omitempty"`
	AttachmentID *uuid.UUID `json:"attachment_id"`
	FileName     *string    `json:"file_name,omitempty"`
	FileType     *string    `json:"file_type,omitempty"`
	FileSize     *int64     `json:"file_size,omitempty"`
	LastModified string     `json:"last_modified"`
}

// repository implements the Repository interface for PostgreSQL
type repository struct {
	pool *pgxpool.Pool
}

// NewRepository creates a new storage repository
func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{
		pool: pool,
	}
}

// GetFolderByID retrieves a folder by its ID
func (r *repository) GetFolderByID(ctx context.Context, folderID uuid.UUID) (*domain.Folder, error) {
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

// GetRootFolders retrieves root folders for a user with pagination
func (r *repository) GetRootFolders(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*domain.Folder, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM folders
		WHERE owner_id = $1 AND is_root_folder = true
	`

	var total int
	err := r.pool.QueryRow(ctx, countQuery, ownerID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count root folders: %w", err)
	}

	// Get folders ordered by updated_at DESC (most recent first)
	query := `
		SELECT id, name, path, is_root_folder, parent_folder_id, owner_id, created_at, updated_at
		FROM folders
		WHERE owner_id = $1 AND is_root_folder = true
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get root folders: %w", err)
	}
	defer rows.Close()

	var folders []*domain.Folder
	for rows.Next() {
		var folder domain.Folder
		err := rows.Scan(
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
			return nil, 0, fmt.Errorf("failed to scan folder: %w", err)
		}
		folders = append(folders, &folder)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating folders: %w", err)
	}

	return folders, total, nil
}

// GetSubfolders retrieves subfolders of a parent folder with pagination
func (r *repository) GetSubfolders(ctx context.Context, parentFolderID uuid.UUID, limit, offset int) ([]*domain.Folder, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM folders
		WHERE parent_folder_id = $1
	`

	var total int
	err := r.pool.QueryRow(ctx, countQuery, parentFolderID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count subfolders: %w", err)
	}

	// Get subfolders ordered by updated_at DESC
	query := `
		SELECT id, name, path, is_root_folder, parent_folder_id, owner_id, created_at, updated_at
		FROM folders
		WHERE parent_folder_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, parentFolderID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get subfolders: %w", err)
	}
	defer rows.Close()

	var folders []*domain.Folder
	for rows.Next() {
		var folder domain.Folder
		err := rows.Scan(
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
			return nil, 0, fmt.Errorf("failed to scan folder: %w", err)
		}
		folders = append(folders, &folder)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating folders: %w", err)
	}

	return folders, total, nil
}

// GetFolderContents retrieves folder information along with its subfolders and documents
func (r *repository) GetFolderContents(ctx context.Context, folderID uuid.UUID) (*FolderContents, error) {
	// Get folder info
	folder, err := r.GetFolderByID(ctx, folderID)
	if err != nil {
		return nil, err
	}

	// Get subfolders (no pagination, get all)
	subfolders, _, err := r.GetSubfolders(ctx, folderID, 1000, 0)
	if err != nil {
		return nil, err
	}

	// Get documents (no pagination, get all)
	documents, _, err := r.GetDocumentsByFolderID(ctx, folderID, 1000, 0)
	if err != nil {
		return nil, err
	}

	return &FolderContents{
		Folder:     folder,
		Subfolders: subfolders,
		Documents:  documents,
	}, nil
}

// GetDocumentByID retrieves a document with its current attachment
func (r *repository) GetDocumentByID(ctx context.Context, documentID uuid.UUID) (*DocumentWithAttachment, error) {
	query := `
		SELECT 
			d.id, d.title, d.description, d.type, d.category_id, d.folder_id, 
			d.barcode, d.registrant_id, d.current_department_id, d.status, 
			d.created_at, d.updated_at,
			da.id, da.document_id, da.file_name, da.file_path, da.file_size, 
			da.file_type, da.version, da.is_current, da.uploaded_by, da.created_at
		FROM documents d
		LEFT JOIN document_attachments da ON d.id = da.document_id AND da.is_current = true
		WHERE d.id = $1
	`

	var doc DocumentWithAttachment
	doc.Document = &domain.Document{}
	var attachment domain.DocumentAttachment
	var hasAttachment bool

	err := r.pool.QueryRow(ctx, query, documentID).Scan(
		&doc.ID,
		&doc.Title,
		&doc.Description,
		&doc.Type,
		&doc.CategoryID,
		&doc.FolderID,
		&doc.Barcode,
		&doc.RegistrantID,
		&doc.CurrentDepartmentID,
		&doc.Status,
		&doc.CreatedAt,
		&doc.UpdatedAt,
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
			return nil, fmt.Errorf("document not found")
		}
		// Check if it's a NULL attachment (LEFT JOIN result)
		hasAttachment = attachment.ID != uuid.Nil
	} else {
		hasAttachment = attachment.ID != uuid.Nil
	}

	if hasAttachment {
		doc.Attachment = &attachment
	}

	return &doc, nil
}

// GetDocumentsByFolderID retrieves documents in a folder with their current attachments
func (r *repository) GetDocumentsByFolderID(ctx context.Context, folderID uuid.UUID, limit, offset int) ([]*DocumentWithAttachment, int, error) {
	// Get total count
	countQuery := `
		SELECT COUNT(*)
		FROM documents
		WHERE folder_id = $1
	`

	var total int
	err := r.pool.QueryRow(ctx, countQuery, folderID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Get documents ordered by updated_at DESC
	query := `
		SELECT 
			d.id, d.title, d.description, d.type, d.category_id, d.folder_id, 
			d.barcode, d.registrant_id, d.current_department_id, d.status, 
			d.created_at, d.updated_at,
			da.id, da.document_id, da.file_name, da.file_path, da.file_size, 
			da.file_type, da.version, da.is_current, da.uploaded_by, da.created_at
		FROM documents d
		LEFT JOIN document_attachments da ON d.id = da.document_id AND da.is_current = true
		WHERE d.folder_id = $1
		ORDER BY d.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, folderID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get documents: %w", err)
	}
	defer rows.Close()

	var documents []*DocumentWithAttachment
	for rows.Next() {
		var doc DocumentWithAttachment
		doc.Document = &domain.Document{}
		var attachment domain.DocumentAttachment

		err := rows.Scan(
			&doc.ID,
			&doc.Title,
			&doc.Description,
			&doc.Type,
			&doc.CategoryID,
			&doc.FolderID,
			&doc.Barcode,
			&doc.RegistrantID,
			&doc.CurrentDepartmentID,
			&doc.Status,
			&doc.CreatedAt,
			&doc.UpdatedAt,
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
			return nil, 0, fmt.Errorf("failed to scan document: %w", err)
		}

		// Check if attachment exists (LEFT JOIN might return NULLs)
		if attachment.ID != uuid.Nil {
			doc.Attachment = &attachment
		}

		documents = append(documents, &doc)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating documents: %w", err)
	}

	return documents, total, nil
}

// GetAllDocuments retrieves all documents for a user
func (r *repository) GetAllDocuments(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]*DocumentWithAttachment, int, error) {
	// Get total count - documents where user is registrant
	countQuery := `
		SELECT COUNT(*)
		FROM documents
		WHERE registrant_id = $1
	`

	var total int
	err := r.pool.QueryRow(ctx, countQuery, ownerID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count documents: %w", err)
	}

	// Get documents ordered by updated_at DESC
	query := `
		SELECT 
			d.id, d.title, d.description, d.type, d.category_id, d.folder_id, 
			d.barcode, d.registrant_id, d.current_department_id, d.status, 
			d.created_at, d.updated_at,
			da.id, da.document_id, da.file_name, da.file_path, da.file_size, 
			da.file_type, da.version, da.is_current, da.uploaded_by, da.created_at
		FROM documents d
		LEFT JOIN document_attachments da ON d.id = da.document_id AND da.is_current = true
		WHERE d.registrant_id = $1
		ORDER BY d.updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.pool.Query(ctx, query, ownerID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get documents: %w", err)
	}
	defer rows.Close()

	var documents []*DocumentWithAttachment
	for rows.Next() {
		var doc DocumentWithAttachment
		doc.Document = &domain.Document{}
		var attachment domain.DocumentAttachment

		err := rows.Scan(
			&doc.ID,
			&doc.Title,
			&doc.Description,
			&doc.Type,
			&doc.CategoryID,
			&doc.FolderID,
			&doc.Barcode,
			&doc.RegistrantID,
			&doc.CurrentDepartmentID,
			&doc.Status,
			&doc.CreatedAt,
			&doc.UpdatedAt,
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
			return nil, 0, fmt.Errorf("failed to scan document: %w", err)
		}

		// Check if attachment exists
		if attachment.ID != uuid.Nil {
			doc.Attachment = &attachment
		}

		documents = append(documents, &doc)
	}

	if err = rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating documents: %w", err)
	}

	return documents, total, nil
}

// GetRecentFiles retrieves recently modified files for a user
func (r *repository) GetRecentFiles(ctx context.Context, ownerID uuid.UUID, limit int) ([]*RecentFile, error) {
	query := `
		SELECT 
			d.id AS document_id,
			d.title,
			d.folder_id,
			f.name AS folder_name,
			f.path AS folder_path,
			da.id AS attachment_id,
			da.file_name,
			da.file_type,
			da.file_size,
			GREATEST(d.updated_at, da.created_at) AS last_modified
		FROM documents d
		LEFT JOIN folders f ON d.folder_id = f.id
		LEFT JOIN document_attachments da ON d.id = da.document_id AND da.is_current = true
		WHERE d.registrant_id = $1
		ORDER BY last_modified DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, query, ownerID, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent files: %w", err)
	}
	defer rows.Close()

	var files []*RecentFile
	for rows.Next() {
		var file RecentFile
		err := rows.Scan(
			&file.DocumentID,
			&file.Title,
			&file.FolderID,
			&file.FolderName,
			&file.FolderPath,
			&file.AttachmentID,
			&file.FileName,
			&file.FileType,
			&file.FileSize,
			&file.LastModified,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan recent file: %w", err)
		}

		files = append(files, &file)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating recent files: %w", err)
	}

	return files, nil
}
