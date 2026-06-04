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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to PostgreSQL
	pgClient, err := postgres.NewClient(ctx, cfg.Database.PostgresDSN)
	if err != nil {
		logger.FatalWithErr("Failed to connect to PostgreSQL", err)
	}
	defer pgClient.Close()

	// Initialize repositories
	userRepo := user.NewPostgresRepository(pgClient.Pool)

	// Step 1: Seed master / reference data
	// (roles → departments → sectors → doc_types)
	logger.Info("Seeding master data...")
	if err := seed.SeedMasterData(ctx, pgClient.Pool); err != nil {
		logger.FatalWithErr("Failed to seed master data", err)
	}
	logger.Info("✓ Master data seeded successfully!")

	// Step 2: Seed system users
	// (depends on roles being present)
	logger.Info("Seeding system users...")
	if err := seed.SeedUsers(ctx, userRepo, pgClient.Pool); err != nil {
		logger.FatalWithErr("Failed to seed users", err)
	}
	logger.Info("✓ Users seeded successfully!")

	logger.Info("🎉 Database seeding complete!")
}

