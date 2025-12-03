# E-Document Backend

A production-ready RESTful API backend for E-Document Management System built with Go, Echo framework, and MongoDB. This project follows industry-standard architectural patterns and best practices.

## Architecture & Design Patterns

This project implements **Clean Architecture** with the following patterns:

### Architectural Patterns
- **Clean Architecture / Layered Architecture**: Clear separation between Handler, Service, and Repository layers
- **Dependency Injection**: All dependencies injected through constructors using interfaces
- **Repository Pattern**: Abstracts data access logic from business logic

### Design Patterns
- **Factory Pattern**: `NewRepository()`, `NewService()`, `NewHandler()` constructors
- **Middleware Pattern**: Authentication, logging, rate limiting
- **DTO Pattern**: Separate request/response objects from domain models
- **Strategy Pattern**: Context-based timeout strategies for different operations
- **Singleton Pattern**: Validator instance, Logger instance

### SOLID Principles
- ✅ **Single Responsibility**: Each struct has one clear purpose
- ✅ **Open/Closed**: Extensible through interfaces
- ✅ **Liskov Substitution**: Interfaces are properly substitutable
- ✅ **Interface Segregation**: Focused, minimal interfaces
- ✅ **Dependency Inversion**: Depends on abstractions (interfaces), not concretions

## Project Structure

```
E-Document-backend/
├── cmd/
│   ├── api/
│   │   └── main.go              # Application entry point
│   ├── migrate/
│   │   └── main.go              # Database migration runner
│   └── seed/
│       └── main.go              # Database seeder
│
├── internal/
│   ├── app/                     # Application modules
│   │   ├── auth/
│   │   │   ├── handler.go       # Auth HTTP handlers
│   │   │   └── service.go       # Auth business logic
│   │   └── user/
│   │       ├── handler.go       # User HTTP handlers
│   │       ├── service.go       # User business logic
│   │       └── repository.go    # User data access
│   │
│   ├── config/
│   │   └── config.go            # Configuration management
│   │
│   ├── domain/
│   │   └── user.go              # Domain models & DTOs
│   │
│   ├── logger/
│   │   └── logger.go            # Structured logging
│   │
│   ├── middleware/
│   │   ├── auth.go              # JWT authentication
│   │   ├── logger.go            # Request/response logging
│   │   └── ratelimit.go         # Rate limiting
│   │
│   ├── migration/
│   │   └── migration.go         # Database migrations
│   │
│   ├── platform/
│   │   └── mongodb/
│   │       └── client.go        # MongoDB connection & config
│   │
│   └── util/
│       ├── error-code.go        # Error code constants
│       ├── errors.go            # Custom error types & helpers
│       ├── response.go          # API response helpers
│       └── validator.go         # Validation utilities
│
├── docs/                        # Swagger documentation (auto-generated)
├── migrations/                  # Database migration files
├── tmp/                         # Temporary files (Air)
│
├── .env                         # Environment variables (not in git)
├── .env.example                 # Environment template
├── .air.toml                    # Air hot reload config
├── .gitignore                   # Git ignore rules
├── Makefile                     # Build & development commands
├── go.mod                       # Go dependencies
└── go.sum                       # Dependency checksums
```

## Layer Responsibilities

### Handler Layer (Presentation)
- HTTP request/response handling
- Input validation using `go-playground/validator`
- Error handling and response formatting
- Authentication/authorization checks

### Service Layer (Business Logic)
- Business rules and validation
- Data transformation
- Orchestration of multiple repositories
- Transaction management
- Context timeout handling (5s default)

### Repository Layer (Data Access)
- Database operations (CRUD)
- Query building
- Data mapping
- Error wrapping

## Prerequisites

- Go 1.24 or higher
- MongoDB Atlas account (or local MongoDB)
- Make (optional, for Makefile commands)

## Features

### Security & Authentication
- ✅ JWT-based authentication (access & refresh tokens)
- ✅ Password hashing with bcrypt
- ✅ Protected routes with middleware
- ✅ Rate limiting (20 req/s, burst 50)
- ✅ CORS configuration
- ✅ Context timeout protection (5s)

### Code Quality & Best Practices
- ✅ Clean Architecture pattern
- ✅ Dependency Injection
- ✅ Interface-based programming
- ✅ Structured error handling with custom error types
- ✅ Input validation using `go-playground/validator/v10`
- ✅ Consistent error responses
- ✅ Graceful shutdown
- ✅ Request ID tracking
- ✅ Comprehensive logging

### API Features
- ✅ RESTful API design
- ✅ Pagination support
- ✅ Search functionality
- ✅ Swagger documentation
- ✅ Health check endpoint
- ✅ Exclude current user from user listing

### Database
- ✅ MongoDB with official Go driver
- ✅ Connection pooling
- ✅ Database migration support
- ✅ Seeding utilities
- ✅ Context-aware queries

## Setup

1. **Clone the repository**

2. **Copy environment file**
   ```bash
   cp .env.example .env
   ```

3. **Update .env with your configuration**
   ```env
   # Server Configuration
   PORT=5000

   # Database Configuration
   MONGO_URI=your_mongodb_connection_string
   DB_NAME=e_document_db

   # JWT Configuration
   JWT_SECRET=your-secret-key-here
   JWT_ACCESS_EXPIRATION=15m
   JWT_REFRESH_EXPIRATION=7d

   # Logger Configuration
   LOG_LEVEL=debug
   LOG_PRETTY=true
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

The server will start on `http://localhost:5000`

## API Documentation

After starting the server, access Swagger documentation at:
```
http://localhost:5000/swagger/index.html
```

## API Endpoints

### Health Check
- `GET /api/health` - Check server status

### Authentication
- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login user
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout user
- `GET /api/v1/auth/me` - Get current user profile

### User Management (Protected)
- `POST /api/v1/users` - Create a new user
- `GET /api/v1/users` - Get all users (paginated, excludes current user)
  - Query params: `page`, `limit`, `search`
- `GET /api/v1/users/:id` - Get user by ID
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

## Request Validation

The API uses `go-playground/validator` for automatic validation:

```go
type CreateUserRequest struct {
    Username  string   `json:"username" validate:"required"`
    Email     string   `json:"email" validate:"required,email"`
    Password  string   `json:"password" validate:"required,min=6"`
    Role      UserRole `json:"role" validate:"required,oneof=Director DepartmentManager SectorManager Employee"`
    Phone     string   `json:"phone" validate:"required,e164"`
    FirstName string   `json:"first_name" validate:"required"`
    LastName  string   `json:"last_name" validate:"required"`
}
```

Validation errors return user-friendly messages:
```json
{
  "success": false,
  "message": "Validation failed",
  "error": {
    "code": "INVALID_INPUT",
    "detail": "Email must be a valid email address; Password must be at least 6 characters"
  }
}
```

## Error Handling

The API uses custom error types for consistent error responses:

```go
// Helper functions for common errors
util.NewNotFoundError("User", id)
util.NewAlreadyExistsError("User", "email", email)
util.NewValidationError("Invalid input")
util.NewDatabaseError("create user", err)
util.NewUnauthorizedError("Invalid token")
util.NewInternalError("Something went wrong")
```

Example error response:
```json
{
  "success": false,
  "message": "User not found",
  "error": {
    "code": "USER_NOT_FOUND",
    "detail": "User with identifier 507f1f77bcf86cd799439011 was not found"
  }
}
```

## Testing with Postman

1. Import `E-Document-API.postman_collection.json`
2. Import `E-Document.postman_environment.json`
3. Start testing the API endpoints

## Development

The project uses:
- **Echo v4**: High-performance, minimalist web framework
- **MongoDB Driver**: Official Go driver for MongoDB
- **JWT-Go**: JSON Web Token implementation
- **godotenv**: Environment variable management
- **bcrypt**: Password hashing
- **go-playground/validator**: Struct validation
- **Swagger**: API documentation
- **Air**: Hot reload for development

## Best Practices Implemented

### Code Organization
- Domain-Driven Design structure
- Clear separation of concerns
- Package organization by feature

### Error Handling
- Custom error types with error codes
- Consistent error responses
- Wrapped errors for better debugging

### Security
- Password hashing with bcrypt
- JWT token authentication
- CORS configuration
- Rate limiting
- Input validation

### Performance
- Connection pooling
- Parallel database queries (where applicable)
- Context timeout protection
- Efficient pagination

### Code Quality
- Interface-based design for testability
- Dependency injection
- Consistent naming conventions
- Comprehensive comments

## Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| PORT | Server port | 5000 | No |
| MONGO_URI | MongoDB connection string | - | Yes |
| DB_NAME | Database name | e_document_db | Yes |
| JWT_SECRET | Secret key for JWT | - | Yes |
| JWT_ACCESS_EXPIRATION | Access token expiration | 15m | No |
| JWT_REFRESH_EXPIRATION | Refresh token expiration | 7d | No |
| LOG_LEVEL | Logging level (debug, info, warn, error) | info | No |
| LOG_PRETTY | Pretty print logs | false | No |

## Project Timeline & Improvements

### Recent Improvements
- ✅ Implemented custom error handling system
- ✅ Added go-playground/validator for validation
- ✅ Added context timeout for database operations (5s)
- ✅ Fixed pagination to exclude current user from counts
- ✅ Improved error messages for better DX
- ✅ Added comprehensive validation messages

### Future Enhancements
- [ ] Unit tests with mocks
- [ ] Integration tests
- [ ] Docker containerization
- [ ] CI/CD pipeline
- [ ] API versioning strategy
- [ ] Caching layer (Redis)
- [ ] Background job processing
- [ ] File upload support
- [ ] Email notifications

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

MIT

---

**Built with ❤️ using Go and Clean Architecture principles**
