package user

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// Service defines the interface for user business logic
type Service interface {
	CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*domain.UserResponse, error)
	GetAllUsers(ctx context.Context, page, limit int, search string) ([]domain.UserResponse, int, error)
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

// NOTE CreateUser creates a new user
func (s *service) CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.UserResponse, error) {
	// Normalize email and username to lowercase for consistent checking
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))
	normalizedUsername := strings.ToLower(strings.TrimSpace(req.Username))

	// Check if user with email already exists
	existingEmail, _ := s.repo.FindByEmail(ctx, normalizedEmail)
	if existingEmail != nil {
		return nil, util.ErrorResponse(
			"Email already exists",
			util.EMAIL_ALREADY_EXISTS,
			400,
			fmt.Sprintf("user with email %s already exists", normalizedEmail),
		)
	}

	// Check if user with username already exists
	existingUsername, _ := s.repo.FindByUsername(ctx, normalizedUsername)
	if existingUsername != nil {
		return nil, util.ErrorResponse(
			"Username already exists",
			util.USER_ALREADY_EXISTS,
			400,
			fmt.Sprintf("user with username %s already exists", normalizedUsername),
		)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, util.ErrorResponse(
			"Failed to hash password",
			util.INTERNAL_SERVER_ERROR,
			500,
			err.Error(),
		)
	}

	// Create user object
	user := &domain.User{
		Username: normalizedUsername,
		Email:    normalizedEmail,
		Password: string(hashedPassword),
	}

	// Save to database
	if err := s.repo.Create(ctx, user); err != nil {
		return nil, util.ErrorResponse(
			"Failed to create user",
			util.DATABASE_ERROR,
			500,
			err.Error(),
		)
	}

	response := user.ToResponse()
	return &response, nil
}

// NOTE GetUserByID retrieves a user by ID
func (s *service) GetUserByID(ctx context.Context, id string) (*domain.UserResponse, error) {
	user, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.ErrorResponse(
			"User not found",
			util.USER_NOT_FOUND,
			404,
			fmt.Sprintf("user with id %s not found", id),
		)
	}

	response := user.ToResponse()
	return &response, nil
}

// NOTE GetAllUsers retrieves all users with pagination
func (s *service) GetAllUsers(ctx context.Context, page, limit int, search string) ([]domain.UserResponse, int, error) {
	// Calculate skip
	skip := (page - 1) * limit

	// Use WaitGroup and channels to run count and find in parallel
	var wg sync.WaitGroup
	var total int
	var users []domain.User
	var countErr, findErr error

	wg.Add(2)

	// Get total count in parallel
	go func() {
		defer wg.Done()
		total, countErr = s.repo.Count(ctx, search)
	}()

	// Get paginated users in parallel
	go func() {
		defer wg.Done()
		users, findErr = s.repo.FindAll(ctx, skip, limit, search)
	}()

	// Wait for both operations to complete
	wg.Wait()

	// Check for errors
	if countErr != nil {
		return nil, 0, util.ErrorResponse(
			"Failed to count users",
			util.DATABASE_ERROR,
			500,
			countErr.Error(),
		)
	}

	if findErr != nil {
		return nil, 0, util.ErrorResponse(
			"Failed to fetch users",
			util.DATABASE_ERROR,
			500,
			findErr.Error(),
		)
	}

	// Convert to responses
	responses := make([]domain.UserResponse, len(users))
	for i, user := range users {
		responses[i] = user.ToResponse()
	}

	return responses, total, nil
}

// UpdateUser updates a user by ID
func (s *service) UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.UserResponse, error) {
	// Check if user exists
	existingUser, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.ErrorResponse(
			"User not found",
			util.USER_NOT_FOUND,
			404,
			fmt.Sprintf("user with id %s not found", id),
		)
	}

	// Check if email is being changed and if it already exists
	if req.Email != "" {
		normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))
		if normalizedEmail != existingUser.Email {
			emailUser, _ := s.repo.FindByEmail(ctx, normalizedEmail)
			if emailUser != nil {
				return nil, util.ErrorResponse(
					"Email already exists",
					util.EMAIL_ALREADY_EXISTS,
					400,
					fmt.Sprintf("user with email %s already exists", normalizedEmail),
				)
			}
			existingUser.Email = normalizedEmail
		}
	}

	// Check if username is being changed and if it already exists
	if req.Username != "" {
		normalizedUsername := strings.ToLower(strings.TrimSpace(req.Username))
		if normalizedUsername != existingUser.Username {
			usernameUser, _ := s.repo.FindByUsername(ctx, normalizedUsername)
			if usernameUser != nil {
				return nil, util.ErrorResponse(
					"Username already exists",
					util.USER_ALREADY_EXISTS,
					400,
					fmt.Sprintf("user with username %s already exists", normalizedUsername),
				)
			}
			existingUser.Username = normalizedUsername
		}
	}

	// Update in database
	if err := s.repo.Update(ctx, id, existingUser); err != nil {
		return nil, util.ErrorResponse(
			"Failed to update user",
			util.DATABASE_ERROR,
			500,
			err.Error(),
		)
	}

	// Fetch updated user
	updatedUser, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, util.ErrorResponse(
			"Failed to fetch updated user",
			util.DATABASE_ERROR,
			500,
			err.Error(),
		)
	}

	response := updatedUser.ToResponse()
	return &response, nil
}

// DeleteUser deletes a user by ID
func (s *service) DeleteUser(ctx context.Context, id string) error {
	// Check if user exists first
	_, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return util.ErrorResponse(
			"User not found",
			util.USER_NOT_FOUND,
			404,
			fmt.Sprintf("user with id %s not found", id),
		)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return util.ErrorResponse(
			"Failed to delete user",
			util.DATABASE_ERROR,
			500,
			err.Error(),
		)
	}

	return nil
}
