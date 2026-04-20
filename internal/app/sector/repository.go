package sector

import (
	"context"
	"e-document-backend/internal/domain"
)

type Repository interface {
	FindAll(ctx context.Context, limit, offset int, search string) ([]domain.Sector, error)
	Count(ctx context.Context, search string) (int, error)
	FindByID(ctx context.Context, id string) (*domain.Sector, error)
	FindByDepartmentID(ctx context.Context, deptID string) ([]domain.Sector, error)
	Create(ctx context.Context, sector *domain.Sector) error
	Update(ctx context.Context, id string, sector *domain.Sector) error
	Delete(ctx context.Context, id string) error
}
