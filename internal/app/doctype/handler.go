package doctype

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
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	doctypes := e.Group("/v1/doctypes", authMiddleware)
	doctypes.GET("", h.GetAllDocTypes)
	doctypes.GET("/:id", h.GetDocTypeByID)
	doctypes.POST("", h.CreateDocType)
	doctypes.PUT("/:id", h.UpdateDocType)
	doctypes.DELETE("/:id", h.DeleteDocType)
}

// GetAllDocTypes godoc
//
//	@Summary		Get all document types
//	@Description	Get list of all document types
//	@Tags			DocTypes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	util.Response{data=[]domain.DocTypeResponse}
//	@Failure		401	{object}	util.Response
//	@Router			/v1/doctypes [get]
func (h *Handler) GetAllDocTypes(c echo.Context) error {
	docTypes, err := h.service.GetAllDocTypes(c.Request().Context())
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document types retrieved successfully", docTypes)
}

// GetDocTypeByID godoc
//
//	@Summary		Get document type by ID
//	@Description	Get detailed information of a specific document type
//	@Tags			DocTypes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"DocType ID"
//	@Success		200	{object}	util.Response{data=domain.DocTypeResponse}
//	@Failure		404	{object}	util.Response
//	@Router			/v1/doctypes/{id} [get]
func (h *Handler) GetDocTypeByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	docType, err := h.service.GetDocTypeByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document type retrieved successfully", docType)
}

// CreateDocType godoc
//
//	@Summary		Create a new document type
//	@Description	Create a new document type
//	@Tags			DocTypes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.CreateDocTypeRequest	true	"DocType data"
//	@Success		201		{object}	util.Response{data=domain.DocTypeResponse}
//	@Failure		400		{object}	util.Response
//	@Router			/v1/doctypes [post]
func (h *Handler) CreateDocType(c echo.Context) error {
	var req domain.CreateDocTypeRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
	}

	docType, err := h.service.CreateDocType(c.Request().Context(), req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document type created successfully", docType, 201)
}

// UpdateDocType godoc
//
//	@Summary		Update document type
//	@Description	Update document type information
//	@Tags			DocTypes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int							true	"DocType ID"
//	@Param			request	body		domain.UpdateDocTypeRequest	true	"DocType data"
//	@Success		200		{object}	util.Response{data=domain.DocTypeResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/doctypes/{id} [put]
func (h *Handler) UpdateDocType(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	var req domain.UpdateDocTypeRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
	}

	docType, err := h.service.UpdateDocType(c.Request().Context(), id, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document type updated successfully", docType)
}

// DeleteDocType godoc
//
//	@Summary		Delete document type
//	@Description	Delete a document type
//	@Tags			DocTypes
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"DocType ID"
//	@Success		200	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/doctypes/{id} [delete]
func (h *Handler) DeleteDocType(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	if err := h.service.DeleteDocType(c.Request().Context(), id); err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document type deleted successfully", nil)
}
