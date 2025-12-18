package main

import (
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	_ = godotenv.Load()

	// Get database DSN from environment
	dsn := os.Getenv("POSTGRES_DSN")
	if dsn == "" {
		log.Fatal("POSTGRES_DSN is not set")
		os.Exit(1)
	}

	// Create migration instance
	m, err := migrate.New(
		"file://migrations",
		dsn,
	)
	if err != nil {
		log.Fatal("Failed to create migration instance:", err)
	}

	// Get command from arguments
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Migration up failed:", err)
		}
		fmt.Println("✅ Migrations applied successfully!")

	case "down":
		if err := m.Down(); err != nil && err != migrate.ErrNoChange {
			log.Fatal("Migration down failed:", err)
		}
		fmt.Println("✅ Migrations rolled back successfully!")

	case "force":
		if len(os.Args) < 3 {
			fmt.Println("Usage: migrate force <version>")
			os.Exit(1)
		}
		version := os.Args[2]
		var v int
		_, err := fmt.Sscanf(version, "%d", &v)
		if err != nil {
			log.Fatal("Invalid version number:", err)
		}
		if err := m.Force(v); err != nil {
			log.Fatal("Force migration failed:", err)
		}
		fmt.Printf("✅ Forced migration to version %d\n", v)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatal("Failed to get version:", err)
		}
		if dirty {
			fmt.Printf("Current version: %d (dirty)\n", version)
		} else {
			fmt.Printf("Current version: %d\n", version)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Database Migration Tool")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  migrate up           - Apply all pending migrations")
	fmt.Println("  migrate down         - Rollback all migrations")
	fmt.Println("  migrate force <ver>  - Force set version without running migrations")
	fmt.Println("  migrate version      - Show current migration version")
}
