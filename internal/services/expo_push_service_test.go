package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestExpoPushService_SendNewsPublished_SendsExpectedPayload(t *testing.T) {
	var received []ExpoPushMessage
	var mu sync.Mutex
	var handlerErr error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			handlerErr = fmt.Errorf("expected POST method, got %s", r.Method)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		var payload []ExpoPushMessage
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			handlerErr = fmt.Errorf("failed to decode payload: %w", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mu.Lock()
		received = append(received, payload...)
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := &ExpoPushService{
		client:   server.Client(),
		endpoint: server.URL,
	}

	if err := svc.SendNewsPublished(context.Background(), []string{"ExponentPushToken[abc]"}, 123); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handlerErr != nil {
		t.Fatalf("handler error: %v", handlerErr)
	}

	if len(received) != 1 {
		t.Fatalf("expected 1 push message, got %d", len(received))
	}

	msg := received[0]
	if msg.To != "ExponentPushToken[abc]" {
		t.Errorf("expected token %q, got %q", "ExponentPushToken[abc]", msg.To)
	}
	if msg.Title != newsPushTitle {
		t.Errorf("expected title %q, got %q", newsPushTitle, msg.Title)
	}
	if msg.Body != newsPushBody {
		t.Errorf("expected body %q, got %q", newsPushBody, msg.Body)
	}
	if msg.Data["screen"] != "/(news)/123" {
		t.Errorf("expected screen %q, got %v", "/(news)/123", msg.Data["screen"])
	}
}

func TestExpoPushService_SendNewsPublished_ChunksTokens(t *testing.T) {
	var mu sync.Mutex
	var requestCounts []int
	var handlerErr error

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var payload []ExpoPushMessage
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			handlerErr = fmt.Errorf("failed to decode payload: %w", err)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		mu.Lock()
		requestCounts = append(requestCounts, len(payload))
		mu.Unlock()

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := &ExpoPushService{
		client:   server.Client(),
		endpoint: server.URL,
	}

	tokens := make([]string, 205)
	for i := range tokens {
		tokens[i] = fmt.Sprintf("ExponentPushToken[%03d]", i)
	}

	if err := svc.SendNewsPublished(context.Background(), tokens, 999); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if handlerErr != nil {
		t.Fatalf("handler error: %v", handlerErr)
	}

	if len(requestCounts) != 3 {
		t.Fatalf("expected 3 requests, got %d", len(requestCounts))
	}
	if requestCounts[0] != expoPushBatchSize || requestCounts[1] != expoPushBatchSize || requestCounts[2] != 5 {
		t.Fatalf("unexpected batch sizes: %v", requestCounts)
	}
}

func TestExpoPushService_SendNewsPublished_EmptyTokens(t *testing.T) {
	called := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	svc := &ExpoPushService{
		client:   server.Client(),
		endpoint: server.URL,
	}

	if err := svc.SendNewsPublished(context.Background(), nil, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if called {
		t.Fatal("expected no request when there are no tokens")
	}
}
