package outgoingdoc

import (
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"strconv"
	"strings"
	"time"

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
	docs.GET("/department/:deptId", h.GetOutgoingDocsByDepartment)
	docs.POST("", h.CreateOutgoingDoc)
	docs.POST("/:id/approve", h.ApproveOutgoingDoc)
	docs.PUT("/:id", h.UpdateOutgoingDoc)
	docs.DELETE("/:id", h.DeleteOutgoingDoc)
}

func buildFilter(c echo.Context) DocFilter {
	f := DocFilter{
		DocNo:  c.QueryParam("doc_no"),
		Status: c.QueryParam("status"),
	}

	startStr := c.QueryParam("start_date")
	endStr := c.QueryParam("end_date")

	today := time.Now().In(time.Local).Truncate(24 * time.Hour)

	if startStr != "" {
		if t, err := time.ParseInLocation("2006-01-02", startStr, time.Local); err == nil {
			f.StartDate = &t
		}
	} else {
		f.StartDate = &today
	}

	if endStr != "" {
		if t, err := time.ParseInLocation("2006-01-02", endStr, time.Local); err == nil {
			t = t.Add(24 * time.Hour)
			f.EndDate = &t
		}
	} else {
		tomorrow := today.Add(24 * time.Hour)
		f.EndDate = &tomorrow
	}

	return f
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

	filter := buildFilter(c)

	// Check if user has a department and is not admin
	deptIDStr, ok := c.Get("dept_id").(string)
	if ok && deptIDStr != "" {
		deptID, err := uuid.Parse(deptIDStr)
		if err == nil {
			docs, total, err := h.service.GetOutgoingDocsByDepartment(ctx, deptID, page, limit, filter)
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

	docs, total, err := h.service.GetAllOutgoingDocs(ctx, page, limit, filter)
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

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}

	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 10
	}

	docs, total, err := h.service.GetOutgoingDocsByDepartment(ctx, deptID, page, limit, buildFilter(c))
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

	userIDStr := util.GetUserIDFromContext(c)
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return util.HandleError(c, util.NewUnauthorizedError("invalid user ID in token"))
	}

	var req domain.CreateOutgoingDocRequest
	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
	}

	if req.CreatedBy == nil {
		req.CreatedBy = &userID
	}

	_, err = h.service.CreateOutgoingDocWithParams(ctx, req.DocDetailsID, req.CreatedBy, nil)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Outgoing document created successfully", nil, 201)
}

// ApproveOutgoingDoc godoc
//
//	@Summary		Approve/Reject outgoing document
//	@Description	Owner department head approves or rejects an outgoing document. On approval the sequential recipient flow starts.
//	@Tags			outgoing-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string							true	"Outgoing Document ID (UUID)"
//	@Param			request	body		domain.ApproveDocumentRequest	true	"Approval data"
//	@Success		200		{object}	util.Response{data=domain.OutgoingDocResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/outgoing-docs/{id}/approve [post]
func (h *Handler) ApproveOutgoingDoc(c echo.Context) error {
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

	doc, err := h.service.ApproveOutgoingDoc(ctx, id, req)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Outgoing document evaluated successfully", doc)
}

// updateOutgoingDocPayload is the JSON body accepted by PUT /v1/outgoing-docs/:id.
// A missing receiver_ids field leaves the recipient route untouched; an empty
// list [] clears it. Empty strings for doc_no / description are skipped.
type updateOutgoingDocPayload struct {
	DocNo       string    `json:"doc_no"`
	Description string    `json:"description"`
	ReceiverIDs *[]string `json:"receiver_ids,omitempty"`
}

// UpdateOutgoingDoc godoc
//
//	@Summary		Update outgoing document
//	@Description	Edit metadata (doc_no, description) and/or recipient departments while the doc is still pending. Blocked once the owner department head has approved or rejected.
//	@Tags			outgoing-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string					true	"Outgoing Document ID (UUID)"
//	@Param			request	body		updateOutgoingDocPayload	true	"Editable fields"
//	@Success		200		{object}	util.Response{data=domain.OutgoingDocResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/outgoing-docs/{id} [put]
func (h *Handler) UpdateOutgoingDoc(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	var payload updateOutgoingDocPayload
	if err := c.Bind(&payload); err != nil {
		return util.HandleError(c, util.NewInvalidInputError("request body", "invalid JSON format"))
	}

	updaterIDStr := util.GetUserIDFromContext(c)
	updaterID, err := uuid.Parse(updaterIDStr)
	if err != nil {
		return util.HandleError(c, util.NewUnauthorizedError("invalid user ID in token"))
	}

	req := UpdateOutgoingDocRequest{
		DocNo:   payload.DocNo,
		DocName: payload.Description,
	}
	if payload.ReceiverIDs != nil {
		parsed := make([]uuid.UUID, 0, len(*payload.ReceiverIDs))
		for _, s := range *payload.ReceiverIDs {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			u, perr := uuid.Parse(s)
			if perr != nil {
				return util.HandleError(c, util.NewInvalidInputError("receiver_ids", "must contain valid UUIDs"))
			}
			parsed = append(parsed, u)
		}
		req.RecipientDeptIDs = &parsed
	}

	doc, err := h.service.UpdateOutgoingDoc(ctx, id, req, &updaterID)
	if err != nil {
		return util.HandleError(c, err)
	}
	return util.OKResponse(c, "Outgoing document updated successfully", doc)
}

// DeleteOutgoingDoc godoc
//
//	@Summary		Delete outgoing document (soft delete)
//	@Description	Soft-deletes the outgoing doc (sets deleted_at). Only allowed while the doc is still pending; the owner-approval gate acting on the doc locks it in the audit trail.
//	@Tags			outgoing-docs
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"Outgoing Document ID (UUID)"
//	@Success		200	{object}	util.Response
//	@Failure		400	{object}	util.Response
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/outgoing-docs/{id} [delete]
func (h *Handler) DeleteOutgoingDoc(c echo.Context) error {
	ctx := c.Request().Context()
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return util.HandleError(c, util.NewInvalidInputError("id", "must be a valid UUID"))
	}

	updaterIDStr := util.GetUserIDFromContext(c)
	updaterID, err := uuid.Parse(updaterIDStr)
	if err != nil {
		return util.HandleError(c, util.NewUnauthorizedError("invalid user ID in token"))
	}

	if err := h.service.DeleteOutgoingDoc(ctx, id, &updaterID); err != nil {
		return util.HandleError(c, err)
	}
	return util.OKResponse(c, "Outgoing document deleted successfully", nil)
}
