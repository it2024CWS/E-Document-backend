package file

import (
	"context"
	"time"
)

// Service defines business logic for file operations
type Service interface {
	GeneratePresignedURL(ctx context.Context, objectPath string, expirySeconds int64) (string, int64, error)
}

// storageClient defines the minimal interface we need from MinIO client
type storageClient interface {
	GetPresignedURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error)
}

// service implements Service
type service struct {
	storage storageClient
}

// NewService creates a new file service
func NewService(storage storageClient) Service {
	return &service{
		storage: storage,
	}
}

// GeneratePresignedURL contains the main logic for creating a presigned URL
func (s *service) GeneratePresignedURL(ctx context.Context, objectPath string, expirySeconds int64) (string, int64, error) {
	// Default expiry: 1 hour
	if expirySeconds <= 0 {
		expirySeconds = 3600
	}

	url, err := s.storage.GetPresignedURL(ctx, objectPath, time.Duration(expirySeconds)*time.Second)
	if err != nil {
		return "", 0, err
	}

	return url, expirySeconds, nil
}
