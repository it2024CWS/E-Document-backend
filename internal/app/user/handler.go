package user

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"

	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for user operations
type Handler struct {
	service Service
}

// NewHandler creates a new user handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers user routes
func (h *Handler) RegisterRoutes(e *echo.Group) {
	users := e.Group("/users")
	users.POST("", h.CreateUser)
	users.GET("", h.GetAllUsers)
	users.GET("/:id", h.GetUserByID)
	users.PUT("/:id", h.UpdateUser)
	users.DELETE("/:id", h.DeleteUser)
}

// CreateUser handles POST /users
func (h *Handler) CreateUser(c echo.Context) error {
	var req domain.CreateUserRequest

	if err := c.Bind(&req); err != nil {
		return util.BadRequestResponse(c, "Invalid request body", err)
	}

	// Validate request
	if req.Name == "" || req.Email == "" || req.Password == "" {
		return util.BadRequestResponse(c, "Name, email, and password are required", nil)
	}

	user, err := h.service.CreateUser(c.Request().Context(), req)
	if err != nil {
		return util.BadRequestResponse(c, "Failed to create user", err)
	}

	return util.CreatedResponse(c, "User created successfully", user)
}

// GetAllUsers handles GET /users
func (h *Handler) GetAllUsers(c echo.Context) error {
	users, err := h.service.GetAllUsers(c.Request().Context())
	if err != nil {
		return util.InternalServerErrorResponse(c, "Failed to fetch users", err)
	}

	return util.OKResponse(c, "Users retrieved successfully", users)
}

// GetUserByID handles GET /users/:id
func (h *Handler) GetUserByID(c echo.Context) error {
	id := c.Param("id")

	user, err := h.service.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return util.NotFoundResponse(c, "User not found", err)
	}

	return util.OKResponse(c, "User retrieved successfully", user)
}

// UpdateUser handles PUT /users/:id
func (h *Handler) UpdateUser(c echo.Context) error {
	id := c.Param("id")

	var req domain.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return util.BadRequestResponse(c, "Invalid request body", err)
	}

	user, err := h.service.UpdateUser(c.Request().Context(), id, req)
	if err != nil {
		return util.BadRequestResponse(c, "Failed to update user", err)
	}

	return util.OKResponse(c, "User updated successfully", user)
}

// DeleteUser handles DELETE /users/:id
func (h *Handler) DeleteUser(c echo.Context) error {
	id := c.Param("id")

	if err := h.service.DeleteUser(c.Request().Context(), id); err != nil {
		return util.NotFoundResponse(c, "Failed to delete user", err)
	}

	return util.OKResponse(c, "User deleted successfully", nil)
}
