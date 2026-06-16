package folder

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"

	"github.com/google/uuid"
)

// Service defines the business logic interface for folders
type Service interface {
	GetAllFolders(ctx context.Context, userID string) ([]domain.FolderResponse, error)
	GetFolderByID(ctx context.Context, id uuid.UUID) (*domain.FolderResponse, error)
	CreateFolder(ctx context.Context, userID string, req domain.CreateFolderRequest) (*domain.FolderResponse, error)
	UpdateFolder(ctx context.Context, id uuid.UUID, req domain.CreateFolderRequest) (*domain.FolderResponse, error)
	DeleteFolder(ctx context.Context, id uuid.UUID) error
}

type service struct {
	repo Repository
}

// NewService creates a new folder service
func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAllFolders(ctx context.Context, userID string) ([]domain.FolderResponse, error) {
	ownerID, err := uuid.Parse(userID)
	if err != nil {
		return nil, util.NewInvalidInputError("user_id", "invalid user ID")
	}
	folders, err := s.repo.FindAll(ctx, ownerID)
	if err != nil {
		return nil, util.NewDatabaseError("get all folders", err)
	}
	if folders == nil {
		folders = []domain.FolderResponse{}
	}
	return folders, nil
}

func (s *service) GetFolderByID(ctx context.Context, id uuid.UUID) (*domain.FolderResponse, error) {
	folder, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Folder", "not found")
	}
	return folder, nil
}

func (s *service) CreateFolder(ctx context.Context, userID string, req domain.CreateFolderRequest) (*domain.FolderResponse, error) {
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	userUUID := parseUserUUID(userID)
	var parentFolderID *uuid.UUID
	if req.ParentFolderID != "" {
		if pID, err := uuid.Parse(req.ParentFolderID); err == nil {
			parentFolderID = &pID
		}
	}

	folder := &domain.Folder{
		FolderName:     req.FolderName,
		FolderPath:     req.FolderPath,
		UserID:         userUUID,
		ParentFolderID: parentFolderID,
	}

	if err := s.repo.Create(ctx, folder); err != nil {
		return nil, util.NewDatabaseError("create folder", err)
	}

	return s.repo.FindByID(ctx, folder.ID)
}

func (s *service) UpdateFolder(ctx context.Context, id uuid.UUID, req domain.CreateFolderRequest) (*domain.FolderResponse, error) {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return nil, util.NewNotFoundError("Folder", "not found")
	}

	updated, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, util.NewDatabaseError("update folder", err)
	}
	return updated, nil
}

func (s *service) DeleteFolder(ctx context.Context, id uuid.UUID) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return util.NewNotFoundError("Folder", "not found")
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return util.NewDatabaseError("delete folder", err)
	}
	return nil
}
