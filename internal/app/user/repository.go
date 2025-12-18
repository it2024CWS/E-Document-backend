package user

import (
	"context"
	"e-document-backend/internal/domain"
)

// Repository defines the interface for user data access
type Repository interface {
	Create(ctx context.Context, user *domain.User) error
	FindByID(ctx context.Context, id string) (*domain.User, error)
	FindByEmail(ctx context.Context, email string) (*domain.User, error)
	FindByUsername(ctx context.Context, username string) (*domain.User, error)
	FindAll(ctx context.Context, skip int, limit int, search string, currentUserID string) ([]domain.User, error)
	Count(ctx context.Context, search string, currentUserID string) (int, error)
	Update(ctx context.Context, id string, user *domain.User) error
	Delete(ctx context.Context, id string) error
}
