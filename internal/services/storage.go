package services

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type StorageService interface {
	UploadFile(file multipart.File, filename string, contentType string) (string, error)
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

func (s *supabaseStorageService) UploadFile(file multipart.File, filename string, contentType string) (string, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	// Move cursor back to the beginning just in case
	file.Seek(0, io.SeekStart)

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
