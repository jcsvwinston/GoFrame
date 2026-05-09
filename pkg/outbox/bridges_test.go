package outbox

import (
	"context"
	"strings"
	"testing"
)

func TestBridgeRegistry(t *testing.T) {
	registry := NewBridgeRegistry()

	// Test registering a bridge
	mockBridge := &mockBridge{name: "test-bridge"}
	err := registry.Register(mockBridge)
	if err != nil {
		t.Fatalf("register bridge: %v", err)
	}

	// Test getting a bridge
	bridge, ok := registry.Get("test-bridge")
	if !ok {
		t.Fatal("bridge not found")
	}
	if bridge.Name() != "test-bridge" {
		t.Fatalf("unexpected bridge name: %s", bridge.Name())
	}

	// Test duplicate registration
	err = registry.Register(mockBridge)
	if err == nil {
		t.Fatal("expected error for duplicate registration")
	}

	// Test listing bridges
	bridges := registry.List()
	if len(bridges) != 1 {
		t.Fatalf("expected 1 bridge, got %d", len(bridges))
	}
}

func TestRouter(t *testing.T) {
	router := NewRouter()

	// Test adding routes
	router.AddRoute("billing.*", "webhook-billing")
	router.AddRoute("orders.created", "kafka-orders")

	// Test matching patterns
	matches := router.Match("billing.invoice.created")
	if len(matches) != 1 || matches[0] != "webhook-billing" {
		t.Fatalf("expected webhook-billing for billing.invoice.created, got %v", matches)
	}

	matches = router.Match("orders.created")
	if len(matches) != 1 || matches[0] != "kafka-orders" {
		t.Fatalf("expected kafka-orders for orders.created, got %v", matches)
	}

	matches = router.Match("orders.updated")
	if len(matches) != 0 {
		t.Fatalf("expected no matches for orders.updated, got %v", matches)
	}

	// Test wildcard
	router.AddRoute("*", "default-bridge")
	matches = router.Match("any.topic")
	if len(matches) != 1 || matches[0] != "default-bridge" {
		t.Fatalf("expected default-bridge for wildcard, got %v", matches)
	}
}

func TestWebhookBridge(t *testing.T) {
	// This test would require a test HTTP server
	// For now, we'll just test the configuration
	cfg := WebhookConfig{
		Name: "test-webhook",
		URL:  "http://localhost:8080/webhook",
		Headers: map[string]string{
			"Authorization": "Bearer test-token",
		},
	}

	bridge, err := NewWebhookBridge(cfg)
	if err != nil {
		t.Fatalf("create webhook bridge: %v", err)
	}

	if bridge.Name() != "test-webhook" {
		t.Fatalf("unexpected bridge name: %s", bridge.Name())
	}

	if err := bridge.Close(); err != nil {
		t.Fatalf("close bridge: %v", err)
	}
}

func TestWebhookBridgeValidation(t *testing.T) {
	// Test missing name
	_, err := NewWebhookBridge(WebhookConfig{URL: "http://localhost"})
	if err == nil {
		t.Fatal("expected error for missing name")
	}

	// Test missing URL
	_, err = NewWebhookBridge(WebhookConfig{Name: "test"})
	if err == nil {
		t.Fatal("expected error for missing URL")
	}
}

func TestKafkaBridge(t *testing.T) {
	cfg := KafkaConfig{
		Name:    "test-kafka",
		Brokers: []string{"localhost:9092"},
		Topic:   "events",
	}

	_, err := NewKafkaBridge(cfg)
	if err == nil {
		t.Fatal("expected disabled kafka bridge error")
	}
	if !strings.Contains(err.Error(), "disabled") {
		t.Fatalf("expected disabled kafka bridge error, got %v", err)
	}
}

func TestKafkaBridgeValidation(t *testing.T) {
	// Test missing name
	_, err := NewKafkaBridge(KafkaConfig{Brokers: []string{"localhost:9092"}, Topic: "events"})
	if err == nil {
		t.Fatal("expected error for missing name")
	}

	// Test missing brokers
	_, err = NewKafkaBridge(KafkaConfig{Name: "test", Topic: "events"})
	if err == nil {
		t.Fatal("expected error for missing brokers")
	}

	// Test missing topic
	_, err = NewKafkaBridge(KafkaConfig{Name: "test", Brokers: []string{"localhost:9092"}})
	if err == nil {
		t.Fatal("expected error for missing topic")
	}
}

// mockBridge is a test implementation of Bridge
type mockBridge struct {
	name    string
	sendErr error
	healthy bool
	closed  bool
}

func (m *mockBridge) Name() string {
	return m.name
}

func (m *mockBridge) Send(ctx context.Context, msg Message) error {
	return m.sendErr
}

func (m *mockBridge) Healthy(ctx context.Context) error {
	if !m.healthy {
		return &testError{msg: "unhealthy"}
	}
	return nil
}

func (m *mockBridge) Close() error {
	m.closed = true
	return nil
}

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
