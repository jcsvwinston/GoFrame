package outbox

import (
	"context"
	"fmt"
)

// KafkaBridge is reserved for a future Kafka implementation.
// It is intentionally disabled until the package wires a maintained Kafka
// client and real delivery/health semantics. Applications must not configure
// Kafka as a production bridge in this release.
type KafkaBridge struct {
	name    string
	brokers []string
	topic   string
}

// KafkaConfig configures a Kafka bridge.
//
// Brokers is a list of Kafka broker addresses in the format "host:port".
// Topic is the Kafka topic to which messages will be published.
type KafkaConfig struct {
	Name    string
	Brokers []string
	Topic   string
}

// NewKafkaBridge validates Kafka bridge configuration and then returns a clear
// error because Kafka delivery is not implemented in this release.
func NewKafkaBridge(cfg KafkaConfig) (*KafkaBridge, error) {
	if cfg.Name == "" {
		return nil, fmt.Errorf("kafka: name is required")
	}
	if len(cfg.Brokers) == 0 {
		return nil, fmt.Errorf("kafka: brokers are required")
	}
	if cfg.Topic == "" {
		return nil, fmt.Errorf("kafka: topic is required")
	}

	return nil, fmt.Errorf("kafka: bridge is experimental and disabled; use a supported bridge or add a real Kafka implementation")
}

// Name returns the bridge name.
func (b *KafkaBridge) Name() string {
	return b.name
}

// Send returns an explicit unsupported error.
func (b *KafkaBridge) Send(ctx context.Context, msg Message) error {
	return fmt.Errorf("kafka: bridge is experimental and disabled")
}

// Healthy returns an explicit unsupported error.
func (b *KafkaBridge) Healthy(ctx context.Context) error {
	return fmt.Errorf("kafka: bridge is experimental and disabled")
}

// Close is a no-op because no Kafka resources are acquired.
func (b *KafkaBridge) Close() error {
	return nil
}
