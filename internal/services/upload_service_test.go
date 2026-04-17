package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"strings"
	"testing"
)

// mockStorageService implements StorageService for testing
type mockStorageService struct {
	uploadedFiles []mockUploadedFile
	deletedURLs   []string
	deleteError   error
}

type mockUploadedFile struct {
	Path        string
	ContentType string
}

func (m *mockStorageService) UploadFile(file io.Reader, filename string, contentType string) (string, error) {
	m.uploadedFiles = append(m.uploadedFiles, mockUploadedFile{
		Path:        filename,
		ContentType: contentType,
	})
	publicURL := fmt.Sprintf("https://supabase.example.com/storage/v1/object/public/test-bucket/%s", filename)
	return publicURL, nil
}

func (m *mockStorageService) DeleteFile(fileURL string) error {
	if m.deleteError != nil {
		return m.deleteError
	}
	m.deletedURLs = append(m.deletedURLs, fileURL)
	return nil
}

// createMultipartFiles creates []*multipart.FileHeader by writing to a multipart
// request body and then parsing it with http.Request.MultipartForm
func createMultipartFiles(files []struct {
	Name        string
	Size        int64
	ContentType string
}) []*multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	for _, f := range files {
		h := make(textproto.MIMEHeader)
		h.Set("Content-Type", f.ContentType)
		h.Set("Content-Disposition", fmt.Sprintf(`form-data; name="files"; filename="%s"`, f.Name))

		part, _ := writer.CreatePart(h)
		content := bytes.Repeat([]byte("x"), int(f.Size))
		part.Write(content)
	}
	writer.Close()

	// Create an HTTP request with the multipart body and parse it
	req, _ := http.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.ParseMultipartForm(100 << 20)

	if req.MultipartForm == nil || req.MultipartForm.File == nil {
		return nil
	}
	return req.MultipartForm.File["files"]
}

// ===================== TESTS =====================

func TestUploadService_NoFiles(t *testing.T) {
	mock := &mockStorageService{}
	svc := NewUploadService(mock)

	urls, err := svc.UploadServiceRequestFiles(1, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if urls != nil {
		t.Errorf("expected nil URLs for no files, got %v", urls)
	}
	if len(mock.uploadedFiles) != 0 {
		t.Errorf("expected no uploads, got %d", len(mock.uploadedFiles))
	}
}

func TestUploadService_MaxFilesExceeded(t *testing.T) {
	mock := &mockStorageService{}
	svc := NewUploadService(mock)

	// Create more than MaxFilesPerRequest files
	fileSpecs := make([]struct {
		Name        string
		Size        int64
		ContentType string
	}, MaxFilesPerRequest+1)
	for i := range fileSpecs {
		fileSpecs[i] = struct {
			Name        string
			Size        int64
			ContentType string
		}{fmt.Sprintf("file%d.jpg", i), 100, "image/jpeg"}
	}

	files := createMultipartFiles(fileSpecs)
	urls, err := svc.UploadServiceRequestFiles(1, files)
	if err == nil {
		t.Fatal("expected error for exceeding max files, got nil")
	}
	if !strings.Contains(err.Error(), "maximum") {
		t.Errorf("expected max files error, got: %v", err)
	}
	if urls != nil {
		t.Errorf("expected nil URLs, got %v", urls)
	}
}

func TestUploadService_FileTooLarge(t *testing.T) {
	mock := &mockStorageService{}
	svc := NewUploadService(mock)

	files := createMultipartFiles([]struct {
		Name        string
		Size        int64
		ContentType string
	}{{"big.jpg", MaxFileSizeBytes + 1, "image/jpeg"}})

	_, err := svc.UploadServiceRequestFiles(1, files)
	if err == nil {
		t.Fatal("expected error for file too large, got nil")
	}
	if !strings.Contains(err.Error(), "file size exceeds") {
		t.Errorf("expected file size error, got: %v", err)
	}
}

func TestUploadService_DisallowedExtension(t *testing.T) {
	mock := &mockStorageService{}
	svc := NewUploadService(mock)

	files := createMultipartFiles([]struct {
		Name        string
		Size        int64
		ContentType string
	}{{"malware.exe", 100, "application/octet-stream"}})

	_, err := svc.UploadServiceRequestFiles(1, files)
	if err == nil {
		t.Fatal("expected error for disallowed extension, got nil")
	}
	if !strings.Contains(err.Error(), "is not allowed") {
		t.Errorf("expected extension not allowed error, got: %v", err)
	}
}

func TestUploadService_DisallowedMIMEType(t *testing.T) {
	mock := &mockStorageService{}
	svc := NewUploadService(mock)

	files := createMultipartFiles([]struct {
		Name        string
		Size        int64
		ContentType string
	}{{"file.jpg", 100, "video/mp4"}})

	_, err := svc.UploadServiceRequestFiles(1, files)
	if err == nil {
		t.Fatal("expected error for disallowed MIME type, got nil")
	}
	if !strings.Contains(err.Error(), "is not allowed") {
		t.Errorf("expected content type not allowed error, got: %v", err)
	}
}

func TestUploadService_TotalSizeExceeded(t *testing.T) {
	mock := &mockStorageService{}
	svc := NewUploadService(mock)

	// Use small content but override Size to simulate large files (avoids allocating 40MB in memory)
	sizePerFile := int64(8 << 20) // 8MB each, 5 * 8MB = 40MB > 30MB limit
	fileSpecs := []struct {
		Name        string
		Size        int64
		ContentType string
	}{
		{"f1.jpg", 100, "image/jpeg"},
		{"f2.jpg", 100, "image/jpeg"},
		{"f3.jpg", 100, "image/jpeg"},
		{"f4.jpg", 100, "image/jpeg"},
		{"f5.jpg", 100, "image/jpeg"},
	}

	files := createMultipartFiles(fileSpecs)
	for i := range files {
		files[i].Size = sizePerFile // override reported size to simulate large files
	}

	_, err := svc.UploadServiceRequestFiles(1, files)
	if err == nil {
		t.Fatal("expected error for total size exceeded, got nil")
	}
	if !strings.Contains(err.Error(), "total file size exceeds") {
		t.Errorf("expected total size error, got: %v", err)
	}
}

func TestUploadService_Success(t *testing.T) {
	mock := &mockStorageService{}
	svc := NewUploadService(mock)

	files := createMultipartFiles([]struct {
		Name        string
		Size        int64
		ContentType string
	}{
		{"photo1.jpg", 100, "image/jpeg"},
		{"photo2.png", 200, "image/png"},
		{"doc.pdf", 300, "application/pdf"},
	})

	urls, err := svc.UploadServiceRequestFiles(5, files)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(urls) != 3 {
		t.Fatalf("expected 3 URLs, got %d", len(urls))
	}
	if len(mock.uploadedFiles) != 3 {
		t.Fatalf("expected 3 uploaded files, got %d", len(mock.uploadedFiles))
	}

	// Verify paths use correct format: service_requests/{userID}/{uuid}.{ext}
	for _, uploaded := range mock.uploadedFiles {
		if !strings.HasPrefix(uploaded.Path, "service_requests/5/") {
			t.Errorf("expected path prefix 'service_requests/5/', got %q", uploaded.Path)
		}
		// Should contain a UUID (36 chars) plus extension
		parts := strings.Split(strings.TrimPrefix(uploaded.Path, "service_requests/5/"), ".")
		if len(parts) != 2 {
			t.Errorf("expected path with uuid.ext format, got %q", uploaded.Path)
		}
		uuidPart := parts[0]
		if len(uuidPart) != 36 {
			t.Errorf("expected UUID (36 chars), got %d chars in %q", len(uuidPart), uuidPart)
		}
	}

	// Verify content types are valid
	for _, uploaded := range mock.uploadedFiles {
		if !AllowedMIMETypes[uploaded.ContentType] && uploaded.ContentType != "application/octet-stream" {
			t.Errorf("unexpected content type %q for path %q", uploaded.ContentType, uploaded.Path)
		}
	}

	// Verify returned URLs
	for _, url := range urls {
		if !strings.HasPrefix(url, "https://supabase.example.com/storage/v1/object/public/test-bucket/service_requests/5/") {
			t.Errorf("unexpected URL format: %q", url)
		}
	}
}

func TestUploadService_RollbackOnUploadFailure(t *testing.T) {
	failMock := &failingMockStorage{failAfterN: 1}
	svc := NewUploadService(failMock)

	files := createMultipartFiles([]struct {
		Name        string
		Size        int64
		ContentType string
	}{
		{"first.jpg", 100, "image/jpeg"},
		{"second.jpg", 100, "image/jpeg"},
	})

	urls, err := svc.UploadServiceRequestFiles(1, files)
	if err == nil {
		t.Fatal("expected error when upload fails, got nil")
	}
	if urls != nil {
		t.Errorf("expected nil URLs on failure, got %v", urls)
	}

	// First file should have been uploaded then rolled back
	if len(failMock.uploaded) != 1 {
		t.Fatalf("expected 1 upload attempt before failure, got %d", len(failMock.uploaded))
	}
	if len(failMock.deleted) != 1 {
		t.Fatalf("expected 1 rollback deletion, got %d", len(failMock.deleted))
	}
}

// failingMockStorage succeeds for the first N uploads, then fails
type failingMockStorage struct {
	uploaded   []string
	deleted    []string
	failAfterN int
	uploadCount int
}

func (m *failingMockStorage) UploadFile(file io.Reader, filename string, contentType string) (string, error) {
	m.uploadCount++
	if m.uploadCount > m.failAfterN {
		return "", fmt.Errorf("simulated upload failure")
	}
	m.uploaded = append(m.uploaded, filename)
	return fmt.Sprintf("https://supabase.example.com/storage/v1/object/public/test-bucket/%s", filename), nil
}

func (m *failingMockStorage) DeleteFile(fileURL string) error {
	m.deleted = append(m.deleted, fileURL)
	return nil
}

func TestUploadService_ExtensionCaseInsensitive(t *testing.T) {
	mock := &mockStorageService{}
	svc := NewUploadService(mock)

	// .JPG uppercase should be accepted (case-insensitive extension check)
	files := createMultipartFiles([]struct {
		Name        string
		Size        int64
		ContentType string
	}{{"photo.JPG", 100, "image/jpeg"}})

	urls, err := svc.UploadServiceRequestFiles(1, files)
	if err != nil {
		t.Fatalf("expected uppercase extension to be accepted, got error: %v", err)
	}
	if len(urls) != 1 {
		t.Errorf("expected 1 URL, got %d", len(urls))
	}
}

func TestUploadService_AllowedFileTypes(t *testing.T) {
	mock := &mockStorageService{}
	svc := NewUploadService(mock)

	// Test all allowed extensions with their correct MIME types
	allowedTypes := []struct {
		Ext  string
		MIME string
	}{
		{".jpg", "image/jpeg"},
		{".jpeg", "image/jpeg"},
		{".png", "image/png"},
		{".webp", "image/webp"},
		{".gif", "image/gif"},
		{".pdf", "application/pdf"},
	}
	for _, ft := range allowedTypes {
		files := createMultipartFiles([]struct {
			Name        string
			Size        int64
			ContentType string
		}{{"file" + ft.Ext, 100, ft.MIME}})

		_, err := svc.UploadServiceRequestFiles(1, files)
		if err != nil {
			t.Errorf("extension %q should be allowed but got error: %v", ft.Ext, err)
		}
	}
}

func TestRollbackFiles(t *testing.T) {
	mock := &mockStorageService{}
	svc := NewUploadService(mock)

	urls := []string{
		"https://supabase.example.com/storage/v1/object/public/test-bucket/service_requests/1/abc.jpg",
		"https://supabase.example.com/storage/v1/object/public/test-bucket/service_requests/1/def.png",
	}

	svc.RollbackFiles(urls)

	if len(mock.deletedURLs) != 2 {
		t.Fatalf("expected 2 deletions, got %d", len(mock.deletedURLs))
	}
	if mock.deletedURLs[0] != urls[0] {
		t.Errorf("expected deletion of %q, got %q", urls[0], mock.deletedURLs[0])
	}
	if mock.deletedURLs[1] != urls[1] {
		t.Errorf("expected deletion of %q, got %q", urls[1], mock.deletedURLs[1])
	}
}

func TestRollbackFiles_Empty(t *testing.T) {
	mock := &mockStorageService{}
	svc := NewUploadService(mock)

	svc.RollbackFiles(nil)

	if len(mock.deletedURLs) != 0 {
		t.Errorf("expected 0 deletions for empty list, got %d", len(mock.deletedURLs))
	}
}

func TestRollbackFiles_BestEffortOnDeleteFailure(t *testing.T) {
	failMock := &failingDeleteMock{failOnURL: "url2"}
	svc := NewUploadService(failMock)

	urls := []string{"url1", "url2", "url3"}
	// Should not panic or return error even if one deletion fails
	svc.RollbackFiles(urls)

	if len(failMock.deleted) != 2 {
		t.Errorf("expected 2 successful deletions, got %d", len(failMock.deleted))
	}
}

type failingDeleteMock struct {
	deleted   []string
	failOnURL string
}

func (m *failingDeleteMock) UploadFile(file io.Reader, filename string, contentType string) (string, error) {
	return "url", nil
}

func (m *failingDeleteMock) DeleteFile(fileURL string) error {
	if fileURL == m.failOnURL {
		return fmt.Errorf("simulated delete failure")
	}
	m.deleted = append(m.deleted, fileURL)
	return nil
}

func TestParseAttachmentURLs_Valid(t *testing.T) {
	input := json.RawMessage(`["https://example.com/a.jpg","https://example.com/b.png"]`)
	urls := ParseAttachmentURLs(input)

	if len(urls) != 2 {
		t.Fatalf("expected 2 URLs, got %d", len(urls))
	}
	if urls[0] != "https://example.com/a.jpg" {
		t.Errorf("expected first URL to be 'https://example.com/a.jpg', got %q", urls[0])
	}
	if urls[1] != "https://example.com/b.png" {
		t.Errorf("expected second URL to be 'https://example.com/b.png', got %q", urls[1])
	}
}

func TestParseAttachmentURLs_Empty(t *testing.T) {
	urls := ParseAttachmentURLs(nil)
	if urls != nil {
		t.Errorf("expected nil for nil input, got %v", urls)
	}

	urls = ParseAttachmentURLs(json.RawMessage{})
	if urls != nil {
		t.Errorf("expected nil for empty input, got %v", urls)
	}
}

func TestParseAttachmentURLs_InvalidJSON(t *testing.T) {
	urls := ParseAttachmentURLs(json.RawMessage(`{invalid}`))
	if urls != nil {
		t.Errorf("expected nil for invalid JSON, got %v", urls)
	}
}

func TestParseAttachmentURLs_EmptyArray(t *testing.T) {
	urls := ParseAttachmentURLs(json.RawMessage(`[]`))
	if len(urls) != 0 {
		t.Errorf("expected empty slice for empty array, got %v", urls)
	}
}

func TestFileUploadError_Format(t *testing.T) {
	err := FileUploadError{Filename: "test.exe", Reason: "extension not allowed"}
	expected := `file "test.exe": extension not allowed`
	if err.Error() != expected {
		t.Errorf("expected %q, got %q", expected, err.Error())
	}
}
