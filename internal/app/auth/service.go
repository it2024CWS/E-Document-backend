package auth

import (
	"context"
	"e-document-backend/internal/app/user"
	"e-document-backend/internal/config"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"fmt"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// AuthResult contains tokens and response for internal use
type AuthResult struct {
	Response     *domain.AuthResponse
	AccessToken  string
	RefreshToken string
}

// Service defines the interface for authentication business logic
type Service interface {
	Login(ctx context.Context, req domain.LoginRequest) (*AuthResult, error)
	RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error)
	GetProfile(ctx context.Context, userID string) (*domain.UserResponse, error)
	ValidateAccessToken(tokenString string) (*domain.TokenClaims, error)
	ValidateRefreshToken(tokenString string) (*domain.TokenClaims, error)
}

// service implements the Service interface
type service struct {
	userRepo user.Repository
	cfg      *config.Config
}

// NewService creates a new auth service
func NewService(userRepo user.Repository, cfg *config.Config) Service {
	return &service{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// Login authenticates a user with username/email and password
func (s *service) Login(ctx context.Context, req domain.LoginRequest) (*AuthResult, error) {
	// Normalize username or email to lowercase
	usernameOrEmail := strings.ToLower(strings.TrimSpace(req.UsernameOrEmail))

	// Try to find user by email first, then by username
	var user *domain.User
	var err error

	// Check if it looks like an email (contains @)
	if strings.Contains(usernameOrEmail, "@") {
		user, err = s.userRepo.FindByEmail(ctx, usernameOrEmail)
	} else {
		user, err = s.userRepo.FindByUsername(ctx, usernameOrEmail)
	}

	if err != nil {
		return nil, util.ErrorResponse(
			"Invalid credentials",
			util.INVALID_CREDENTIALS,
			401,
			"username/email or password is incorrect",
		)
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, util.ErrorResponse(
			"Invalid credentials",
			util.INCORRECT_PASSWORD,
			401,
			"username/email or password is incorrect",
		)
	}

	// Generate tokens
	accessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, util.ErrorResponse(
			"Failed to generate access token",
			util.INTERNAL_SERVER_ERROR,
			500,
			err.Error(),
		)
	}

	refreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, util.ErrorResponse(
			"Failed to generate refresh token",
			util.INTERNAL_SERVER_ERROR,
			500,
			err.Error(),
		)
	}

	result := &AuthResult{
		Response: &domain.AuthResponse{
			User: user.ToResponse(),
		},
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}

	return result, nil
}

// RefreshToken generates new tokens using a valid refresh token
func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*AuthResult, error) {
	// Validate refresh token
	claims, err := s.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, util.ErrorResponse(
			"Invalid refresh token",
			util.INVALID_TOKEN,
			401,
			err.Error(),
		)
	}

	// Get user from database
	user, err := s.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		return nil, util.ErrorResponse(
			"User not found",
			util.USER_NOT_FOUND,
			404,
			"user associated with token not found",
		)
	}

	// Generate new tokens
	newAccessToken, err := s.generateAccessToken(user)
	if err != nil {
		return nil, util.ErrorResponse(
			"Failed to generate access token",
			util.INTERNAL_SERVER_ERROR,
			500,
			err.Error(),
		)
	}

	newRefreshToken, err := s.generateRefreshToken(user)
	if err != nil {
		return nil, util.ErrorResponse(
			"Failed to generate refresh token",
			util.INTERNAL_SERVER_ERROR,
			500,
			err.Error(),
		)
	}

	result := &AuthResult{
		Response: &domain.AuthResponse{
			User: user.ToResponse(),
		},
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
	}

	return result, nil
}

// GetProfile retrieves user profile by user ID
func (s *service) GetProfile(ctx context.Context, userID string) (*domain.UserResponse, error) {
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, util.ErrorResponse(
			"User not found",
			util.USER_NOT_FOUND,
			404,
			fmt.Sprintf("user with id %s not found", userID),
		)
	}

	response := user.ToResponse()
	return &response, nil
}

// buildUserClaims creates JWT claims for a user
func (s *service) buildUserClaims(user *domain.User, tokenType string, expiry int64) jwt.MapClaims {
	return jwt.MapClaims{
		"user_id":       user.ID.Hex(),
		"username":      user.Username,
		"email":         user.Email,
		"phone":         user.Phone,
		"first_name":    user.FirstName,
		"last_name":     user.LastName,
		"role":          user.Role.String(),
		"department_id": user.DepartmentID,
		"sector_id":     user.SectorID,
		"type":          tokenType,
		"exp":           time.Now().Add(time.Duration(expiry) * time.Second).Unix(),
		"iat":           time.Now().Unix(),
	}
}

// generateToken creates and signs a JWT token
func (s *service) generateToken(claims jwt.MapClaims, secret string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

// generateAccessToken creates a new access token for the user
func (s *service) generateAccessToken(user *domain.User) (string, error) {
	claims := s.buildUserClaims(user, "access", s.cfg.JWT.AccessTokenExpiry)
	return s.generateToken(claims, s.cfg.JWT.AccessTokenSecret)
}

// generateRefreshToken creates a new refresh token for the user
func (s *service) generateRefreshToken(user *domain.User) (string, error) {
	claims := s.buildUserClaims(user, "refresh", s.cfg.JWT.RefreshTokenExpiry)
	return s.generateToken(claims, s.cfg.JWT.RefreshTokenSecret)
}

// parseTokenClaims extracts TokenClaims from JWT MapClaims
func parseTokenClaims(claims jwt.MapClaims) *domain.TokenClaims {
	userID, _ := claims["user_id"].(string)
	username, _ := claims["username"].(string)
	email, _ := claims["email"].(string)
	phone, _ := claims["phone"].(string)
	firstName, _ := claims["first_name"].(string)
	lastName, _ := claims["last_name"].(string)
	role, _ := claims["role"].(string)
	departmentID, _ := claims["department_id"].(string)
	sectorID, _ := claims["sector_id"].(string)
	tokenType, _ := claims["type"].(string)

	return &domain.TokenClaims{
		UserID:       userID,
		Username:     username,
		Email:        email,
		Phone:        phone,
		FirstName:    firstName,
		LastName:     lastName,
		Role:         role,
		DepartmentID: departmentID,
		SectorID:     sectorID,
		Type:         tokenType,
	}
}

// validateToken validates a JWT token with the given secret and expected type
func (s *service) validateToken(tokenString, secret, expectedType string) (*domain.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenType, ok := claims["type"].(string)
		if !ok || tokenType != expectedType {
			return nil, fmt.Errorf("invalid token type: expected %s, got %s", expectedType, tokenType)
		}

		return parseTokenClaims(claims), nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateAccessToken validates and parses an access token
func (s *service) ValidateAccessToken(tokenString string) (*domain.TokenClaims, error) {
	return s.validateToken(tokenString, s.cfg.JWT.AccessTokenSecret, "access")
}

// ValidateRefreshToken validates and parses a refresh token
func (s *service) ValidateRefreshToken(tokenString string) (*domain.TokenClaims, error) {
	return s.validateToken(tokenString, s.cfg.JWT.RefreshTokenSecret, "refresh")
}
