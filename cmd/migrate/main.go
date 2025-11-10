package main

import (
	"context"
	"e-document-backend/internal/config"
	"e-document-backend/internal/migration"
	"e-document-backend/internal/platform/mongodb"
	"e-document-backend/migrations"
	"flag"
	"log"
	"os"
	"time"
)

func main() {
	// Define command flags
	upCmd := flag.Bool("up", false, "Run all pending migrations")
	downCmd := flag.Bool("down", false, "Rollback the last migration")
	statusCmd := flag.Bool("status", false, "Show migration status")

	flag.Parse()

	// Validate command
	if !*upCmd && !*downCmd && !*statusCmd {
		log.Println("Usage:")
		log.Println("  go run cmd/migrate/main.go -up       # Run all pending migrations")
		log.Println("  go run cmd/migrate/main.go -down     # Rollback last migration")
		log.Println("  go run cmd/migrate/main.go -status   # Show migration status")
		log.Println("")
		log.Println("Or use make commands:")
		log.Println("  make migrate-up")
		log.Println("  make migrate-down")
		log.Println("  make migrate-status")
		os.Exit(1)
	}

	// Load configuration
	cfg := config.Load()

	// Connect to MongoDB
	mongoClient, err := mongodb.NewClient(cfg.Database.MongoURI, cfg.Database.DBName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect()

	// Create migration runner
	runner := migration.NewRunner(mongoClient.Database)

	// Register all migrations
	migrations.RegisterMigrations(runner)

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Execute command
	switch {
	case *upCmd:
		if err := runner.Up(ctx); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}

	case *downCmd:
		if err := runner.Down(ctx); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}

	case *statusCmd:
		if err := runner.Status(ctx); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}
	}
}
