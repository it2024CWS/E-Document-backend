package folder_file_manage

import (
	"e-document-backend/internal/util"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for storage operations
type Handler struct {
	service Service
}

// NewHandler creates a new storage handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers storage routes
func (h *Handler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	storage := e.Group("/v1/storage", authMiddleware)

	// Folder routes
	storage.GET("/folders/root", h.GetRootFolders)
	storage.GET("/folders/:id", h.GetFolder)
	storage.GET("/folders/:id/contents", h.GetFolderContents)
	storage.GET("/folders/:id/subfolders", h.GetSubfolders)
	storage.GET("/folders/:id/documents", h.GetDocumentsByFolder)

	// Document routes
	storage.GET("/documents", h.GetAllDocuments)
	storage.GET("/documents/:id", h.GetDocument)

	// Recent files
	storage.GET("/recent", h.GetRecentFiles)
}

// GetRootFolders godoc
// @Summary		Get root folders
// @Description	Get all root folders for the authenticated user
// @Tags		Storage
// @Produce		json
// @Security	BearerAuth
// @Param		page		query		int		false	"Page number"		default(1)
// @Param		page_size	query		int		false	"Items per page"	default(20)
// @Success		200			{object}	util.Response
// @Failure		401			{object}	util.Response
// @Failure		500			{object}	util.Response
// @Router		/v1/storage/folders/root [get]
func (h *Handler) GetRootFolders(c echo.Context) error {
	// Get user ID from context
	userID := c.Get("user_id").(string)
	ownerID, err := uuid.Parse(userID)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid user ID", util.INVALID_INPUT, 400, err.Error()))
	}

	// Get pagination params
	page := 1
	pageSize := 20
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.QueryParam("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	// Get root folders
	folders, total, err := h.service.GetRootFolders(c.Request().Context(), ownerID, page, pageSize)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Failed to get root folders", util.INTERNAL_SERVER_ERROR, 500, err.Error()))
	}

	// Calculate pagination info
	totalPages := (total + pageSize - 1) / pageSize
	pagination := util.PaginationInfo{
		CurrentPage:  page,
		TotalPages:   totalPages,
		TotalItems:   total,
		ItemsPerPage: pageSize,
	}

	return util.OKResponseWithPagination(c, "Root folders retrieved successfully", folders, pagination)
}

// GetFolder godoc
// @Summary		Get folder details
// @Description	Get folder information by ID
// @Tags		Storage
// @Produce		json
// @Security	BearerAuth
// @Param		id	path		string	true	"Folder ID"
// @Success		200	{object}	util.Response
// @Failure		400	{object}	util.Response
// @Failure		401	{object}	util.Response
// @Failure		404	{object}	util.Response
// @Router		/v1/storage/folders/{id} [get]
func (h *Handler) GetFolder(c echo.Context) error {
	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid folder ID", util.INVALID_INPUT, 400, err.Error()))
	}

	folder, err := h.service.GetFolder(c.Request().Context(), folderID)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Folder not found", util.VALIDATION_ERROR, 404, err.Error()))
	}

	return util.OKResponse(c, "Folder retrieved successfully", folder)
}

// GetFolderContents godoc
// @Summary		Get folder contents
// @Description	Get folder information with subfolders and documents
// @Tags		Storage
// @Produce		json
// @Security	BearerAuth
// @Param		id	path		string	true	"Folder ID"
// @Success		200	{object}	util.Response
// @Failure		400	{object}	util.Response
// @Failure		401	{object}	util.Response
// @Failure		404	{object}	util.Response
// @Router		/v1/storage/folders/{id}/contents [get]
func (h *Handler) GetFolderContents(c echo.Context) error {
	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid folder ID", util.INVALID_INPUT, 400, err.Error()))
	}

	contents, err := h.service.GetFolderContents(c.Request().Context(), folderID)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Failed to get folder contents", util.INTERNAL_SERVER_ERROR, 500, err.Error()))
	}

	return util.OKResponse(c, "Folder contents retrieved successfully", contents)
}

// GetSubfolders godoc
// @Summary		Get subfolders
// @Description	Get subfolders of a folder with pagination
// @Tags		Storage
// @Produce		json
// @Security	BearerAuth
// @Param		id			path		string	true	"Folder ID"
// @Param		page		query		int		false	"Page number"		default(1)
// @Param		page_size	query		int		false	"Items per page"	default(20)
// @Success		200			{object}	util.Response
// @Failure		400			{object}	util.Response
// @Failure		401			{object}	util.Response
// @Router		/v1/storage/folders/{id}/subfolders [get]
func (h *Handler) GetSubfolders(c echo.Context) error {
	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid folder ID", util.INVALID_INPUT, 400, err.Error()))
	}

	// Get pagination params
	page := 1
	pageSize := 20
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.QueryParam("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	folders, pagination, err := h.service.GetSubfolders(c.Request().Context(), folderID, page, pageSize)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Failed to get subfolders", util.INTERNAL_SERVER_ERROR, 500, err.Error()))
	}

	return c.JSON(200, map[string]interface{}{
		"success":    true,
		"message":    "Subfolders retrieved successfully",
		"data":       folders,
		"pagination": pagination,
	})
}

// GetDocumentsByFolder godoc
// @Summary		Get documents in a folder
// @Description	Get all documents in a specific folder with pagination
// @Tags		Storage
// @Produce		json
// @Security	BearerAuth
// @Param		id			path		string	true	"Folder ID"
// @Param		page		query		int		false	"Page number"		default(1)
// @Param		page_size	query		int		false	"Items per page"	default(20)
// @Success		200			{object}	util.Response
// @Failure		400			{object}	util.Response
// @Failure		401			{object}	util.Response
// @Router		/v1/storage/folders/{id}/documents [get]
func (h *Handler) GetDocumentsByFolder(c echo.Context) error {
	folderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid folder ID", util.INVALID_INPUT, 400, err.Error()))
	}

	// Get pagination params
	page := 1
	pageSize := 20
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.QueryParam("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	documents, pagination, err := h.service.GetDocumentsByFolder(c.Request().Context(), folderID, page, pageSize)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Failed to get documents", util.INTERNAL_SERVER_ERROR, 500, err.Error()))
	}

	return c.JSON(200, map[string]interface{}{
		"success":    true,
		"message":    "Documents retrieved successfully",
		"data":       documents,
		"pagination": pagination,
	})
}

// GetAllDocuments godoc
// @Summary		Get all documents
// @Description	Get all documents for the authenticated user with pagination
// @Tags		Storage
// @Produce		json
// @Security	BearerAuth
// @Param		page		query		int		false	"Page number"		default(1)
// @Param		page_size	query		int		false	"Items per page"	default(20)
// @Success		200			{object}	util.Response
// @Failure		401			{object}	util.Response
// @Failure		500			{object}	util.Response
// @Router		/v1/storage/documents [get]
func (h *Handler) GetAllDocuments(c echo.Context) error {
	// Get user ID from context
	userID := c.Get("user_id").(string)
	ownerID, err := uuid.Parse(userID)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid user ID", util.INVALID_INPUT, 400, err.Error()))
	}

	// Get pagination params
	page := 1
	pageSize := 20
	if p := c.QueryParam("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := c.QueryParam("page_size"); ps != "" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 100 {
			pageSize = parsed
		}
	}

	documents, pagination, err := h.service.GetAllDocuments(c.Request().Context(), ownerID, page, pageSize)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Failed to get documents", util.INTERNAL_SERVER_ERROR, 500, err.Error()))
	}

	return c.JSON(200, map[string]interface{}{
		"success":    true,
		"message":    "Documents retrieved successfully",
		"data":       documents,
		"pagination": pagination,
	})
}

// GetDocument godoc
// @Summary		Get document details
// @Description	Get document information with current attachment by ID
// @Tags		Storage
// @Produce		json
// @Security	BearerAuth
// @Param		id	path		string	true	"Document ID"
// @Success		200	{object}	util.Response
// @Failure		400	{object}	util.Response
// @Failure		401	{object}	util.Response
// @Failure		404	{object}	util.Response
// @Router		/v1/storage/documents/{id} [get]
func (h *Handler) GetDocument(c echo.Context) error {
	documentID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid document ID", util.INVALID_INPUT, 400, err.Error()))
	}

	document, err := h.service.GetDocument(c.Request().Context(), documentID)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Document not found", util.VALIDATION_ERROR, 404, err.Error()))
	}

	return util.OKResponse(c, "Document retrieved successfully", document)
}

// GetRecentFiles godoc
// @Summary		Get recent files
// @Description	Get recently modified files for the authenticated user
// @Tags		Storage
// @Produce		json
// @Security	BearerAuth
// @Param		limit	query		int		false	"Number of files to return"	default(10)
// @Success		200		{object}	util.Response
// @Failure		401		{object}	util.Response
// @Failure		500		{object}	util.Response
// @Router		/v1/storage/recent [get]
func (h *Handler) GetRecentFiles(c echo.Context) error {
	// Get user ID from context
	userID := c.Get("user_id").(string)
	ownerID, err := uuid.Parse(userID)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid user ID", util.INVALID_INPUT, 400, err.Error()))
	}

	// Get limit param
	limit := 10
	if l := c.QueryParam("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 50 {
			limit = parsed
		}
	}

	files, err := h.service.GetRecentFiles(c.Request().Context(), ownerID, limit)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Failed to get recent files", util.INTERNAL_SERVER_ERROR, 500, err.Error()))
	}

	return util.OKResponse(c, "Recent files retrieved successfully", files)
}
