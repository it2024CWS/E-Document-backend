package role

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for role operations
type Handler struct {
	service Service
}

// NewHandler creates a new role handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers role routes
func (h *Handler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	roles := e.Group("/v1/roles", authMiddleware)
	roles.GET("", h.GetAllRoles)
	roles.GET("/:id", h.GetRoleByID)
	roles.POST("", h.CreateRole)
	roles.PUT("/:id", h.UpdateRole)
	roles.DELETE("/:id", h.DeleteRole)
}

// GetAllRoles godoc
//
//	@Summary		Get all user roles
//	@Description	Get list of all user roles
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	util.Response{data=[]domain.RoleResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		500	{object}	util.Response
//	@Router			/v1/roles [get]
func (h *Handler) GetAllRoles(c echo.Context) error {
	roles, err := h.service.GetAllRoles(c.Request().Context())
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Roles retrieved successfully", roles)
}

// GetRoleByID godoc
//
//	@Summary		Get role by ID
//	@Description	Get detailed information of a specific role
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Role ID"
//	@Success		200	{object}	util.Response{data=domain.RoleResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/roles/{id} [get]
func (h *Handler) GetRoleByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	role, err := h.service.GetRoleByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Role retrieved successfully", role)
}

// CreateRole godoc
//
//	@Summary		Create a new role
//	@Description	Create a new user role
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.CreateRoleRequest	true	"Role data"
//	@Success		201		{object}	util.Response{data=domain.RoleResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Router			/v1/roles [post]
func (h *Handler) CreateRole(c echo.Context) error {
	var req domain.CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
	}

	role, err := h.service.CreateRole(c.Request().Context(), req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Role created successfully", role, 201)
}

// UpdateRole godoc
//
//	@Summary		Update role
//	@Description	Update role information
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"Role ID"
//	@Param			request	body		domain.UpdateRoleRequest	true	"Role data"
//	@Success		200		{object}	util.Response{data=domain.RoleResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/roles/{id} [put]
func (h *Handler) UpdateRole(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	var req domain.UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
	}

	role, err := h.service.UpdateRole(c.Request().Context(), id, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Role updated successfully", role)
}

// DeleteRole godoc
//
//	@Summary		Delete role
//	@Description	Delete a role
//	@Tags			Roles
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Role ID"
//	@Success		200	{object}	util.Response
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/roles/{id} [delete]
func (h *Handler) DeleteRole(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	if err := h.service.DeleteRole(c.Request().Context(), id); err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Role deleted successfully", nil)
}
