package incomingdoc

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"

	"strconv"

	"github.com/google/uuid"
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
	docs := e.Group("/v1/incoming-docs", middleware...)
	docs.GET("", h.GetAllIncomingDocs)
	docs.GET("/:id", h.GetIncomingDocByID)
	docs.GET("/receiver/:receiverId", h.GetIncomingDocsByReceiver)
	docs.GET("/department/:deptId", h.GetIncomingDocsByDepartment)
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

	// Get pagination params from query
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 10
	}

	// Check if user has a department and is not admin
	deptIDStr, ok := c.Get("dept_id").(string)
	if ok && deptIDStr != "" {
		deptID, err := uuid.Parse(deptIDStr)
		if err == nil {
			docs, total, err := h.service.GetIncomingDocsByDepartment(ctx, deptID, page, limit)
			if err != nil {
				return util.HandleError(c, err)
			}
			return util.OKResponseWithPagination(c, "Incoming documents retrieved successfully", docs, util.PaginationInfo{
				CurrentPage:  page,
				TotalPages:   (total + limit - 1) / limit,
				TotalItems:   total,
				ItemsPerPage: limit,
			})
		}
	}

	// No dept — return all docs except ones this user sent themselves.
	userIDStr, _ := c.Get("user_id").(string)
	userID, parseErr := uuid.Parse(userIDStr)
	if parseErr == nil {
		docs, total, err := h.service.GetAllIncomingDocsExcludingSender(ctx, userID, page, limit)
		if err != nil {
			return util.HandleError(c, err)
		}
		return util.OKResponseWithPagination(c, "Incoming documents retrieved successfully", docs, util.PaginationInfo{
			CurrentPage:  page,
			TotalPages:   (total + limit - 1) / limit,
			TotalItems:   total,
			ItemsPerPage: limit,
		})
	}

	docs, total, err := h.service.GetAllIncomingDocs(ctx, page, limit)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponseWithPagination(c, "Incoming documents retrieved successfully", docs, util.PaginationInfo{
		CurrentPage:  page,
		TotalPages:   (total + limit - 1) / limit,
		TotalItems:   total,
		ItemsPerPage: limit,
	})
}

// GetIncomingDocByID godoc
//
//	@Summary		Get incoming document by ID
//	@Description	Get incoming document details by ID
//	@Tags			incoming-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Incoming Document ID (UUID)"
//	@Success		200	{object}	util.Response{data=domain.IncomingDocResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/incoming-docs/{id} [get]
func (h *Handler) GetIncomingDocByID(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	doc, err := h.service.GetIncomingDocByID(ctx, id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Incoming document retrieved successfully", doc)
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
	receiverID, err := uuid.Parse(c.Param("receiverId"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("receiverId", "must be a valid UUID"))
	}

	docs, err := h.service.GetIncomingDocsByReceiver(ctx, receiverID)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Incoming documents for receiver retrieved successfully", docs)
}

// GetIncomingDocsByDepartment godoc
//
//	@Summary		Get incoming documents by department
//	@Description	Get incoming documents assigned to a specific department (UUID)
//	@Tags			incoming-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			deptId	path		string	true	"Department UUID"
//	@Success		200		{object}	util.Response{data=[]domain.IncomingDocResponse}
//	@Failure		401		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/incoming-docs/department/{deptId} [get]
func (h *Handler) GetIncomingDocsByDepartment(c echo.Context) error {
	ctx := c.Request().Context()
	deptID, err := uuid.Parse(c.Param("deptId"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("deptId", "must be a valid UUID"))
	}

	// Get pagination params from query
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 10
	}

	docs, total, err := h.service.GetIncomingDocsByDepartment(ctx, deptID, page, limit)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponseWithPagination(c, "Incoming documents for department retrieved successfully", docs, util.PaginationInfo{
		CurrentPage:  page,
		TotalPages:   (total + limit - 1) / limit,
		TotalItems:   total,
		ItemsPerPage: limit,
	})
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

	return util.OKResponse(c, "Incoming documents by status retrieved successfully", docs)
}

// ReceiveDocument godoc
//
//	@Summary		Receive incoming document
//	@Description	Mark an incoming document as received (by secretary)
//	@Tags			incoming-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string								true	"Incoming Document ID (UUID)"
//	@Param			request	body		domain.ReceiveIncomingDocRequest	true	"Receive data"
//	@Success		200		{object}	util.Response{data=domain.IncomingDocResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/incoming-docs/{id}/receive [post]
func (h *Handler) ReceiveDocument(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	var reqBody domain.ReceiveIncomingDocRequest
	if err := c.Bind(&reqBody); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
	}

	// Get receiver ID from context
	receiverID := c.Get("user_id").(string)

	// Create service request object
	req := domain.ReceiveDocumentRequest{
		IncomingDocID: id,
		ReceiverID:    receiverID,
		Remark:        reqBody.Remark,
	}

	doc, err := h.service.ReceiveDocument(ctx, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document received successfully", doc)
}

// ApproveDocument godoc
//
//	@Summary		Approve/Reject incoming document
//	@Description	Approve or reject an incoming document (by department head or director)
//	@Tags			incoming-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"Incoming Document ID (UUID)"
//	@Param			request	body		domain.ApproveDocumentRequest	true	"Approval data"
//	@Success		200		{object}	util.Response{data=domain.IncomingDocResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/incoming-docs/{id}/approve [post]
func (h *Handler) ApproveDocument(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	var req domain.ApproveDocumentRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request", "invalid request body"))
	}

	if req.ApproverID == "" {
		req.ApproverID = c.Get("user_id").(string)
	}

	doc, err := h.service.ApproveDocument(ctx, id, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Document evaluated successfully", doc)
}
