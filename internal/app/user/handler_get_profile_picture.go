package user

import (
	"e-document-backend/internal/util"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// GetProfilePicture godoc
//
//	@Summary		Get profile picture URL
//	@Description	Get a temporary presigned URL to access user's profile picture (valid for 1 hour)
//	@Tags			Users
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		307	{string}	string	"Redirects to presigned URL"
//	@Success		200	{object}	map[string]string{url=string}	"Returns presigned URL"
//	@Failure		400	{object}	util.Response
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/users/{id}/profile-picture [get]
func (h *Handler) GetProfilePicture(c echo.Context) error {
	id := c.Param("id")

	// Get user
	user, err := h.service.GetUserByID(c.Request().Context(), id)
	if err != nil {
		// Reuse standard error handling (will return util.Response)
		return util.HandleError(c, err)
	}

	// Check if user has profile picture
	if user.ProfilePicture == "" {
		return util.HandleError(c, util.ErrorResponse(
			"User does not have a profile picture",
			util.INVALID_INPUT,
			http.StatusBadRequest,
			"user does not have a profile picture",
		))
	}

	// Generate presigned URL (valid for 1 hour)
	presignedURL, err := h.storageClient.GetPresignedURL(c.Request().Context(), user.ProfilePicture, 1*time.Hour)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse(
			"Failed to generate presigned URL",
			util.INTERNAL_SERVER_ERROR,
			http.StatusInternalServerError,
			err.Error(),
		))
	}

	// Check if client wants redirect or JSON
	acceptHeader := c.Request().Header.Get("Accept")
	if acceptHeader == "application/json" {
		// Return JSON with URL
		return c.JSON(http.StatusOK, map[string]interface{}{
			"success": true,
			"message": "Profile picture URL retrieved successfully",
			"data": map[string]string{
				"url":        presignedURL,
				"expires_in": "1 hour",
			},
		})
	}

	// Default: Redirect to presigned URL (for browsers and img tags)
	return c.Redirect(http.StatusTemporaryRedirect, presignedURL)
}
