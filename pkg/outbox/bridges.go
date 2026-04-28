// Package outbox provides a transactional outbox pattern implementation
// with support for external message delivery through configurable bridges.
//
// The bridge system allows outbox messages to be delivered to external systems
// like Kafka, webhooks, RabbitMQ, etc. based on topic-based routing rules.
package outbox

import (
	"context"
	"fmt"
	"sync"
)

// Bridge defines an external message delivery destination (Kafka, Webhook, RabbitMQ, etc.).
//
// Implementations of this interface can be registered with a BridgeRegistry
// and used by the Dispatcher to deliver outbox messages to external systems.
// The router determines which bridges receive a message based on topic patterns.
type Bridge interface {
	// Name returns the unique identifier for this bridge.
	// This name is used for registration and routing configuration.
	Name() string

	// Send delivers an outbox message to the external system.
	// The context can be used for cancellation and timeout control.
	// Returns an error if delivery fails, which will trigger retry logic.
	Send(ctx context.Context, msg Message) error

	// Healthy checks if the bridge is operational.
	// This is called during health checks and can be used to verify
	// connectivity to the external system.
	Healthy(ctx context.Context) error

	// Close gracefully shuts down the bridge.
	// Called during application shutdown to release resources.
	Close() error
}

// BridgeRegistry manages a collection of registered bridges.
//
// The registry provides thread-safe operations for registering, retrieving,
// and closing bridges. It is used by the ManagedOutbox to coordinate
// multiple external delivery destinations.
type BridgeRegistry struct {
	bridges map[string]Bridge
	mu      sync.RWMutex
}

// NewBridgeRegistry creates an empty bridge registry.
func NewBridgeRegistry() *BridgeRegistry {
	return &BridgeRegistry{
		bridges: make(map[string]Bridge),
	}
}

// Register adds a bridge to the registry.
//
// The bridge name must be unique and non-empty. This method is thread-safe
// and will return an error if a bridge with the same name is already registered.
func (r *BridgeRegistry) Register(bridge Bridge) error {
	if bridge == nil {
		return fmt.Errorf("outbox: cannot register nil bridge")
	}
	name := bridge.Name()
	if name == "" {
		return fmt.Errorf("outbox: bridge name cannot be empty")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.bridges[name]; exists {
		return fmt.Errorf("outbox: bridge %q already registered", name)
	}

	r.bridges[name] = bridge
	return nil
}

// Get retrieves a bridge by name.
//
// Returns the bridge and true if found, nil and false otherwise.
// This method is thread-safe.
func (r *BridgeRegistry) Get(name string) (Bridge, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bridge, ok := r.bridges[name]
	return bridge, ok
}

// List returns all registered bridges.
//
// The returned slice is a copy and safe to modify.
// This method is thread-safe.
func (r *BridgeRegistry) List() []Bridge {
	r.mu.RLock()
	defer r.mu.RUnlock()

	bridges := make([]Bridge, 0, len(r.bridges))
	for _, bridge := range r.bridges {
		bridges = append(bridges, bridge)
	}
	return bridges
}

// Close shuts down all registered bridges.
//
// This method calls Close() on each registered bridge and collects any errors.
// If any bridge fails to close, a combined error is returned.
// This method is thread-safe.
func (r *BridgeRegistry) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for name, bridge := range r.bridges {
		if err := bridge.Close(); err != nil {
			errs = append(errs, fmt.Errorf("outbox: close bridge %q: %w", name, err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("outbox: bridge registry close errors: %v", errs)
	}
	return nil
}

// TopicRouting defines how messages are routed to bridges based on topic patterns.
//
// The pattern field supports wildcard matching:
//   - "*" matches any single segment
//   - "prefix.*" matches any topic with the given prefix
//   - Exact string matches the topic exactly
//
// Example patterns:
//   - "billing.*" matches "billing.invoice.created", "billing.payment.received"
//   - "orders.created" matches only "orders.created"
//   - "*" matches all topics
type TopicRouting struct {
	Pattern string   `json:"pattern"` // e.g., "billing.*" or "orders.created"
	Bridges []string `json:"bridges"` // bridge names to route matching messages to
}

// Router determines which bridges should receive a message based on topic patterns.
//
// The router maintains a list of routing rules and matches incoming message topics
// against these rules to determine which bridges should receive the message.
// Multiple bridges can receive the same message, enabling fan-out patterns.
type Router struct {
	routes []TopicRouting
	mu     sync.RWMutex
}

// NewRouter creates an empty topic router.
func NewRouter() *Router {
	return &Router{
		routes: make([]TopicRouting, 0),
	}
}

// AddRoute adds a topic pattern to bridge mapping.
//
// When a message with a matching topic is dispatched, all bridges listed
// in bridgeNames will receive the message. This method is thread-safe.
func (r *Router) AddRoute(pattern string, bridgeNames ...string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.routes = append(r.routes, TopicRouting{
		Pattern: pattern,
		Bridges: bridgeNames,
	})
}

// Match returns the list of bridge names that match the given topic.
//
// The topic is matched against all registered routing rules in order.
// Bridge names are deduplicated so each bridge appears at most once in the result.
// This method is thread-safe.
func (r *Router) Match(topic string) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matched []string
	seen := make(map[string]struct{})

	for _, route := range r.routes {
		if matchPattern(topic, route.Pattern) {
			for _, bridgeName := range route.Bridges {
				if _, exists := seen[bridgeName]; !exists {
					matched = append(matched, bridgeName)
					seen[bridgeName] = struct{}{}
				}
			}
		}
	}

	return matched
}

// matchPattern checks if a topic matches a pattern with wildcard support.
//
// Supported patterns:
//   - "*" matches any topic
//   - "prefix.*" matches any topic with the given prefix (e.g., "billing.*" matches "billing.invoice.created")
//   - Exact string matches the topic exactly
//
// This is a simple implementation that does not support multi-segment wildcards ("**").
// Future versions may add more sophisticated pattern matching.
func matchPattern(topic, pattern string) bool {
	if pattern == "*" || pattern == topic {
		return true
	}

	// Simple wildcard matching for single segment
	if pattern == "*" {
		return true
	}

	// Prefix matching for patterns like "billing.*"
	if len(pattern) > 2 && pattern[len(pattern)-2:] == ".*" {
		prefix := pattern[:len(pattern)-2]
		return topic == prefix || (len(topic) > len(prefix)+1 && topic[:len(prefix)] == prefix && topic[len(prefix)] == '.')
	}

	// Exact match
	return topic == pattern
}
