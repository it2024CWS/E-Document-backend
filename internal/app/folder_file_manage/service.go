package folder_file_manage

import (
	"context"
	"e-document-backend/internal/domain"

	"github.com/google/uuid"
)

// Service defines business logic for storage operations
type Service interface {
	// Folder operations
	GetFolder(ctx context.Context, folderID uuid.UUID) (*domain.Folder, error)
	GetRootFolders(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*domain.Folder, int, error)
	GetSubfolders(ctx context.Context, parentFolderID uuid.UUID, page, pageSize int) ([]*domain.Folder, int, error)
	GetFolderContents(ctx context.Context, folderID uuid.UUID) (*FolderContents, error)

	// Document operations
	GetDocument(ctx context.Context, documentID uuid.UUID) (*DocumentWithAttachment, error)
	GetDocumentsByFolder(ctx context.Context, folderID uuid.UUID, page, pageSize int) ([]*DocumentWithAttachment, int, error)
	GetAllDocuments(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*DocumentWithAttachment, int, error)

	// Recent files
	GetRecentFiles(ctx context.Context, ownerID uuid.UUID, limit int) ([]*RecentFile, error)
}

// service implements Service
type service struct {
	repo Repository
}

// NewService creates a new storage service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// GetFolder retrieves folder details
func (s *service) GetFolder(ctx context.Context, folderID uuid.UUID) (*domain.Folder, error) {
	return s.repo.GetFolderByID(ctx, folderID)
}

// GetRootFolders retrieves root folders with pagination
func (s *service) GetRootFolders(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*domain.Folder, int, error) {
	// Calculate offset
	offset := (page - 1) * pageSize
	var total int
	// Get folders with count
	folders, total, err := s.repo.GetRootFolders(ctx, ownerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return folders, total, nil
}

// GetSubfolders retrieves subfolders with pagination
func (s *service) GetSubfolders(ctx context.Context, parentFolderID uuid.UUID, page, pageSize int) ([]*domain.Folder, int, error) {
	// Calculate offset
	offset := (page - 1) * pageSize

	// Get subfolders with count
	folders, total, err := s.repo.GetSubfolders(ctx, parentFolderID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return folders, total, nil
}

// GetFolderContents retrieves folder contents (subfolders + documents)
func (s *service) GetFolderContents(ctx context.Context, folderID uuid.UUID) (*FolderContents, error) {
	return s.repo.GetFolderContents(ctx, folderID)
}

// GetDocument retrieves document details
func (s *service) GetDocument(ctx context.Context, documentID uuid.UUID) (*DocumentWithAttachment, error) {
	return s.repo.GetDocumentByID(ctx, documentID)
}

// GetDocumentsByFolder retrieves documents in a folder with pagination
func (s *service) GetDocumentsByFolder(ctx context.Context, folderID uuid.UUID, page, pageSize int) ([]*DocumentWithAttachment, int, error) {
	// Calculate offset
	offset := (page - 1) * pageSize

	// Get documents with count
	documents, total, err := s.repo.GetDocumentsByFolderID(ctx, folderID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

// GetAllDocuments retrieves all documents for a user with pagination
func (s *service) GetAllDocuments(ctx context.Context, ownerID uuid.UUID, page, pageSize int) ([]*DocumentWithAttachment, int, error) {
	// Calculate offset
	offset := (page - 1) * pageSize

	// Get documents with count
	documents, total, err := s.repo.GetAllDocuments(ctx, ownerID, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	return documents, total, nil
}

// GetRecentFiles retrieves recently modified files
func (s *service) GetRecentFiles(ctx context.Context, ownerID uuid.UUID, limit int) ([]*RecentFile, error) {
	return s.repo.GetRecentFiles(ctx, ownerID, limit)
}
