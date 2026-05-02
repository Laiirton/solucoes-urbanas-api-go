package services

import (
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

// Upload configuration constants
const (
	MaxFilesPerRequest     = 5
	MaxFileSizeBytes       = 10 << 20 // 10 MB per file
	MaxTotalFilesSizeBytes = 30 << 20 // 30 MB total across all files
)

// AllowedMIMETypes defines which MIME types are permitted for service request attachments
var AllowedMIMETypes = map[string]bool{
	"image/jpeg":      true,
	"image/png":       true,
	"image/webp":      true,
	"image/gif":       true,
	"application/pdf":                                                           true,
	"application/msword":                                                        true,
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document":   true,
	"application/vnd.ms-excel":                                                  true,
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet":         true,
	"text/plain":                                                                true,
	"text/csv":                                                                  true,
}

// AllowedExtensions defines which file extensions are permitted
var AllowedExtensions = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
	".pdf":  true,
	".doc":  true,
	".docx": true,
	".xls":  true,
	".xlsx": true,
	".txt":  true,
	".csv":  true,
}

// FileUploadError represents a validation error for a specific file
type FileUploadError struct {
	Filename string
	Reason   string
}

func (e FileUploadError) Error() string {
	return fmt.Sprintf("file %q: %s", e.Filename, e.Reason)
}

// UploadService handles business logic for file uploads with validation and rollback
type UploadService struct {
	storage StorageService
}

// NewUploadService creates a new UploadService
func NewUploadService(storage StorageService) *UploadService {
	return &UploadService{storage: storage}
}

// UploadServiceRequestFiles validates and uploads multiple files for a service request.
// Returns the list of public URLs on success, or a detailed error on failure.
// Files are closed properly inside the loop (no defer-in-loop bug).
func (s *UploadService) UploadServiceRequestFiles(userID int64, files []*multipart.FileHeader) ([]string, error) {
	if len(files) == 0 {
		return nil, nil
	}

	if len(files) > MaxFilesPerRequest {
		return nil, fmt.Errorf("maximum %d files allowed, got %d", MaxFilesPerRequest, len(files))
	}

	// Validate all files before uploading any
	var totalSize int64
	for _, fh := range files {
		if fh.Size > MaxFileSizeBytes {
			return nil, FileUploadError{Filename: fh.Filename, Reason: fmt.Sprintf("file size exceeds %d MB limit", MaxFileSizeBytes/(1<<20))}
		}
		totalSize += fh.Size

		ext := strings.ToLower(filepath.Ext(fh.Filename))
		if !AllowedExtensions[ext] {
			return nil, FileUploadError{Filename: fh.Filename, Reason: fmt.Sprintf("file extension %q is not allowed", ext)}
		}

		contentType := fh.Header.Get("Content-Type")
		if contentType != "" && !AllowedMIMETypes[contentType] {
			return nil, FileUploadError{Filename: fh.Filename, Reason: fmt.Sprintf("content type %q is not allowed", contentType)}
		}
	}

	if totalSize > MaxTotalFilesSizeBytes {
		return nil, fmt.Errorf("total file size exceeds %d MB limit", MaxTotalFilesSizeBytes/(1<<20))
	}

	// Upload files one by one, closing each file handle immediately after use
	var uploadedURLs []string
	userIDStr := strconv.FormatInt(userID, 10)

	for _, fh := range files {
		file, err := fh.Open()
		if err != nil {
			// Rollback already uploaded files
			s.RollbackFiles(uploadedURLs)
			return nil, FileUploadError{Filename: fh.Filename, Reason: "failed to open file"}
		}

		// Determine content type
		contentType := fh.Header.Get("Content-Type")
		if contentType == "" || !AllowedMIMETypes[contentType] {
			contentType = "application/octet-stream"
		}

		// Generate UUID-based path: service_requests/{userID}/{uuid}.{ext}
		ext := filepath.Ext(fh.Filename)
		filePath := fmt.Sprintf("service_requests/%s/%s%s", userIDStr, uuid.New().String(), ext)

		publicURL, err := s.storage.UploadFile(file, filePath, contentType)
		file.Close() // Close immediately after use, not deferred

		if err != nil {
			// Rollback already uploaded files
			s.RollbackFiles(uploadedURLs)
			return nil, fmt.Errorf("failed to upload file %q: %w", fh.Filename, err)
		}

		uploadedURLs = append(uploadedURLs, publicURL)
	}

	return uploadedURLs, nil
}

// RollbackFiles deletes previously uploaded files from storage to prevent orphans
func (s *UploadService) RollbackFiles(urls []string) {
	for _, url := range urls {
		if err := s.storage.DeleteFile(url); err != nil {
			// Log but don't fail — rollback is best-effort
			log.Printf("warning: failed to rollback file %s: %v", url, err)
		}
	}
}

// ParseAttachmentURLs parses the attachments JSON and returns the list of URLs.
// Returns nil if the JSON is empty or cannot be parsed.
func ParseAttachmentURLs(attachmentsJSON json.RawMessage) []string {
	if len(attachmentsJSON) == 0 {
		return nil
	}

	var urls []string
	if err := json.Unmarshal(attachmentsJSON, &urls); err != nil {
		log.Printf("warning: failed to parse attachments JSON: %v", err)
		return nil
	}

	return urls
}
