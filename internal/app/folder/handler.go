package folder

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"net/http"

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
//	@Summary		Get all folders
//	@Description	Get all folders
//	@Tags			folders
//	@Security		BearerAuth
//	@Success		200	{object}	util.Response{data=[]domain.FolderResponse}
//	@Router			/v1/folders [get]
func (h *Handler) GetAllFolders(c echo.Context) error {
	folders, err := h.service.GetAllFolders(c.Request().Context())
	if err != nil {
		return util.HandleError(c, err)
	}
	return c.JSON(http.StatusOK, util.NewOKResponse(folders))
}

// GetFolderByID godoc
//
//	@Summary		Get folder by ID
//	@Tags			folders
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Folder ID"
//	@Success		200	{object}	util.Response{data=domain.FolderResponse}
//	@Router			/v1/folders/{id} [get]
func (h *Handler) GetFolderByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}
	folder, err := h.service.GetFolderByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}
	return c.JSON(http.StatusOK, util.NewOKResponse(folder))
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
	return c.JSON(http.StatusCreated, util.NewResponse(http.StatusCreated, "Folder created successfully", folder))
}

// UpdateFolder godoc
//
//	@Summary		Update folder
//	@Tags			folders
//	@Security		BearerAuth
//	@Param			id		path	int						true	"Folder ID"
//	@Param			request	body	domain.CreateFolderRequest	true	"Folder data"
//	@Success		200		{object}	util.Response{data=domain.FolderResponse}
//	@Router			/v1/folders/{id} [put]
func (h *Handler) UpdateFolder(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
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
	return c.JSON(http.StatusOK, util.NewOKResponse(folder))
}

// DeleteFolder godoc
//
//	@Summary		Delete folder
//	@Tags			folders
//	@Security		BearerAuth
//	@Param			id	path	int	true	"Folder ID"
//	@Success		200	{object}	util.Response
//	@Router			/v1/folders/{id} [delete]
func (h *Handler) DeleteFolder(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}
	if err := h.service.DeleteFolder(c.Request().Context(), id); err != nil {
		return util.HandleError(c, err)
	}
	return c.JSON(http.StatusOK, util.NewOKResponse(nil))
}
