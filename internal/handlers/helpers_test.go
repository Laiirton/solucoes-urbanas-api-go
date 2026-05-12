package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     interface{}
		status   int
		expected string
	}{
		{
			name:     "map response",
			data:     map[string]interface{}{"key": "value"},
			status:   http.StatusOK,
			expected: `{"key":"value"}` + "\n",
		},
		{
			name:     "slice response",
			data:     []string{"a", "b"},
			status:   http.StatusOK,
			expected: `["a","b"]` + "\n",
		},
		{
			name:     "nil response",
			data:     nil,
			status:   http.StatusOK,
			expected: `null` + "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			respondJSON(w, tt.status, tt.data)
			resp := w.Result()
			if resp.StatusCode != tt.status {
				t.Errorf("expected status %d, got %d", tt.status, resp.StatusCode)
			}
			if resp.Header.Get("Content-Type") != "application/json" {
				t.Errorf("expected Content-Type application/json, got %s", resp.Header.Get("Content-Type"))
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	w := httptest.NewRecorder()
	respondError(w, http.StatusBadRequest, "bad request")
	resp := w.Result()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", resp.StatusCode)
	}

	var errResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}
	if errResp["error"] != "bad request" {
		t.Errorf("expected error 'bad request', got '%s'", errResp["error"])
	}
}

func TestParsePagination(t *testing.T) {
	tests := []struct {
		url      string
		expected int
	}{
		{"/?page=1", 1},
		{"/?page=5", 5},
		{"/?page=0", 1},
		{"/?page=10", 10},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, tt.url, nil)
			page, _ := parsePagination(r)
			if page != tt.expected {
				t.Errorf("expected page %d, got %d", tt.expected, page)
			}
		})
	}
}

func TestParsePaginationDefault(t *testing.T) {
	r := httptest.NewRequest(http.MethodGet, "/", nil)
	page, limit := parsePagination(r)
	if page != 1 {
		t.Errorf("expected default page 1, got %d", page)
	}
	if limit != 20 {
		t.Errorf("expected default limit 20, got %d", limit)
	}
}
