package role

import (
	"context"
	"e-document-backend/internal/domain"
)

// Repository defines the interface for role data access
type Repository interface {
	FindAll(ctx context.Context) ([]domain.UserRole, error)
	FindByID(ctx context.Context, id int) (*domain.UserRole, error)
	Create(ctx context.Context, role *domain.UserRole) error
	Update(ctx context.Context, id int, role *domain.UserRole) error
	Delete(ctx context.Context, id int) error
}
