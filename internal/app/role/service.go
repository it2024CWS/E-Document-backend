package role

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"
	"strings"
	"sync"
)

type Service interface {
	CreateRole(ctx context.Context, req domain.CreateRoleRequest) (*domain.RoleResponse, error)
	GetRoleByID(ctx context.Context, id string) (*domain.RoleResponse, error)
	GetRoleByName(ctx context.Context, name string) (*domain.RoleResponse, error)
	GetAllRoles(ctx context.Context, skip int, limit int, search string) ([]domain.RoleResponse, int, error)
	UpdateRole(ctx context.Context, id string, req domain.UpdateRoleRequest) (*domain.RoleResponse, error)
	DeleteRole(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// NOTE - CreateRole method
func (s *service) CreateRole(ctx context.Context, req domain.CreateRoleRequest) (*domain.RoleResponse, error) {
	normalizedRoleName := strings.ToLower(strings.TrimSpace(req.Name))

	existingRole, _ := s.repo.FindByName(ctx, normalizedRoleName)
	if existingRole != nil {
		return nil, util.ErrorResponse(
			"Role already exists",
			util.ROLE_ALREADY_EXISTS,
			400,
			fmt.Sprintf("role with name %s already exists", normalizedRoleName),
		)
	}

	role := &domain.Role{
		Name: normalizedRoleName,
	}

	err := s.repo.Create(ctx, role)
	if err != nil {
		return nil, util.ErrorResponse(
			"Failed to create role",
			util.INTERNAL_SERVER_ERROR,
			500,
			err.Error(),
		)
	}

	response := role.ToResponse()
	return &response, nil

}

// NOTE - GetRoleByID method
func (s *service) GetRoleByID(ctx context.Context, id string) (*domain.RoleResponse, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.ErrorResponse(
			"Role not found",
			util.ROLE_NOT_FOUND,
			404,
			fmt.Sprintf("role with id %s not found", id),
		)
	}

	response := role.ToResponse()
	return &response, nil
}

// NOTE - GetRoleByName method
func (s *service) GetRoleByName(ctx context.Context, name string) (*domain.RoleResponse, error) {
	role, err := s.repo.FindByName(ctx, name)
	if err != nil {
		return nil, util.ErrorResponse(
			"Role not found",
			util.ROLE_NOT_FOUND,
			404,
			fmt.Sprintf("role with name %s not found", name),
		)
	}

	response := role.ToResponse()
	return &response, nil
}

// NOTE - GetAllRoles method
func (s *service) GetAllRoles(ctx context.Context, skip int, limit int, search string) ([]domain.RoleResponse, int, error) {
	skip = (skip - 1) * limit
	var wg sync.WaitGroup
	var roles []domain.Role
	var total int
	var rolesErr, countErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		roles, rolesErr = s.repo.FindAll(ctx, skip, limit, search)
	}()

	go func() {
		defer wg.Done()
		total, countErr = s.repo.Count(ctx, search)
	}()

	wg.Wait()

	if rolesErr != nil {
		return nil, 0, util.ErrorResponse(
			"Failed to retrieve roles",
			util.INTERNAL_SERVER_ERROR,
			500,
			rolesErr.Error(),
		)
	}

	if countErr != nil {
		return nil, 0, util.ErrorResponse(
			"Failed to count roles",
			util.INTERNAL_SERVER_ERROR,
			500,
			countErr.Error(),
		)
	}

	var responses = make([]domain.RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = role.ToResponse()
	}

	return responses, total, nil
}

func (s *service) UpdateRole(ctx context.Context, id string, req domain.UpdateRoleRequest) (*domain.RoleResponse, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.ErrorResponse(
			"Role not found",
			util.ROLE_NOT_FOUND,
			404,
			fmt.Sprintf("role with id %s not found", id),
		)
	}

	if req.Name != "" {
		role.Name = strings.ToLower(strings.TrimSpace(req.Name))
	}

	err = s.repo.Update(ctx, id, role)
	if err != nil {
		return nil, util.ErrorResponse(
			"Failed to update role",
			util.INTERNAL_SERVER_ERROR,
			500,
			err.Error(),
		)
	}

	response := role.ToResponse()
	return &response, nil
}

func (s *service) DeleteRole(ctx context.Context, id string) error {
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return util.ErrorResponse(
			"Role not found",
			util.ROLE_NOT_FOUND,
			404,
			fmt.Sprintf("role with id %s not found", id),
		)
	}

	err = s.repo.Delete(ctx, id)
	if err != nil {
		return util.ErrorResponse(
			"Failed to delete role",
			util.INTERNAL_SERVER_ERROR,
			500,
			err.Error(),
		)
	}

	return nil
}
