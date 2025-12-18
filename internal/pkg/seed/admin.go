package seed

import (
	"context"
	"e-document-backend/internal/app/user"
	"e-document-backend/internal/config"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/logger"

	"golang.org/x/crypto/bcrypt"
)

// SeedAdmin creates an admin user if it doesn't exist
func SeedAdmin(ctx context.Context, userRepo user.Repository, cfg *config.Config) error {
	// Check if admin user already exists
	existingUser, err := userRepo.FindByEmail(ctx, cfg.Admin.Email)
	if err == nil && existingUser != nil {
		logger.Info("Admin user already exists, skipping seed")
		return nil
	}

	// Hash the admin password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.Admin.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	// Create admin user
	adminUser := &domain.User{
		Username:  cfg.Admin.Username,
		Email:     cfg.Admin.Email,
		Password:  string(hashedPassword),
		FirstName: "Admin",
		LastName:  "User",
		Phone:     "000-000-0000",
		Role:      domain.RoleDirector, // Admin has Director role
	}

	if err := userRepo.Create(ctx, adminUser); err != nil {
		return err
	}

	logger.Infof("Admin user created successfully: %s (%s)", adminUser.Username, adminUser.Email)
	return nil
}
