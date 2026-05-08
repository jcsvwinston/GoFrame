package plugins

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestExecuteRequestSuccess(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-based executable test is unix-only")
	}

	// Skip on macOS due to potential process execution restrictions in test environments
	if runtime.GOOS == "darwin" && os.Getenv("CI") != "true" {
		t.Skip("skipping on macOS outside CI due to process execution restrictions")
	}

	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "goframe-plugin-success")
	writePluginRuntimeExecutable(t, pluginPath, `#!/bin/sh
cat >/dev/null
echo '{"version":"v1","ok":true,"output":{"accepted":true}}'
exit 0
`)

	request, err := NewRequestEnvelope(
		"sendgrid",
		CapabilityMailSend,
		5*time.Second,
		MailSendPayload{
			From:    "noreply@example.com",
			To:      []string{"dev@example.com"},
			Subject: "hello",
			Body:    "world",
		},
		map[string]string{"env": "test"},
	)
	if err != nil {
		t.Fatalf("NewRequestEnvelope failed: %v", err)
	}

	response, err := ExecuteRequest(context.Background(), pluginPath, request, 5*time.Second)
	if err != nil {
		t.Fatalf("ExecuteRequest failed: %v", err)
	}
	if !response.OK {
		t.Fatalf("expected ok response, got: %+v", response)
	}

	output, err := DecodeMailSendOutput(response.Output)
	if err != nil {
		t.Fatalf("DecodeMailSendOutput failed: %v", err)
	}
	if !output.Accepted {
		t.Fatalf("expected accepted=true output, got: %+v", output)
	}
}

func TestExecuteRequestErrorExitCodeMapping(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-based executable test is unix-only")
	}

	// Skip on macOS due to potential process execution restrictions in test environments
	if runtime.GOOS == "darwin" && os.Getenv("CI") != "true" {
		t.Skip("skipping on macOS outside CI due to process execution restrictions")
	}

	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "goframe-plugin-error")
	writePluginRuntimeExecutable(t, pluginPath, `#!/bin/sh
cat >/dev/null
echo '{"version":"v1","ok":false,"retriable":true,"error":{"code":"PROVIDER_RATE_LIMIT","message":"rate limited"}}'
exit 20
`)

	request, err := NewRequestEnvelope(
		"sendgrid",
		CapabilityMailSend,
		5*time.Second,
		MailSendPayload{
			From:    "noreply@example.com",
			To:      []string{"dev@example.com"},
			Subject: "hello",
			Body:    "world",
		},
		nil,
	)
	if err != nil {
		t.Fatalf("NewRequestEnvelope failed: %v", err)
	}

	_, err = ExecuteRequest(context.Background(), pluginPath, request, 5*time.Second)
	if err == nil {
		t.Fatal("expected ExecuteRequest to fail for non-zero exit")
	}

	execErr, ok := err.(*ExecutionError)
	if !ok {
		t.Fatalf("expected ExecutionError, got: %T (%v)", err, err)
	}
	if execErr.ExitCode != ExitCodeTransient {
		t.Fatalf("expected exit code %d, got %d", ExitCodeTransient, execErr.ExitCode)
	}
	if !execErr.Retriable {
		t.Fatalf("expected retriable=true, got false: %+v", execErr)
	}
	if execErr.Code != "PROVIDER_RATE_LIMIT" {
		t.Fatalf("unexpected error code: %s", execErr.Code)
	}
}

func TestExecuteRequestTimeout(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-based executable test is unix-only")
	}

	// Skip on macOS due to potential process execution restrictions in test environments
	if runtime.GOOS == "darwin" && os.Getenv("CI") != "true" {
		t.Skip("skipping on macOS outside CI due to process execution restrictions")
	}

	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "goframe-plugin-timeout")
	writePluginRuntimeExecutable(t, pluginPath, `#!/bin/sh
sleep 2
cat >/dev/null
echo '{"version":"v1","ok":true}'
exit 0
`)

	request, err := NewRequestEnvelope("sendgrid", CapabilityMailSend, time.Second, map[string]string{"x": "y"}, nil)
	if err != nil {
		t.Fatalf("NewRequestEnvelope failed: %v", err)
	}

	_, err = ExecuteRequest(context.Background(), pluginPath, request, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected timeout error")
	}

	execErr, ok := err.(*ExecutionError)
	if !ok {
		t.Fatalf("expected ExecutionError, got: %T (%v)", err, err)
	}
	if execErr.Code != "DEADLINE_EXCEEDED" {
		t.Fatalf("unexpected timeout code: %s", execErr.Code)
	}
	if !execErr.Retriable {
		t.Fatalf("expected timeout to be retriable: %+v", execErr)
	}
}

func TestNewRequestEnvelopeAndSchemaTypes(t *testing.T) {
	payload := MailSendPayload{
		From:    "noreply@example.com",
		To:      []string{"dev@example.com"},
		Subject: "hello",
		Body:    "world",
	}
	request, err := NewRequestEnvelope("MailGun", CapabilityMailSend, 2*time.Second, payload, map[string]string{"trace_id": "abc"})
	if err != nil {
		t.Fatalf("NewRequestEnvelope failed: %v", err)
	}
	if request.Version != EnvelopeVersionV1 {
		t.Fatalf("unexpected envelope version: %s", request.Version)
	}
	if request.Provider != "mailgun" {
		t.Fatalf("expected normalized provider, got: %s", request.Provider)
	}
	if request.Capability != CapabilityMailSend {
		t.Fatalf("unexpected capability: %s", request.Capability)
	}
	if request.TimeoutMS != 2000 {
		t.Fatalf("unexpected timeout_ms: %d", request.TimeoutMS)
	}
	if strings.TrimSpace(request.RequestID) == "" {
		t.Fatal("expected request_id to be set")
	}
	if request.Metadata["trace_id"] != "abc" {
		t.Fatalf("unexpected metadata: %+v", request.Metadata)
	}

	var decoded MailSendPayload
	if err := json.Unmarshal(request.Payload, &decoded); err != nil {
		t.Fatalf("unmarshal request payload failed: %v", err)
	}
	if decoded.Subject != "hello" {
		t.Fatalf("unexpected decoded payload: %+v", decoded)
	}
}

func writePluginRuntimeExecutable(t *testing.T, path, body string) {
	t.Helper()

	if err := os.WriteFile(path, []byte(strings.TrimSpace(body)+"\n"), 0o755); err != nil {
		t.Fatalf("write plugin executable failed: %v", err)
	}
	if runtime.GOOS != "windows" {
		if err := os.Chmod(path, 0o755); err != nil {
			t.Fatalf("chmod plugin executable failed: %v", err)
		}
	}
}

func TestExecutionError(t *testing.T) {
	t.Run("Error method", func(t *testing.T) {
		err := &ExecutionError{
			ExitCode:  20,
			Retriable: true,
			Code:      "TEST_ERROR",
			Message:   "test message",
			Stderr:    "test stderr",
			Cause:     nil,
		}

		msg := err.Error()
		if !strings.Contains(msg, "exit=20") {
			t.Errorf("Expected exit code in error message: %s", msg)
		}
		if !strings.Contains(msg, "code=TEST_ERROR") {
			t.Errorf("Expected code in error message: %s", msg)
		}
		if !strings.Contains(msg, "message=test message") {
			t.Errorf("Expected message in error message: %s", msg)
		}
		if !strings.Contains(msg, "stderr=test stderr") {
			t.Errorf("Expected stderr in error message: %s", msg)
		}
	})

	t.Run("Error method with nil receiver", func(t *testing.T) {
		var err *ExecutionError
		msg := err.Error()
		if msg != "" {
			t.Errorf("Expected empty string for nil receiver, got: %s", msg)
		}
	})

	t.Run("Unwrap method", func(t *testing.T) {
		cause := &testError{msg: "underlying error"}
		err := &ExecutionError{
			Cause: cause,
		}

		unwrapped := err.Unwrap()
		if unwrapped != cause {
			t.Errorf("Expected unwrapped cause, got: %v", unwrapped)
		}
	})

	t.Run("Unwrap method with nil receiver", func(t *testing.T) {
		var err *ExecutionError
		unwrapped := err.Unwrap()
		if unwrapped != nil {
			t.Errorf("Expected nil for nil receiver, got: %v", unwrapped)
		}
	})
}

func TestDecodeResponseEnvelope(t *testing.T) {
	t.Run("valid envelope", func(t *testing.T) {
		raw := `{"version":"v1","ok":true,"output":{"result":"success"}}`
		envelope, err := decodeResponseEnvelope(raw)
		if err != nil {
			t.Fatalf("decodeResponseEnvelope failed: %v", err)
		}
		if !envelope.OK {
			t.Error("Expected OK=true")
		}
		if envelope.Version != "v1" {
			t.Errorf("Expected version v1, got %s", envelope.Version)
		}
	})

	t.Run("empty response", func(t *testing.T) {
		_, err := decodeResponseEnvelope("")
		if err == nil {
			t.Error("Expected error for empty response")
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		_, err := decodeResponseEnvelope("{invalid json}")
		if err == nil {
			t.Error("Expected error for invalid JSON")
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		_, err := decodeResponseEnvelope("   ")
		if err == nil {
			t.Error("Expected error for whitespace only")
		}
	})
}

func TestExtractExitCode(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		code := extractExitCode(nil)
		if code != ExitCodeSuccess {
			t.Errorf("Expected ExitCodeSuccess (0), got %d", code)
		}
	})

	t.Run("non-exit error", func(t *testing.T) {
		err := &testError{msg: "some error"}
		code := extractExitCode(err)
		if code != -1 {
			t.Errorf("Expected -1 for non-exit error, got %d", code)
		}
	})
}

func TestRetriableByExitCode(t *testing.T) {
	tests := []struct {
		code     int
		expected bool
	}{
		{ExitCodeTransient, true},
		{ExitCodeTimeout, true},
		{ExitCodeInternal, true},
		{ExitCodeSuccess, false},
		{ExitCodeValidation, false},
		{ExitCodeRejected, false},
		{99, false},
		{-1, false},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("exit code %d", tt.code), func(t *testing.T) {
			result := retriableByExitCode(tt.code)
			if result != tt.expected {
				t.Errorf("Expected retriable=%v for exit code %d, got %v", tt.expected, tt.code, result)
			}
		})
	}
}

func TestMergeMessage(t *testing.T) {
	tests := []struct {
		name     string
		current  string
		next     string
		expected string
	}{
		{"both non-empty", "first error", "second error", "first error; second error"},
		{"current empty", "", "second error", "second error"},
		{"next empty", "first error", "", "first error"},
		{"both empty", "", "", ""},
		{"whitespace handling", "  first  ", "  second  ", "first; second"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := mergeMessage(tt.current, tt.next)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestExecuteRequestValidation(t *testing.T) {
	t.Run("empty binary path", func(t *testing.T) {
		request := RequestEnvelope{Version: EnvelopeVersionV1}
		_, err := ExecuteRequest(context.Background(), "", request, 5*time.Second)
		if err == nil {
			t.Error("Expected error for empty binary path")
		}
	})

	t.Run("default timeout", func(t *testing.T) {
		request := RequestEnvelope{Version: EnvelopeVersionV1}
		// This will fail with "no such file" but should not fail on timeout validation
		_, err := ExecuteRequest(context.Background(), "/nonexistent", request, 0)
		if err == nil {
			t.Error("Expected error for nonexistent binary")
		}
	})

	t.Run("default version", func(t *testing.T) {
		request := RequestEnvelope{}
		// This will fail with "no such file" but should not fail on version validation
		_, err := ExecuteRequest(context.Background(), "/nonexistent", request, 5*time.Second)
		if err == nil {
			t.Error("Expected error for nonexistent binary")
		}
	})

	t.Run("nil context", func(t *testing.T) {
		request := RequestEnvelope{Version: EnvelopeVersionV1}
		_, err := ExecuteRequest(context.TODO(), "/nonexistent", request, 5*time.Second)
		if err == nil {
			t.Error("Expected error for nonexistent binary")
		}
	})
}

func TestExecuteRequestResponseOKFalse(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("shell-based executable test is unix-only")
	}

	if runtime.GOOS == "darwin" && os.Getenv("CI") != "true" {
		t.Skip("skipping on macOS outside CI due to process execution restrictions")
	}

	dir := t.TempDir()
	pluginPath := filepath.Join(dir, "goframe-plugin-false")
	writePluginRuntimeExecutable(t, pluginPath, `#!/bin/sh
cat >/dev/null
echo '{"version":"v1","ok":false,"error":{"code":"REJECTED","message":"validation failed"}}'
exit 0
`)

	request, err := NewRequestEnvelope("test", "mail.send", 5*time.Second, nil, nil)
	if err != nil {
		t.Fatalf("NewRequestEnvelope failed: %v", err)
	}

	_, err = ExecuteRequest(context.Background(), pluginPath, request, 5*time.Second)
	if err == nil {
		t.Fatal("expected error when ok=false")
	}

	execErr, ok := err.(*ExecutionError)
	if !ok {
		t.Fatalf("expected ExecutionError, got: %T", err)
	}
	if execErr.Code != "REJECTED" {
		t.Errorf("Expected code REJECTED, got %s", execErr.Code)
	}
}

// testError is a simple error type for testing
type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
