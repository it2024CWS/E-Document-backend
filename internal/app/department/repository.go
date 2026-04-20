package department

import (
	"context"
	"e-document-backend/internal/domain"
)

// Repository defines the interface for department data access
type Repository interface {
	FindAll(ctx context.Context, limit, offset int, search string) ([]domain.Department, error)
	Count(ctx context.Context, search string) (int, error)
	FindByID(ctx context.Context, id string) (*domain.Department, error)
	Create(ctx context.Context, dept *domain.Department) error
	Update(ctx context.Context, id string, dept *domain.Department) error
	Delete(ctx context.Context, id string) error
}
