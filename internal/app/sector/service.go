package sector

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"
)

type Service interface {
	GetAllSectors(ctx context.Context) ([]domain.SectorResponse, error)
	GetSectorByID(ctx context.Context, id int) (*domain.SectorResponse, error)
	GetSectorsByDepartment(ctx context.Context, deptID int) ([]domain.SectorResponse, error)
	CreateSector(ctx context.Context, req domain.CreateSectorRequest) (*domain.SectorResponse, error)
	UpdateSector(ctx context.Context, id int, req domain.UpdateSectorRequest) (*domain.SectorResponse, error)
	DeleteSector(ctx context.Context, id int) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAllSectors(ctx context.Context) ([]domain.SectorResponse, error) {
	sectors, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, util.NewDatabaseError("get all sectors", err)
	}

	responses := make([]domain.SectorResponse, len(sectors))
	for i, sector := range sectors {
		responses[i] = sector.ToResponse()
	}

	return responses, nil
}

func (s *service) GetSectorByID(ctx context.Context, id int) (*domain.SectorResponse, error) {
	sector, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Sector", fmt.Sprintf("%d", id))
	}

	response := sector.ToResponse()
	return &response, nil
}

func (s *service) GetSectorsByDepartment(ctx context.Context, deptID int) ([]domain.SectorResponse, error) {
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

	response := sector.ToResponse()
	return &response, nil
}

func (s *service) UpdateSector(ctx context.Context, id int, req domain.UpdateSectorRequest) (*domain.SectorResponse, error) {
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	existingSector, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Sector", fmt.Sprintf("%d", id))
	}

	if req.Name != "" {
		existingSector.Name = req.Name
	}
	if req.DeptID != 0 {
		existingSector.DeptID = req.DeptID
	}

	if err := s.repo.Update(ctx, id, existingSector); err != nil {
		return nil, util.NewDatabaseError("update sector", err)
	}

	updatedSector, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewDatabaseError("get updated sector", err)
	}

	response := updatedSector.ToResponse()
	return &response, nil
}

func (s *service) DeleteSector(ctx context.Context, id int) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return util.NewNotFoundError("Sector", fmt.Sprintf("%d", id))
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return util.NewDatabaseError("delete sector", err)
	}

	return nil
}
