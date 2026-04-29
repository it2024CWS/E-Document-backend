package outgoingdoc

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"

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
	docs := e.Group("/v1/outgoing-docs", middleware...)
	docs.GET("", h.GetAllOutgoingDocs)
	docs.GET("/:id", h.GetOutgoingDocByID)
	docs.GET("/user/:userId", h.GetOutgoingDocsByUser)
	docs.GET("/department/:deptId", h.GetOutgoingDocsByDepartment)
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
			docs, total, err := h.service.GetOutgoingDocsByDepartment(ctx, deptID, page, limit)
			if err != nil {
				return util.HandleError(c, err)
			}
			return util.OKResponseWithPagination(c, "Outgoing documents retrieved successfully", docs, util.PaginationInfo{
				CurrentPage:  page,
				TotalPages:   (total + limit - 1) / limit,
				TotalItems:   total,
				ItemsPerPage: limit,
			})
		}
	}

	docs, total, err := h.service.GetAllOutgoingDocs(ctx, page, limit)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponseWithPagination(c, "Outgoing documents retrieved successfully", docs, util.PaginationInfo{
		CurrentPage:  page,
		TotalPages:   (total + limit - 1) / limit,
		TotalItems:   total,
		ItemsPerPage: limit,
	})
}

// GetOutgoingDocByID godoc
//
//	@Summary		Get outgoing document by ID
//	@Description	Get outgoing document details by ID
//	@Tags			outgoing-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Outgoing Document ID (UUID)"
//	@Success		200	{object}	util.Response{data=domain.OutgoingDocResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/outgoing-docs/{id} [get]
func (h *Handler) GetOutgoingDocByID(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	doc, err := h.service.GetOutgoingDocByID(ctx, id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Outgoing document retrieved successfully", doc)
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
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("userId", "must be a valid UUID"))
	}

	docs, err := h.service.GetOutgoingDocsByUser(ctx, userID)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Outgoing documents for user retrieved successfully", docs)
}

// GetOutgoingDocsByDepartment godoc
//
//	@Summary		Get outgoing documents by department
//	@Description	Get outgoing documents for a specific department
//	@Tags			outgoing-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			deptId	path		string	true	"Department UUID"
//	@Success		200		{object}	util.Response{data=[]domain.OutgoingDocResponse}
//	@Failure		401		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/outgoing-docs/department/{deptId} [get]
func (h *Handler) GetOutgoingDocsByDepartment(c echo.Context) error {
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

	docs, total, err := h.service.GetOutgoingDocsByDepartment(ctx, deptID, page, limit)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponseWithPagination(c, "Outgoing documents for department retrieved successfully", docs, util.PaginationInfo{
		CurrentPage:  page,
		TotalPages:   (total + limit - 1) / limit,
		TotalItems:   total,
		ItemsPerPage: limit,
	})
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

	return util.OKResponse(c, "Outgoing document created successfully", doc, 201)
}
