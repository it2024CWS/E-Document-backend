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
func (h *Handler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	users := e.Group("/v1/users", authMiddleware)
	users.POST("", h.CreateUser)
	users.GET("", h.GetAllUsers)
	users.GET("/:id", h.GetUserByID)
	users.PUT("/:id", h.UpdateUser)
	users.DELETE("/:id", h.DeleteUser)
}

// CreateUser godoc
//
//	@Summary		Create a new user
//	@Description	Create a new user account
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			body	body		domain.CreateUserRequest	true	"User information"
//	@Success		201		{object}	util.Response{data=domain.UserResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Router			/v1/users [post]
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

	return util.OKResponse(c, "User created successfully", user, 201)
}

// GetAllUsers godoc
//
//	@Summary		Get all users
//	@Description	Get list of all users with pagination and search
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int		false	"Page number"	default(1)
//	@Param			limit	query		int		false	"Items per page"	default(10)
//	@Param			search	query		string	false	"Search by username or email"
//	@Success		200		{object}	util.Response{data=util.PaginatedData}
//	@Failure		401		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/users [get]
func (h *Handler) GetAllUsers(c echo.Context) error {
	// Get pagination params from query
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	search := c.QueryParam("search")

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

	users, total, err := h.service.GetAllUsers(c.Request().Context(), pageNum, limitNum, search)
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

// GetUserByID godoc
//
//	@Summary		Get user by ID
//	@Description	Get detailed information of a specific user
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	util.Response{data=domain.UserResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/users/{id} [get]
func (h *Handler) GetUserByID(c echo.Context) error {
	id := c.Param("id")

	user, err := h.service.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "User retrieved successfully", user)
}

// UpdateUser godoc
//
//	@Summary		Update user
//	@Description	Update user information
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"User ID"
//	@Param			body	body		domain.UpdateUserRequest	true	"Updated user information"
//	@Success		200		{object}	util.Response{data=domain.UserResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/users/{id} [put]
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

// DeleteUser godoc
//
//	@Summary		Delete user
//	@Description	Delete a user account
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	util.Response
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/users/{id} [delete]
func (h *Handler) DeleteUser(c echo.Context) error {
	id := c.Param("id")

	if err := h.service.DeleteUser(c.Request().Context(), id); err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "User deleted successfully", nil)
}
