package sector

import (
	"context"
	"e-document-backend/internal/domain"

	"github.com/google/uuid"
)

type Repository interface {
	FindAll(ctx context.Context) ([]domain.Sector, error)
	FindByID(ctx context.Context, id uuid.UUID) (*domain.Sector, error)
	FindByDepartmentID(ctx context.Context, deptID uuid.UUID) ([]domain.Sector, error)
	Create(ctx context.Context, sector *domain.Sector) error
	Update(ctx context.Context, id uuid.UUID, sector *domain.Sector) error
	Delete(ctx context.Context, id uuid.UUID) error
}
