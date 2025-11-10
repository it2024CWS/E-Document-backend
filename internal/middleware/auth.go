package middleware

import (
	"e-document-backend/internal/util"
	"strings"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware validates authentication tokens
func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return util.HandleError(c, util.ErrorResponse("unauthorized", util.UNAUTHORIZED, 401, "Missing authorization header"))
			}

			// Check if it starts with "Bearer "
			if !strings.HasPrefix(authHeader, "Bearer ") {
				return util.HandleError(c, util.ErrorResponse("unauthorized", util.INVALID_TOKEN, 401, "Invalid authorization header format"))
			}

			// Extract token
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if token == "" {
				return util.HandleError(c, util.ErrorResponse("unauthorized", util.INVALID_TOKEN, 401, "Missing token"))
			}

			// TODO: Implement actual token validation (JWT, etc.)
			// For now, we'll just check if token is not empty
			// In production, you should validate JWT or other token types

			// Store token in context for later use
			c.Set("token", token)

			return next(c)
		}
	}
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't fail if no auth is provided
func OptionalAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				c.Set("token", token)
			}
			return next(c)
		}
	}
}
