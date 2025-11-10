package middleware

import (
	"bytes"
	"e-document-backend/internal/logger"
	"encoding/json"
	"io"
	"time"

	"github.com/labstack/echo/v4"
)

// maskSensitiveFields masks sensitive fields in request body
func maskSensitiveFields(body []byte) string {
	if len(body) == 0 {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal(body, &data); err != nil {
		// If not JSON, return as is (but truncate if too long)
		bodyStr := string(body)
		if len(bodyStr) > 500 {
			return bodyStr[:500] + "... (truncated)"
		}
		return bodyStr
	}

	// Sensitive field names to mask
	sensitiveFields := []string{
		"password", "Password", "PASSWORD",
		"passwd", "Passwd", "PASSWD",
		"secret", "Secret", "SECRET",
		"token", "Token", "TOKEN",
		"api_key", "apiKey", "API_KEY",
		"access_token", "accessToken", "ACCESS_TOKEN",
		"refresh_token", "refreshToken", "REFRESH_TOKEN",
	}

	// Mask sensitive fields
	for _, field := range sensitiveFields {
		if _, exists := data[field]; exists {
			data[field] = "***MASKED***"
		}
	}

	// Convert back to JSON string
	maskedBody, err := json.Marshal(data)
	if err != nil {
		return string(body)
	}

	return string(maskedBody)
}

// LoggerMiddleware logs HTTP requests and responses
func LoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// Start timer
			start := time.Now()

			// Read request body
			var requestBody []byte
			var maskedBody string
			if req.Body != nil && req.Method != "GET" && req.Method != "DELETE" {
				requestBody, _ = io.ReadAll(req.Body)
				// Restore request body for next middleware
				req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
				// Mask sensitive fields
				maskedBody = maskSensitiveFields(requestBody)
			}

			// Log request
			logEvent := logger.Logger.Info().
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Str("query", req.URL.RawQuery).
				Str("ip", c.RealIP()).
				Str("user_agent", req.UserAgent())

			if maskedBody != "" {
				logEvent.Str("body", maskedBody)
			}

			logEvent.Msg("Incoming request")

			// Process request
			err := next(c)

			// Calculate duration
			duration := time.Since(start)

			// Log response
			if err != nil {
				// Log error response
				logger.Logger.Error().
					Err(err).
					Str("method", req.Method).
					Str("path", req.URL.Path).
					Int("status", res.Status).
					Dur("duration", duration).
					Str("duration_human", duration.String()).
					Msg("Request failed")
			} else {
				// Log success response
				logEvent := logger.Logger.Info().
					Str("method", req.Method).
					Str("path", req.URL.Path).
					Int("status", res.Status).
					Dur("duration", duration).
					Str("duration_human", duration.String())

				// Color code based on status
				if res.Status >= 500 {
					logEvent = logger.Logger.Error().
						Str("method", req.Method).
						Str("path", req.URL.Path).
						Int("status", res.Status).
						Dur("duration", duration).
						Str("duration_human", duration.String())
				} else if res.Status >= 400 {
					logEvent = logger.Logger.Warn().
						Str("method", req.Method).
						Str("path", req.URL.Path).
						Int("status", res.Status).
						Dur("duration", duration).
						Str("duration_human", duration.String())
				}

				logEvent.Msg("Request completed")
			}

			return err
		}
	}
}

// DetailedLoggerMiddleware logs HTTP requests and responses with body (for debugging)
func DetailedLoggerMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()
			res := c.Response()

			// Start timer
			start := time.Now()

			// Read request body
			var requestBody []byte
			var maskedBody string
			if req.Body != nil {
				requestBody, _ = io.ReadAll(req.Body)
				// Restore request body for next middleware
				req.Body = io.NopCloser(bytes.NewBuffer(requestBody))
				// Mask sensitive fields
				maskedBody = maskSensitiveFields(requestBody)
			}

			// Log request with body
			logger.Logger.Debug().
				Str("method", req.Method).
				Str("path", req.URL.Path).
				Str("query", req.URL.RawQuery).
				Str("ip", c.RealIP()).
				Str("user_agent", req.UserAgent()).
				Str("body", maskedBody).
				Msg("Incoming request (detailed)")

			// Process request
			err := next(c)

			// Calculate duration
			duration := time.Since(start)

			// Log response with details
			if err != nil {
				logger.Logger.Error().
					Err(err).
					Str("method", req.Method).
					Str("path", req.URL.Path).
					Int("status", res.Status).
					Dur("duration", duration).
					Str("duration_human", duration.String()).
					Msg("Request failed (detailed)")
			} else {
				logger.Logger.Debug().
					Str("method", req.Method).
					Str("path", req.URL.Path).
					Int("status", res.Status).
					Dur("duration", duration).
					Str("duration_human", duration.String()).
					Int64("response_size", res.Size).
					Msg("Request completed (detailed)")
			}

			return err
		}
	}
}

// RequestIDMiddleware adds a unique request ID to each request
func RequestIDMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = generateRequestID()
			}

			// Set request ID in response header
			c.Response().Header().Set("X-Request-ID", requestID)

			// Store request ID in context for later use
			c.Set("request_id", requestID)

			return next(c)
		}
	}
}

// generateRequestID generates a simple unique request ID
func generateRequestID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString generates a random string of specified length
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}
