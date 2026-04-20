package main

import (
	"context"
	"e-document-backend/internal/app/auth"
	"e-document-backend/internal/app/department"
	"e-document-backend/internal/app/doctype"
	"e-document-backend/internal/app/document"
	"e-document-backend/internal/app/file"
	"e-document-backend/internal/app/folder"
	"e-document-backend/internal/app/incomingdoc"
	"e-document-backend/internal/app/outgoingdoc"
	"e-document-backend/internal/app/role"
	"e-document-backend/internal/app/sector"
	"e-document-backend/internal/app/upload"
	"e-document-backend/internal/app/user"
	"e-document-backend/internal/config"
	"e-document-backend/internal/logger"
	customMiddleware "e-document-backend/internal/middleware"
	"e-document-backend/internal/pkg/seed"
	"e-document-backend/internal/pkg/storage"
	"e-document-backend/internal/platform/postgres"
	"net/http"
	"os"
	"os/signal"
	"time"

	_ "e-document-backend/docs" // Import generated docs

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

//	@title			E-Document API
//	@version		1.0
//	@description	API for E-Document Management System
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.email	support@edocument.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:5001
//	@BasePath	/api

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Type "Bearer" followed by a space and JWT token.

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger.Init(logger.Config{
		Level:      logger.LogLevel(cfg.Logger.Level),
		Pretty:     cfg.Logger.Pretty,
		TimeFormat: time.RFC3339,
	})

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:5173", "http://localhost:5001"}, // ⚠️ ต้องระบุ origin ชัดเจน ไม่ใช่ "*"
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodPatch, http.MethodOptions, http.MethodHead},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"ngrok-skip-browser-warning",
			// TUS protocol headers
			"Upload-Offset",
			"Upload-Length",
			"Upload-Metadata",
			"Upload-Defer-Length",
			"Upload-Concat",
			"Tus-Resumable",
			"Tus-Version",
			"Tus-Max-Size",
			"Tus-Extension",
		},
		AllowCredentials: true,
		ExposeHeaders: []string{
			"Set-Cookie",
			// TUS protocol headers
			"Upload-Offset",
			"Upload-Length",
			"Location",
			"Tus-Resumable",
			"Tus-Version",
			"Tus-Max-Size",
			"Tus-Extension",
		},
	}))

	// Request ID middleware (adds unique ID to each request)
	e.Use(customMiddleware.RequestIDMiddleware())

	// Logger middleware (logs all requests and responses)
	if cfg.Logger.Level == "debug" {
		// Detailed logging with request/response body for development
		e.Use(customMiddleware.DetailedLoggerMiddleware())
	} else {
		// Standard logging for production
		e.Use(customMiddleware.LoggerMiddleware())
	}

	// Rate limiting middleware
	e.Use(customMiddleware.RateLimitMiddleware(customMiddleware.RateLimitConfig{
		RequestsPerSecond: 20,
		BurstSize:         50,
	}))

	// Connect to PostgreSQL
	ctx := context.Background()
	println("Connecting to PostgreSQL...", cfg.Database.PostgresDSN)
	pgClient, err := postgres.NewClient(ctx, cfg.Database.PostgresDSN)
	if err != nil {
		logger.FatalWithErr("Failed to connect to PostgreSQL", err)
	}
	defer pgClient.Close()

	// Initialize MinIO client for file storage
	minioConfig := storage.LoadConfigFromEnv()
	minioClient, err := storage.NewMinIOClient(minioConfig)
	if err != nil {
		logger.FatalWithErr("Failed to initialize MinIO client", err)
	}
	logger.Info("MinIO client initialized successfully")

	// Initialize user module (Handler-Service-Repository)
	userRepo := user.NewPostgresRepository(pgClient.Pool)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService, minioClient)

	// Initialize file module (Service-Handler) for generating presigned URLs for files in MinIO
	fileService := file.NewService(minioClient)
	fileHandler := file.NewHandler(fileService)

	// Seed admin user if it doesn't exist
	if err := seed.SeedAdmin(ctx, userRepo, pgClient.Pool, cfg); err != nil {
		logger.Warnf("Failed to seed admin user: %v", err)
	}

	// Initialize auth module (Handler-Service)
	authService := auth.NewService(userRepo, cfg)
	authHandler := auth.NewHandler(authService)

	// Initialize role module
	roleRepo := role.NewPostgresRepository(pgClient.Pool)
	roleService := role.NewService(roleRepo)
	roleHandler := role.NewHandler(roleService)

	// Initialize department module
	departmentRepo := department.NewPostgresRepository(pgClient.Pool)
	departmentService := department.NewService(departmentRepo)
	departmentHandler := department.NewHandler(departmentService)

	// Initialize sector module
	sectorRepo := sector.NewPostgresRepository(pgClient.Pool)
	sectorService := sector.NewService(sectorRepo)
	sectorHandler := sector.NewHandler(sectorService)

	// Initialize doctype module
	doctypeRepo := doctype.NewPostgresRepository(pgClient.Pool)
	doctypeService := doctype.NewService(doctypeRepo)
	doctypeHandler := doctype.NewHandler(doctypeService)

	// Initialize incomingdoc module
	incomingdocRepo := incomingdoc.NewPostgresRepository(pgClient.Pool)
	incomingdocService := incomingdoc.NewService(incomingdocRepo)
	incomingdocHandler := incomingdoc.NewHandler(incomingdocService)

	// Initialize outgoingdoc module
	outgoingdocRepo := outgoingdoc.NewPostgresRepository(pgClient.Pool)
	outgoingdocService := outgoingdoc.NewService(outgoingdocRepo)
	outgoingdocHandler := outgoingdoc.NewHandler(outgoingdocService)

	// Initialize document module
	documentRepo := document.NewPostgresRepository(pgClient.Pool)
	documentService := document.NewService(documentRepo, incomingdocRepo, outgoingdocRepo)
	documentHandler := document.NewHandler(documentService)

	// Initialize folder module
	folderRepo := folder.NewPostgresRepository(pgClient.Pool)
	folderService := folder.NewService(folderRepo)
	folderHandler := folder.NewHandler(folderService)

	// Initialize upload module
	uploadRepo := upload.NewPostgresRepository(pgClient.Pool)
	uploadService := upload.NewService(uploadRepo)
	tusConfig := upload.LoadTusConfigFromEnv()
	uploadHandler, err := upload.NewHandler(uploadService, tusConfig)
	if err != nil {
		logger.FatalWithErr("Failed to initialize upload handler", err)
	}

	// API routes
	api := e.Group("/api")

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Health check endpoint
	api.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Server is running",
			"data": map[string]string{
				"status": "healthy",
				"time":   time.Now().Format(time.RFC3339),
			},
		})
	})

	// Register user routes
	userHandler.RegisterRoutes(api, customMiddleware.AuthMiddleware(authService))
	// Register file routes
	fileHandler.RegisterRoutes(api, customMiddleware.AuthMiddleware(authService))
	// Register auth routes (with middleware for protected routes)
	authHandler.RegisterRoutes(api, customMiddleware.AuthMiddleware(authService))
	// Register role routes
	roleHandler.RegisterRoutes(api, customMiddleware.AuthMiddleware(authService))
	// Register department routes
	departmentHandler.RegisterRoutes(api, customMiddleware.AuthMiddleware(authService))
	// Register sector routes
	sectorHandler.RegisterRoutes(api, customMiddleware.AuthMiddleware(authService))
	// Register doctype routes
	doctypeHandler.RegisterRoutes(api, customMiddleware.AuthMiddleware(authService))
	// Register document routes
	documentHandler.RegisterRoutes(api, customMiddleware.AuthMiddleware(authService))
	// Register folder routes
	folderHandler.RegisterRoutes(api, customMiddleware.AuthMiddleware(authService))
	// Register incomingdoc routes
	incomingdocHandler.RegisterRoutes(api, customMiddleware.AuthMiddleware(authService))
	// Register outgoingdoc routes
	outgoingdocHandler.RegisterRoutes(api, customMiddleware.AuthMiddleware(authService))
	// Register upload routes
	uploadHandler.RegisterRoutes(api, customMiddleware.AuthMiddleware(authService))

	// Start server
	go func() {
		if err := e.Start(":" + cfg.Server.Port); err != nil && err != http.ErrServerClosed {
			logger.FatalWithErr("Failed to start server", err)
		}
	}()

	logger.Infof("Server started on port %s \n swagger available at http://localhost:%s/swagger/index.html", cfg.Server.Port, cfg.Server.Port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	logger.Info("Shutting down server...")

	// Gracefully shutdown the server with a timeout of 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		logger.FatalWithErr("Server forced to shutdown", err)
	}

	logger.Info("Server exited")
}
