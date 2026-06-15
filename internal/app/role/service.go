package role

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"
	"sync"
)

// Service defines the interface for role business logic
type Service interface {
	GetAllRoles(ctx context.Context, page, limit int, search string) ([]domain.RoleResponse, int, error)
	GetRoleByID(ctx context.Context, id string) (*domain.RoleResponse, error)
	CreateRole(ctx context.Context, req domain.CreateRoleRequest) (*domain.RoleResponse, error)
	UpdateRole(ctx context.Context, id string, req domain.UpdateRoleRequest) (*domain.RoleResponse, error)
	DeleteRole(ctx context.Context, id string) error
}

type service struct {
	repo Repository
}

// NewService creates a new role service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// GetAllRoles retrieves all roles with pagination and search
func (s *service) GetAllRoles(ctx context.Context, page, limit int, search string) ([]domain.RoleResponse, int, error) {
	offset := (page - 1) * limit

	var wg sync.WaitGroup
	var total int
	var roles []domain.UserRole
	var countErr, findErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		total, countErr = s.repo.Count(ctx, search)
	}()

	go func() {
		defer wg.Done()
		roles, findErr = s.repo.FindAll(ctx, limit, offset, search)
	}()

	wg.Wait()

	if countErr != nil {
		return nil, 0, util.NewDatabaseError("count roles", countErr)
	}

	if findErr != nil {
		return nil, 0, util.NewDatabaseError("get all roles", findErr)
	}

	responses := make([]domain.RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = role.ToResponse()
	}

	return responses, total, nil
}

// GetRoleByID retrieves a role by ID
func (s *service) GetRoleByID(ctx context.Context, id string) (*domain.RoleResponse, error) {
	role, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Role", fmt.Sprintf("%s", id))
	}

	response := role.ToResponse()
	return &response, nil
}

// CreateRole creates a new role
func (s *service) CreateRole(ctx context.Context, req domain.CreateRoleRequest) (*domain.RoleResponse, error) {
	// Validate request
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	// Create role object
	role := &domain.UserRole{
		RoleName:    req.RoleName,
		Description: req.Description,
	}

	// Save to database
	if err := s.repo.Create(ctx, role); err != nil {
		return nil, util.NewDatabaseError("create role", err)
	}

	response := role.ToResponse()
	return &response, nil
}

// UpdateRole updates a role
func (s *service) UpdateRole(ctx context.Context, id string, req domain.UpdateRoleRequest) (*domain.RoleResponse, error) {
	// Validate request
	if err := util.ValidateStruct(&req); err != nil {
		return nil, err
	}

	// Check if role exists
	existingRole, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewNotFoundError("Role", fmt.Sprintf("%d", id))
	}

	// Update fields if provided
	if req.RoleName != "" {
		existingRole.RoleName = req.RoleName
	}
	if req.Description != "" {
		existingRole.Description = req.Description
	}

	// Update in database
	if err := s.repo.Update(ctx, id, existingRole); err != nil {
		return nil, util.NewDatabaseError("update role", err)
	}

	// Fetch updated role
	updatedRole, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.NewDatabaseError("get updated role", err)
	}

	response := updatedRole.ToResponse()
	return &response, nil
}

// DeleteRole deletes a role
func (s *service) DeleteRole(ctx context.Context, id string) error {
	// Check if role exists
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return util.NewNotFoundError("Role", fmt.Sprintf("%s", id))
	}

	// Delete from database
	if err := s.repo.Delete(ctx, id); err != nil {
		return util.NewDatabaseError("delete role", err)
	}

	return nil
}
