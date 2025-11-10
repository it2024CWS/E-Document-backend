# E-Document Backend

A RESTful API backend for E-Document system built with Go, Echo framework, and MongoDB.

## Project Structure

```
E-Document-backend/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point
├── internal/
│   ├── app/
│   │   └── user/
│   │       ├── handler.go       # HTTP handlers
│   │       ├── service.go       # Business logic
│   │       └── repository.go    # Data access
│   ├── config/
│   │   └── config.go            # Configuration loader
│   ├── domain/
│   │   └── user.go              # Domain models
│   ├── middleware/
│   │   ├── auth.go              # Authentication
│   │   └── ratelimit.go         # Rate limiting
│   ├── platform/
│   │   └── mongodb/
│   │       └── client.go        # MongoDB connection
│   └── util/
│       └── response.go          # API response helpers
├── .env                          # Environment variables (not in git)
├── .env.example                  # Environment template
├── .air.toml                     # Air hot reload config
├── Makefile                      # Build commands
└── go.mod                        # Go dependencies
```

## Prerequisites

- Go 1.21 or higher
- MongoDB Atlas account (or local MongoDB)
- Make (optional, for Makefile commands)

## Setup

1. **Clone the repository**

2. **Copy environment file**
   ```bash
   cp .env.example .env
   ```

3. **Update .env with your configuration**
   ```env
   PORT=8081
   MONGO_URI=your_mongodb_connection_string
   DB_NAME=e_document_db
   ```

4. **Install dependencies**
   ```bash
   go mod tidy
   ```

5. **Seed the database (optional)**
   ```bash
   make seed
   ```
   This creates an admin user:
   - Email: `admin@edocument.com`
   - Password: `password`
   - Name: `admin`

## Running the Project

### Option 1: Using Make (Recommended)

```bash
# Run in development mode
make dev

# Show all available commands
make help

# Seed database with admin user
make seed

# Build the application
make build

# Run tests
make test

# Clean build artifacts
make clean
```

### Option 2: Direct Go Command

```bash
go run cmd/api/main.go
```

### Option 3: With Hot Reload (Air)

```bash
# Install Air (first time only)
make install-air

# Run with hot reload
make air
```

## API Endpoints

### Health Check
- `GET /api/v1/health` - Check server status

### User Management
- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users` - Get all users
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

## Testing with Postman

1. Import `E-Document-API.postman_collection.json`
2. Import `E-Document.postman_environment.json`
3. Start testing the API endpoints

## Architecture

This project follows **Layered Architecture** pattern:

- **Handler Layer**: Handles HTTP requests/responses
- **Service Layer**: Contains business logic
- **Repository Layer**: Manages data access

## Features

- RESTful API design
- MongoDB integration
- Environment-based configuration
- Rate limiting
- Authentication middleware (ready to use)
- CORS enabled
- Graceful shutdown
- Hot reload support (Air)

## Development

The project uses:
- **Echo v4**: High-performance web framework
- **MongoDB Driver**: Official Go driver for MongoDB
- **godotenv**: Environment variable management
- **Air**: Hot reload for development

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Server port | 8080 |
| MONGO_URI | MongoDB connection string | - |
| DB_NAME | Database name | e_document_db |

## License

MIT
