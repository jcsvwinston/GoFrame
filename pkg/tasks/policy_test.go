package tasks

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/hibiken/asynq"
)

func TestDefaultEnqueuePolicy(t *testing.T) {
	policy := DefaultEnqueuePolicy()
	if policy.MaxRetry != -1 {
		t.Fatalf("DefaultEnqueuePolicy MaxRetry = %d, want -1", policy.MaxRetry)
	}
}

func TestEnqueuePolicyValidate(t *testing.T) {
	tests := []EnqueuePolicy{
		{MaxRetry: -2},
		{MaxRetry: -1, Timeout: -time.Second},
		{MaxRetry: -1, ProcessIn: -time.Second},
		{MaxRetry: -1, Retention: -time.Second},
	}

	for _, tc := range tests {
		if err := tc.Validate(); err == nil {
			t.Fatalf("expected validation error for %#v", tc)
		}
	}
}

func TestManagerEnqueueJSONCtxWithPolicy(t *testing.T) {
	redisServer := miniredis.RunT(t)

	manager, err := NewManager(Config{RedisURL: "redis://" + redisServer.Addr()}, nil)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer manager.Close()

	policy := DefaultEnqueuePolicy()
	policy.Queue = "critical"
	policy.MaxRetry = 0
	policy.Timeout = 2 * time.Minute
	policy.ProcessIn = 30 * time.Second
	policy.Retention = 5 * time.Minute

	info, err := manager.EnqueueJSONCtxWithPolicy(context.Background(), "emails.send_welcome", map[string]string{
		"email": "alice@example.com",
	}, policy)
	if err != nil {
		t.Fatalf("EnqueueJSONCtxWithPolicy failed: %v", err)
	}

	if info.Queue != "critical" {
		t.Fatalf("queue = %q, want critical", info.Queue)
	}

	inspector, err := newInspectorForTest("redis://" + redisServer.Addr())
	if err != nil {
		t.Fatalf("newInspectorForTest failed: %v", err)
	}
	defer inspector.Close()

	got, err := inspector.GetTaskInfo("critical", info.ID)
	if err != nil {
		t.Fatalf("GetTaskInfo failed: %v", err)
	}

	if got.Type != "emails.send_welcome" {
		t.Fatalf("type = %q, want emails.send_welcome", got.Type)
	}
	if got.MaxRetry != 0 {
		t.Fatalf("max retry = %d, want 0", got.MaxRetry)
	}
	if got.Timeout != 2*time.Minute {
		t.Fatalf("timeout = %s, want 2m", got.Timeout)
	}
	if got.Retention != 5*time.Minute {
		t.Fatalf("retention = %s, want 5m", got.Retention)
	}
	if got.State != asynq.TaskStateScheduled {
		t.Fatalf("state = %q, want scheduled", got.State)
	}
	if got.NextProcessAt.IsZero() {
		t.Fatal("expected scheduled task to have next process time")
	}
}

func TestManagerEnqueueJSONCtxWithPolicyValidation(t *testing.T) {
	mgr := &Manager{}
	_, err := mgr.EnqueueJSONCtxWithPolicy(context.Background(), "test.task", map[string]string{"key": "value"}, EnqueuePolicy{MaxRetry: -2})
	if err == nil {
		t.Fatal("expected policy validation error")
	}
}

func newInspectorForTest(redisURL string) (*asynq.Inspector, error) {
	opt, err := redisClientOptFromURL(redisURL)
	if err != nil {
		return nil, err
	}
	return asynq.NewInspector(opt), nil
}

func TestEnqueuePolicyOptions_ValidationError(t *testing.T) {
	_, err := (EnqueuePolicy{MaxRetry: -3}).Options()
	if err == nil {
		t.Fatal("expected options validation error")
	}
}

func TestManagerEnqueueJSONWithPolicy_NilManager(t *testing.T) {
	var mgr *Manager
	_, err := mgr.EnqueueJSONWithPolicy("test.task", map[string]string{"key": "value"}, DefaultEnqueuePolicy())
	if !errors.Is(err, ErrNilManager) {
		t.Fatalf("expected ErrNilManager, got %v", err)
	}
}
