package outbox

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// WebhookBridge delivers outbox messages via HTTP webhooks.
//
// This bridge sends outbox messages as HTTP POST requests to a configured URL.
// The message payload is serialized as JSON and includes the message ID, topic,
// payload, status, and metadata. Custom headers can be configured for authentication
// and other purposes.
//
// Example usage:
//
//	bridge, err := outbox.NewWebhookBridge(outbox.WebhookConfig{
//	    Name: "notifications",
//	    URL:  "https://api.example.com/webhooks",
//	    Headers: map[string]string{
//	        "Authorization": "Bearer token",
//	    },
//	})
type WebhookBridge struct {
	name    string
	url     string
	headers map[string]string
	client  *http.Client
}

// WebhookConfig configures a webhook bridge.
//
// The URL field is required and must be a valid HTTP/HTTPS endpoint.
// Headers can be used for authentication (e.g., Bearer tokens) or custom metadata.
// Timeout defaults to 30 seconds if not specified.
type WebhookConfig struct {
	Name    string
	URL     string
	Headers map[string]string
	Timeout time.Duration
}

// NewWebhookBridge creates a new webhook bridge.
//
// Returns an error if the name or URL is empty. The HTTP client is configured
// with the specified timeout (default 30 seconds).
func NewWebhookBridge(cfg WebhookConfig) (*WebhookBridge, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("webhook: name is required")
	}
	if cfg.URL == "" {
		return nil, fmt.Errorf("webhook: url is required")
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}

	return &WebhookBridge{
		name:    cfg.Name,
		url:     cfg.URL,
		headers: cfg.Headers,
		client: &http.Client{
			Timeout: timeout,
		},
	}, nil
}

// Name returns the bridge name.
func (b *WebhookBridge) Name() string {
	return b.name
}

// Send delivers a message via HTTP POST.
//
// The message is serialized as JSON with the following structure:
//
//	{
//	  "id": "message-id",
//	  "topic": "event.topic",
//	  "payload": { ... },
//	  "status": "pending",
//	  "attempts": 1,
//	  "available_at": "2024-01-01T00:00:00Z",
//	  "created_at": "2024-01-01T00:00:00Z"
//	}
//
// Returns an error if the HTTP request fails or returns a non-2xx status code.
// The response body is included in error messages for debugging.
func (b *WebhookBridge) Send(ctx context.Context, msg Message) error {
	payload := map[string]interface{}{
		"id":           msg.ID,
		"topic":        msg.Topic,
		"payload":      msg.Payload,
		"status":       msg.Status,
		"attempts":     msg.Attempts,
		"available_at": msg.AvailableAt,
		"created_at":   msg.CreatedAt,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("webhook: marshal payload: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", b.url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("webhook: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for key, value := range b.headers {
		req.Header.Set(key, value)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("webhook: unexpected status %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// Healthy checks if the webhook endpoint is reachable.
//
// This method performs a GET request to the configured URL and checks the response.
// Status codes 2xx-4xx are considered healthy (the endpoint is responding).
// Status code 5xx indicates a server error and is considered unhealthy.
//
// Note: Some webhook endpoints may not support GET requests. In such cases,
// consider disabling health checks or implementing a custom health check endpoint.
func (b *WebhookBridge) Healthy(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", b.url, nil)
	if err != nil {
		return fmt.Errorf("webhook: create health check request: %w", err)
	}

	resp, err := b.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook: health check failed: %w", err)
	}
	defer resp.Body.Close()

	// Accept any 2xx-4xx status as healthy (5xx indicates server error)
	if resp.StatusCode >= 500 {
		return fmt.Errorf("webhook: endpoint unhealthy (status %d)", resp.StatusCode)
	}

	return nil
}

// Close closes the HTTP client.
//
// This method closes idle connections to release resources. It is safe to call
// multiple times.
func (b *WebhookBridge) Close() error {
	b.client.CloseIdleConnections()
	return nil
}
