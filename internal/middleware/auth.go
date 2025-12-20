package middleware

import (
	"e-document-backend/internal/app/auth"
	"e-document-backend/internal/util"
	"strings"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware validates authentication tokens (supports both Bearer and Cookie)
func AuthMiddleware(authService auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Skip authentication for OPTIONS requests (CORS preflight)
			if c.Request().Method == "OPTIONS" {
				return next(c)
			}

			var token string

			// Try to get token from Authorization header first
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				// Try to get token from cookie
				if cookie, err := c.Cookie("accessToken"); err == nil {
					token = cookie.Value
				}
			}

			if token == "" {
				return util.HandleError(c, util.ErrorResponse(
					"Unauthorized",
					util.UNAUTHORIZED,
					401,
					"Missing authentication token",
				))
			}

			// Validate token using auth service
			claims, err := authService.ValidateAccessToken(token)
			if err != nil {
				return util.HandleError(c, util.ErrorResponse(
					"Unauthorized",
					util.INVALID_TOKEN,
					401,
					"Invalid or expired token",
				))
			}

			// Store user information in context
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
			c.Set("token", token)

			return next(c)
		}
	}
}

// OptionalAuthMiddleware is similar to AuthMiddleware but doesn't fail if no auth is provided
func OptionalAuthMiddleware(authService auth.Service) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			var token string

			// Try to get token from Authorization header first
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
				token = strings.TrimPrefix(authHeader, "Bearer ")
			} else {
				// Try to get token from cookie
				if cookie, err := c.Cookie("accessToken"); err == nil {
					token = cookie.Value
				}
			}

			if token != "" {
				// Validate token if present
				if claims, err := authService.ValidateAccessToken(token); err == nil {
					// Store user information in context
					c.Set("user_id", claims.UserID)
					c.Set("username", claims.Username)
					c.Set("email", claims.Email)
					c.Set("token", token)
				}
			}

			return next(c)
		}
	}
}
