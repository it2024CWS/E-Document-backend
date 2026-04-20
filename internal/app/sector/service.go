package sector

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
)

type Service interface {
	GetAllSectors(ctx context.Context, page, limit int, search string) ([]domain.SectorResponse, int, error)
	GetSectorByID(ctx context.Context, id string) (*domain.SectorResponse, error)
	GetSectorsByDepartmentID(ctx context.Context, deptID string) ([]domain.SectorResponse, error)
	CreateSector(ctx context.Context, req domain.CreateSectorRequest) (*domain.SectorResponse, error)
	UpdateSector(ctx context.Context, id string, req domain.UpdateSectorRequest) (*domain.SectorResponse, error)
	DeleteSector(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAllSectors(ctx context.Context, page, limit int, search string) ([]domain.SectorResponse, int, error) {
	offset := (page - 1) * limit

	total, err := s.repo.Count(ctx, search)
	if err != nil {
		return nil, 0, util.NewDatabaseError("count sectors", err)
	}

	sectors, err := s.repo.FindAll(ctx, limit, offset, search)
	if err != nil {
		return nil, 0, util.NewDatabaseError("get all sectors", err)
	}

	responses := make([]domain.SectorResponse, len(sectors))
	for i, sector := range sectors {
		responses[i] = sector.ToResponse()
	}

	return responses, total, nil
}

func (s *service) GetSectorByID(ctx context.Context, id string) (*domain.SectorResponse, error) {
	sector, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Sector", id)
	}

	response := sector.ToResponse()
	return &response, nil
}

func (s *service) GetSectorsByDepartmentID(ctx context.Context, deptID string) ([]domain.SectorResponse, error) {
	sectors, err := s.repo.FindByDepartmentID(ctx, deptID)
	if err != nil {
		return nil, util.NewDatabaseError("get sectors by department", err)
	}

	responses := make([]domain.SectorResponse, len(sectors))
	for i, sector := range sectors {
		responses[i] = sector.ToResponse()
	}

	return responses, nil
}

func (s *service) CreateSector(ctx context.Context, req domain.CreateSectorRequest) (*domain.SectorResponse, error) {
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	sector := &domain.Sector{
		Name:   req.Name,
		DeptID: req.DeptID,
	}

	if err := s.repo.Create(ctx, sector); err != nil {
		return nil, util.NewDatabaseError("create sector", err)
	}

	// Fetch with joined info
	return s.GetSectorByID(ctx, sector.ID.String())
}

func (s *service) UpdateSector(ctx context.Context, id string, req domain.UpdateSectorRequest) (*domain.SectorResponse, error) {
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	existingSector, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Sector", id)
	}

	if req.Name != "" {
		existingSector.Name = req.Name
	}
	if req.DeptID != nil {
		existingSector.DeptID = *req.DeptID
	}

	if err := s.repo.Update(ctx, id, existingSector); err != nil {
		return nil, util.NewDatabaseError("update sector", err)
	}

	return s.GetSectorByID(ctx, id)
}

func (s *service) DeleteSector(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return util.NewNotFoundError("Sector", id)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return util.NewDatabaseError("delete sector", err)
	}

	return nil
}
