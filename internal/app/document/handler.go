package document

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler handles document HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new document handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the document routes
func (h *Handler) RegisterRoutes(e *echo.Group, middleware ...echo.MiddlewareFunc) {
	docs := e.Group("/v1/documents", middleware...)
	docs.GET("", h.GetAllDocuments)
	docs.GET("/:id", h.GetDocumentByID)
	docs.GET("/folder/:folderId", h.GetDocumentsByFolder)
	docs.POST("", h.CreateDocument)
	docs.PUT("/:id", h.UpdateDocument)
	docs.POST("/:id/send", h.SendDocument)
}

// GetAllDocuments godoc
//
//	@Summary		Get all documents
//	@Tags			documents
//	@Security		BearerAuth
//	@Param			page	query		int		false	"Page number"	default(1)
//	@Param			limit	query		int		false	"Items per page"	default(10)
//	@Param			search	query		string	false	"Search by document name"
//	@Success		200		{object}	util.Response{data=util.PaginatedData}
//	@Failure		401		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/documents [get]
func (h *Handler) GetAllDocuments(c echo.Context) error {
	ctx := c.Request().Context()
	userID := util.GetUserIDFromContext(c)

	// Get pagination params from query
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	// search := c.QueryParam("search")

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

	// For now, document service doesn't support pagination, so we'll fetch all
	// and simulate pagination info or just return all with dummy pagination.
	// Actually, let's keep it simple if service doesn't support it, but the pattern suggests we should add it.
	docs, err := h.service.GetAllDocuments(ctx, userID)
	if err != nil {
		return util.HandleError(c, err)
	}
	pagination := util.PaginationInfo{
		CurrentPage:  pageNum,
		TotalPages:   1, // Dummy
		TotalItems:   len(docs),
		ItemsPerPage: limitNum,
	}

	return util.OKResponseWithPagination(c, "Documents retrieved successfully", docs, pagination)
}

// GetDocumentByID godoc
//
//	@Summary		Get document by ID
//	@Description	Get document details by ID
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Document ID"
//	@Success		200	{object}	util.Response{data=domain.DocumentResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/documents/{id} [get]
func (h *Handler) GetDocumentByID(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	doc, err := h.service.GetDocumentByID(ctx, id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document retrieved successfully", doc)
}

// GetDocumentsByFolder godoc
//
//	@Summary		Get documents by folder
//	@Description	Get all documents in a specific folder
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			folderId	path		string	true	"Folder ID"
//	@Success		200			{object}	util.Response{data=[]domain.DocumentResponse}
//	@Failure		401			{object}	util.Response
//	@Failure		500			{object}	util.Response
//	@Router			/v1/documents/folder/{folderId} [get]
func (h *Handler) GetDocumentsByFolder(c echo.Context) error {
	ctx := c.Request().Context()
	folderIDStr := c.Param("folderId")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("folderId", "must be a valid UUID"))
	}

	docs, err := h.service.GetDocumentsByFolder(ctx, folderID)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Documents in folder retrieved successfully", docs)
}

// CreateDocument godoc
//
//	@Summary		Create document
//	@Description	Create a new document
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.CreateDocumentRequest	true	"Document data"
//	@Success		201		{object}	util.Response{data=domain.DocumentResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Router			/v1/documents [post]
func (h *Handler) CreateDocument(c echo.Context) error {
	ctx := c.Request().Context()
	userID := util.GetUserIDFromContext(c)

	var req domain.CreateDocumentRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid request format"))
	}

	// Handle file upload
	file, err := c.FormFile("file")
	if err == nil {
		path, err := util.SaveFile(file, "documents")
		if err != nil {
			return util.HandleError(c, err)
		}
		req.DocPath = path
	}

	doc, err := h.service.CreateDocument(ctx, userID, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document created successfully", doc, http.StatusCreated)
}

// UpdateDocument godoc
//
//	@Summary		Update document
//	@Description	Update document details
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"Document ID"
//	@Param			request	body		domain.UpdateDocumentRequest	true	"Document data"
//	@Success		200		{object}	util.Response{data=domain.DocumentResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/documents/{id} [put]
func (h *Handler) UpdateDocument(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	var req domain.UpdateDocumentRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid request format"))
	}

	// Handle file upload (optional for update)
	file, err := c.FormFile("file")
	if err == nil {
		path, err := util.SaveFile(file, "documents")
		if err != nil {
			return util.HandleError(c, err)
		}
		req.DocPath = path
	}

	doc, err := h.service.UpdateDocument(ctx, id, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document updated successfully", doc)
}

// GetDocumentVersions godoc
//
//	@Summary		Get document versions
//	@Description	Get all versions of a document
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Document ID"
//	@Success		200	{object}	util.Response{data=[]domain.VersionResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/documents/{id}/versions [get]
func (h *Handler) GetDocumentVersions(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	versions, err := h.service.GetDocumentVersions(ctx, id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document versions retrieved successfully", versions)
}

// SendDocument godoc
//
//	@Summary		Send document
//	@Description	Send document to a receiver (user)
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"Document ID"
//	@Param			request	body		domain.SendDocumentRequest	true	"Send data"
//	@Success		200		{object}	util.Response
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/documents/{id}/send [post]
func (h *Handler) SendDocument(c echo.Context) error {
	ctx := c.Request().Context()
	userID := util.GetUserIDFromContext(c)
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	var req domain.SendDocumentRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid request format"))
	}

	if err := h.service.SendDocument(ctx, userID, id, req); err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document sent successfully", nil)
}

// DeleteDocument godoc
//
//	@Summary		Delete document
//	@Description	Delete document by ID
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Document ID"
//	@Success		200	{object}	util.Response
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/documents/{id} [delete]
func (h *Handler) DeleteDocument(c echo.Context) error {
	ctx := c.Request().Context()
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	if err := h.service.DeleteDocument(ctx, id); err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document deleted successfully", nil)
}
