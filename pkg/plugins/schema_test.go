package plugins

import (
	"encoding/json"
	"testing"
	"time"
)

func TestNewRequestEnvelope(t *testing.T) {
	t.Run("valid envelope", func(t *testing.T) {
		payload := map[string]string{"key": "value"}
		metadata := map[string]string{"trace": "12345"}

		envelope, err := NewRequestEnvelope("test-provider", "mail.send", 5*time.Second, payload, metadata)
		if err != nil {
			t.Fatalf("NewRequestEnvelope failed: %v", err)
		}

		if envelope.Version != EnvelopeVersionV1 {
			t.Errorf("Expected version %s, got %s", EnvelopeVersionV1, envelope.Version)
		}
		if envelope.Provider != "test-provider" {
			t.Errorf("Expected provider test-provider, got %s", envelope.Provider)
		}
		if envelope.Capability != "mail.send" {
			t.Errorf("Expected capability mail.send, got %s", envelope.Capability)
		}
		if envelope.TimeoutMS != 5000 {
			t.Errorf("Expected timeout 5000ms, got %d", envelope.TimeoutMS)
		}
		if envelope.Metadata["trace"] != "12345" {
			t.Errorf("Expected metadata trace=12345, got %s", envelope.Metadata["trace"])
		}
	})

	t.Run("empty provider", func(t *testing.T) {
		_, err := NewRequestEnvelope("", "mail.send", 5*time.Second, nil, nil)
		if err == nil {
			t.Error("Expected error for empty provider")
		}
	})

	t.Run("empty capability", func(t *testing.T) {
		_, err := NewRequestEnvelope("test-provider", "", 5*time.Second, nil, nil)
		if err == nil {
			t.Error("Expected error for empty capability")
		}
	})

	t.Run("default timeout", func(t *testing.T) {
		envelope, err := NewRequestEnvelope("test-provider", "mail.send", 0, nil, nil)
		if err != nil {
			t.Fatalf("NewRequestEnvelope failed: %v", err)
		}
		if envelope.TimeoutMS != 10000 {
			t.Errorf("Expected default timeout 10000ms, got %d", envelope.TimeoutMS)
		}
	})

	t.Run("normalization", func(t *testing.T) {
		envelope, err := NewRequestEnvelope("  Test-Provider  ", "  Mail.Send  ", 5*time.Second, nil, nil)
		if err != nil {
			t.Fatalf("NewRequestEnvelope failed: %v", err)
		}
		if envelope.Provider != "test-provider" {
			t.Errorf("Expected normalized provider, got %s", envelope.Provider)
		}
		if envelope.Capability != "mail.send" {
			t.Errorf("Expected normalized capability, got %s", envelope.Capability)
		}
	})

	t.Run("no metadata", func(t *testing.T) {
		envelope, err := NewRequestEnvelope("test-provider", "mail.send", 5*time.Second, nil, nil)
		if err != nil {
			t.Fatalf("NewRequestEnvelope failed: %v", err)
		}
		if envelope.Metadata != nil {
			t.Error("Expected nil metadata when not provided")
		}
	})

	t.Run("marshal error", func(t *testing.T) {
		// Channel is unmarshalable
		ch := make(chan int)
		_, err := NewRequestEnvelope("test-provider", "mail.send", 5*time.Second, ch, nil)
		if err == nil {
			t.Error("Expected error for unmarshalable payload")
		}
	})
}

func TestDecodeMailSendOutput(t *testing.T) {
	t.Run("valid output", func(t *testing.T) {
		raw := json.RawMessage(`{"accepted":true,"provider_request_id":"msg-123"}`)
		output, err := DecodeMailSendOutput(raw)
		if err != nil {
			t.Fatalf("DecodeMailSendOutput failed: %v", err)
		}
		if !output.Accepted {
			t.Error("Expected accepted=true")
		}
		if output.ProviderRequestID != "msg-123" {
			t.Errorf("Expected provider_request_id=msg-123, got %s", output.ProviderRequestID)
		}
	})

	t.Run("empty output", func(t *testing.T) {
		output, err := DecodeMailSendOutput(json.RawMessage{})
		if err != nil {
			t.Fatalf("DecodeMailSendOutput failed: %v", err)
		}
		if !output.Accepted {
			t.Error("Expected accepted=true for empty output")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		raw := json.RawMessage(`{invalid json}`)
		_, err := DecodeMailSendOutput(raw)
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})
}

func TestCloneMetadata(t *testing.T) {
	t.Run("clone metadata", func(t *testing.T) {
		original := map[string]string{"key1": "value1", "key2": "value2"}
		cloned := cloneMetadata(original)

		if len(cloned) != len(original) {
			t.Errorf("Expected %d entries, got %d", len(original), len(cloned))
		}
		if cloned["key1"] != "value1" {
			t.Errorf("Expected key1=value1, got %s", cloned["key1"])
		}

		// Verify it's a copy
		cloned["key1"] = "modified"
		if original["key1"] == "modified" {
			t.Error("Clone should not modify original")
		}
	})

	t.Run("empty metadata", func(t *testing.T) {
		cloned := cloneMetadata(map[string]string{})
		if len(cloned) != 0 {
			t.Errorf("Expected empty map, got %d entries", len(cloned))
		}
	})
}

func TestSchemaConstants(t *testing.T) {
	if EnvelopeVersionV1 != "v1" {
		t.Errorf("Expected EnvelopeVersionV1=v1, got %s", EnvelopeVersionV1)
	}

	if CapabilityMailSend != "mail.send" {
		t.Errorf("Expected CapabilityMailSend=mail.send, got %s", CapabilityMailSend)
	}
	if CapabilityQueuePublish != "queue.publish" {
		t.Errorf("Expected CapabilityQueuePublish=queue.publish, got %s", CapabilityQueuePublish)
	}
	if CapabilityWebhookDeliver != "webhook.deliver" {
		t.Errorf("Expected CapabilityWebhookDeliver=webhook.deliver, got %s", CapabilityWebhookDeliver)
	}

	if ExitCodeSuccess != 0 {
		t.Errorf("Expected ExitCodeSuccess=0, got %d", ExitCodeSuccess)
	}
	if ExitCodeValidation != 10 {
		t.Errorf("Expected ExitCodeValidation=10, got %d", ExitCodeValidation)
	}
	if ExitCodeTransient != 20 {
		t.Errorf("Expected ExitCodeTransient=20, got %d", ExitCodeTransient)
	}
	if ExitCodeRejected != 30 {
		t.Errorf("Expected ExitCodeRejected=30, got %d", ExitCodeRejected)
	}
	if ExitCodeTimeout != 40 {
		t.Errorf("Expected ExitCodeTimeout=40, got %d", ExitCodeTimeout)
	}
	if ExitCodeInternal != 50 {
		t.Errorf("Expected ExitCodeInternal=50, got %d", ExitCodeInternal)
	}
}

func TestRequestEnvelopeJSON(t *testing.T) {
	envelope := RequestEnvelope{
		Version:    EnvelopeVersionV1,
		RequestID:  "req-123",
		Timestamp:  "2024-01-01T00:00:00Z",
		Capability: CapabilityMailSend,
		Provider:   "test-provider",
		TimeoutMS:  5000,
		Metadata:   map[string]string{"trace": "12345"},
		Payload:    json.RawMessage(`{"test":true}`),
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded RequestEnvelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if decoded.Provider != "test-provider" {
		t.Errorf("Expected provider test-provider, got %s", decoded.Provider)
	}
}

func TestResponseEnvelopeJSON(t *testing.T) {
	envelope := ResponseEnvelope{
		Version:           EnvelopeVersionV1,
		RequestID:         "req-123",
		OK:                true,
		ProviderRequestID: "msg-456",
		Retriable:         false,
		Output:            json.RawMessage(`{"result":"success"}`),
		Error: &ResponseError{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid input",
		},
		Metrics: &ResponseMetrics{
			DurationMS: 100,
		},
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var decoded ResponseEnvelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if !decoded.OK {
		t.Error("Expected OK=true")
	}
	if decoded.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("Expected error code VALIDATION_ERROR, got %s", decoded.Error.Code)
	}
}
