package role

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"strconv"

	"github.com/labstack/echo/v4"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	roles := e.Group("/v1/roles", authMiddleware)
	roles.POST("", h.CreateRole)
	roles.GET("", h.GetAllRoles)
	roles.GET("/:id", h.GetRoleByID)
	roles.PUT("/:id", h.UpdateRole)
	roles.DELETE("/:id", h.DeleteRole)
}

// NOTE - CreateRole
func (h *Handler) CreateRole(c echo.Context) error {
	var req domain.CreateRoleRequest

	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid request body", util.INVALID_INPUT, 400, err.Error()))
	}

	// Validate request
	if req.Name == "" {
		return util.HandleError(c, util.ErrorResponse("Validation failed", util.MISSING_REQUIRED_FIELD, 400, "Role name is required"))
	}

	role, err := h.service.CreateRole(c.Request().Context(), req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Role created successfully", role, 201)
}

// NOTE - GetAllRoles
func (h *Handler) GetAllRoles(c echo.Context) error {
	// Get pagination params from query
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	search := c.QueryParam("search")

	pageNum := 1
	limitNum := 10

	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageNum = p
		}
	}
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			limitNum = l
		}
	}

	roles, total, err := h.service.GetAllRoles(c.Request().Context(), pageNum, limitNum, search)
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

	return util.OKResponseWithPagination(c, "Roles retrieved successfully", roles, pagination)
}

// NOTE - GetRoleByID
func (h *Handler) GetRoleByID(c echo.Context) error {
	id := c.Param("id")

	role, err := h.service.GetRoleByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Role retrieved successfully", role)
}

// NOTE - UpdateRole
func (h *Handler) UpdateRole(c echo.Context) error {
	id := c.Param("id")
	var req domain.UpdateRoleRequest

	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid request body", util.INVALID_INPUT, 400, err.Error()))
	}

	role, err := h.service.UpdateRole(c.Request().Context(), id, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Role updated successfully", role)
}

// NOTE - DeleteRole
func (h *Handler) DeleteRole(c echo.Context) error {
	id := c.Param("id")

	if err := h.service.DeleteRole(c.Request().Context(), id); err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Role deleted successfully", nil)
}
