package user

import (
	"context"
	"e-document-backend/internal/domain"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// Service defines the interface for user business logic
type Service interface {
	CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*domain.UserResponse, error)
	GetAllUsers(ctx context.Context) ([]domain.UserResponse, error)
	UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.UserResponse, error)
	DeleteUser(ctx context.Context, id string) error
}

// service implements the Service interface
type service struct {
	repo Repository
}

// NewService creates a new user service
func NewService(repo Repository) Service {
	return &service{
		repo: repo,
	}
}

// CreateUser creates a new user
func (s *service) CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.UserResponse, error) {
	// Check if user with email already exists
	existingUser, _ := s.repo.FindByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user object
	user := &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	// Save to database
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// GetUserByID retrieves a user by ID
func (s *service) GetUserByID(ctx context.Context, id string) (*domain.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := user.ToResponse()
	return &response, nil
}

// GetAllUsers retrieves all users
func (s *service) GetAllUsers(ctx context.Context) ([]domain.UserResponse, error) {
	users, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]domain.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, nil
}

// UpdateUser updates a user by ID
func (s *service) UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.UserResponse, error) {
	// Check if user exists
	existingUser, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check if email is being changed and if it already exists
	if req.Email != "" && req.Email != existingUser.Email {
		emailUser, _ := s.repo.FindByEmail(ctx, req.Email)
		if emailUser != nil {
			return nil, fmt.Errorf("user with email %s already exists", req.Email)
		}
		existingUser.Email = req.Email
	}

	// Update fields
	if req.Name != "" {
		existingUser.Name = req.Name
	}

	// Update in database
	if err := s.repo.Update(ctx, id, existingUser); err != nil {
		return nil, err
	}

	// Fetch updated user
	updatedUser, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}

	response := updatedUser.ToResponse()
	return &response, nil
}

// DeleteUser deletes a user by ID
func (s *service) DeleteUser(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
