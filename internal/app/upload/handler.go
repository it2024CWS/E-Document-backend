package upload

import (
	"archive/zip"
	"context"
	"e-document-backend/internal/util"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
	"github.com/tus/tusd/v2/pkg/filelocker"
	tusd "github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/s3store"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	awscreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Handler handles HTTP requests for file upload operations
type Handler struct {
	service     Service
	tusHandler  *tusd.UnroutedHandler
	tusConfig   TusConfig
	bucket      string
	minioClient *minio.Client
}

// TusConfig holds tusd configuration
type TusConfig struct {
	BasePath    string
	S3Endpoint  string
	S3AccessKey string
	S3SecretKey string
	S3Bucket    string
	S3UseSSL    bool
	StorageDir  string // Local storage directory for file locker
}

// LoadTusConfigFromEnv loads tusd configuration from environment variables
func LoadTusConfigFromEnv() TusConfig {
	return TusConfig{
		BasePath:    getEnvWithDefault("TUSD_BASE_PATH", "/api/v1/upload"),
		S3Endpoint:  os.Getenv("MINIO_ENDPOINT"),
		S3AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		S3SecretKey: os.Getenv("MINIO_SECRET_KEY"),
		S3Bucket:    os.Getenv("MINIO_BUCKET"),
		S3UseSSL:    os.Getenv("MINIO_USE_SSL") == "true",
		StorageDir:  getEnvWithDefault("TUSD_STORAGE_DIR", "./tmp/tusd"),
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// NewHandler creates a new upload handler with tusd integration
func NewHandler(service Service, tusConfig TusConfig) (*Handler, error) {
	h := &Handler{
		service:   service,
		tusConfig: tusConfig,
		bucket:    tusConfig.S3Bucket,
	}

	// Initialize MinIO client
	minioClient, err := minio.New(tusConfig.S3Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(tusConfig.S3AccessKey, tusConfig.S3SecretKey, ""),
		Secure: tusConfig.S3UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create minio client: %w", err)
	}
	h.minioClient = minioClient

	// Initialize tusd handler
	if err := h.initTusHandler(); err != nil {
		return nil, fmt.Errorf("failed to initialize tusd handler: %w", err)
	}

	return h, nil
}

// initTusHandler initializes the tusd handler with S3 store
func (h *Handler) initTusHandler() error {
	// Create AWS config for MinIO
	protocol := "http"
	if h.tusConfig.S3UseSSL {
		protocol = "https"
	}
	endpointURL := fmt.Sprintf("%s://%s", protocol, h.tusConfig.S3Endpoint)

	// Load AWS config with custom endpoint resolver for MinIO
	cfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithRegion("us-east-1"), // MinIO doesn't care about region
		config.WithCredentialsProvider(awscreds.NewStaticCredentialsProvider(
			h.tusConfig.S3AccessKey,
			h.tusConfig.S3SecretKey,
			"",
		)),
	)
	if err != nil {
		return fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint for MinIO
	s3Client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpointURL)
		o.UsePathStyle = true // Required for MinIO
	})

	// Create S3 store for tusd
	store := s3store.New(h.tusConfig.S3Bucket, s3Client)

	// Create storage directory for file locker if it doesn't exist
	if err := os.MkdirAll(h.tusConfig.StorageDir, 0755); err != nil {
		return fmt.Errorf("failed to create storage directory: %w", err)
	}

	// Create file locker for concurrent upload handling
	locker := filelocker.New(h.tusConfig.StorageDir)

	// Create tusd unrouted handler (for custom routing with Echo)
	composer := tusd.NewStoreComposer()
	store.UseIn(composer)
	locker.UseIn(composer)

	tusHandler, err := tusd.NewUnroutedHandler(tusd.Config{
		StoreComposer:           composer,
		NotifyCompleteUploads:   true,
		RespectForwardedHeaders: true,
	})
	if err != nil {
		return fmt.Errorf("failed to create tusd unrouted handler: %w", err)
	}

	h.tusHandler = tusHandler

	// Start goroutine to handle completed uploads
	go h.handleCompleteUploads()

	log.Info().
		Str("base_path", h.tusConfig.BasePath).
		Str("bucket", h.tusConfig.S3Bucket).
		Msg("tusd handler initialized successfully")

	return nil
}

// handleCompleteUploads processes completed uploads
func (h *Handler) handleCompleteUploads() {
	log.Info().Msg("Starting to listen for completed uploads...")
	for {
		log.Debug().Msg("Waiting for upload completion event...")
		event := <-h.tusHandler.CompleteUploads
		log.Info().
			Str("upload_id", event.Upload.ID).
			Int64("size", event.Upload.Size).
			Msg("Received upload completion event")
		go h.processCompletedUpload(event)
	}
}

// processCompletedUpload handles the post-upload logic
func (h *Handler) processCompletedUpload(event tusd.HookEvent) {
	ctx := context.Background()
	upload := event.Upload

	log.Info().
		Str("upload_id", upload.ID).
		Int64("size", upload.Size).
		Interface("metadata", upload.MetaData).
		Msg("Processing completed upload")

	// Extract metadata
	relativePath := upload.MetaData["relative_path"]
	ownerIDStr := upload.MetaData["owner_id"]
	parentFolderIDStr := upload.MetaData["parent_folder_id"]
	fileType := upload.MetaData["file_type"]
	fileName := upload.MetaData["filename"]

	// Validate required metadata
	if ownerIDStr == "" {
		log.Error().Str("upload_id", upload.ID).Msg("Missing owner_id in metadata")
		return
	}

	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		log.Error().Err(err).Str("owner_id", ownerIDStr).Msg("Invalid owner_id format")
		return
	}

	// Parse parent_folder_id if provided
	var parentFolderID *uuid.UUID
	if parentFolderIDStr != "" {
		if parsed, err := uuid.Parse(parentFolderIDStr); err == nil {
			parentFolderID = &parsed
		}
	}

	// Use relative_path if provided, otherwise use filename
	if relativePath == "" && fileName != "" {
		relativePath = fileName
	}

	if relativePath == "" {
		log.Error().Str("upload_id", upload.ID).Msg("Missing relative_path and filename in metadata")
		return
	}

	// Build the MinIO file path (the actual location in S3)
	// tusd stores files with the upload ID as the object key
	filePath := upload.ID

	// If storage has prefix info, use it
	if upload.Storage != nil {
		if key, ok := upload.Storage["Key"]; ok {
			filePath = key
		}
	}

	// Process the upload
	params := ProcessUploadParams{
		RelativePath:   relativePath,
		ParentFolderID: parentFolderID,
		OwnerID:        ownerID,
		FilePath:       filePath,
		FileSize:       upload.Size,
		FileType:       fileType,
		UploadID:       upload.ID,
	}

	result, err := h.service.ProcessUploadComplete(ctx, params)
	if err != nil {
		log.Error().Err(err).
			Str("upload_id", upload.ID).
			Str("relative_path", relativePath).
			Msg("Failed to process upload")
		return
	}

	log.Info().
		Str("upload_id", upload.ID).
		Str("document_id", result.Document.ID.String()).
		Str("attachment_id", result.Attachment.ID.String()).
		Int("folders_created", len(result.Folders)).
		Msg("Upload processed successfully")
}

// locationFixerWriter wraps http.ResponseWriter to fix Location header
type locationFixerWriter struct {
	http.ResponseWriter
	req *http.Request
}

func (w *locationFixerWriter) WriteHeader(statusCode int) {
	// Fix Location header for 201 Created responses
	if statusCode == 201 && w.req.Method == "POST" {
		location := w.Header().Get("Location")
		log.Info().
			Str("original_location", location).
			Msg("locationFixerWriter: Original Location header")

		if location != "" {
			var uploadID string

			// Extract upload ID from location (could be absolute or relative URL)
			if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
				// Absolute URL - malformed like: http://localhost:5000uploadid
				// Find ":5000" and extract everything after it
				idx := strings.Index(location, ":5000")
				if idx >= 0 {
					// Everything after ":5000" is the upload ID
					uploadID = location[idx+5:] // +5 to skip ":5000"
				} else {
					// Try normal URL parsing
					if u, err := url.Parse(location); err == nil {
						uploadID = strings.TrimPrefix(u.Path, "/")
					}
				}
			} else {
				// Relative URL
				uploadID = strings.TrimPrefix(location, "/")
			}

			// Set corrected Location header (always relative)
			if uploadID != "" {
				fixedLocation := "/api/v1/upload/files/" + uploadID
				w.Header().Set("Location", fixedLocation)
				log.Info().
					Str("upload_id", uploadID).
					Str("fixed_location", fixedLocation).
					Msg("locationFixerWriter: Fixed Location header")
			}
		}
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

// RegisterRoutes registers upload routes with tusd handler
func (h *Handler) RegisterRoutes(e *echo.Group, authMiddleware echo.MiddlewareFunc) {
	// Create upload group WITH auth middleware
	upload := e.Group("/v1/upload", authMiddleware)

	// Middleware to inject owner_id from JWT (extracted by authMiddleware)
	injectOwnerID := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if c.Request().Method == "POST" {
				// Get user_id from context (set by authMiddleware)
				userID, ok := c.Get("user_id").(string)
				if !ok || userID == "" {
					return echo.NewHTTPError(401, "Unauthorized: user_id not found in context")
				}

				// Inject owner_id into Upload-Metadata header
				metadata := c.Request().Header.Get("Upload-Metadata")
				if !strings.Contains(metadata, "owner_id") {
					ownerIDEncoded := base64.StdEncoding.EncodeToString([]byte(userID))
					if metadata != "" {
						metadata += ", "
					}
					metadata += "owner_id " + ownerIDEncoded
					c.Request().Header.Set("Upload-Metadata", metadata)
				}
				log.Debug().
					Str("user_id", userID).
					Str("metadata", metadata).
					Msg("Injected owner_id into upload metadata")
			}
			return next(c)
		}
	}

	// Wrapper to fix Location header using custom ResponseWriter
	wrapWithLocationFixer := func(handler http.Handler) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Wrap the response writer
			wrapped := &locationFixerWriter{
				ResponseWriter: c.Response().Writer,
				req:            c.Request(),
			}
			c.Response().Writer = wrapped

			// Call the handler
			handler.ServeHTTP(c.Response(), c.Request())
			return nil
		}
	}

	// Wrapper to fix request path for tusd handlers (PATCH, HEAD, DELETE)
	// Echo provides path param :id, but tusd expects path to be /uploadid
	wrapTusHandler := func(handler http.Handler) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get upload ID from path parameter
			uploadID := c.Param("id")
			if uploadID != "" {
				// Rewrite request URL path to what tusd expects: /uploadid
				c.Request().URL.Path = "/" + uploadID
				log.Debug().
					Str("original_path", c.Path()).
					Str("rewritten_path", c.Request().URL.Path).
					Str("upload_id", uploadID).
					Msg("Rewrote request path for tusd handler")
			}
			handler.ServeHTTP(c.Response(), c.Request())
			return nil
		}
	}

	// Map individual TUS protocol methods to unrouted handler
	// POST /files - Create new upload
	upload.POST("/files", wrapWithLocationFixer(http.HandlerFunc(h.tusHandler.PostFile)), injectOwnerID)
	// POST /files/ - Also handle with trailing slash
	upload.POST("/files/", wrapWithLocationFixer(http.HandlerFunc(h.tusHandler.PostFile)), injectOwnerID)
	// HEAD /files/:id - Get upload status
	upload.HEAD("/files/:id", wrapTusHandler(http.HandlerFunc(h.tusHandler.HeadFile)))
	// PATCH /files/:id - Upload file chunk
	upload.PATCH("/files/:id", wrapTusHandler(http.HandlerFunc(h.tusHandler.PatchFile)))
	// DELETE /files/:id - Terminate upload
	upload.DELETE("/files/:id", wrapTusHandler(http.HandlerFunc(h.tusHandler.DelFile)))

	// Info endpoint
	upload.GET("/info", h.GetUploadInfo)

	// Download endpoint
	upload.GET("/download/:id", h.DownloadFile)

	// Download folder as ZIP endpoint
	upload.GET("/download/folder/:id", h.DownloadFolder)
}

// UploadInfoResponse represents the response for upload info endpoint
type UploadInfoResponse struct {
	TusVersion string   `json:"tus_version" example:"1.0.0"`
	MaxSize    int64    `json:"max_size" example:"0"`
	Extensions []string `json:"extensions" example:"creation,creation-defer-length,termination,concatenation"`
	UploadPath string   `json:"upload_path" example:"/api/v1/upload/files"`
}

// GetUploadInfo godoc
// @Summary		Get upload service info
// @Description	Returns information about the TUS upload service including supported version and extensions
// @Tags		Upload
// @Produce		json
// @Security	BearerAuth
// @Success		200		{object}	util.Response{data=UploadInfoResponse}
// @Failure		401		{object}	util.Response
// @Router		/v1/upload/info [get]
func (h *Handler) GetUploadInfo(c echo.Context) error {
	return util.OKResponse(c, "Upload service info", UploadInfoResponse{
		TusVersion: "1.0.0",
		MaxSize:    0, // No limit
		Extensions: []string{"creation", "creation-defer-length", "termination", "concatenation"},
		UploadPath: h.tusConfig.BasePath + "/files",
	})
}

// DownloadFile godoc
// @Summary		Download a file
// @Description	Downloads a file by attachment ID with original filename
// @Tags		Upload
// @Produce		application/octet-stream
// @Security	BearerAuth
// @Param		id	path		string	true	"Attachment ID"
// @Success		200	{file}		binary
// @Failure		400	{object}	util.Response
// @Failure		404	{object}	util.Response
// @Failure		500	{object}	util.Response
// @Router		/v1/upload/download/{id} [get]
func (h *Handler) DownloadFile(c echo.Context) error {
	// Get attachment ID from URL parameter
	attachmentIDStr := c.Param("id")
	attachmentID, err := uuid.Parse(attachmentIDStr)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid attachment ID", util.INVALID_INPUT, 400, "The provided attachment ID is not a valid UUID"))
	}

	// Get attachment details from database
	attachment, err := h.service.GetAttachment(c.Request().Context(), attachmentID)
	if err != nil {
		log.Error().Err(err).Str("attachment_id", attachmentIDStr).Msg("Failed to get attachment")
		return util.HandleError(c, util.ErrorResponse("Attachment not found", util.VALIDATION_ERROR, 404, fmt.Sprintf("No attachment found with ID: %s", attachmentIDStr)))
	}

	// Download file from MinIO using file_path (upload ID)
	object, err := h.minioClient.GetObject(
		c.Request().Context(),
		h.bucket,
		attachment.FilePath, // This is the upload ID
		minio.GetObjectOptions{},
	)
	if err != nil {
		log.Error().Err(err).
			Str("attachment_id", attachmentIDStr).
			Str("file_path", attachment.FilePath).
			Msg("Failed to get object from MinIO")
		return util.HandleError(c, util.ErrorResponse("Failed to download file", util.INTERNAL_SERVER_ERROR, 500, "Could not retrieve file from storage"))
	}
	defer object.Close()

	// Get object info to verify it exists
	stat, err := object.Stat()
	if err != nil {
		log.Error().Err(err).
			Str("attachment_id", attachmentIDStr).
			Str("file_path", attachment.FilePath).
			Msg("Failed to get object stat from MinIO")
		return util.HandleError(c, util.ErrorResponse("File not found in storage", util.VALIDATION_ERROR, 404, "The file exists in database but not found in storage"))
	}

	// Set response headers with original filename
	c.Response().Header().Set("Content-Type", attachment.FileType)
	c.Response().Header().Set("Content-Disposition", encodeFilename(attachment.FileName))
	c.Response().Header().Set("Content-Length", fmt.Sprintf("%d", stat.Size))

	// Stream the file to client
	return c.Stream(200, attachment.FileType, object)
}

// DownloadFolder godoc
// @Summary		Download a folder as ZIP
// @Description	Downloads all files in a folder (including subfolders) as a ZIP archive
// @Tags		Upload
// @Produce		application/zip
// @Security	BearerAuth
// @Param		id	path		string	true	"Folder ID"
// @Success		200	{file}		binary
// @Failure		400	{object}	util.Response
// @Failure		404	{object}	util.Response
// @Failure		500	{object}	util.Response
// @Router		/v1/upload/download/folder/{id} [get]
func (h *Handler) DownloadFolder(c echo.Context) error {
	// Get folder ID from URL parameter
	folderIDStr := c.Param("id")
	folderID, err := uuid.Parse(folderIDStr)
	if err != nil {
		return util.HandleError(c, util.ErrorResponse("Invalid folder ID", util.INVALID_INPUT, 400, "The provided folder ID is not a valid UUID"))
	}

	// Get folder details to use the folder name
	folder, err := h.service.GetFolder(c.Request().Context(), folderID)
	if err != nil {
		log.Error().Err(err).Str("folder_id", folderIDStr).Msg("Failed to get folder details")
		return util.HandleError(c, util.ErrorResponse("Folder not found", util.VALIDATION_ERROR, 404, fmt.Sprintf("No folder found with ID: %s", folderIDStr)))
	}

	// Get all attachments in the folder (recursively)
	attachments, err := h.service.GetFolderAttachments(c.Request().Context(), folderID)
	if err != nil {
		log.Error().Err(err).Str("folder_id", folderIDStr).Msg("Failed to get folder attachments")
		return util.HandleError(c, util.ErrorResponse("Failed to get folder contents", util.INTERNAL_SERVER_ERROR, 500, "Could not retrieve folder contents"))
	}

	if len(attachments) == 0 {
		return util.HandleError(c, util.ErrorResponse("Empty folder", util.VALIDATION_ERROR, 404, "No files found in this folder"))
	}

	// Set response headers for ZIP download using folder name
	c.Response().Header().Set("Content-Type", "application/zip")
	c.Response().Header().Set("Content-Disposition", encodeFilename(folder.Name+".zip"))
	c.Response().WriteHeader(200)

	// Create ZIP writer that writes directly to response
	zipWriter := zip.NewWriter(c.Response().Writer)
	defer zipWriter.Close()

	// Track added files to avoid duplicates
	addedFiles := make(map[string]bool)

	// Add each file to the ZIP
	for _, attachment := range attachments {
		// Skip if already added (shouldn't happen, but just in case)
		if addedFiles[attachment.FileName] {
			log.Warn().Str("filename", attachment.FileName).Msg("Duplicate filename detected, skipping")
			continue
		}

		// Download file from MinIO
		object, err := h.minioClient.GetObject(
			c.Request().Context(),
			h.bucket,
			attachment.FilePath,
			minio.GetObjectOptions{},
		)
		if err != nil {
			log.Error().Err(err).
				Str("file_path", attachment.FilePath).
				Str("filename", attachment.FileName).
				Msg("Failed to get object from MinIO, skipping file")
			continue // Skip this file and continue with others
		}

		// Create file in ZIP
		writer, err := zipWriter.Create(attachment.FileName)
		if err != nil {
			log.Error().Err(err).Str("filename", attachment.FileName).Msg("Failed to create file in ZIP")
			object.Close()
			continue
		}

		// Copy file content to ZIP
		_, err = io.Copy(writer, object)
		object.Close()

		if err != nil {
			log.Error().Err(err).Str("filename", attachment.FileName).Msg("Failed to copy file to ZIP")
			continue
		}

		addedFiles[attachment.FileName] = true
		log.Debug().Str("filename", attachment.FileName).Msg("Added file to ZIP")
	}

	log.Info().
		Str("folder_id", folderIDStr).
		Int("files_count", len(addedFiles)).
		Msg("Folder download completed")

	return nil
}

// PreCreateMiddleware is called before creating an upload
// Can be used to validate metadata and inject owner_id from JWT
func (h *Handler) PreCreateMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Get user ID from context (set by auth middleware)
		userID := c.Get("user_id")
		if userID != nil {
			// Add owner_id to the Upload-Metadata header if not present
			metadata := c.Request().Header.Get("Upload-Metadata")
			if !strings.Contains(metadata, "owner_id") {
				ownerIDEncoded := base64.StdEncoding.EncodeToString([]byte(userID.(string)))
				if metadata != "" {
					metadata += ", "
				}
				metadata += "owner_id " + ownerIDEncoded
				c.Request().Header.Set("Upload-Metadata", metadata)
			}
		}
		return next(c)
	}
}

// GetFileTypeFromPath returns the file type based on extension
func GetFileTypeFromPath(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	mimeTypes := map[string]string{
		".pdf":  "application/pdf",
		".doc":  "application/msword",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xls":  "application/vnd.ms-excel",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".ppt":  "application/vnd.ms-powerpoint",
		".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",
		".txt":  "text/plain",
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".mp4":  "video/mp4",
		".mp3":  "audio/mpeg",
		".zip":  "application/zip",
		".rar":  "application/x-rar-compressed",
	}

	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}
	return "application/octet-stream"
}

// encodeFilename encodes a filename for Content-Disposition header using RFC 5987
// This ensures proper support for Unicode characters (Thai, Lao, Chinese, etc.)
func encodeFilename(filename string) string {
	// Use RFC 5987 encoding: filename*=UTF-8''encoded_filename
	// This properly handles Unicode characters
	return fmt.Sprintf(`attachment; filename*=UTF-8''%s`, url.PathEscape(filename))
}
