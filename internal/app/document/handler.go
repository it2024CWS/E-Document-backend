package document

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"net/http"
	"strconv"

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
	docs := e.Group("/documents", middleware...)
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
//	@Description	Get all documents for the authenticated user
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	util.Response{data=[]domain.DocumentResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		500	{object}	util.Response
//	@Router			/v1/documents [get]
func (h *Handler) GetAllDocuments(c echo.Context) error {
	ctx := c.Request().Context()
	userID := util.GetUserIDFromContext(c)

	docs, err := h.service.GetAllDocuments(ctx, userID)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(docs))
}

// GetDocumentByID godoc
//
//	@Summary		Get document by ID
//	@Description	Get document details by ID
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Document ID"
//	@Success		200	{object}	util.Response{data=domain.DocumentResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/documents/{id} [get]
func (h *Handler) GetDocumentByID(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	doc, err := h.service.GetDocumentByID(ctx, id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(doc))
}

// GetDocumentsByFolder godoc
//
//	@Summary		Get documents by folder
//	@Description	Get all documents in a specific folder
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			folderId	path		int	true	"Folder ID"
//	@Success		200			{object}	util.Response{data=[]domain.DocumentResponse}
//	@Failure		401			{object}	util.Response
//	@Failure		500			{object}	util.Response
//	@Router			/v1/documents/folder/{folderId} [get]
func (h *Handler) GetDocumentsByFolder(c echo.Context) error {
	ctx := c.Request().Context()
	folderID, err := strconv.Atoi(c.Param("folderId"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("folderId", "must be a valid integer"))
	}

	docs, err := h.service.GetDocumentsByFolder(ctx, folderID)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(docs))
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

	return c.JSON(http.StatusCreated, util.NewResponse(http.StatusCreated, "Document created successfully", doc))
}

// UpdateDocument godoc
//
//	@Summary		Update document
//	@Description	Update document details
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Document ID"
//	@Param			request	body		domain.UpdateDocumentRequest	true	"Document data"
//	@Success		200		{object}	util.Response{data=domain.DocumentResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/documents/{id} [put]
func (h *Handler) UpdateDocument(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
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

	return c.JSON(http.StatusOK, util.NewOKResponse(doc))
}

// GetDocumentVersions godoc
//
//	@Summary		Get document versions
//	@Description	Get all versions of a document
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Document ID"
//	@Success		200	{object}	util.Response{data=[]domain.VersionResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/documents/{id}/versions [get]
func (h *Handler) GetDocumentVersions(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	versions, err := h.service.GetDocumentVersions(ctx, id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(versions))
}

// SendDocument godoc
//
//	@Summary		Send document
//	@Description	Send document to a receiver (user)
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"Document ID"
//	@Param			request	body		domain.SendDocumentRequest	true	"Send data"
//	@Success		200		{object}	util.Response
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/documents/{id}/send [post]
func (h *Handler) SendDocument(c echo.Context) error {
	ctx := c.Request().Context()
	userID := util.GetUserIDFromContext(c)
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	var req domain.SendDocumentRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid request format"))
	}

	if err := h.service.SendDocument(ctx, userID, id, req); err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(map[string]string{"message": "Document sent successfully"}))
}

// DeleteDocument godoc
//
//	@Summary		Delete document
//	@Description	Delete document by ID
//	@Tags			documents
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Document ID"
//	@Success		200	{object}	util.Response
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/documents/{id} [delete]
func (h *Handler) DeleteDocument(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	if err := h.service.DeleteDocument(ctx, id); err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(nil))
}
