# E-Document Backend Setup Guide

## Quick Start

### 1. Clone and Install Dependencies

```bash
# Install Go dependencies
go mod download

# Install Air for hot reload (optional)
go install github.com/cosmtrek/air@latest
```

### 2. Setup Environment

Copy `.env.example` to `.env` and configure:

```bash
cp .env.example .env
```

Update the following values in `.env`:

```env
# Server
PORT=5000

# MongoDB (use your MongoDB Atlas URI or local MongoDB)
MONGO_URI=mongodb://localhost:27017
DB_NAME=e_document_db

# JWT Secrets (generate strong secrets for production)
JWT_ACCESS_SECRET=your-super-secret-access-key
JWT_REFRESH_SECRET=your-super-secret-refresh-key
JWT_ACCESS_EXPIRY=3600
JWT_REFRESH_EXPIRY=604800

# MinIO (for file storage)
MINIO_ENDPOINT=localhost:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=edocument-files
MINIO_USE_SSL=false
MINIO_PUBLIC_URL=http://localhost:9000

# Admin User (for seeding)
ADMIN_NAME=admin
ADMIN_EMAIL=admin@edocument.com
ADMIN_PASSWORD=password

# Logger
LOG_LEVEL=info
LOG_PRETTY=true
```

### 3. Start Services with Docker Compose

```bash
# Start MongoDB and MinIO
docker-compose up -d

# Check if services are running
docker ps
```

You should see:
- `edocument-mongodb` on port 27017
- `edocument-minio` on ports 9000 (API) and 9001 (Console)

### 4. Run Database Migrations (if any)

```bash
go run cmd/migrate/main.go
```

### 5. Seed Initial Data

```bash
go run cmd/seed/main.go
```

This creates an admin user with credentials from `.env`.

### 6. Start the API Server

**Option A: Using Go run**
```bash
go run cmd/api/main.go
```

**Option B: Using Air (with hot reload)**
```bash
air
```

**Option C: Build and run**
```bash
go build -o api.exe cmd/api/main.go
./api.exe
```

### 7. Verify Installation

Open your browser and check:

- **API Health**: http://localhost:5000/api/health
- **Swagger Docs**: http://localhost:5000/swagger/index.html
- **MinIO Console**: http://localhost:9001

## Available Services

| Service | URL | Credentials |
|---------|-----|-------------|
| API Server | http://localhost:5000 | - |
| Swagger UI | http://localhost:5000/swagger/index.html | - |
| MinIO Console | http://localhost:9001 | admin:minioadmin |
| MongoDB | mongodb://localhost:27017 | - |

## API Endpoints

### Authentication

- `POST /api/v1/auth/register` - Register new user
- `POST /api/v1/auth/login` - Login (web)
- `POST /api/v1/auth/mobile/login` - Login (mobile)
- `POST /api/v1/auth/refresh` - Refresh access token
- `POST /api/v1/auth/logout` - Logout
- `GET /api/v1/auth/me` - Get current user

### Users

- `POST /api/v1/users` - Create user (authenticated)
- `GET /api/v1/users` - Get all users with pagination (authenticated)
- `GET /api/v1/users/:id` - Get user by ID (authenticated)
- `PUT /api/v1/users/:id` - Update user (authenticated)
- `DELETE /api/v1/users/:id` - Delete user (authenticated)

### Profile Picture (NEW)

- `POST /api/v1/users/:id/profile-picture` - Upload profile picture
- `DELETE /api/v1/users/:id/profile-picture` - Delete profile picture

## Testing the APIs

### 1. Register/Login

```bash
# Register
curl -X POST http://localhost:5000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123",
    "first_name": "Test",
    "last_name": "User",
    "phone": "+66812345678",
    "role": "Employee"
  }'

# Login (Web - returns cookie)
curl -X POST http://localhost:5000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "usernameOrEmail": "test@example.com",
    "password": "password123"
  }' \
  -c cookies.txt

# Login (Mobile - returns tokens)
curl -X POST http://localhost:5000/api/v1/auth/mobile/login \
  -H "Content-Type: application/json" \
  -d '{
    "usernameOrEmail": "test@example.com",
    "password": "password123"
  }'
```

### 2. Upload Profile Picture

```bash
# Get your user ID from login response or /auth/me endpoint
USER_ID="your-user-id-here"
TOKEN="your-access-token"

# Upload profile picture
curl -X POST http://localhost:5000/api/v1/users/${USER_ID}/profile-picture \
  -H "Authorization: Bearer ${TOKEN}" \
  -F "file=@/path/to/your/image.jpg"
```

### 3. Access Profile Picture

After upload, you'll get a URL like:
```
http://localhost:9000/edocument-files/profiles/1733136000_profile.jpg
```

You can access this directly in your browser or mobile app.

## Project Structure

```
E-Document-backend/
├── cmd/
│   ├── api/          # Main application entry point
│   ├── migrate/      # Database migrations
│   └── seed/         # Database seeding
├── internal/
│   ├── app/          # Application modules
│   │   ├── auth/     # Authentication module
│   │   └── user/     # User management module
│   ├── config/       # Configuration
│   ├── domain/       # Domain models
│   ├── logger/       # Logging utilities
│   ├── middleware/   # HTTP middlewares
│   ├── pkg/
│   │   └── storage/  # MinIO storage client
│   ├── platform/
│   │   └── mongodb/  # MongoDB client
│   └── util/         # Utility functions
├── docs/             # Documentation and Swagger
├── docker-compose.yml
├── .env
├── .env.example
└── go.mod
```

## Development

### Hot Reload with Air

```bash
# Install Air
go install github.com/cosmtrek/air@latest

# Run with hot reload
air
```

Air will watch for file changes and automatically rebuild and restart the server.

### Generate Swagger Docs

```bash
# Install swag
go install github.com/swaggo/swag/cmd/swag@latest

# Generate docs
swag init -g cmd/api/main.go -o docs

# View docs at http://localhost:5000/swagger/index.html
```

## Production Deployment

### Environment Variables

Update these for production:

1. **JWT Secrets**: Generate strong random secrets
2. **MinIO Credentials**: Use strong passwords
3. **MongoDB URI**: Use production database
4. **SSL/TLS**: Enable HTTPS
5. **CORS**: Update allowed origins

### Docker Production Build

```dockerfile
FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o api cmd/api/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/api .
COPY --from=builder /app/.env .
CMD ["./api"]
```

### Security Checklist

- [ ] Use strong JWT secrets
- [ ] Enable HTTPS/TLS
- [ ] Update MinIO credentials
- [ ] Configure proper CORS origins
- [ ] Use environment variables for secrets
- [ ] Enable rate limiting
- [ ] Implement proper logging
- [ ] Set up monitoring
- [ ] Regular backups for MongoDB and MinIO
- [ ] Use reverse proxy (Nginx/Caddy)

## Troubleshooting

### Port Already in Use

```bash
# Find process using port 5000
lsof -i :5000

# Kill the process
kill -9 <PID>
```

### MongoDB Connection Failed

1. Check if MongoDB is running: `docker ps`
2. Check MongoDB logs: `docker logs edocument-mongodb`
3. Verify MONGO_URI in `.env`

### MinIO Connection Failed

1. Check if MinIO is running: `docker ps`
2. Check MinIO logs: `docker logs edocument-minio`
3. Verify MinIO endpoint in `.env`
4. Access MinIO Console: http://localhost:9001

### Build Errors

```bash
# Clean Go cache
go clean -cache -modcache

# Re-download dependencies
go mod download
go mod tidy

# Rebuild
go build cmd/api/main.go
```

## Additional Documentation

- [MinIO Setup Guide](docs/minio-setup.md)
- [Mobile Integration Guide](docs/mobile-integration.md)
- [Testing Examples](docs/testing-examples.md)

## Support

For issues or questions, please open an issue on GitHub or contact the development team.
