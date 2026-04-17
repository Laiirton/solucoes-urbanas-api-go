package services

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSupabaseStorageService_UploadFile_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/storage/v1/object/") {
			t.Errorf("expected storage path prefix, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header to be set")
		}
		if r.Header.Get("Content-Type") == "" {
			t.Error("expected Content-Type header to be set")
		}
		body, _ := io.ReadAll(r.Body)
		if len(body) == 0 {
			t.Error("expected request body to contain file data")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewSupabaseStorageService(server.URL, "test-key", "test-bucket")
	content := strings.NewReader("hello world")

	publicURL, err := svc.UploadFile(content, "service_requests/1/test.jpg", "image/jpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expectedURL := server.URL + "/storage/v1/object/public/test-bucket/service_requests/1/test.jpg"
	if publicURL != expectedURL {
		t.Errorf("expected URL %q, got %q", expectedURL, publicURL)
	}
}

func TestSupabaseStorageService_UploadFile_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("internal server error"))
	}))
	defer server.Close()

	svc := NewSupabaseStorageService(server.URL, "test-key", "test-bucket")
	content := strings.NewReader("hello world")

	_, err := svc.UploadFile(content, "test.jpg", "image/jpeg")
	if err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
	if !strings.Contains(err.Error(), "supabase storage error") {
		t.Errorf("expected supabase storage error, got: %v", err)
	}
}

func TestSupabaseStorageService_DeleteFile_Success(t *testing.T) {
	var requestMethod string
	var requestPath string
	var apikeySet bool

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestMethod = r.Method
		requestPath = r.URL.Path
		apikeySet = r.Header.Get("apikey") != ""
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	svc := NewSupabaseStorageService(server.URL, "test-key", "test-bucket")
	fileURL := server.URL + "/storage/v1/object/public/test-bucket/service_requests/1/abc.jpg"

	err := svc.DeleteFile(fileURL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestMethod != http.MethodDelete {
		t.Errorf("expected DELETE method, got %s", requestMethod)
	}
	expectedPath := "/storage/v1/object/test-bucket/service_requests/1/abc.jpg"
	if requestPath != expectedPath {
		t.Errorf("expected path %q, got %q", expectedPath, requestPath)
	}
	if !apikeySet {
		t.Error("expected apikey header to be set on DELETE request")
	}
}

func TestSupabaseStorageService_DeleteFile_WrongBucket(t *testing.T) {
	svc := NewSupabaseStorageService("https://example.supabase.co", "test-key", "my-bucket")

	// URL that belongs to a different bucket
	wrongURL := "https://example.supabase.co/storage/v1/object/public/other-bucket/file.jpg"
	err := svc.DeleteFile(wrongURL)
	if err == nil {
		t.Fatal("expected error for URL from different bucket, got nil")
	}
	if !strings.Contains(err.Error(), "does not belong to this bucket") {
		t.Errorf("expected bucket mismatch error, got: %v", err)
	}
}

func TestSupabaseStorageService_DeleteFile_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("access denied"))
	}))
	defer server.Close()

	svc := NewSupabaseStorageService(server.URL, "test-key", "test-bucket")
	fileURL := server.URL + "/storage/v1/object/public/test-bucket/file.jpg"

	err := svc.DeleteFile(fileURL)
	if err == nil {
		t.Fatal("expected error for server error response, got nil")
	}
	if !strings.Contains(err.Error(), "supabase storage delete error") {
		t.Errorf("expected supabase delete error, got: %v", err)
	}
}

func TestSupabaseStorageService_UploadFile_UsesReader(t *testing.T) {
	var receivedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := NewSupabaseStorageService(server.URL, "test-key", "test-bucket")

	// Test that io.Reader works (not just multipart.File)
	content := strings.NewReader("test content from io.Reader")
	_, err := svc.UploadFile(content, "test/file.txt", "text/plain")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if receivedBody != "test content from io.Reader" {
		t.Errorf("expected body %q, got %q", "test content from io.Reader", receivedBody)
	}
}
