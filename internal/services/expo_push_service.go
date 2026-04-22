package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const (
	expoPushEndpoint  = "https://exp.host/--/api/v2/push/send"
	expoPushBatchSize = 100
)

type ExpoPushMessage struct {
	To    string         `json:"to"`
	Title string         `json:"title"`
	Body  string         `json:"body"`
	Data  map[string]any `json:"data,omitempty"`
}

type ExpoPushService struct {
	client   *http.Client
	endpoint string
}

func NewExpoPushService() *ExpoPushService {
	return &ExpoPushService{
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
		endpoint: expoPushEndpoint,
	}
}

func (s *ExpoPushService) SendNewsPublished(ctx context.Context, tokens []string, newsID int64, title, body string) error {
	if len(tokens) == 0 {
		return nil
	}

	screen := fmt.Sprintf("/(news)/%d", newsID)
	for _, batch := range chunkStrings(tokens, expoPushBatchSize) {
		messages := make([]ExpoPushMessage, 0, len(batch))
		for _, token := range batch {
			token = strings.TrimSpace(token)
			if token == "" {
				continue
			}

			messages = append(messages, ExpoPushMessage{
				To:    token,
				Title: title,
				Body:  body,
				Data: map[string]any{
					"screen": screen,
				},
			})
		}

		if len(messages) == 0 {
			continue
		}

		if err := s.sendBatch(ctx, messages); err != nil {
			return err
		}
	}

	return nil
}

func (s *ExpoPushService) sendBatch(ctx context.Context, messages []ExpoPushMessage) error {
	payload, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("failed to encode expo push payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create expo push request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send expo push request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("expo push error: %s (status %d)", strings.TrimSpace(string(body)), resp.StatusCode)
	}

	return nil
}

func chunkStrings(values []string, size int) [][]string {
	if size <= 0 || len(values) == 0 {
		return nil
	}

	chunks := make([][]string, 0, (len(values)+size-1)/size)
	for start := 0; start < len(values); start += size {
		end := start + size
		if end > len(values) {
			end = len(values)
		}
		chunks = append(chunks, values[start:end])
	}

	return chunks
}
