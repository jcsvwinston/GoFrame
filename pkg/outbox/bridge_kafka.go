package outbox

import (
	"context"
	"fmt"
)

// KafkaBridge delivers outbox messages to Apache Kafka.
//
// This is a placeholder implementation that demonstrates the interface structure.
// A production implementation would use a Kafka client library such as:
//   - github.com/segmentio/kafka-go
//   - github.com/IBM/sarama
//   - github.com/Shopify/sarama
//
// To implement a production Kafka bridge:
// 1. Import a Kafka client library
// 2. Create a Kafka writer/producer in NewKafkaBridge
// 3. Serialize the message payload in Send()
// 4. Handle delivery acknowledgments and retries
// 5. Implement proper health checks
// 6. Close the producer in Close()
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

// NewKafkaBridge creates a new Kafka bridge (placeholder).
//
// This placeholder validates the configuration but does not create a real Kafka connection.
// Returns an error if name, brokers, or topic are empty.
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

	return &KafkaBridge{
		name:    cfg.Name,
		brokers: cfg.Brokers,
		topic:   cfg.Topic,
	}, nil
}

// Name returns the bridge name.
func (b *KafkaBridge) Name() string {
	return b.name
}

// Send delivers a message to Kafka (placeholder).
//
// TODO: Implement actual Kafka delivery using a Kafka client library.
// A production implementation would:
// 1. Create a Kafka writer/producer
// 2. Serialize the message payload (likely as JSON or Avro)
// 3. Send to the configured topic with a message key
// 4. Handle delivery acknowledgments
// 5. Return an error if delivery fails
func (b *KafkaBridge) Send(ctx context.Context, msg Message) error {
	return fmt.Errorf("kafka: not yet implemented (placeholder)")
}

// Healthy checks if Kafka brokers are reachable (placeholder).
//
// TODO: Implement actual Kafka health check.
// A production implementation would:
// 1. Connect to the configured brokers
// 2. Query cluster metadata
// 3. Verify the topic exists
// 4. Check broker connectivity
func (b *KafkaBridge) Healthy(ctx context.Context) error {
	return fmt.Errorf("kafka: health check not yet implemented (placeholder)")
}

// Close closes the Kafka connection (placeholder).
//
// TODO: Close Kafka writer/producer and release resources.
func (b *KafkaBridge) Close() error {
	return nil
}
