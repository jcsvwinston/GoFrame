package observe

import (
	"context"
	"log/slog"
	"testing"
)

func TestNewLogger(t *testing.T) {
	// Should not panic for any level/format combination
	for _, level := range []string{"debug", "info", "warn", "error", "invalid"} {
		for _, format := range []string{"json", "text", "invalid"} {
			l := NewLogger(level, format)
			if l == nil {
				t.Errorf("NewLogger(%q, %q) returned nil", level, format)
			}
		}
	}
}

func TestContextHelpers(t *testing.T) {
	ctx := context.Background()

	ctx = CtxWithRequestID(ctx, "req-123")
	ctx = CtxWithUserID(ctx, "user-456")
	ctx = CtxWithTraceID(ctx, "trace-789")

	if v := RequestIDFromCtx(ctx); v != "req-123" {
		t.Errorf("RequestID: expected req-123, got %s", v)
	}
	if v := UserIDFromCtx(ctx); v != "user-456" {
		t.Errorf("UserID: expected user-456, got %s", v)
	}
	if v := TraceIDFromCtx(ctx); v != "trace-789" {
		t.Errorf("TraceID: expected trace-789, got %s", v)
	}
}

func TestContextHelpers_Empty(t *testing.T) {
	ctx := context.Background()
	if v := RequestIDFromCtx(ctx); v != "" {
		t.Errorf("expected empty, got %s", v)
	}
	if v := UserIDFromCtx(ctx); v != "" {
		t.Errorf("expected empty, got %s", v)
	}
	if v := TraceIDFromCtx(ctx); v != "" {
		t.Errorf("expected empty, got %s", v)
	}
}

func TestWithContext(t *testing.T) {
	logger := NewLogger("info", "json")
	ctx := CtxWithRequestID(context.Background(), "req-1")
	enriched := WithContext(ctx, logger)
	if enriched == nil {
		t.Error("WithContext returned nil")
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"debug", "DEBUG"},
		{"DEBUG", "DEBUG"},
		{"info", "INFO"},
		{"", "INFO"},
		{"invalid", "INFO"},
		{"warn", "WARN"},
		{"warning", "WARN"},
		{"error", "ERROR"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			lvl := parseLevel(tc.input)
			if lvl.String() != tc.want {
				t.Errorf("parseLevel(%q)=%q; want %q", tc.input, lvl.String(), tc.want)
			}
		})
	}
}

func TestWithContext_AllFields(t *testing.T) {
	logger := NewLogger("info", "json")
	ctx := context.Background()
	ctx = CtxWithRequestID(ctx, "req-1")
	ctx = CtxWithUserID(ctx, "user-1")
	ctx = CtxWithTraceID(ctx, "trace-1")

	// Should not panic and should return enriched logger
	enriched := WithContext(ctx, logger)
	if enriched == nil {
		t.Fatal("expected non-nil enriched logger")
	}

	// Enriched logger should be different from original
	if enriched == logger {
		t.Error("expected enriched logger to be different from original")
	}
}

func TestWithContext_NilLogger(t *testing.T) {
	ctx := CtxWithRequestID(context.Background(), "req-1")
	// Should not panic with nil logger
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("WithContext panicked with nil logger: %v", r)
		}
	}()
	_ = WithContext(ctx, nil)
}

func TestContextHelpers_NilContext(t *testing.T) {
	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("context helpers panicked with nil context: %v", r)
		}
	}()

	// Note: context.Value on nil context panics in Go, so we test with empty context
	ctx := context.Background()
	if v := RequestIDFromCtx(ctx); v != "" {
		t.Errorf("expected empty for empty context, got %s", v)
	}
	if v := UserIDFromCtx(ctx); v != "" {
		t.Errorf("expected empty for empty context, got %s", v)
	}
	if v := TraceIDFromCtx(ctx); v != "" {
		t.Errorf("expected empty for empty context, got %s", v)
	}
}

func TestContextHelpers_Isolation(t *testing.T) {
	// Verify that different context chains don't leak values
	ctx1 := CtxWithRequestID(context.Background(), "req-1")
	ctx2 := CtxWithRequestID(context.Background(), "req-2")

	if RequestIDFromCtx(ctx1) != "req-1" {
		t.Errorf("ctx1: expected req-1, got %s", RequestIDFromCtx(ctx1))
	}
	if RequestIDFromCtx(ctx2) != "req-2" {
		t.Errorf("ctx2: expected req-2, got %s", RequestIDFromCtx(ctx2))
	}
}

func TestContextHelpers_Chaining(t *testing.T) {
	ctx := context.Background()
	ctx = CtxWithRequestID(ctx, "req-1")
	ctx = CtxWithUserID(ctx, "user-1")

	// Both values should be accessible
	if RequestIDFromCtx(ctx) != "req-1" {
		t.Errorf("expected req-1, got %s", RequestIDFromCtx(ctx))
	}
	if UserIDFromCtx(ctx) != "user-1" {
		t.Errorf("expected user-1, got %s", UserIDFromCtx(ctx))
	}
}

func TestContextHelpers_Overwrite(t *testing.T) {
	ctx := CtxWithRequestID(context.Background(), "req-1")
	ctx = CtxWithRequestID(ctx, "req-2")

	// Last write should win
	if RequestIDFromCtx(ctx) != "req-2" {
		t.Errorf("expected req-2 (overwrite), got %s", RequestIDFromCtx(ctx))
	}
}

func TestNewLogger_EnabledLevels(t *testing.T) {
	// Debug logger should log debug messages
	debugLogger := NewLogger("debug", "json")
	if !debugLogger.Enabled(nil, slog.LevelDebug) {
		t.Error("debug logger should enable debug level")
	}

	// Error logger should NOT log debug messages
	errorLogger := NewLogger("error", "json")
	if errorLogger.Enabled(nil, slog.LevelDebug) {
		t.Error("error logger should NOT enable debug level")
	}
}
