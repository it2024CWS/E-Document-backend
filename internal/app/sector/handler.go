package sector

import (
	"strconv"

	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"

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
	sectors := e.Group("/v1/sectors", authMiddleware)
	sectors.GET("", h.GetAllSectors)
	sectors.GET("/:id", h.GetSectorByID)
	sectors.POST("", h.CreateSector)
	sectors.PUT("/:id", h.UpdateSector)
	sectors.DELETE("/:id", h.DeleteSector)

	// Department-specific routes
	sectors.GET("/department/:deptId", h.GetSectorsByDepartment)
}

// GetAllSectors godoc
//
//	@Summary		Get all sectors
//	@Description	Get all sectors with pagination and search
//	@Tags			sectors
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int		false	"Page number"
//	@Param			limit	query		int		false	"Items per page"
//	@Param			search	query		string	false	"Search by sector name"
//	@Success		200		{object}	util.Response{data=[]domain.SectorResponse,pagination=util.PaginationInfo}
//	@Failure		401		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/sectors [get]
func (h *Handler) GetAllSectors(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 10
	}

	search := c.QueryParam("search")

	sectors, total, err := h.service.GetAllSectors(c.Request().Context(), page, limit, search)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponseWithPagination(c, "Sectors retrieved successfully", sectors, util.PaginationInfo{
		CurrentPage:  page,
		TotalPages:   (total + limit - 1) / limit,
		TotalItems:   total,
		ItemsPerPage: limit,
	})
}

// GetSectorByID godoc
//
//	@Summary		Get sector by ID
//	@Description	Get a single sector by its ID
//	@Tags			sectors
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Sector ID"
//	@Success		200	{object}	util.Response{data=domain.SectorResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Failure		500	{object}	util.Response
//	@Router			/v1/sectors/{id} [get]
func (h *Handler) GetSectorByID(c echo.Context) error {
	id := c.Param("id")
	if err := uuid.Validate(id); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid uuid"))
	}

	sector, err := h.service.GetSectorByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Sector retrieved successfully", sector)
}

// GetSectorsByDepartment godoc
//
//	@Summary		Get sectors by department ID
//	@Description	Get all sectors belonging to a specific department
//	@Tags			sectors
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			deptId	path		string	true	"Department ID"
//	@Success		200		{object}	util.Response{data=[]domain.SectorResponse}
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/sectors/department/{deptId} [get]
func (h *Handler) GetSectorsByDepartment(c echo.Context) error {
	deptID := c.Param("deptId")
	if err := uuid.Validate(deptID); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("deptId", "must be a valid uuid"))
	}

	sectors, err := h.service.GetSectorsByDepartmentID(c.Request().Context(), deptID)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Sectors retrieved successfully", sectors)
}

// CreateSector godoc
//
//	@Summary		Create a new sector
//	@Description	Create a new sector
//	@Tags			Sectors
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.CreateSectorRequest	true	"Sector data"
//	@Success		201		{object}	util.Response{data=domain.SectorResponse}
//	@Failure		400		{object}	util.Response
//	@Router			/v1/sectors [post]
func (h *Handler) CreateSector(c echo.Context) error {
	var req domain.CreateSectorRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
	}

	sector, err := h.service.CreateSector(c.Request().Context(), req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Sector created successfully", sector, 201)
}

// UpdateSector godoc
//
//	@Summary		Update sector
//	@Description	Update a sector by its ID
//	@Tags			sectors
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string						true	"Sector ID"
//	@Param			request	body		domain.UpdateSectorRequest	true	"Update sector data"
//	@Success		200		{object}	util.Response{data=domain.SectorResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/sectors/{id} [put]
func (h *Handler) UpdateSector(c echo.Context) error {
	id := c.Param("id")
	if err := uuid.Validate(id); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid uuid"))
	}

	var req domain.UpdateSectorRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request_body", "invalid request body"))
	}

	sector, err := h.service.UpdateSector(c.Request().Context(), id, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Sector updated successfully", sector)
}

// DeleteSector godoc
//
//	@Summary		Delete sector
//	@Description	Delete a sector by its ID
//	@Tags			sectors
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Sector ID"
//	@Success		200	{object}	util.Response
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Failure		500	{object}	util.Response
//	@Router			/v1/sectors/{id} [delete]
func (h *Handler) DeleteSector(c echo.Context) error {
	id := c.Param("id")
	if err := uuid.Validate(id); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid uuid"))
	}

	if err := h.service.DeleteSector(c.Request().Context(), id); err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Sector deleted successfully", nil)
}
