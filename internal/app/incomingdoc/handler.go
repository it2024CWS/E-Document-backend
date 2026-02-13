package incomingdoc

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

// Handler handles incoming document HTTP requests
type Handler struct {
	service Service
}

// NewHandler creates a new incoming document handler
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the incoming document routes
func (h *Handler) RegisterRoutes(e *echo.Group, middleware ...echo.MiddlewareFunc) {
	docs := e.Group("/incoming-docs", middleware...)
	docs.GET("", h.GetAllIncomingDocs)
	docs.GET("/:id", h.GetIncomingDocByID)
	docs.GET("/receiver/:receiverId", h.GetIncomingDocsByReceiver)
	docs.GET("/status/:status", h.GetIncomingDocsByStatus)
	docs.POST("/:id/receive", h.ReceiveDocument)
	docs.POST("/:id/approve", h.ApproveDocument)
}

// GetAllIncomingDocs godoc
//
//	@Summary		Get all incoming documents
//	@Description	Get all incoming documents (for admin/staff)
//	@Tags			incoming-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Success		200	{object}	util.Response{data=[]domain.IncomingDocResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		500	{object}	util.Response
//	@Router			/v1/incoming-docs [get]
func (h *Handler) GetAllIncomingDocs(c echo.Context) error {
	ctx := c.Request().Context()

	docs, err := h.service.GetAllIncomingDocs(ctx)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(docs))
}

// GetIncomingDocByID godoc
//
//	@Summary		Get incoming document by ID
//	@Description	Get incoming document details by ID
//	@Tags			incoming-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		int	true	"Incoming Document ID"
//	@Success		200	{object}	util.Response{data=domain.IncomingDocResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/incoming-docs/{id} [get]
func (h *Handler) GetIncomingDocByID(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	doc, err := h.service.GetIncomingDocByID(ctx, id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(doc))
}

// GetIncomingDocsByReceiver godoc
//
//	@Summary		Get incoming documents by receiver
//	@Description	Get incoming documents assigned to a specific receiver (user UUID)
//	@Tags			incoming-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			receiverId	path		string	true	"Receiver User UUID"
//	@Success		200			{object}	util.Response{data=[]domain.IncomingDocResponse}
//	@Failure		401			{object}	util.Response
//	@Failure		500			{object}	util.Response
//	@Router			/v1/incoming-docs/receiver/{receiverId} [get]
func (h *Handler) GetIncomingDocsByReceiver(c echo.Context) error {
	ctx := c.Request().Context()
	receiverID := c.Param("receiverId")
	// UUID validation could be added here

	docs, err := h.service.GetIncomingDocsByReceiver(ctx, receiverID)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(docs))
}

// GetIncomingDocsByStatus godoc
//
//	@Summary		Get incoming documents by status
//	@Description	Get incoming documents filtered by status
//	@Tags			incoming-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			status	path		string	true	"Status (pending, received, approved, rejected)"
//	@Success		200		{object}	util.Response{data=[]domain.IncomingDocResponse}
//	@Failure		401		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/incoming-docs/status/{status} [get]
func (h *Handler) GetIncomingDocsByStatus(c echo.Context) error {
	ctx := c.Request().Context()
	status := c.Param("status")

	docs, err := h.service.GetIncomingDocsByStatus(ctx, status)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(docs))
}

// ReceiveDocument godoc
//
//	@Summary		Receive incoming document
//	@Description	Mark an incoming document as received (by secretary)
//	@Tags			incoming-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"Incoming Document ID"
//	@Param			request	body		domain.ReceiveIncomingDocRequest	true	"Receive data"
//	@Success		200		{object}	util.Response{data=domain.IncomingDocResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/incoming-docs/{id}/receive [post]
func (h *Handler) ReceiveDocument(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	// Get receiver ID from context (authenticated user)
	receiverID := util.GetUserIDFromContext(c)

	var reqBody domain.ReceiveIncomingDocRequest
	if err := c.Bind(&reqBody); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
	}

	// Create service request object with both ID and body data
	req := domain.ReceiveDocumentRequest{
		IncomingDocID: id,
		ReceiverID:    receiverID, // Use string UUID directly
		Remark:        reqBody.Remark,
	}

	doc, err := h.service.ReceiveDocument(ctx, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(doc))
}

// ApproveDocument godoc
//
//	@Summary		Approve/Reject incoming document
//	@Description	Approve or reject an incoming document (by department head or director)
//	@Tags			incoming-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int								true	"Incoming Document ID"
//	@Param			request	body		domain.ApproveIncomingDocRequest	true	"Approval data"
//	@Success		200		{object}	util.Response{data=domain.IncomingDocResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/incoming-docs/{id}/approve [post]
func (h *Handler) ApproveDocument(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid integer"))
	}

	// Get approver ID from context
	approverID := util.GetUserIDFromContext(c)

	var reqBody domain.ApproveIncomingDocRequest
	if err := c.Bind(&reqBody); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
	}

	// Determine status based on boolean Approved field
	status := "approved"
	if !reqBody.Approved {
		status = "rejected"
	}

	// Create service request object
	req := domain.ApproveDocumentRequest{
		ApproverID: approverID,
		Remark:     reqBody.Remark,
		Status:     status,
	}

	doc, err := h.service.ApproveDocument(ctx, id, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return c.JSON(http.StatusOK, util.NewOKResponse(doc))
}
