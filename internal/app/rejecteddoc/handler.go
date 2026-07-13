package rejecteddoc

import (
	"e-document-backend/internal/util"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// Handler handles rejected-document report HTTP requests.
type Handler struct {
	service Service
}

// NewHandler creates a new rejected-document handler.
func NewHandler(service Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the rejected-document routes.
func (h *Handler) RegisterRoutes(e *echo.Group, middleware ...echo.MiddlewareFunc) {
	docs := e.Group("/v1/rejected-docs", middleware...)
	docs.GET("", h.GetRejectedDocs)
}

// GetRejectedDocs godoc
//
//	@Summary		Get rejected documents report
//	@Description	Report of rejected documents (inbound + outbound). Admin and Secretary roles can filter by any department via dept_id; other roles are scoped to their own department.
//	@Tags			rejected-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		false	"Page number (default 1)"
//	@Param			limit		query		int		false	"Items per page (default 10)"
//	@Param			dept_id		query		string	false	"Department ID (Admin/Secretary only; ignored for other roles)"
//	@Param			source		query		string	false	"Filter by source: inbound | outbound"
//	@Param			start_date	query		string	false	"Start date (YYYY-MM-DD)"
//	@Param			end_date	query		string	false	"End date (YYYY-MM-DD)"
//	@Success		200			{object}	util.Response{data=[]domain.RejectedDocResponse}
//	@Failure		401			{object}	util.Response
//	@Failure		500			{object}	util.Response
//	@Router			/v1/rejected-docs [get]
func (h *Handler) GetRejectedDocs(c echo.Context) error {
	ctx := c.Request().Context()

	roleName, _ := c.Get("role_name").(string)
	canFilterAnyDept := strings.EqualFold(roleName, "Admin") ||
		strings.EqualFold(roleName, "Secretary")

	// Department scoping:
	// - Admin / Secretary: dept_id query param is an optional filter (empty = all departments).
	// - Others: always scoped to the caller's own department from the JWT.
	var deptID *uuid.UUID
	if canFilterAnyDept {
		if v := strings.TrimSpace(c.QueryParam("dept_id")); v != "" {
			parsed, err := uuid.Parse(v)
			if err != nil {
				return util.HandleError(c, util.NewValidationError("invalid dept_id"))
			}
			deptID = &parsed
		}
	} else {
		deptIDStr, ok := c.Get("dept_id").(string)
		if !ok || deptIDStr == "" {
			return util.HandleError(c, util.NewUnauthorizedError("department not assigned to user"))
		}
		parsed, err := uuid.Parse(deptIDStr)
		if err != nil {
			return util.HandleError(c, util.NewUnauthorizedError("invalid department in token"))
		}
		deptID = &parsed
	}

	source := strings.ToLower(strings.TrimSpace(c.QueryParam("source")))
	if source != "" && source != "inbound" && source != "outbound" {
		return util.HandleError(c, util.NewValidationError("source must be 'inbound' or 'outbound'"))
	}

	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page <= 0 {
		page = 1
	}
	limit, _ := strconv.Atoi(c.QueryParam("limit"))
	if limit <= 0 {
		limit = 10
	}

	var start, end *time.Time
	if v := c.QueryParam("start_date"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02", v, time.Local); err == nil {
			start = &t
		}
	}
	if v := c.QueryParam("end_date"); v != "" {
		if t, err := time.ParseInLocation("2006-01-02", v, time.Local); err == nil {
			t = t.Add(24 * time.Hour)
			end = &t
		}
	}

	docs, total, err := h.service.GetRejectedDocs(ctx, deptID, source, start, end, page, limit)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponseWithPagination(c, "Rejected documents retrieved successfully", docs, util.PaginationInfo{
		CurrentPage:  page,
		TotalPages:   (total + limit - 1) / limit,
		TotalItems:   total,
		ItemsPerPage: limit,
	})
}
