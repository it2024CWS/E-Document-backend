package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"golang.org/x/time/rate"
)

// RateLimitConfig holds the configuration for rate limiting
type RateLimitConfig struct {
	RequestsPerSecond int
	BurstSize         int
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(config RateLimitConfig) echo.MiddlewareFunc {
	// Default configuration
	if config.RequestsPerSecond == 0 {
		config.RequestsPerSecond = 20 // 20 requests per second
	}
	if config.BurstSize == 0 {
		config.BurstSize = 50 // burst size of 50
	}

	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{
				Rate:      rate.Limit(config.RequestsPerSecond),
				Burst:     config.BurstSize,
				ExpiresIn: 60, // expire in 60 seconds
			},
		),
		IdentifierExtractor: func(c echo.Context) (string, error) {
			// Use IP address as identifier
			return c.RealIP(), nil
		},
		ErrorHandler: func(c echo.Context, err error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"success": false,
				"message": "Rate limit exceeded",
				"error":   "Too many requests, please try again later",
			})
		},
		DenyHandler: func(c echo.Context, identifier string, err error) error {
			return c.JSON(http.StatusTooManyRequests, map[string]interface{}{
				"success": false,
				"message": "Rate limit exceeded",
				"error":   "Too many requests, please try again later",
			})
		},
	})
}
