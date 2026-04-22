package folder

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler handles folder HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new folder handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the folder routes
func (h *Handler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	folders := e.Group("/v1/folders", authMiddleware)
	folders.GET("", h.GetAllFolders)
	folders.GET("/:id", h.GetFolderByID)
	folders.POST("", h.CreateFolder)
	folders.PUT("/:id", h.UpdateFolder)
	folders.DELETE("/:id", h.DeleteFolder)
}

// GetAllFolders godoc
//
//	@Param			page	query		int		false	"Page number"	default(1)
//	@Param			limit	query		int		false	"Items per page"	default(10)
//	@Param			search	query		string	false	"Search by folder name"
//	@Success		200		{object}	util.Response{data=util.PaginatedData}
//	@Failure		401		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/folders [get]
func (h *Handler) GetAllFolders(c echo.Context) error {
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

	folders, err := h.service.GetAllFolders(c.Request().Context())
	if err != nil {
		return util.HandleError(c, err)
	}

	pagination := util.PaginationInfo{
		CurrentPage:  pageNum,
		TotalPages:   1, // Dummy
		TotalItems:   len(folders),
		ItemsPerPage: limitNum,
	}

	return util.OKResponseWithPagination(c, "Folders retrieved successfully", folders, pagination)
}

// GetFolderByID godoc
//
//	@Summary		Get folder by ID
//	@Tags			folders
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Folder ID"
//	@Success		200	{object}	util.Response{data=domain.FolderResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/folders/{id} [get]
func (h *Handler) GetFolderByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	folder, err := h.service.GetFolderByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Folder retrieved successfully", folder)
}

// CreateFolder godoc
//
//	@Summary		Create folder
//	@Tags			folders
//	@Security		BearerAuth
//	@Param			request	body	domain.CreateFolderRequest	true	"Folder data"
//	@Success		201	{object}	util.Response{data=domain.FolderResponse}
//	@Router			/v1/folders [post]
func (h *Handler) CreateFolder(c echo.Context) error {
	userID := util.GetUserIDFromContext(c)
	var req domain.CreateFolderRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid request format"))
	}
	folder, err := h.service.CreateFolder(c.Request().Context(), userID, req)
	if err != nil {
		return util.HandleError(c, err)
	}
	return util.OKResponse(c, "Folder created successfully", folder, http.StatusCreated)
}

// UpdateFolder godoc
//
//	@Summary		Update folder
//	@Tags			folders
//	@Security		BearerAuth
//	@Param			id		path	string						true	"Folder ID"
//	@Param			request	body	domain.CreateFolderRequest	true	"Folder data"
//	@Success		200		{object}	util.Response{data=domain.FolderResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/folders/{id} [put]
func (h *Handler) UpdateFolder(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	var req domain.CreateFolderRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid request format"))
	}

	folder, err := h.service.UpdateFolder(c.Request().Context(), id, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Folder updated successfully", folder)
}

// DeleteFolder godoc
//
//	@Summary		Delete folder
//	@Tags			folders
//	@Security		BearerAuth
//	@Param			id	path	string	true	"Folder ID"
//	@Success		200	{object}	util.Response
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/folders/{id} [delete]
func (h *Handler) DeleteFolder(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	if err := h.service.DeleteFolder(c.Request().Context(), id); err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Folder deleted successfully", nil)
}
