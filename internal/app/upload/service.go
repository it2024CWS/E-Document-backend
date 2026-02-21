package upload

import (
	"context"
	"e-document-backend/internal/domain"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

// Service defines business logic for upload operations
type Service interface {
	// ProcessUploadComplete handles the post-upload logic: folder creation, document, attachment
	ProcessUploadComplete(ctx context.Context, params ProcessUploadParams) (*ProcessUploadResult, error)

	// GetAttachment retrieves attachment details by ID
	GetAttachment(ctx context.Context, attachmentID uuid.UUID) (*domain.DocumentAttachment, error)

	// GetFolderAttachments retrieves all attachments in a folder (recursively)
	GetFolderAttachments(ctx context.Context, folderID uuid.UUID) ([]*domain.DocumentAttachment, error)

	// GetFolder retrieves folder details by ID
	GetFolder(ctx context.Context, folderID uuid.UUID) (*domain.Folder, error)
}

// ProcessUploadParams contains parameters for processing an upload
type ProcessUploadParams struct {
	RelativePath   string     // e.g., "Photos/2024/beach.jpg"
	ParentFolderID *uuid.UUID // optional: if provided, use as root
	OwnerID        uuid.UUID  // required: owner of the folders/documents
	FilePath       string     // MinIO object path
	FileSize       int64      // file size in bytes
	FileType       string     // file MIME type
	UploadID       string     // tusd upload ID
}

// ProcessUploadResult contains the result of processing an upload
type ProcessUploadResult struct {
	Document   *domain.Document           `json:"document"`
	Attachment *domain.DocumentAttachment `json:"attachment"`
	Folders    []*domain.Folder           `json:"folders"`
}

// service implements Service
type service struct {
	repo Repository
}

// NewService creates a new upload service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// ProcessUploadComplete handles the complete upload processing with transaction
func (s *service) ProcessUploadComplete(ctx context.Context, params ProcessUploadParams) (*ProcessUploadResult, error) {
	// Start transaction
	tx, err := s.repo.BeginTx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Ensure transaction is rolled back on error
	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Error().Err(rbErr).Msg("failed to rollback transaction")
			}
		}
	}()

	result := &ProcessUploadResult{
		Folders: make([]*domain.Folder, 0),
	}

	// Parse the relative path
	pathParts := parsePath(params.RelativePath)
	if len(pathParts) == 0 {
		err = fmt.Errorf("invalid relative path: %s", params.RelativePath)
		return nil, err
	}

	// The last part is the filename, everything before is folder path
	fileName := pathParts[len(pathParts)-1]
	folderParts := pathParts[:len(pathParts)-1]

	// Process folder hierarchy
	var currentParentID *uuid.UUID = params.ParentFolderID
	var currentPath string

	for i, folderName := range folderParts {
		// Build the path for this folder level
		if currentPath == "" {
			currentPath = folderName
		} else {
			currentPath = currentPath + "/" + folderName
		}

		// Determine if this is a root folder
		// It's a root folder ONLY if:
		// 1. It's the first folder in the path AND
		// 2. No parent_folder_id was provided from the client
		isRootFolder := i == 0 && params.ParentFolderID == nil

		// Try to find existing folder
		folder, findErr := s.repo.FindFolderByNameAndParent(ctx, tx, folderName, currentParentID, params.OwnerID)
		if findErr != nil {
			err = findErr
			return nil, err
		}

		if folder == nil {
			// Create new folder
			folder = &domain.Folder{
				Name:           folderName,
				Path:           currentPath,
				IsRootFolder:   isRootFolder,
				ParentFolderID: currentParentID,
				OwnerID:        params.OwnerID,
			}

			if createErr := s.repo.CreateFolder(ctx, tx, folder); createErr != nil {
				err = createErr
				return nil, err
			}

			log.Info().
				Str("folder_name", folderName).
				Str("path", currentPath).
				Bool("is_root", isRootFolder).
				Msg("Created new folder")
		}

		result.Folders = append(result.Folders, folder)
		currentParentID = &folder.ID
	}

	// Create document - use the filename as title
	titleWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	doc := &domain.Document{
		Title:        titleWithoutExt,
		Type:         domain.DocumentTypeGeneral,
		FolderID:     currentParentID, // Last folder in the hierarchy
		RegistrantID: &params.OwnerID,
		Status:       domain.DocumentStatusDraft,
	}

	if createErr := s.repo.CreateDocument(ctx, tx, doc); createErr != nil {
		err = createErr
		return nil, err
	}
	result.Document = doc

	log.Info().
		Str("document_id", doc.ID.String()).
		Str("title", doc.Title).
		Msg("Created new document")

	// Create attachment
	attachment := &domain.DocumentAttachment{
		DocumentID: doc.ID,
		FileName:   fileName,
		FilePath:   params.FilePath,
		FileSize:   params.FileSize,
		FileType:   params.FileType,
		Version:    1,
		IsCurrent:  true,
		UploadedBy: &params.OwnerID,
	}

	if createErr := s.repo.CreateAttachment(ctx, tx, attachment); createErr != nil {
		err = createErr
		return nil, err
	}
	result.Attachment = attachment

	log.Info().
		Str("attachment_id", attachment.ID.String()).
		Str("file_name", attachment.FileName).
		Str("file_path", attachment.FilePath).
		Int64("file_size", attachment.FileSize).
		Msg("Created new attachment")

	// Commit transaction
	if commitErr := tx.Commit(ctx); commitErr != nil {
		err = commitErr
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// parsePath splits a path string into individual parts, handling both / and \ separators
func parsePath(path string) []string {
	// Normalize path separators
	normalized := strings.ReplaceAll(path, "\\", "/")

	// Remove leading/trailing slashes
	normalized = strings.Trim(normalized, "/")

	if normalized == "" {
		return []string{}
	}

	// Split by /
	parts := strings.Split(normalized, "/")

	// Filter out empty parts
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		if part != "" {
			result = append(result, part)
		}
	}

	return result
}

// GetAttachment retrieves attachment details by ID
func (s *service) GetAttachment(ctx context.Context, attachmentID uuid.UUID) (*domain.DocumentAttachment, error) {
	return s.repo.GetAttachmentByID(ctx, attachmentID)
}

// GetFolderAttachments retrieves all attachments in a folder (recursively)
func (s *service) GetFolderAttachments(ctx context.Context, folderID uuid.UUID) ([]*domain.DocumentAttachment, error) {
	return s.repo.GetAttachmentsByFolderID(ctx, folderID)
}

// GetFolder retrieves folder details by ID
func (s *service) GetFolder(ctx context.Context, folderID uuid.UUID) (*domain.Folder, error) {
	return s.repo.GetFolderByID(ctx, folderID)
}
