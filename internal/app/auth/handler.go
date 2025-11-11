package auth

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handler defines the auth HTTP handler
type Handler struct {
	service Service
}

// NewHandler creates a new auth handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Login handles user login
// @Summary User login
// @Description Authenticate user with username/email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body domain.LoginRequest true "Login credentials"
// @Success 200 {object} domain.AuthResponse
// @Failure 400 {object} util.ErrorDetail
// @Failure 401 {object} util.ErrorDetail
// @Router /auth/login [post]
func (h *Handler) Login(c echo.Context) error {
	var req domain.LoginRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.ErrorResponse(
			"Invalid request body",
			util.VALIDATION_ERROR,
			400,
			err.Error(),
		))
	}

	// Validate required fields
	if req.UsernameOrEmail == "" || req.Password == "" {
		return util.HandleError(c, util.ErrorResponse(
			"Username/email and password are required",
			util.MISSING_REQUIRED_FIELD,
			400,
			"both usernameOrEmail and password fields are required",
		))
	}

	result, err := h.service.Login(c.Request().Context(), req)
	if err != nil {
		return util.HandleError(c, err)
	}

	// Set cookies
	h.setCookies(c, result.AccessToken, result.RefreshToken)

	return c.JSON(http.StatusOK, result)
}

// RefreshToken handles token refresh
// @Summary Refresh access token
// @Description Get new access token using refresh token
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} domain.AuthResponse
// @Failure 400 {object} util.ErrorDetail
// @Failure 401 {object} util.ErrorDetail
// @Router /auth/refresh [post]
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
		return util.HandleError(c, util.ErrorResponse(
			"Refresh token is required",
			util.MISSING_REQUIRED_FIELD,
			400,
			"refresh token must be provided in cookie or request body",
		))
	}

	result, err := h.service.RefreshToken(c.Request().Context(), refreshToken)
	if err != nil {
		return util.HandleError(c, err)
	}

	// Set new cookies
	h.setCookies(c, result.AccessToken, result.RefreshToken)

	return c.JSON(http.StatusOK, result)
}

// GetProfile handles getting user profile
// @Summary Get user profile
// @Description Get current user profile information
// @Tags auth
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} domain.UserResponse
// @Failure 401 {object} util.ErrorDetail
// @Failure 404 {object} util.ErrorDetail
// @Router /auth/profile [get]
func (h *Handler) GetProfile(c echo.Context) error {
	// Get user ID from context (set by auth middleware)
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return util.HandleError(c, util.ErrorResponse(
			"Unauthorized",
			util.UNAUTHORIZED,
			401,
			"user not authenticated",
		))
	}

	result, err := h.service.GetProfile(c.Request().Context(), userID)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, result)
}

// Logout handles user logout
// @Summary User logout
// @Description Clear authentication cookies
// @Tags auth
// @Accept json
// @Produce json
// @Success 200 {object} map[string]string
// @Router /auth/logout [post]
func (h *Handler) Logout(c echo.Context) error {
	// Clear cookies
	h.clearCookies(c)

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
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
