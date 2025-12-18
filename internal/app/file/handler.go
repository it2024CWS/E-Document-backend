package file

import (
	"e-document-backend/internal/util"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for file operations (presigned URLs for download/view)
type Handler struct {
	service Service
}

// NewHandler creates a new file handler
func NewHandler(service Service) *Handler {
	return &Handler{
		service: service,
	}
}

// RegisterRoutes registers file routes
func (h *Handler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	files := e.Group("/v1/files", authMiddleware)

	// Generate presigned URL by object path (key stored in DB)
	files.GET("/presign", h.GetPresignedURL)
}

// GetPresignedURLRequest represents query params for presign endpoint
type GetPresignedURLRequest struct {
	ObjectPath string `query:"object_path" validate:"required"`
	Expiry     int64  `query:"expiry"` // seconds, optional (default 3600)
}

// GetPresignedURLResponse represents response for presign endpoint
type GetPresignedURLResponse struct {
	URL       string `json:"url"`
	ExpiresIn int64  `json:"expires_in"` // seconds
}

// GetPresignedURL godoc
//
//	@Summary		Generate presigned URL for file
//	@Description	Generate a temporary presigned URL from MinIO for downloading or viewing a file by its object path (key).
//	@Tags			Files
//	@Produce		json
//	@Security		BearerAuth
//	@Param			object_path	query		string	true	"Object path (key) in MinIO bucket (e.g. documents/12345_file.pdf)"
//	@Param			expiry		query		int		false	"Expiry time in seconds (default: 3600)"
//	@Success		200			{object}	util.Response{data=GetPresignedURLResponse}
//	@Failure		400			{object}	util.Response
//	@Failure		401			{object}	util.Response
//	@Failure		500			{object}	util.Response
//	@Router			/v1/files/presign [get]
func (h *Handler) GetPresignedURL(c echo.Context) error {
	var req GetPresignedURLRequest

	if err := c.Bind(&req); err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid query parameters", util.INVALID_INPUT, http.StatusBadRequest, err.Error()))
	}

	if req.ObjectPath == "" {
		return util.HandleError(c, util.ErrorResponse("Validation failed", util.MISSING_REQUIRED_FIELD, http.StatusBadRequest, "object_path is required"))
	}

	url, expirySeconds, err := h.service.GeneratePresignedURL(c.Request().Context(), req.ObjectPath, req.Expiry)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Failed to generate presigned URL", util.INTERNAL_SERVER_ERROR, http.StatusInternalServerError, err.Error()))
	}

	resp := GetPresignedURLResponse{
		URL:       url,
		ExpiresIn: expirySeconds,
	}

	return util.OKResponse(c, "Presigned URL generated successfully", resp)
}
