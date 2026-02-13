package util

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SaveFile saves a multipart file to the specified directory with a unique name
// Returns the relative path to the saved file
func SaveFile(file *multipart.FileHeader, subDir string) (string, error) {
	// Open the file
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("failed to open file: %w", err)
	}
	defer src.Close()

	// Create upload directory if it doesn't exist
	// Base upload dir could be "uploads" or configured via env
	baseDir := "uploads"
	uploadDir := filepath.Join(baseDir, subDir)
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Generate unique filename
	// Original filename: example.pdf
	// New filename: uuid-timestamp.pdf
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		// Try to guess from content type or just default?
		// For now let's keep it simple.
	}

	uniqueID := uuid.New().String()
	timestamp := time.Now().Unix()
	newFilename := fmt.Sprintf("%s-%d%s", uniqueID, timestamp, ext)

	dstPath := filepath.Join(uploadDir, newFilename)

	// Create destination file
	dst, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("failed to copy file content: %w", err)
	}

	// Return relative path (replace backslashes with forward slashes for URL compatibility if needed,
	// but strictly for storage path usually OS separator is fine.
	// However, if saving to DB to be served via API, forward slashes are better).
	relativePath := filepath.Join(subDir, newFilename)
	return strings.ReplaceAll(relativePath, "\\", "/"), nil
}
