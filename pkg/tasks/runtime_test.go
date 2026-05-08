package tasks

import (
	"testing"
)

func TestNormalizeQueueAction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
		valid    bool
	}{
		{"pause lowercase", "pause", "pause", true},
		{"pause uppercase", "PAUSE", "pause", true},
		{"pause mixed case", "PaUsE", "pause", true},
		{"pause with spaces", "  pause  ", "pause", true},
		{"unpause", "unpause", "unpause", true},
		{"retry", "retry", "retry", true},
		{"archive-retry", "archive-retry", "archive-retry", true},
		{"retry-archived", "retry-archived", "retry-archived", true},
		{"purge-archived", "purge-archived", "purge-archived", true},
		{"invalid action", "invalid", "invalid", false},
		{"empty string", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, valid := NormalizeQueueAction(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
			if valid != tt.valid {
				t.Errorf("Expected valid=%v, got %v", tt.valid, valid)
			}
		})
	}
}

func TestSupportedQueueActions(t *testing.T) {
	actions := SupportedQueueActions()
	if len(actions) != 6 {
		t.Errorf("Expected 6 actions, got %d", len(actions))
	}

	expectedActions := []string{
		QueueActionPause,
		QueueActionUnpause,
		QueueActionRetry,
		QueueActionArchiveRetry,
		QueueActionRetryArchived,
		QueueActionPurgeArchived,
	}

	for _, expected := range expectedActions {
		found := false
		for _, action := range actions {
			if action == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected action %s not found in supported actions", expected)
		}
	}
}

func TestRuntimeSnapshot(t *testing.T) {
	snapshot := RuntimeSnapshot{
		Enabled:     true,
		GeneratedAt: "2024-01-01T00:00:00Z",
		Queues:      []RuntimeQueueSnapshot{},
		Schedules:   []RuntimeScheduleSnapshot{},
		Servers:     []RuntimeServerSnapshot{},
		Workers:     []RuntimeWorkerSnapshot{},
	}

	if !snapshot.Enabled {
		t.Error("Expected Enabled=true")
	}
	if snapshot.GeneratedAt != "2024-01-01T00:00:00Z" {
		t.Errorf("Expected generated_at=2024-01-01T00:00:00Z, got %s", snapshot.GeneratedAt)
	}
}

func TestRuntimeQueueSnapshot(t *testing.T) {
	queue := RuntimeQueueSnapshot{
		Name:      "default",
		Paused:    false,
		Size:      100,
		Pending:   50,
		Active:    10,
		Completed: 40,
	}

	if queue.Name != "default" {
		t.Errorf("Expected name=default, got %s", queue.Name)
	}
	if queue.Size != 100 {
		t.Errorf("Expected size=100, got %d", queue.Size)
	}
}

func TestRuntimeScheduleSnapshot(t *testing.T) {
	schedule := RuntimeScheduleSnapshot{
		ID:       "schedule-1",
		Spec:     "@every 1h",
		TaskType: "cleanup",
	}

	if schedule.ID != "schedule-1" {
		t.Errorf("Expected id=schedule-1, got %s", schedule.ID)
	}
	if schedule.Spec != "@every 1h" {
		t.Errorf("Expected spec=@every 1h, got %s", schedule.Spec)
	}
}

func TestRuntimeServerSnapshot(t *testing.T) {
	server := RuntimeServerSnapshot{
		ID:          "server-1",
		Host:        "localhost",
		PID:         1234,
		Status:      "running",
		Concurrency: 10,
		Queues:      map[string]int{"default": 1},
	}

	if server.ID != "server-1" {
		t.Errorf("Expected id=server-1, got %s", server.ID)
	}
	if server.PID != 1234 {
		t.Errorf("Expected pid=1234, got %d", server.PID)
	}
}

func TestRuntimeWorkerSnapshot(t *testing.T) {
	worker := RuntimeWorkerSnapshot{
		ServerID: "server-1",
		Host:     "localhost",
		PID:      1234,
		Queue:    "default",
		TaskID:   "task-1",
		TaskType: "process",
	}

	if worker.TaskID != "task-1" {
		t.Errorf("Expected task_id=task-1, got %s", worker.TaskID)
	}
	if worker.TaskType != "process" {
		t.Errorf("Expected task_type=process, got %s", worker.TaskType)
	}
}

func TestQueueActionResult(t *testing.T) {
	result := QueueActionResult{
		Enabled:     true,
		GeneratedAt: "2024-01-01T00:00:00Z",
		Queue:       "default",
		Action:      "pause",
		Applied:     true,
		Affected:    5,
		Message:     "Queue paused successfully",
	}

	if !result.Enabled {
		t.Error("Expected Enabled=true")
	}
	if result.Queue != "default" {
		t.Errorf("Expected queue=default, got %s", result.Queue)
	}
	if result.Affected != 5 {
		t.Errorf("Expected affected=5, got %d", result.Affected)
	}
}

func TestQueueActionConstants(t *testing.T) {
	if QueueActionPause != "pause" {
		t.Errorf("Expected QueueActionPause=pause, got %s", QueueActionPause)
	}
	if QueueActionUnpause != "unpause" {
		t.Errorf("Expected QueueActionUnpause=unpause, got %s", QueueActionUnpause)
	}
	if QueueActionRetry != "retry" {
		t.Errorf("Expected QueueActionRetry=retry, got %s", QueueActionRetry)
	}
	if QueueActionArchiveRetry != "archive-retry" {
		t.Errorf("Expected QueueActionArchiveRetry=archive-retry, got %s", QueueActionArchiveRetry)
	}
	if QueueActionRetryArchived != "retry-archived" {
		t.Errorf("Expected QueueActionRetryArchived=retry-archived, got %s", QueueActionRetryArchived)
	}
	if QueueActionPurgeArchived != "purge-archived" {
		t.Errorf("Expected QueueActionPurgeArchived=purge-archived, got %s", QueueActionPurgeArchived)
	}
}
