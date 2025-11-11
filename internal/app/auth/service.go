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

// Service defines the interface for authentication business logic
type Service interface {
	Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error)
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
func (s *service) Login(ctx context.Context, req domain.LoginRequest) (*domain.AuthResponse, error) {
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

	response := &domain.AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    s.cfg.JWT.AccessTokenExpiry,
	}

	return response, nil
}

// RefreshToken generates new tokens using a valid refresh token
func (s *service) RefreshToken(ctx context.Context, refreshToken string) (*domain.AuthResponse, error) {
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

	response := &domain.AuthResponse{
		User:         user.ToResponse(),
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    s.cfg.JWT.AccessTokenExpiry,
	}

	return response, nil
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

// generateAccessToken creates a new access token for the user
func (s *service) generateAccessToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID.Hex(),
		"username": user.Username,
		"email":    user.Email,
		"type":     "access",
		"exp":      time.Now().Add(time.Duration(s.cfg.JWT.AccessTokenExpiry) * time.Second).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.AccessTokenSecret))
}

// generateRefreshToken creates a new refresh token for the user
func (s *service) generateRefreshToken(user *domain.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID.Hex(),
		"username": user.Username,
		"email":    user.Email,
		"type":     "refresh",
		"exp":      time.Now().Add(time.Duration(s.cfg.JWT.RefreshTokenExpiry) * time.Second).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.RefreshTokenSecret))
}

// ValidateAccessToken validates and parses an access token
func (s *service) ValidateAccessToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.AccessTokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenType, ok := claims["type"].(string)
		if !ok || tokenType != "access" {
			return nil, fmt.Errorf("invalid token type")
		}

		userID, _ := claims["user_id"].(string)
		username, _ := claims["username"].(string)
		email, _ := claims["email"].(string)

		return &domain.TokenClaims{
			UserID:   userID,
			Username: username,
			Email:    email,
			Type:     tokenType,
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}

// ValidateRefreshToken validates and parses a refresh token
func (s *service) ValidateRefreshToken(tokenString string) (*domain.TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.RefreshTokenSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		tokenType, ok := claims["type"].(string)
		if !ok || tokenType != "refresh" {
			return nil, fmt.Errorf("invalid token type")
		}

		userID, _ := claims["user_id"].(string)
		username, _ := claims["username"].(string)
		email, _ := claims["email"].(string)

		return &domain.TokenClaims{
			UserID:   userID,
			Username: username,
			Email:    email,
			Type:     tokenType,
		}, nil
	}

	return nil, fmt.Errorf("invalid token")
}
