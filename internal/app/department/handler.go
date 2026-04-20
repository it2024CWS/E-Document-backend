package department

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service Service
}

func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	departments := e.Group("/v1/departments", authMiddleware)
	departments.GET("", h.GetAllDepartments)
	departments.GET("/:id", h.GetDepartmentByID)
	departments.POST("", h.CreateDepartment)
	departments.PUT("/:id", h.UpdateDepartment)
	departments.DELETE("/:id", h.DeleteDepartment)
}

// GetAllDepartments godoc
//
//	@Summary		Get all departments
//	@Description	Get all departments with pagination and search
//	@Tags			departments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int		false	"Page number"
//	@Param			limit	query		int		false	"Items per page"
//	@Param			search	query		string	false	"Search by department name"
//	@Success		200		{object}	util.Response{data=[]domain.DepartmentResponse,pagination=util.PaginationInfo}
//	@Failure		401		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/departments [get]
func (h *Handler) GetAllDepartments(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 10
	}

	search := c.QueryParam("search")

	departments, total, err := h.service.GetAllDepartments(c.Request().Context(), page, limit, search)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponseWithPagination(c, "Departments retrieved successfully", departments, util.PaginationInfo{
		CurrentPage:  page,
		TotalPages:   (total + limit - 1) / limit,
		TotalItems:   total,
		ItemsPerPage: limit,
	})
}

// GetDepartmentByID godoc
//
//	@Summary		Get department by ID
//	@Description	Get a single department by its ID
//	@Tags			departments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Department ID"
//	@Success		200	{object}	util.Response{data=domain.DepartmentResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Failure		500	{object}	util.Response
//	@Router			/v1/departments/{id} [get]
func (h *Handler) GetDepartmentByID(c echo.Context) error {
	id := c.Param("id")
	if err := uuid.Validate(id); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid uuid"))
	}

	department, err := h.service.GetDepartmentByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Department retrieved successfully", department)
}

// CreateDepartment godoc
//
//	@Summary		Create a new department
//	@Description	Create a new department
//	@Tags			Departments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.CreateDepartmentRequest	true	"Department data"
//	@Success		201		{object}	util.Response{data=domain.DepartmentResponse}
//	@Failure		400		{object}	util.Response
//	@Router			/v1/departments [post]
func (h *Handler) CreateDepartment(c echo.Context) error {
	var req domain.CreateDepartmentRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
	}

	department, err := h.service.CreateDepartment(c.Request().Context(), req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Department created successfully", department, 201)
}

// UpdateDepartment godoc
//
//	@Summary		Update department
//	@Description	Update a department by its ID
//	@Tags			departments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"Department ID"
//	@Param			request	body		domain.UpdateDepartmentRequest	true	"Update department data"
//	@Success		200		{object}	util.Response{data=domain.DepartmentResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/departments/{id} [put]
func (h *Handler) UpdateDepartment(c echo.Context) error {
	id := c.Param("id")
	if err := uuid.Validate(id); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid uuid"))
	}

	var req domain.UpdateDepartmentRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request_body", "invalid request body"))
	}

	department, err := h.service.UpdateDepartment(c.Request().Context(), id, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Department updated successfully", department)
}

// DeleteDepartment godoc
//
//	@Summary		Delete department
//	@Description	Delete a department by its ID
//	@Tags			departments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Department ID"
//	@Success		200	{object}	util.Response
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Failure		500	{object}	util.Response
//	@Router			/v1/departments/{id} [delete]
func (h *Handler) DeleteDepartment(c echo.Context) error {
	id := c.Param("id")
	if err := uuid.Validate(id); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid uuid"))
	}

	if err := h.service.DeleteDepartment(c.Request().Context(), id); err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Department deleted successfully", nil)
}
