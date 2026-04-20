package role

import (
	"context"
	"e-document-backend/internal/domain"
)

// Repository defines the interface for role data access
type Repository interface {
	FindAll(ctx context.Context, limit, offset int, search string) ([]domain.UserRole, error)
	Count(ctx context.Context, search string) (int, error)
	FindByID(ctx context.Context, id string) (*domain.UserRole, error)
	Create(ctx context.Context, role *domain.UserRole) error
	Update(ctx context.Context, id string, role *domain.UserRole) error
	Delete(ctx context.Context, id string) error
}
