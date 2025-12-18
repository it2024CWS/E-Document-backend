package main

import (
	"context"
	"e-document-backend/internal/app/user"
	"e-document-backend/internal/config"
	"e-document-backend/internal/logger"
	"e-document-backend/internal/pkg/seed"
	"e-document-backend/internal/platform/postgres"
	"time"
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

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Connect to PostgreSQL
	pgClient, err := postgres.NewClient(ctx, cfg.Database.PostgresDSN)
	if err != nil {
		logger.FatalWithErr("Failed to connect to PostgreSQL", err)
	}
	defer pgClient.Close()

	// Initialize repositories
	userRepo := user.NewPostgresRepository(pgClient.Pool)

	// Seed admin user using shared seeder
	if err := seed.SeedAdmin(ctx, userRepo, cfg); err != nil {
		logger.FatalWithErr("Failed to seed admin user", err)
	}

	logger.Info("✓ Admin user seeded successfully!")
}
