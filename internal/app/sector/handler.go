package sector

import (
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
	departments := e.Group("/v1/departments", authMiddleware)
	departments.GET("/:id/sectors", h.GetSectorsByDepartment)
}

// GetAllSectors godoc
//
//	@Summary		Get all sectors
//	@Description	Get list of all sectors
//	@Tags			Sectors
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	util.Response{data=[]domain.SectorResponse}
//	@Failure		401	{object}	util.Response
//	@Router			/v1/sectors [get]
func (h *Handler) GetAllSectors(c echo.Context) error {
	sectors, err := h.service.GetAllSectors(c.Request().Context())
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Sectors retrieved successfully", sectors)
}

// GetSectorByID godoc
//
//	@Summary		Get sector by ID
//	@Description	Get detailed information of a specific sector
//	@Tags			Sectors
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Sector ID"
//	@Success		200	{object}	util.Response{data=domain.SectorResponse}
//	@Failure		404	{object}	util.Response
//	@Router			/v1/sectors/{id} [get]
func (h *Handler) GetSectorByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	sector, err := h.service.GetSectorByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Sector retrieved successfully", sector)
}

// GetSectorsByDepartment godoc
//
//	@Summary		Get sectors by department
//	@Description	Get all sectors belonging to a department
//	@Tags			Departments
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Department ID"
//	@Success		200	{object}	util.Response{data=[]domain.SectorResponse}
//	@Failure		404	{object}	util.Response
//	@Router			/v1/departments/{id}/sectors [get]
func (h *Handler) GetSectorsByDepartment(c echo.Context) error {
	deptID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	sectors, err := h.service.GetSectorsByDepartment(c.Request().Context(), deptID)
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
//	@Description	Update sector information
//	@Tags			Sectors
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Sector ID"
//	@Param			request	body		domain.UpdateSectorRequest	true	"Sector data"
//	@Success		200		{object}	util.Response{data=domain.SectorResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/sectors/{id} [put]
func (h *Handler) UpdateSector(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	var req domain.UpdateSectorRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
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
//	@Description	Delete a sector
//	@Tags			Sectors
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Sector ID"
//	@Success		200	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/sectors/{id} [delete]
func (h *Handler) DeleteSector(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	if err := h.service.DeleteSector(c.Request().Context(), id); err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Sector deleted successfully", nil)
}
