package storage

import (
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/rs/zerolog/log"
)

// MinIOConfig holds MinIO configuration
type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	PublicURL string
}

// MinIOClient handles file operations with MinIO
type MinIOClient struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

// NewMinIOClient creates a new MinIO client
func NewMinIOClient(config MinIOConfig) (*MinIOClient, error) {
	// Initialize MinIO client
	minioClient, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKey, config.SecretKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize MinIO client: %w", err)
	}

	// Check if bucket exists, create if not
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, config.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, config.Bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create bucket: %w", err)
		}
		log.Info().Str("bucket", config.Bucket).Msg("Bucket created successfully")
	}

	return &MinIOClient{
		client:    minioClient,
		bucket:    config.Bucket,
		publicURL: config.PublicURL,
	}, nil
}

// LoadConfigFromEnv loads MinIO configuration from environment variables
func LoadConfigFromEnv() MinIOConfig {
	useSSL := false
	if os.Getenv("MINIO_USE_SSL") == "true" {
		useSSL = true
	}

	return MinIOConfig{
		Endpoint:  os.Getenv("MINIO_ENDPOINT"),
		AccessKey: os.Getenv("MINIO_ACCESS_KEY"),
		SecretKey: os.Getenv("MINIO_SECRET_KEY"),
		Bucket:    os.Getenv("MINIO_BUCKET"),
		UseSSL:    useSSL,
		PublicURL: os.Getenv("MINIO_PUBLIC_URL"),
	}
}

// UploadFile uploads a file to MinIO and returns the object path (not full URL)
func (m *MinIOClient) UploadFile(ctx context.Context, file *multipart.FileHeader, folder string) (string, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s/%d_%s%s", folder, time.Now().Unix(), strings.ReplaceAll(file.Filename, " ", "_"), "")
	if ext != "" {
		filename = fmt.Sprintf("%s/%d_%s", folder, time.Now().Unix(), strings.ReplaceAll(file.Filename, " ", "_"))
	}

	// Detect content type
	contentType := file.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to MinIO
	_, err = m.client.PutObject(ctx, m.bucket, filename, src, file.Size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return only the object path (not full URL)
	return filename, nil
}

// UploadFileFromReader uploads a file from an io.Reader to MinIO
func (m *MinIOClient) UploadFileFromReader(ctx context.Context, reader io.Reader, filename string, size int64, contentType string, folder string) (string, error) {
	// Generate unique filename with folder
	ext := filepath.Ext(filename)
	uniqueFilename := fmt.Sprintf("%s/%d_%s", folder, time.Now().Unix(), strings.ReplaceAll(filename, " ", "_"))
	if ext == "" {
		uniqueFilename = fmt.Sprintf("%s/%d_%s", folder, time.Now().Unix(), strings.ReplaceAll(filename, " ", "_"))
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Upload to MinIO
	_, err := m.client.PutObject(ctx, m.bucket, uniqueFilename, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	// Return the file URL
	fileURL := fmt.Sprintf("%s/%s/%s", m.publicURL, m.bucket, uniqueFilename)
	return fileURL, nil
}

// DeleteFile deletes a file from MinIO using object path
func (m *MinIOClient) DeleteFile(ctx context.Context, objectPath string) error {
	if objectPath == "" {
		return fmt.Errorf("empty object path")
	}

	err := m.client.RemoveObject(ctx, m.bucket, objectPath, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

// GetFile retrieves a file from MinIO using object path
func (m *MinIOClient) GetFile(ctx context.Context, objectPath string) (*minio.Object, error) {
	if objectPath == "" {
		return nil, fmt.Errorf("empty object path")
	}

	object, err := m.client.GetObject(ctx, m.bucket, objectPath, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get file: %w", err)
	}

	return object, nil
}

// GetPresignedURL generates a presigned URL for temporary file access
func (m *MinIOClient) GetPresignedURL(ctx context.Context, objectPath string, expiry time.Duration) (string, error) {
	if objectPath == "" {
		return "", fmt.Errorf("empty object path")
	}

	// Generate presigned URL with expiry time
	presignedURL, err := m.client.PresignedGetObject(ctx, m.bucket, objectPath, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// ValidateImageFile checks if the uploaded file is a valid image
func ValidateImageFile(file *multipart.FileHeader) error {
	// Check file size (max 5MB)
	maxSize := int64(5 * 1024 * 1024) // 5MB
	if file.Size > maxSize {
		return fmt.Errorf("file size exceeds 5MB limit")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(file.Filename))
	validExtensions := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".webp": true,
	}

	if !validExtensions[ext] {
		return fmt.Errorf("invalid file type. Allowed: jpg, jpeg, png, gif, webp")
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
		return fmt.Errorf("invalid content type. Must be an image")
	}

	return nil
}
