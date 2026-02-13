package outgoingdoc

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler handles outgoing document HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new outgoing document handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the outgoing document routes
func (h *Handler) RegisterRoutes(e *echo.Group, middleware ...echo.MiddlewareFunc) {
	docs := e.Group("/outgoing-docs", middleware...)
	docs.GET("", h.GetAllOutgoingDocs)
	docs.GET("/:id", h.GetOutgoingDocByID)
	docs.GET("/user/:userId", h.GetOutgoingDocsByUser)
	docs.POST("", h.CreateOutgoingDoc)
}

// GetAllOutgoingDocs godoc
//
//	@Summary		Get all outgoing documents
//	@Description	Get all outgoing documents
//	@Tags			outgoing-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	util.Response{data=[]domain.OutgoingDocResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		500	{object}	util.Response
//	@Router			/v1/outgoing-docs [get]
func (h *Handler) GetAllOutgoingDocs(c echo.Context) error {
	ctx := c.Request().Context()

	docs, err := h.service.GetAllOutgoingDocs(ctx)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(docs))
}

// GetOutgoingDocByID godoc
//
//	@Summary		Get outgoing document by ID
//	@Description	Get outgoing document details by ID
//	@Tags			outgoing-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Outgoing Document ID"
//	@Success		200	{object}	util.Response{data=domain.OutgoingDocResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/outgoing-docs/{id} [get]
func (h *Handler) GetOutgoingDocByID(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	doc, err := h.service.GetOutgoingDocByID(ctx, id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(doc))
}

// GetOutgoingDocsByUser godoc
//
//	@Summary		Get outgoing documents by user
//	@Description	Get outgoing documents for a specific user
//	@Tags			outgoing-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			userId	path		string	true	"User UUID"
//	@Success		200		{object}	util.Response{data=[]domain.OutgoingDocResponse}
//	@Failure		401		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/outgoing-docs/user/{userId} [get]
func (h *Handler) GetOutgoingDocsByUser(c echo.Context) error {
	ctx := c.Request().Context()
	userID := c.Param("userId")
	// UUID validation could be added here

	docs, err := h.service.GetOutgoingDocsByUser(ctx, userID)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(docs))
}

// CreateOutgoingDoc godoc
//
//	@Summary		Create outgoing document
//	@Description	Create a new outgoing document record
//	@Tags			outgoing-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		domain.CreateOutgoingDocRequest	true	"Outgoing Document data"
//	@Success		201		{object}	util.Response{data=domain.OutgoingDocResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Router			/v1/outgoing-docs [post]
func (h *Handler) CreateOutgoingDoc(c echo.Context) error {
	ctx := c.Request().Context()

	// Get current user ID
	userIDStr := util.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return util.HandleError(c, util.NewUnauthorizedError("invalid user ID in token"))
	}

	var req domain.CreateOutgoingDocRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
	}

	// Ensure user ID is set to current user if not provided (though in this case we might force it)
	if req.UserID == nil {
		req.UserID = &userID
	}

	doc, err := h.service.CreateOutgoingDoc(ctx, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusCreated, util.NewResponse(http.StatusCreated, "Outgoing document created successfully", doc))
}
