package main

import (
	"context"
	"e-document-backend/internal/app/auth"
	"e-document-backend/internal/app/user"
	"e-document-backend/internal/config"
	"e-document-backend/internal/logger"
	customMiddleware "e-document-backend/internal/middleware"
	"e-document-backend/internal/platform/mongodb"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
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

	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

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

	// Connect to MongoDB
	mongoClient, err := mongodb.NewClient(cfg.Database.MongoURI, cfg.Database.DBName)
	if err != nil {
		logger.FatalWithErr("Failed to connect to MongoDB", err)
	}
	defer mongoClient.Disconnect()

	// Initialize user module (Handler-Service-Repository)
	userRepo := user.NewRepository(mongoClient.Database)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

	// Initialize auth module (Handler-Service)
	authService := auth.NewService(userRepo, cfg)
	authHandler := auth.NewHandler(authService)

	// API routes
	api := e.Group("/api/v1")

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
	userHandler.RegisterRoutes(api)

	// Register auth routes
	authGroup := api.Group("/auth")
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/refresh", authHandler.RefreshToken)
	authGroup.POST("/logout", authHandler.Logout)

	// Protected routes (requires authentication)
	protected := api.Group("")
	protected.Use(customMiddleware.AuthMiddleware(authService))
	protected.GET("/auth/profile", authHandler.GetProfile)

	// You can add more protected routes here with auth middleware
	// Example:
	// protected.GET("/protected", someHandler)

	// Start server
	go func() {
		if err := e.Start(":" + cfg.Server.Port); err != nil && err != http.ErrServerClosed {
			logger.FatalWithErr("Failed to start server", err)
		}
	}()

	logger.Infof("Server started on port %s", cfg.Server.Port)

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
