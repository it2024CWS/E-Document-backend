package main

import (
	"context"
	"e-document-backend/internal/app/user"
	customMiddleware "e-document-backend/internal/middleware"
	"e-document-backend/internal/platform/mongodb"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Create Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Rate limiting middleware
	e.Use(customMiddleware.RateLimitMiddleware(customMiddleware.RateLimitConfig{
		RequestsPerSecond: 20,
		BurstSize:         50,
	}))

	// MongoDB configuration
	mongoURI := os.Getenv("MONGO_URI")
	if mongoURI == "" {
		mongoURI = "mongodb+srv://hackterssv:26F0wLf2WShAOuaU@hcu-year--book.uvdycy1.mongodb.net/?retryWrites=true&w=majority&appName=EDocument"
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		dbName = "e_document_db"
	}

	// Connect to MongoDB
	mongoClient, err := mongodb.NewClient(mongoURI, dbName)
	if err != nil {
		log.Fatalf("Failed to connect to MongoDB: %v", err)
	}
	defer mongoClient.Disconnect()

	// Initialize user module (Handler-Service-Repository)
	userRepo := user.NewRepository(mongoClient.Database)
	userService := user.NewService(userRepo)
	userHandler := user.NewHandler(userService)

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

	// You can add protected routes here with auth middleware
	// Example:
	// protected := api.Group("")
	// protected.Use(customMiddleware.AuthMiddleware())
	// protected.GET("/protected", someHandler)

	// Server configuration
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server
	go func() {
		if err := e.Start(":" + port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started on port %s", port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit

	log.Println("Shutting down server...")

	// Gracefully shutdown the server with a timeout of 10 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
