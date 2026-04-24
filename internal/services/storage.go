package services

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type StorageService interface {
	UploadFile(file io.Reader, filename string, contentType string) (string, error)
	DeleteFile(fileURL string) error
}

type supabaseStorageService struct {
	url    string
	key    string
	bucket string
	client *http.Client
}

func NewSupabaseStorageService(url, key, bucket string) StorageService {
	return &supabaseStorageService{
		url:    url,
		key:    key,
		bucket: bucket,
		client: &http.Client{},
	}
}

func (s *supabaseStorageService) UploadFile(file io.Reader, filename string, contentType string) (string, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	endpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.url, s.bucket, filename)
	req, err := http.NewRequest("POST", endpoint, bytes.NewReader(fileBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.key)
	req.Header.Set("Content-Type", contentType)

	resp, err := s.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase storage error: %s - status: %d", string(bodyBytes), resp.StatusCode)
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.url, s.bucket, filename)
	return publicURL, nil
}

func (s *supabaseStorageService) DeleteFile(fileURL string) error {
	// Extract the file path relative to the bucket from the public URL
	// Public URL format: {supabaseURL}/storage/v1/object/public/{bucket}/{filePath}
	publicPrefix := fmt.Sprintf("%s/storage/v1/object/public/%s/", s.url, s.bucket)
	if !strings.HasPrefix(fileURL, publicPrefix) {
		return fmt.Errorf("file URL does not belong to this bucket: %s", fileURL)
	}

	filePath := strings.TrimPrefix(fileURL, publicPrefix)

	endpoint := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.url, s.bucket, filePath)
	req, err := http.NewRequest("DELETE", endpoint, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.key)
	// Supabase requires apikey header for DELETE requests
	req.Header.Set("apikey", s.key)

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute delete request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase storage delete error: %s - status: %d", string(bodyBytes), resp.StatusCode)
	}

	return nil
}
