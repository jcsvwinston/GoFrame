package tasks

import "testing"

func TestInspectRuntime_WithoutRedisURL(t *testing.T) {
	snapshot := InspectRuntime("")
	if snapshot.Enabled {
		t.Fatalf("expected disabled snapshot when redis_url is empty")
	}
	if snapshot.Reason == "" {
		t.Fatalf("expected reason for disabled snapshot")
	}
}

func TestInspectRuntime_InvalidRedisURL(t *testing.T) {
	snapshot := InspectRuntime("http://localhost:6379")
	if snapshot.Enabled {
		t.Fatalf("expected disabled snapshot for invalid redis scheme")
	}
	if snapshot.Reason == "" {
		t.Fatalf("expected reason for invalid redis scheme")
	}
}

func TestOperateQueue_Validation(t *testing.T) {
	if _, err := OperateQueue("", "critical", "pause"); err == nil {
		t.Fatalf("expected error when redis url is empty")
	}
	if _, err := OperateQueue("redis://127.0.0.1:6379/0", "", "pause"); err == nil {
		t.Fatalf("expected error when queue is empty")
	}
	if _, err := OperateQueue("redis://127.0.0.1:6379/0", "critical", "unknown"); err == nil {
		t.Fatalf("expected error for unsupported action")
	}
}

func TestNormalizeQueueAction(t *testing.T) {
	tests := []struct {
		input string
		want  string
		ok    bool
	}{
		{input: "pause", want: QueueActionPause, ok: true},
		{input: "  retry-archived ", want: QueueActionRetryArchived, ok: true},
		{input: "PURGE-ARCHIVED", want: QueueActionPurgeArchived, ok: true},
		{input: "unknown", want: "unknown", ok: false},
	}

	for _, tc := range tests {
		got, ok := NormalizeQueueAction(tc.input)
		if got != tc.want || ok != tc.ok {
			t.Fatalf("NormalizeQueueAction(%q) = (%q, %t), want (%q, %t)", tc.input, got, ok, tc.want, tc.ok)
		}
	}
}

func TestSupportedQueueActions(t *testing.T) {
	got := SupportedQueueActions()
	want := []string{
		QueueActionPause,
		QueueActionUnpause,
		QueueActionRetry,
		QueueActionArchiveRetry,
		QueueActionRetryArchived,
		QueueActionPurgeArchived,
	}
	if len(got) != len(want) {
		t.Fatalf("SupportedQueueActions len = %d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("SupportedQueueActions[%d] = %q, want %q", i, got[i], want[i])
		}
	}

	got[0] = "mutated"
	fresh := SupportedQueueActions()
	if fresh[0] != QueueActionPause {
		t.Fatalf("SupportedQueueActions should return a copy, got %q", fresh[0])
	}
}
