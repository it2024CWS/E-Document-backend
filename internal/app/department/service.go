package department

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
)

type Service interface {
	GetAllDepartments(ctx context.Context, page, limit int, search string) ([]domain.DepartmentResponse, int, error)
	GetDepartmentByID(ctx context.Context, id string) (*domain.DepartmentResponse, error)
	CreateDepartment(ctx context.Context, req domain.CreateDepartmentRequest) (*domain.DepartmentResponse, error)
	UpdateDepartment(ctx context.Context, id string, req domain.UpdateDepartmentRequest) (*domain.DepartmentResponse, error)
	DeleteDepartment(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GetAllDepartments(ctx context.Context, page, limit int, search string) ([]domain.DepartmentResponse, int, error) {
	offset := (page - 1) * limit

	total, err := s.repo.Count(ctx, search)
	if err != nil {
		return nil, 0, util.NewDatabaseError("count departments", err)
	}

	departments, err := s.repo.FindAll(ctx, limit, offset, search)
	if err != nil {
		return nil, 0, util.NewDatabaseError("get all departments", err)
	}

	responses := make([]domain.DepartmentResponse, len(departments))
	for i, dept := range departments {
		responses[i] = dept.ToResponse()
	}

	return responses, total, nil
}

func (s *service) GetDepartmentByID(ctx context.Context, id string) (*domain.DepartmentResponse, error) {
	dept, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Department", id)
	}

	response := dept.ToResponse()
	return &response, nil
}

func (s *service) CreateDepartment(ctx context.Context, req domain.CreateDepartmentRequest) (*domain.DepartmentResponse, error) {
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	dept := &domain.Department{
		DeptName:    req.DeptName,
		Description: req.Description,
	}

	if err := s.repo.Create(ctx, dept); err != nil {
		return nil, util.NewDatabaseError("create department", err)
	}

	response := dept.ToResponse()
	return &response, nil
}

func (s *service) UpdateDepartment(ctx context.Context, id string, req domain.UpdateDepartmentRequest) (*domain.DepartmentResponse, error) {
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	existingDept, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Department", id)
	}

	if req.DeptName != "" {
		existingDept.DeptName = req.DeptName
	}
	if req.Description != "" {
		existingDept.Description = req.Description
	}

	if err := s.repo.Update(ctx, id, existingDept); err != nil {
		return nil, util.NewDatabaseError("update department", err)
	}

	updatedDept, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewDatabaseError("get updated department", err)
	}

	response := updatedDept.ToResponse()
	return &response, nil
}

func (s *service) DeleteDepartment(ctx context.Context, id string) error {
	if _, err := s.repo.FindByID(ctx, id); err != nil {
		return util.NewNotFoundError("Department", id)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return util.NewDatabaseError("delete department", err)
	}

	return nil
}
