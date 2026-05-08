package plugins

import (
	"context"
	"testing"
)

func TestLocalHost(t *testing.T) {
	host := LocalHost{}

	t.Run("CollectInventory", func(t *testing.T) {
		result := host.CollectInventory("", []string{"noop"}, DefaultProbeTimeout)
		if result == nil {
			t.Error("Expected non-nil result")
		}
	})

	t.Run("ProbeCapabilities", func(t *testing.T) {
		_, err := host.ProbeCapabilities(context.Background(), "/nonexistent", DefaultProbeTimeout)
		if err == nil {
			t.Error("Expected error for non-existent binary")
		}
	})

	t.Run("ExecuteRequest", func(t *testing.T) {
		request := RequestEnvelope{Version: EnvelopeVersionV1}
		_, err := host.ExecuteRequest(context.Background(), "/nonexistent", request, DefaultProbeTimeout)
		if err == nil {
			t.Error("Expected error for non-existent binary")
		}
	})
}

func TestHostInterface(t *testing.T) {
	// Compile-time check that LocalHost implements Host interface
	var _ Host = LocalHost{}
}
