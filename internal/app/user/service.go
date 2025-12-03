package user

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	dbTimeout = 5 * time.Second // Database operation timeout
)

// Service defines the interface for user business logic
type Service interface {
	CreateUser(ctx context.Context, req domain.CreateUserRequest) (*domain.UserResponse, error)
	GetUserByID(ctx context.Context, id string) (*domain.UserResponse, error)
	GetAllUsers(ctx context.Context, page, limit int, search string, currentUserID string) ([]domain.UserResponse, int, error)
	UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.UserResponse, error)
	UpdateProfilePicture(ctx context.Context, id string, profilePictureURL string) (*domain.UserResponse, error)
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
	// Create context with timeout for database operations
	dbCtx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	// Normalize email and username to lowercase for consistent checking
	normalizedEmail := strings.ToLower(strings.TrimSpace(req.Email))
	normalizedUsername := strings.ToLower(strings.TrimSpace(req.Username))

	// Check if user with email already exists
	existingEmail, _ := s.repo.FindByEmail(dbCtx, normalizedEmail)
	if existingEmail != nil {
		return nil, util.NewAlreadyExistsError("User", "email", normalizedEmail)
	}

	// Check if user with username already exists
	existingUsername, _ := s.repo.FindByUsername(dbCtx, normalizedUsername)
	if existingUsername != nil {
		return nil, util.NewAlreadyExistsError("User", "username", normalizedUsername)
	}

	// Validate role
	if !req.Role.IsValid() {
		return nil, util.NewInvalidInputError("Role", "must be Director, DepartmentManager, SectorManager, or Employee")
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, util.NewInternalError(fmt.Sprintf("failed to hash password: %v", err))
	}

	// Create user object
	user := &domain.User{
		Username:     normalizedUsername,
		Email:        normalizedEmail,
		Password:     string(hashedPassword),
		Role:         req.Role,
		Phone:        req.Phone,
		FirstName:    req.FirstName,
		LastName:     req.LastName,
		DepartmentID: req.DepartmentID,
		SectorID:     req.SectorID,
	}

	// Save to database
	if err := s.repo.Create(dbCtx, user); err != nil {
		return nil, util.NewDatabaseError("create user", err)
	}

	response := user.ToResponse()
	return &response, nil
}

// NOTE GetUserByID retrieves a user by ID
func (s *service) GetUserByID(ctx context.Context, id string) (*domain.UserResponse, error) {
	// Create context with timeout for database operations
	dbCtx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	user, err := s.repo.FindByID(dbCtx, id)
	if err != nil {
		return nil, util.NewNotFoundError("User", id)
	}

	response := user.ToResponse()
	return &response, nil
}

// NOTE GetAllUsers retrieves all users with pagination (excluding current user)
func (s *service) GetAllUsers(ctx context.Context, page, limit int, search string, currentUserID string) ([]domain.UserResponse, int, error) {
	// Create context with timeout for database operations
	dbCtx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	// Calculate skip
	skip := (page - 1) * limit

	// Use WaitGroup and channels to run count and find in parallel
	var wg sync.WaitGroup
	var total int
	var users []domain.User
	var countErr, findErr error

	wg.Add(2)

	// Get total count in parallel (excluding current user)
	go func() {
		defer wg.Done()
		total, countErr = s.repo.Count(dbCtx, search, currentUserID)
	}()

	// Get paginated users in parallel (excluding current user)
	go func() {
		defer wg.Done()
		users, findErr = s.repo.FindAll(dbCtx, skip, limit, search, currentUserID)
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

// NOTE UpdateUser updates a user by ID
func (s *service) UpdateUser(ctx context.Context, id string, req domain.UpdateUserRequest) (*domain.UserResponse, error) {
	// Create context with timeout for database operations
	dbCtx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	// Check if user exists
	existingUser, err := s.repo.FindByID(dbCtx, id)
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
			emailUser, _ := s.repo.FindByEmail(dbCtx, normalizedEmail)
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
			usernameUser, _ := s.repo.FindByUsername(dbCtx, normalizedUsername)
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

	// Check if role is being changed and validate it
	if req.Role != "" {
		if !req.Role.IsValid() {
			return nil, util.ErrorResponse(
				"Invalid role",
				util.INVALID_INPUT,
				400,
				"role must be Director, DepartmentManager, SectorManager, or Employee",
			)
		}
		existingUser.Role = req.Role
	}

	// Update phone if provided
	if req.Phone != "" {
		existingUser.Phone = req.Phone
	}

	// Update first name if provided
	if req.FirstName != "" {
		existingUser.FirstName = req.FirstName
	}

	// Update last name if provided
	if req.LastName != "" {
		existingUser.LastName = req.LastName
	}

	// Update password if provided
	if req.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
		if err != nil {
			return nil, util.ErrorResponse(
				"Failed to hash password",
				util.INTERNAL_SERVER_ERROR,
				500,
				err.Error(),
			)
		}
		existingUser.Password = string(hashedPassword)
	}

	// Update in database
	if err := s.repo.Update(dbCtx, id, existingUser); err != nil {
		return nil, util.ErrorResponse(
			"Failed to update user",
			util.DATABASE_ERROR,
			500,
			err.Error(),
		)
	}

	// Fetch updated user
	updatedUser, err := s.repo.FindByID(dbCtx, id)
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

// NOTE UpdateProfilePicture updates a user's profile picture
func (s *service) UpdateProfilePicture(ctx context.Context, id string, profilePictureURL string) (*domain.UserResponse, error) {
	// Create context with timeout for database operations
	dbCtx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	// Check if user exists
	existingUser, err := s.repo.FindByID(dbCtx, id)
	if err != nil {
		return nil, util.ErrorResponse(
			"User not found",
			util.USER_NOT_FOUND,
			404,
			fmt.Sprintf("user with id %s not found", id),
		)
	}

	// Update profile picture
	existingUser.ProfilePicture = profilePictureURL
	fmt.Println(existingUser)
	// Update in database
	if err := s.repo.Update(dbCtx, id, existingUser); err != nil {
		return nil, util.ErrorResponse(
			"Failed to update profile picture",
			util.DATABASE_ERROR,
			500,
			err.Error(),
		)
	}

	// Fetch updated user
	updatedUser, err := s.repo.FindByID(dbCtx, id)
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
	// Create context with timeout for database operations
	dbCtx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	// Check if user exists first
	_, err := s.repo.FindByID(dbCtx, id)
	if err != nil {
		return util.ErrorResponse(
			"User not found",
			util.USER_NOT_FOUND,
			404,
			fmt.Sprintf("user with id %s not found", id),
		)
	}

	if err := s.repo.Delete(dbCtx, id); err != nil {
		return util.ErrorResponse(
			"Failed to delete user",
			util.DATABASE_ERROR,
			500,
			err.Error(),
		)
	}

	return nil
}
