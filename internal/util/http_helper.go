package util

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// Helper functions for standardizing responses in the new modules

// NewOKResponse creates a standard success response with data
func NewOKResponse(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"data":    data,
	}
}

// NewResponse creates a custom response
func NewResponse(statusCode int, message string, data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": statusCode >= 200 && statusCode < 300,
		"message": message,
		"data":    data,
	}
}

// GetUserIDFromContext extracts user ID from context
func GetUserIDFromContext(c echo.Context) string {
	// 1. Try to get user_id directly from context (set by middleware)
	if userID, ok := c.Get("user_id").(string); ok && userID != "" {
		return userID
	}

	// 2. Fallback to extracting from JWT token if stored as "user" object
	user := c.Get("user")
	if user == nil {
		return ""
	}

	token, ok := user.(*jwt.Token)
	if !ok {
		return ""
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return ""
	}

	userID, ok := claims["user_id"].(string)
	if !ok {
		return ""
	}

	return userID
}
