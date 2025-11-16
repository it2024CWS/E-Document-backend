package auth

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for auth operations
type Handler struct {
	service Service
}

// NewHandler creates a new auth handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers auth routes
func (h *Handler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	auth := e.Group("/v1/auth")

	// Public routes
	auth.POST("/login", h.Login)
	auth.POST("/refresh", h.RefreshToken)
	auth.POST("/logout", h.Logout)

	// Protected routes (requires authentication)
	auth.GET("/profile", h.GetProfile, authMiddleware)
} // Login handles POST /auth/login

// NOTE - Login
func (h *Handler) Login(c echo.Context) error {
	var req domain.LoginRequest

	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid request body", util.INVALID_INPUT, 400, err.Error()))
	}

	// Validate request
	if req.UsernameOrEmail == "" || req.Password == "" {
		return util.HandleError(c, util.ErrorResponse("Validation failed", util.MISSING_REQUIRED_FIELD, 400, "Username/email and password are required"))
	}

	result, err := h.service.Login(c.Request().Context(), req)
	if err != nil {
		return util.HandleError(c, err)
	}

	// Set cookies
	h.setCookies(c, result.AccessToken, result.RefreshToken)

	return util.OKResponse(c, "Login successful", result.Response.User)
}

// NOTE - Refresh Token
// RefreshToken handles POST /auth/refresh
func (h *Handler) RefreshToken(c echo.Context) error {
	// Try to get refresh token from cookie first, then from body
	refreshToken := h.getRefreshTokenFromCookie(c)

	if refreshToken == "" {
		var req domain.RefreshTokenRequest
		if err := c.Bind(&req); err == nil && req.RefreshToken != "" {
			refreshToken = req.RefreshToken
		}
	}

	if refreshToken == "" {
		return util.HandleError(c, util.ErrorResponse("Validation failed", util.MISSING_REQUIRED_FIELD, 400, "Refresh token is required"))
	}

	result, err := h.service.RefreshToken(c.Request().Context(), refreshToken)
	if err != nil {
		return util.HandleError(c, err)
	}

	// Set new cookies
	h.setCookies(c, result.AccessToken, result.RefreshToken)

	return util.OKResponse(c, "Token refreshed successfully", result.Response.User)
}

// NOTE - Get Profile
// GetProfile handles GET /auth/profile
func (h *Handler) GetProfile(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return util.HandleError(c, util.ErrorResponse("Unauthorized", util.UNAUTHORIZED, 401, "user not authenticated"))
	}

	result, err := h.service.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Profile retrieved successfully", result)
}

// NOTE - Logout
// Logout handles POST /auth/logout
func (h *Handler) Logout(c echo.Context) error {
	// Clear cookies
	h.clearCookies(c)

	return util.OKResponse(c, "Logged out successfully", nil)
}

// setCookies sets access and refresh tokens as HTTP-only cookies
func (h *Handler) setCookies(c echo.Context, accessToken, refreshToken string) {
	// Set access token cookie (1 hour)
	accessCookie := &http.Cookie{
		Name:     "accessToken",
		Value:    accessToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   3600, // 1 hour
	}

	// Set refresh token cookie (7 days)
	refreshCookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    refreshToken,
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production with HTTPS
		SameSite: http.SameSiteLaxMode,
		MaxAge:   604800, // 7 days
	}

	c.SetCookie(accessCookie)
	c.SetCookie(refreshCookie)
}

// clearCookies removes authentication cookies
func (h *Handler) clearCookies(c echo.Context) {
	accessCookie := &http.Cookie{
		Name:     "accessToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}

	refreshCookie := &http.Cookie{
		Name:     "refreshToken",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	}

	c.SetCookie(accessCookie)
	c.SetCookie(refreshCookie)
}

// getRefreshTokenFromCookie extracts refresh token from cookie
func (h *Handler) getRefreshTokenFromCookie(c echo.Context) string {
	cookie, err := c.Cookie("refreshToken")
	if err != nil {
		return ""
	}
	return cookie.Value
}
