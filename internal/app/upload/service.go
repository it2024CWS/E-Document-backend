package upload

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
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
	GetFolderAttachments(ctx context.Context, folderID int) ([]*domain.DocumentAttachment, error)

	// GetFolder retrieves folder details by ID
	GetFolder(ctx context.Context, folderID int) (*domain.Folder, error)
}

// ProcessUploadParams contains parameters for processing an upload
type ProcessUploadParams struct {
	RelativePath   string     // e.g., "Photos/2024/beach.jpg"
	ParentFolderID *uuid.UUID // optional: if provided, use as root folder
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

	for _, folderName := range folderParts {
		// Build the path for this folder level
		if currentPath == "" {
			currentPath = folderName
		} else {
			currentPath = currentPath + "/" + folderName
		}

		// Try to find existing folder
		folder, findErr := s.repo.FindFolderByNameAndParent(ctx, tx, folderName, currentParentID, params.OwnerID.String())
		if findErr != nil {
			err = findErr
			return nil, err
		}

		if folder == nil {
			// Create new folder
			folder = &domain.Folder{
				FolderName:     folderName,
				FolderPath:     currentPath,
				UserID:         params.OwnerID,
				ParentFolderID: currentParentID,
			}

			if createErr := s.repo.CreateFolder(ctx, tx, folder); createErr != nil {
				err = createErr
				return nil, err
			}

			log.Info().
				Str("folder_name", folderName).
				Str("path", currentPath).
				Msg("Created new folder")
		}

		result.Folders = append(result.Folders, folder)
		currentParentID = &folder.ID
	}

	// -- Document & Version logic --
	// Check if a document with the same name + type already exists in the same folder
	titleWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
	docExt := strings.TrimPrefix(filepath.Ext(fileName), ".")

	existingDoc, findErr := s.repo.FindDocumentByNameAndType(ctx, tx, titleWithoutExt, docExt, currentParentID)
	if findErr != nil {
		err = findErr
		return nil, err
	}

	var doc *domain.Document
	var newVersionNumber int

	if existingDoc == nil {
		// ─── NEW FILE: create document + version 1 ───
		docNo := util.GenerateLALDocNumber()
		doc = &domain.Document{
			DocNo:         docNo,
			DocName:       titleWithoutExt,
			Type:          docExt,
			FolderID:      currentParentID,
			RegistrantID:  &params.OwnerID,
			Status:        domain.StatusNone,
			VersionNumber: 1,
		}

		if createErr := s.repo.CreateDocument(ctx, tx, doc); createErr != nil {
			err = createErr
			return nil, err
		}
		newVersionNumber = 1

		log.Info().
			Str("document_id", doc.ID.String()).
			Str("doc_name", doc.DocName).
			Msg("Created new document (version 1)")
	} else {
		// ─── EXISTING FILE: increment version ───
		latestVersion, vErr := s.repo.GetLatestVersionByDocumentID(ctx, tx, existingDoc.ID)
		if vErr != nil {
			err = vErr
			return nil, err
		}
		newVersionNumber = latestVersion + 1

		// Update the document's current path and version_number
		if updateErr := s.repo.UpdateDocumentVersion(ctx, tx, existingDoc.ID, newVersionNumber, params.FilePath); updateErr != nil {
			err = updateErr
			return nil, err
		}

		// Mark old attachments as not current
		if markErr := s.repo.SetPreviousVersionsNotCurrent(ctx, tx, existingDoc.ID); markErr != nil {
			err = markErr
			return nil, err
		}

		doc = existingDoc
		doc.VersionNumber = newVersionNumber

		log.Info().
			Str("document_id", doc.ID.String()).
			Str("doc_name", doc.DocName).
			Int("new_version", newVersionNumber).
			Msg("Updated existing document to new version")
	}

	result.Document = doc

	// Create attachment (always new for every upload)
	attachment := &domain.DocumentAttachment{
		ID:         uuid.New(),
		DocumentID: doc.ID,
		FileName:   fileName,
		FilePath:   params.FilePath,
		FileSize:   params.FileSize,
		FileType:   params.FileType,
		Version:    newVersionNumber,
		IsCurrent:  true,
		UploadedBy: params.OwnerID,
	}

	if createErr := s.repo.CreateAttachment(ctx, tx, attachment); createErr != nil {
		err = createErr
		return nil, err
	}
	result.Attachment = attachment

	// Create version record (with folder_id if available)
	version := &domain.Version{
		DocID:         doc.ID,
		FolderID:      doc.FolderID,
		VersionNumber: newVersionNumber,
		DocPath:       params.FilePath,
	}
	if createErr := s.repo.CreateVersion(ctx, tx, version); createErr != nil {
		err = createErr
		return nil, err
	}

	log.Info().
		Str("attachment_id", attachment.ID.String()).
		Str("file_name", attachment.FileName).
		Str("file_path", attachment.FilePath).
		Int64("file_size", attachment.FileSize).
		Int("version", newVersionNumber).
		Msg("Created new attachment and version")

	// Commit transaction
	if commitErr := tx.Commit(ctx); commitErr != nil {
		err = commitErr
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// parsePath splits a path string into individual parts, handling both / and \ separators
func parsePath(path string) []string {
	normalized := strings.ReplaceAll(path, "\\", "/")
	normalized = strings.Trim(normalized, "/")
	if normalized == "" {
		return []string{}
	}
	parts := strings.Split(normalized, "/")
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
func (s *service) GetFolderAttachments(ctx context.Context, folderID int) ([]*domain.DocumentAttachment, error) {
	return s.repo.GetAttachmentsByFolderID(ctx, folderID)
}

// GetFolder retrieves folder details by ID
func (s *service) GetFolder(ctx context.Context, folderID int) (*domain.Folder, error) {
	return s.repo.GetFolderByID(ctx, folderID)
}
