package seed

import (
	"context"
	"e-document-backend/internal/app/user"
	"e-document-backend/internal/config"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/logger"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

// SeedAdmin creates an admin user if it doesn't exist
func SeedAdmin(ctx context.Context, userRepo user.Repository, pool *pgxpool.Pool, cfg *config.Config) error {
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
	// Find Director role ID
	var roleID uuid.UUID
	err = pool.QueryRow(ctx, "SELECT id FROM user_roles WHERE role_name = 'Director'").Scan(&roleID)
	if err != nil {
		return fmt.Errorf("failed to find Director role: %w", err)
	}

	isActive := true
	adminUser := &domain.User{
		Username:  cfg.Admin.Username,
		Email:     cfg.Admin.Email,
		Password:  string(hashedPassword),
		Firstname: "Admin",
		Lastname:  "User",
		Phone:     "000-000-0000",
		RoleID:    &roleID, // Director role
		IsActive:  isActive,
	}

	if err := userRepo.Create(ctx, adminUser); err != nil {
		return err
	}

	logger.Infof("Admin user created successfully: %s (%s)", adminUser.Username, adminUser.Email)
	return nil
}
