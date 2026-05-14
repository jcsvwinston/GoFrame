package observe

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"
)

// logLine renders one log record through newLogger (JSON) into a buffer
// and returns the parsed object — the test's window onto the actual
// rendered output, redaction included.
func logLine(t *testing.T, cfg RedactionConfig, msg string, args ...any) map[string]any {
	t.Helper()
	var buf bytes.Buffer
	l := newLogger(&buf, "debug", "json", cfg)
	l.Info(msg, args...)
	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal log line %q: %v", buf.String(), err)
	}
	return out
}

func TestRedaction_DefaultKeysRedacted(t *testing.T) {
	line := logLine(t, RedactionConfig{}, "login attempt",
		"password", "hunter2",
		"authorization", "Bearer abc.def.ghi",
		"token", "secret-token-value",
		"api_key", "sk_live_xxx",
		"cookie", "session=abc",
	)
	for _, key := range []string{"password", "authorization", "token", "api_key", "cookie"} {
		if got := line[key]; got != RedactionPlaceholder {
			t.Errorf("key %q = %v, want %q", key, got, RedactionPlaceholder)
		}
	}
}

func TestRedaction_NonSensitiveKeysUntouched(t *testing.T) {
	line := logLine(t, RedactionConfig{}, "request done",
		"user_id", "u-123",
		"request_id", "r-456",
		"status", 200,
		"path", "/api/widgets",
	)
	if line["user_id"] != "u-123" || line["request_id"] != "r-456" || line["path"] != "/api/widgets" {
		t.Fatalf("non-sensitive keys must pass through unchanged: %+v", line)
	}
	// The built-in attrs survive too.
	if line["msg"] != "request done" || line["level"] != "INFO" {
		t.Fatalf("built-in attrs must pass through: %+v", line)
	}
}

func TestRedaction_CaseInsensitiveKeyMatch(t *testing.T) {
	line := logLine(t, RedactionConfig{}, "m",
		"Authorization", "Bearer x",
		"PASSWORD", "x",
		"Api_Key", "x",
	)
	for _, key := range []string{"Authorization", "PASSWORD", "Api_Key"} {
		if line[key] != RedactionPlaceholder {
			t.Errorf("key %q should be redacted regardless of case, got %v", key, line[key])
		}
	}
}

func TestRedaction_ValueTypeIndependence(t *testing.T) {
	// A sensitive key holding a non-string value is still redacted to the
	// placeholder string — the original type is irrelevant.
	line := logLine(t, RedactionConfig{}, "m",
		"secret", 1234567,
		"token", true,
	)
	if line["secret"] != RedactionPlaceholder || line["token"] != RedactionPlaceholder {
		t.Fatalf("non-string sensitive values must be redacted: %+v", line)
	}
}

func TestRedaction_InsideGroup(t *testing.T) {
	// A sensitive key nested in a group is redacted just like a top-level
	// one — the redactor matches on the key, ignoring group nesting.
	var buf bytes.Buffer
	l := newLogger(&buf, "debug", "json", RedactionConfig{})
	l.Info("m", slog.Group("req", slog.String("password", "hunter2"), slog.String("path", "/x")))

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	req, ok := out["req"].(map[string]any)
	if !ok {
		t.Fatalf("expected a req group object, got %T", out["req"])
	}
	if req["password"] != RedactionPlaceholder {
		t.Errorf("grouped password should be redacted, got %v", req["password"])
	}
	if req["path"] != "/x" {
		t.Errorf("grouped non-sensitive key should pass through, got %v", req["path"])
	}
}

func TestRedaction_WithContextFieldsNotRedacted(t *testing.T) {
	// WithContext adds request_id / user_id / tenant_id / trace_id — none
	// are secrets and none must be redacted.
	var buf bytes.Buffer
	base := newLogger(&buf, "debug", "json", RedactionConfig{})
	ctx := CtxWithUserID(CtxWithRequestID(context.Background(), "r-1"), "u-1")
	WithContext(ctx, base).Info("m")

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["user_id"] != "u-1" || out["request_id"] != "r-1" {
		t.Fatalf("context identity fields must not be redacted: %+v", out)
	}
}

func TestRedaction_ExtraKeys(t *testing.T) {
	cfg := RedactionConfig{ExtraKeys: []string{"ssn", "Card_Number"}}
	line := logLine(t, cfg, "m",
		"ssn", "123-45-6789",
		"card_number", "4111111111111111", // matches case-insensitively
		"password", "x", // default key still redacted
		"order_id", "o-1", // not sensitive
	)
	if line["ssn"] != RedactionPlaceholder || line["card_number"] != RedactionPlaceholder {
		t.Errorf("ExtraKeys should be redacted: %+v", line)
	}
	if line["password"] != RedactionPlaceholder {
		t.Error("default keys must still be redacted when ExtraKeys is set")
	}
	if line["order_id"] != "o-1" {
		t.Error("non-sensitive key must pass through")
	}
}

func TestRedaction_CustomPlaceholder(t *testing.T) {
	line := logLine(t, RedactionConfig{Placeholder: "***"}, "m", "password", "x")
	if line["password"] != "***" {
		t.Fatalf("custom placeholder not applied: %v", line["password"])
	}
}

func TestRedaction_Disabled(t *testing.T) {
	// With redaction disabled the secret value passes through verbatim.
	line := logLine(t, RedactionConfig{Disabled: true}, "m", "password", "hunter2")
	if line["password"] != "hunter2" {
		t.Fatalf("disabled redaction must pass the value through, got %v", line["password"])
	}
	// And newRedactor returns nil so ReplaceAttr is left unset.
	if newRedactor(RedactionConfig{Disabled: true}) != nil {
		t.Fatal("newRedactor must return nil when Disabled is set")
	}
}

func TestRedaction_NewLoggerRedactsByDefault(t *testing.T) {
	// The exported NewLogger must redact — this is the security-by-default
	// guarantee. We cannot easily capture os.Stdout here, so assert via
	// the redactor the zero-value config produces.
	r := newRedactor(RedactionConfig{})
	if r == nil {
		t.Fatal("zero-value RedactionConfig must produce an active redactor")
	}
	got := r(nil, slog.String("password", "hunter2"))
	if got.Value.String() != RedactionPlaceholder {
		t.Fatalf("default redactor must redact password, got %q", got.Value.String())
	}
}

func TestDefaultRedactedKeys_SortedAndComplete(t *testing.T) {
	keys := DefaultRedactedKeys()
	if len(keys) == 0 {
		t.Fatal("DefaultRedactedKeys must not be empty")
	}
	for i := 1; i < len(keys); i++ {
		if keys[i-1] > keys[i] {
			t.Fatalf("DefaultRedactedKeys must be sorted: %q before %q", keys[i-1], keys[i])
		}
	}
	// A few keys the audit specifically called for must be present.
	want := map[string]bool{"authorization": false, "cookie": false, "password": false, "token": false, "secret": false, "api_key": false}
	for _, k := range keys {
		if _, ok := want[k]; ok {
			want[k] = true
		}
	}
	for k, found := range want {
		if !found {
			t.Errorf("DefaultRedactedKeys missing expected key %q", k)
		}
	}

	// The returned slice must be a copy — mutating it must not affect the
	// next call.
	keys[0] = "MUTATED"
	if DefaultRedactedKeys()[0] == "MUTATED" {
		t.Fatal("DefaultRedactedKeys must return a fresh copy each call")
	}
}

func TestRedaction_BuiltinAttrsPassThrough(t *testing.T) {
	line := logLine(t, RedactionConfig{}, "hello world")
	for _, k := range []string{"time", "level", "msg"} {
		if _, ok := line[k]; !ok {
			t.Errorf("built-in attr %q missing from output: %+v", k, line)
		}
	}
	msg, ok := line["msg"].(string)
	if !ok {
		t.Fatalf("msg attr is not a string: %T %v", line["msg"], line["msg"])
	}
	if !strings.Contains(msg, "hello world") {
		t.Fatalf("msg attr mangled: %v", msg)
	}
}

// TestRedaction_ExtraKeysCannotSilenceBuiltins guards the footgun where
// an operator lists a slog built-in key ("time", "msg", …) in ExtraKeys
// — that must NOT redact the timestamp or message, or it would silently
// break log pipelines that key off those fields.
func TestRedaction_ExtraKeysCannotSilenceBuiltins(t *testing.T) {
	cfg := RedactionConfig{ExtraKeys: []string{"time", "level", "msg", "source"}}
	line := logLine(t, cfg, "the message")
	if line["msg"] != "the message" {
		t.Errorf("msg must not be redactable via ExtraKeys, got %v", line["msg"])
	}
	if line["level"] != "INFO" {
		t.Errorf("level must not be redactable via ExtraKeys, got %v", line["level"])
	}
	if line["time"] == RedactionPlaceholder || line["time"] == nil {
		t.Errorf("time must not be redactable via ExtraKeys, got %v", line["time"])
	}
}

// TestRedaction_WithAttrsPathRedacted confirms attrs attached via
// logger.With(...) are also redacted — slog runs ReplaceAttr over
// With-attrs at emit time, not just over per-record attrs.
func TestRedaction_WithAttrsPathRedacted(t *testing.T) {
	var buf bytes.Buffer
	l := newLogger(&buf, "debug", "json", RedactionConfig{})
	l.With("password", "hunter2", "user_id", "u-1").Info("m")

	var out map[string]any
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out["password"] != RedactionPlaceholder {
		t.Errorf("password attached via With() must be redacted, got %v", out["password"])
	}
	if out["user_id"] != "u-1" {
		t.Errorf("non-secret With() attr must pass through, got %v", out["user_id"])
	}
}
