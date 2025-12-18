package user

import (
	"context"
	"e-document-backend/internal/domain"
	"e-document-backend/internal/util"
	"mime/multipart"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
)

// Handler handles HTTP requests for user operations
type Handler struct {
	service       Service
	storageClient interface {
		UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
		DeleteFile(ctx context.Context, objectPath string) error
		GetPresignedURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error)
	}
}

// NewHandler creates a new user handler
func NewHandler(service Service, storageClient interface {
	UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error)
	DeleteFile(ctx context.Context, objectPath string) error
	GetPresignedURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error)
}) *Handler {
	return &Handler{
		service:       service,
		storageClient: storageClient,
	}
}

// RegisterRoutes registers user routes
func (h *Handler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	users := e.Group("/v1/users", authMiddleware)
	users.POST("", h.CreateUser)
	users.GET("", h.GetAllUsers)
	users.GET("/:id", h.GetUserByID)
	users.PUT("/:id", h.UpdateUser)
	users.GET("/:id/profile-picture", h.GetProfilePicture)
	users.POST("/:id/profile-picture", h.UploadProfilePicture)
	users.DELETE("/:id/profile-picture", h.DeleteProfilePicture)
	users.DELETE("/:id", h.DeleteUser)
}

// CreateUser godoc
//
//	@Summary		Create a new user
//	@Description	Create a new user account with optional profile picture
//	@Tags			Users
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			username		formData	string	true	"Username"
//	@Param			email			formData	string	true	"Email"
//	@Param			password		formData	string	true	"Password (min 6 characters)"
//	@Param			first_name		formData	string	false	"First name"
//	@Param			last_name		formData	string	false	"Last name"
//	@Param			phone			formData	string	false	"Phone number (E.164 format)"
//	@Param			role			formData	string	true	"Role (Director, DepartmentManager, SectorManager, Employee)"
//	@Param			department_id	formData	string	false	"Department ID"
//	@Param			sector_id		formData	string	false	"Sector ID"
//	@Param			profile_picture	formData	file	false	"Profile picture (max 5MB, jpg/png/gif/webp)"
//	@Success		201				{object}	util.Response{data=domain.UserResponse}
//	@Failure		400				{object}	util.Response
//	@Failure		401				{object}	util.Response
//	@Router			/v1/users [post]
func (h *Handler) CreateUser(c echo.Context) error {
	// Parse form data
	req := domain.CreateUserRequest{
		Username:     c.FormValue("username"),
		Email:        c.FormValue("email"),
		Password:     c.FormValue("password"),
		FirstName:    c.FormValue("first_name"),
		LastName:     c.FormValue("last_name"),
		Phone:        c.FormValue("phone"),
		Role:         domain.UserRole(c.FormValue("role")),
		DepartmentID: c.FormValue("department_id"),
		SectorID:     c.FormValue("sector_id"),
	}

	// Validate request using validator
	if err := util.ValidateStruct(&req); err != nil {
		return util.HandleError(c, err)
	}

	// Check if profile picture is uploaded
	var profilePictureURL string
	file, err := c.FormFile("profile_picture")
	if err == nil && file != nil {
		// Validate image file
		if err := validateImageFile(file); err != nil {
			return util.HandleError(c, util.ErrorResponse("Invalid profile picture", util.INVALID_INPUT, 400, err.Error()))
		}

		// Upload to MinIO (returns object path, not full URL)
		profilePictureURL, err = h.storageClient.UploadFile(c.Request().Context(), file, "profiles")
		if err != nil {
			return util.HandleError(c, util.ErrorResponse("Failed to upload profile picture", util.INTERNAL_SERVER_ERROR, 500, err.Error()))
		}
	}

	// Create user
	user, err := h.service.CreateUser(c.Request().Context(), req)
	if err != nil {
		// If user creation fails and we uploaded a file, delete it
		if profilePictureURL != "" {
			_ = h.storageClient.DeleteFile(c.Request().Context(), profilePictureURL)
		}
		return util.HandleError(c, err)
	}

	// Update profile picture if uploaded
	if profilePictureURL != "" {
		updatedUser, err := h.service.UpdateProfilePicture(c.Request().Context(), user.ID.String(), profilePictureURL)
		if err != nil {
			// If update fails, delete the uploaded file
			_ = h.storageClient.DeleteFile(c.Request().Context(), profilePictureURL)
			// Return error since profile picture update failed
			return util.HandleError(c, err)
		}
		user = updatedUser
	}

	return util.OKResponse(c, "User created successfully", user, 201)
}

// GetAllUsers godoc
//
//	@Summary		Get all users
//	@Description	Get list of all users with pagination and search
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			page	query		int		false	"Page number"	default(1)
//	@Param			limit	query		int		false	"Items per page"	default(10)
//	@Param			search	query		string	false	"Search by username or email"
//	@Success		200		{object}	util.Response{data=util.PaginatedData}
//	@Failure		401		{object}	util.Response
//	@Failure		500		{object}	util.Response
//	@Router			/v1/users [get]
func (h *Handler) GetAllUsers(c echo.Context) error {
	// Get pagination params from query
	page := c.QueryParam("page")
	limit := c.QueryParam("limit")
	search := c.QueryParam("search")

	// Default values
	pageNum := 1
	limitNum := 10

	// Parse page
	if page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			pageNum = p
		}
	}

	// Parse limit
	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			limitNum = l
		}
	}

	// Get current user ID from JWT context
	currentUserID := ""
	if userID := c.Get("user_id"); userID != nil {
		currentUserID = userID.(string)
	}

	users, total, err := h.service.GetAllUsers(c.Request().Context(), pageNum, limitNum, search, currentUserID)
	if err != nil {
		return util.HandleError(c, err)
	}

	// Calculate pagination info
	totalPages := (total + limitNum - 1) / limitNum
	pagination := util.PaginationInfo{
		CurrentPage:  pageNum,
		TotalPages:   totalPages,
		TotalItems:   total,
		ItemsPerPage: limitNum,
	}

	return util.OKResponseWithPagination(c, "Users retrieved successfully", users, pagination)
}

// GetUserByID godoc
//
//	@Summary		Get user by ID
//	@Description	Get detailed information of a specific user
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	util.Response{data=domain.UserResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/users/{id} [get]
func (h *Handler) GetUserByID(c echo.Context) error {
	id := c.Param("id")

	user, err := h.service.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "User retrieved successfully", user)
}

// UpdateUser godoc
//
//	@Summary		Update user
//	@Description	Update user information with optional profile picture
//	@Tags			Users
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id				path		string	true	"User ID"
//	@Param			username		formData	string	false	"Username"
//	@Param			email			formData	string	false	"Email"
//	@Param			password		formData	string	false	"Password (min 6 characters)"
//	@Param			first_name		formData	string	false	"First name"
//	@Param			last_name		formData	string	false	"Last name"
//	@Param			phone			formData	string	false	"Phone number (E.164 format)"
//	@Param			role			formData	string	false	"Role (Director, DepartmentManager, SectorManager, Employee)"
//	@Param			department_id	formData	string	false	"Department ID"
//	@Param			sector_id		formData	string	false	"Sector ID"
//	@Param			profile_picture	formData	file	false	"Profile picture (max 5MB, jpg/png/gif/webp)"
//	@Success		200				{object}	util.Response{data=domain.UserResponse}
//	@Failure		400				{object}	util.Response
//	@Failure		401				{object}	util.Response
//	@Failure		404				{object}	util.Response
//	@Router			/v1/users/{id} [put]
func (h *Handler) UpdateUser(c echo.Context) error {
	id := c.Param("id")

	// Parse form data
	req := domain.UpdateUserRequest{
		Username:     c.FormValue("username"),
		Email:        c.FormValue("email"),
		Password:     c.FormValue("password"),
		FirstName:    c.FormValue("first_name"),
		LastName:     c.FormValue("last_name"),
		Phone:        c.FormValue("phone"),
		DepartmentID: c.FormValue("department_id"),
		SectorID:     c.FormValue("sector_id"),
	}

	// Parse role if provided
	if roleStr := c.FormValue("role"); roleStr != "" {
		req.Role = domain.UserRole(roleStr)
	}

	// Validate request using validator
	if err := util.ValidateStruct(&req); err != nil {
		return util.HandleError(c, err)
	}

	// Get existing user to check profile picture
	existingUser, err := h.service.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	// Check if new profile picture is uploaded
	var newProfilePictureURL string
	file, err := c.FormFile("profile_picture")
	if err == nil && file != nil {
		// Validate image file
		if err := validateImageFile(file); err != nil {
			return util.HandleError(c, util.ErrorResponse("Invalid profile picture", util.INVALID_INPUT, 400, err.Error()))
		}

		// Upload to MinIO
		newProfilePictureURL, err = h.storageClient.UploadFile(c.Request().Context(), file, "profiles")
		if err != nil {
			return util.HandleError(c, util.ErrorResponse("Failed to upload profile picture", util.INTERNAL_SERVER_ERROR, 500, err.Error()))
		}
	}

	// Update user
	user, err := h.service.UpdateUser(c.Request().Context(), id, req)
	if err != nil {
		// If user update fails and we uploaded a new file, delete it
		if newProfilePictureURL != "" {
			_ = h.storageClient.DeleteFile(c.Request().Context(), newProfilePictureURL)
		}
		return util.HandleError(c, err)
	}

	// Update profile picture if uploaded
	if newProfilePictureURL != "" {
		user, err = h.service.UpdateProfilePicture(c.Request().Context(), id, newProfilePictureURL)
		if err != nil {
			// If update fails, delete the uploaded file
			_ = h.storageClient.DeleteFile(c.Request().Context(), newProfilePictureURL)
			return util.HandleError(c, err)
		}

		// Delete old profile picture if exists and is different
		if existingUser.ProfilePicture != "" && existingUser.ProfilePicture != newProfilePictureURL {
			_ = h.storageClient.DeleteFile(c.Request().Context(), existingUser.ProfilePicture)
		}
	}

	return util.OKResponse(c, "User updated successfully", user)
}

// UploadProfilePicture godoc
//
//	@Summary		Upload profile picture
//	@Description	Upload or update user's profile picture
//	@Tags			Users
//	@Accept			multipart/form-data
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		string	true	"User ID"
//	@Param			file	formData	file	true	"Profile picture (max 5MB, jpg/png/gif/webp)"
//	@Success		200		{object}	util.Response{data=domain.UserResponse}
//	@Failure		400		{object}	util.Response
//	@Failure		401		{object}	util.Response
//	@Failure		404		{object}	util.Response
//	@Router			/v1/users/{id}/profile-picture [post]
func (h *Handler) UploadProfilePicture(c echo.Context) error {
	id := c.Param("id")

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("No file provided", util.INVALID_INPUT, 400, err.Error()))
	}

	// Validate image file
	if err := validateImageFile(file); err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid file", util.INVALID_INPUT, 400, err.Error()))
	}

	// Get existing user to check if they have an old profile picture
	existingUser, err := h.service.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	// Upload new file to MinIO
	fileURL, err := h.storageClient.UploadFile(c.Request().Context(), file, "profiles")
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Failed to upload file", util.INTERNAL_SERVER_ERROR, 500, err.Error()))
	}

	// Update user profile picture in database
	updatedUser, err := h.service.UpdateProfilePicture(c.Request().Context(), id, fileURL)
	if err != nil {
		// If database update fails, try to delete the uploaded file
		_ = h.storageClient.DeleteFile(c.Request().Context(), fileURL)
		return util.HandleError(c, err)
	}

	// Delete old profile picture if exists
	if existingUser.ProfilePicture != "" {
		_ = h.storageClient.DeleteFile(c.Request().Context(), existingUser.ProfilePicture)
	}

	return util.OKResponse(c, "Profile picture uploaded successfully", updatedUser)
}

// DeleteProfilePicture godoc
//
//	@Summary		Delete profile picture
//	@Description	Delete user's profile picture
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	util.Response{data=domain.UserResponse}
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/users/{id}/profile-picture [delete]
func (h *Handler) DeleteProfilePicture(c echo.Context) error {
	id := c.Param("id")

	// Get existing user to check if they have a profile picture
	existingUser, err := h.service.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return util.HandleError(c, err)
	}

	if existingUser.ProfilePicture == "" {
		return util.HandleError(c, util.ErrorResponse("No profile picture to delete", util.INVALID_INPUT, 400, "user does not have a profile picture"))
	}

	// Delete file from MinIO
	if err := h.storageClient.DeleteFile(c.Request().Context(), existingUser.ProfilePicture); err != nil {
		return util.HandleError(c, util.ErrorResponse("Failed to delete file", util.INTERNAL_SERVER_ERROR, 500, err.Error()))
	}

	// Update user profile picture in database (set to empty)
	updatedUser, err := h.service.UpdateProfilePicture(c.Request().Context(), id, "")
	if err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "Profile picture deleted successfully", updatedUser)
}

// DeleteUser godoc
//
//	@Summary		Delete user
//	@Description	Delete a user account
//	@Tags			Users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id	path		string	true	"User ID"
//	@Success		200	{object}	util.Response
//	@Failure		401	{object}	util.Response
//	@Failure		404	{object}	util.Response
//	@Router			/v1/users/{id} [delete]
func (h *Handler) DeleteUser(c echo.Context) error {
	id := c.Param("id")

	if err := h.service.DeleteUser(c.Request().Context(), id); err != nil {
		return util.HandleError(c, err)
	}

	return util.OKResponse(c, "User deleted successfully", nil)
}

// Helper function to validate image files
func validateImageFile(file *multipart.FileHeader) error {
	// Check file size (max 5MB)
	maxSize := int64(5 * 1024 * 1024) // 5MB
	if file.Size > maxSize {
		return util.ErrorResponse("File size exceeds 5MB limit", util.INVALID_INPUT, 400, "")
	}

	// Check MIME type
	contentType := file.Header.Get("Content-Type")
	validMimeTypes := map[string]bool{
		"image/jpeg": true,
		"image/jpg":  true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}

	if !validMimeTypes[contentType] {
		return util.ErrorResponse("Invalid file type. Allowed: jpg, jpeg, png, gif, webp", util.INVALID_INPUT, 400, "")
	}

	return nil
}
