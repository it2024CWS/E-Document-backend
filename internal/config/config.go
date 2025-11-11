package config

import (
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Admin    AdminConfig
	Logger   LoggerConfig
	JWT      JWTConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port string
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	MongoURI string
	DBName   string
}

// AdminConfig holds admin user configuration for seeding
type AdminConfig struct {
	Username string
	Email    string
	Password string
}

// LoggerConfig holds logger configuration
type LoggerConfig struct {
	Level  string
	Pretty bool
}

// JWTConfig holds JWT configuration
type JWTConfig struct {
	AccessTokenSecret  string
	RefreshTokenSecret string
	AccessTokenExpiry  int64 // in seconds
	RefreshTokenExpiry int64 // in seconds
}

// Load loads configuration from .env file and environment variables
func Load() *Config {
	// Load .env file (silently ignore if not found)
	_ = godotenv.Load()

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
		},
		Database: DatabaseConfig{
			MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
			DBName:   getEnv("DB_NAME", "e_document_db"),
		},
		Admin: AdminConfig{
			Username: getEnv("ADMIN_USERNAME", "admin"),
			Email:    getEnv("ADMIN_EMAIL", "admin@example.com"),
			Password: getEnv("ADMIN_PASSWORD", "password"),
		},
		Logger: LoggerConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Pretty: getEnv("LOG_PRETTY", "true") == "true",
		},
		JWT: JWTConfig{
			AccessTokenSecret:  getEnv("JWT_ACCESS_SECRET", "your-access-secret-key"),
			RefreshTokenSecret: getEnv("JWT_REFRESH_SECRET", "your-refresh-secret-key"),
			AccessTokenExpiry:  getEnvAsInt64("JWT_ACCESS_EXPIRY", 3600),    // 1 hour
			RefreshTokenExpiry: getEnvAsInt64("JWT_REFRESH_EXPIRY", 604800), // 7 days
		},
	}
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// getEnvAsInt64 gets an environment variable as int64 or returns a default value
func getEnvAsInt64(key string, defaultValue int64) int64 {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	if intValue, err := strconv.ParseInt(value, 10, 64); err == nil {
		return intValue
	}
	return defaultValue
}
