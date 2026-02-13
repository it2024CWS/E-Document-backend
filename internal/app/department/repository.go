package department

import (
	"context"
	"e-document-backend/internal/domain"
)

// Repository defines the interface for department data access
type Repository interface {
	FindAll(ctx context.Context) ([]domain.Department, error)
	FindByID(ctx context.Context, id int) (*domain.Department, error)
	Create(ctx context.Context, dept *domain.Department) error
	Update(ctx context.Context, id int, dept *domain.Department) error
	Delete(ctx context.Context, id int) error
}
