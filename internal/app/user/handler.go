package user

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"strconv"

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
		return util.HandleError(c, util.ErrorResponse("Invalid request body", util.INVALID_INPUT, 400, err.Error()))
	}

	// Validate request
	if req.Username == "" || req.Email == "" || req.Password == "" {
		return util.HandleError(c, util.ErrorResponse("Validation failed", util.MISSING_REQUIRED_FIELD, 400, "Username, email, and password are required"))
	}

	user, err := h.service.CreateUser(c.Request().Context(), req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.CreatedResponse(c, "User created successfully", user)
}

// GetAllUsers handles GET /users
func (h *Handler) GetAllUsers(c echo.Context) error {
	// Get pagination params from query
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")

	// Default values
	pageNum := 1
	limitNum := 10

	// Parse page
	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageNum = p
		}
	}

	// Parse limit
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			limitNum = l
		}
	}

	users, total, err := h.service.GetAllUsers(c.Request().Context(), pageNum, limitNum)
	if err != nil {
		return util.HandleError(c, err)
	}

	// Calculate pagination info
	totalPages := (total + limitNum - 1) / limitNum
	pagination := util.PaginationInfo{
		CurrentPage:  pageNum,
		TotalPages:   totalPages,
		TotalItems:   total,
		ItemsPerPage: limitNum,
	}

	return util.OKResponseWithPagination(c, "Users retrieved successfully", users, pagination)
}

// GetUserByID handles GET /users/:id
func (h *Handler) GetUserByID(c echo.Context) error {
	id := c.Param("id")

	user, err := h.service.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "User retrieved successfully", user)
}

// UpdateUser handles PUT /users/:id
func (h *Handler) UpdateUser(c echo.Context) error {
	id := c.Param("id")

	var req domain.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid request body", util.INVALID_INPUT, 400, err.Error()))
	}

	user, err := h.service.UpdateUser(c.Request().Context(), id, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "User updated successfully", user)
}

// DeleteUser handles DELETE /users/:id
func (h *Handler) DeleteUser(c echo.Context) error {
	id := c.Param("id")

	if err := h.service.DeleteUser(c.Request().Context(), id); err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "User deleted successfully", nil)
}
