package sector

import (
	"context"
	"e-document-backend/internal/domain"
)

type Repository interface {
	FindAll(ctx context.Context) ([]domain.Sector, error)
	FindByID(ctx context.Context, id int) (*domain.Sector, error)
	FindByDepartmentID(ctx context.Context, deptID int) ([]domain.Sector, error)
	Create(ctx context.Context, sector *domain.Sector) error
	Update(ctx context.Context, id int, sector *domain.Sector) error
	Delete(ctx context.Context, id int) error
}
