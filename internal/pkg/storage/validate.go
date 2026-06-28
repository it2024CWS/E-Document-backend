package storage

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

// allowedDocumentExt is the whitelist of extensions accepted by the document
// upload pipeline (single file, folder upload, version replacement).
// Profile-picture uploads use the stricter ValidateImageFile path.
var allowedDocumentExt = map[string]bool{
	// Images
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".gif":  true,
	".webp": true,
	".bmp":  true,
	".svg":  true,
	// Office documents
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".ppt":  true,
	".pptx": true,
	".txt":  true,
	".csv":  true,
	// Archives
	".zip": true,
	".rar": true,
}

// ValidateDocumentExtension returns an error if filename's extension is not in
// the document whitelist. Comparison is case-insensitive.
func ValidateDocumentExtension(filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return fmt.Errorf("file %q has no extension", filename)
	}
	if !allowedDocumentExt[ext] {
		return fmt.Errorf("unsupported file type %q", ext)
	}
	return nil
}

// AllowedDocumentExtensions returns the whitelist as a sorted slice — useful
// for surfacing the allowed list in error messages or API metadata.
func AllowedDocumentExtensions() []string {
	out := make([]string, 0, len(allowedDocumentExt))
	for ext := range allowedDocumentExt {
		out = append(out, ext)
	}
	sort.Strings(out)
	return out
}
