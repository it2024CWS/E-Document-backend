package main

import (
	"context"
	"e-document-backend/internal/app/role"
	"e-document-backend/internal/app/user"
	"e-document-backend/internal/config"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/logger"
	"e-document-backend/internal/platform/mongodb"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init(logger.Config{
		Level:      logger.LogLevel(cfg.Logger.Level),
		Pretty:     cfg.Logger.Pretty,
		TimeFormat: time.RFC3339,
	})

	logger.Info("Starting database seeding...")

	// Connect to MongoDB
	mongoClient, err := mongodb.NewClient(cfg.Database.MongoURI, cfg.Database.DBName)
	if err != nil {
		logger.FatalWithErr("Failed to connect to MongoDB", err)
	}
	defer mongoClient.Disconnect()

	// Initialize repositories
	userRepo := user.NewRepository(mongoClient.Database)
	roleRepo := role.NewRepository(mongoClient.Database)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create Admin role
	adminRole, err := roleRepo.FindByName(ctx, "Admin")
	if err != nil {
		// Role doesn't exist, create it
		adminRole = &domain.Role{
			Name: "Admin",
		}
		if err := roleRepo.Create(ctx, adminRole); err != nil {
			logger.FatalWithErr("Failed to create Admin role", err)
		}
		logger.Info("✓ Admin role created successfully!")
		logger.Infof("  Role ID: %s", adminRole.ID.Hex())
	} else {
		logger.Info("Admin role already exists.")
	}

	// Check if admin user already exists
	existingAdmin, _ := userRepo.FindByEmail(ctx, cfg.Admin.Email)
	if existingAdmin != nil {
		logger.Info("Admin user already exists. Skipping seed.")
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.Admin.Password), bcrypt.DefaultCost)
	if err != nil {
		logger.FatalWithErr("Failed to hash password", err)
	}

	// Create admin user with admin role
	adminUser := &domain.User{
		Username:  cfg.Admin.Username,
		Email:     cfg.Admin.Email,
		Password:  string(hashedPassword),
		RoleID:    adminRole.ID,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save to database
	if err := userRepo.Create(ctx, adminUser); err != nil {
		logger.FatalWithErr("Failed to create admin user", err)
	}

	logger.Info("✓ Admin user created successfully!")
	logger.Infof("  Username: %s", adminUser.Username)
	logger.Infof("  Email: %s", adminUser.Email)
	logger.Infof("  Password: %s", cfg.Admin.Password)
	logger.Infof("  Role: Admin")
	logger.Infof("  ID: %s", adminUser.ID.Hex())
}
