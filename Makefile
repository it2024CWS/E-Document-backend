.PHONY: help dev run build clean test install-air air seed migrate-up migrate-down migrate-status

# Help command - shows all available commands
help:
	@echo "Available commands:"
	@echo "  make dev             - Run the server in development mode"
	@echo "  make run             - Run the server"
	@echo "  make build           - Build the application"
	@echo "  make clean           - Clean build artifacts"
	@echo "  make test            - Run tests"
	@echo "  make seed            - Seed the database with initial admin user"
	@echo "  make migrate-up      - Run all pending migrations"
	@echo "  make migrate-down    - Rollback the last migration"
	@echo "  make migrate-status  - Show migration status"
	@echo "  make install-air     - Install Air for hot reload"
	@echo "  make air             - Run with Air hot reload"

# Run in development mode
dev:
	@echo "Starting development server..."
	go run cmd/api/main.go

# Run the application
run:
	go run cmd/api/main.go

# Build the application
build:
	@echo "Building application..."
	go build -o bin/server cmd/api/main.go
	@echo "Build complete! Binary: bin/server"

# Clean build artifacts
clean:
	@echo "Cleaning..."
	rm -rf bin/
	go clean

# Run tests
test:
	go test -v ./...

# Install Air for hot reload (development)
install-air:
	@echo "Installing Air..."
	go install github.com/air-verse/air@latest
	@echo "Air installed! Use 'make air' to run with hot reload"

# Run with Air hot reload
air:
	@echo "Starting server with hot reload..."
	air

# Seed the database with initial data
seed:
	@echo "Seeding database..."
	go run cmd/seed/main.go

# Run all pending migrations
migrate-up:
	@echo "Running migrations..."
	go run cmd/migrate/main.go -up

# Rollback the last migration
migrate-down:
	@echo "Rolling back migration..."
	go run cmd/migrate/main.go -down

# Show migration status
migrate-status:
	@echo "Migration status:"
	go run cmd/migrate/main.go -status
