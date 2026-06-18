package rejecteddoc

import (
	"e-document-backend/internal/util"
	"strconv"
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
//	@Description	Report of rejected documents for the caller's own department (inbound + outbound), with optional date-range filters.
//	@Tags			rejected-docs
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page		query		int		false	"Page number (default 1)"
//	@Param			limit		query		int		false	"Items per page (default 10)"
//	@Param			start_date	query		string	false	"Start date (YYYY-MM-DD)"
//	@Param			end_date	query		string	false	"End date (YYYY-MM-DD)"
//	@Success		200			{object}	util.Response{data=[]domain.RejectedDocResponse}
//	@Failure		401			{object}	util.Response
//	@Failure		500			{object}	util.Response
//	@Router			/v1/rejected-docs [get]
func (h *Handler) GetRejectedDocs(c echo.Context) error {
	ctx := c.Request().Context()

	// Scope to the caller's own department
	deptIDStr, ok := c.Get("dept_id").(string)
	if !ok || deptIDStr == "" {
		return util.HandleError(c, util.NewUnauthorizedError("department not assigned to user"))
	}
	parsed, err := uuid.Parse(deptIDStr)
	if err != nil {
		return util.HandleError(c, util.NewUnauthorizedError("invalid department in token"))
	}
	deptID := &parsed

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

	docs, total, err := h.service.GetRejectedDocs(ctx, deptID, start, end, page, limit)
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
