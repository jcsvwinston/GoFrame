package memoryprovider

import (
	"context"
	"testing"
	"time"

	"log/slog"

	"github.com/jcsvwinston/GoFrame/pkg/tasks"
)

func TestNewManager(t *testing.T) {
	logger := slog.Default()
	cfg := tasks.Config{Concurrency: 5}

	manager, err := NewManager(cfg, logger)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	if manager == nil {
		t.Fatal("Expected non-nil manager")
	}
	if manager.concurrency != 5 {
		t.Errorf("Expected concurrency=5, got %d", manager.concurrency)
	}
	if manager.logger == nil {
		t.Error("Expected non-nil logger")
	}
}

func TestNewManagerDefaults(t *testing.T) {
	cfg := tasks.Config{} // Concurrency = 0
	manager, err := NewManager(cfg, nil)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	if manager.concurrency != 10 {
		t.Errorf("Expected default concurrency=10, got %d", manager.concurrency)
	}
	if manager.logger == nil {
		t.Error("Expected default logger")
	}
}

func TestManager_HandleFunc(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)

	t.Run("valid handler", func(t *testing.T) {
		handler := func(ctx context.Context, task tasks.Task) error {
			return nil
		}
		err := manager.HandleFunc("test-task", handler)
		if err != nil {
			t.Fatalf("HandleFunc failed: %v", err)
		}
	})

	t.Run("empty task type", func(t *testing.T) {
		handler := func(ctx context.Context, task tasks.Task) error {
			return nil
		}
		err := manager.HandleFunc("", handler)
		if err != ErrTaskTypeRequired {
			t.Errorf("Expected ErrTaskTypeRequired, got %v", err)
		}
	})

	t.Run("nil handler", func(t *testing.T) {
		err := manager.HandleFunc("test-task", nil)
		if err != ErrNilHandler {
			t.Errorf("Expected ErrNilHandler, got %v", err)
		}
	})
}

func TestManager_EnqueueJSON(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)
	manager.HandleFunc("test-task", func(ctx context.Context, task tasks.Task) error {
		return nil
	})

	t.Run("valid enqueue", func(t *testing.T) {
		payload := map[string]string{"key": "value"}
		id, err := manager.EnqueueJSON("test-task", payload)
		if err != nil {
			t.Fatalf("EnqueueJSON failed: %v", err)
		}
		if id == "" {
			t.Error("Expected non-empty task ID")
		}
	})

	t.Run("empty task type", func(t *testing.T) {
		payload := map[string]string{"key": "value"}
		_, err := manager.EnqueueJSON("", payload)
		if err != ErrTaskTypeRequired {
			t.Errorf("Expected ErrTaskTypeRequired, got %v", err)
		}
	})

	t.Run("invalid payload", func(t *testing.T) {
		// Channel is unmarshalable
		ch := make(chan int)
		_, err := manager.EnqueueJSON("test-task", ch)
		if err == nil {
			t.Error("Expected error for unmarshalable payload")
		}
	})
}

func TestManager_EnqueueJSONWithPolicy(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)
	manager.HandleFunc("test-task", func(ctx context.Context, task tasks.Task) error {
		return nil
	})

	t.Run("with delay", func(t *testing.T) {
		payload := map[string]string{"key": "value"}
		policy := tasks.EnqueuePolicy{
			ProcessIn: 100 * time.Millisecond,
		}
		id, err := manager.EnqueueJSONWithPolicy("test-task", payload, policy)
		if err != nil {
			t.Fatalf("EnqueueJSONWithPolicy failed: %v", err)
		}
		if id == "" {
			t.Error("Expected non-empty task ID")
		}
	})
}

func TestManager_Run(t *testing.T) {
	manager, _ := NewManager(tasks.Config{Concurrency: 2}, nil)

	called := false
	manager.HandleFunc("test-task", func(ctx context.Context, task tasks.Task) error {
		called = true
		return nil
	})

	ctx, cancel := context.WithCancel(context.Background())

	// Start manager in background
	go func() {
		_ = manager.Run(ctx)
	}()

	// Enqueue a task
	_, _ = manager.EnqueueJSON("test-task", map[string]string{"key": "value"})

	// Give worker time to process
	time.Sleep(100 * time.Millisecond)

	cancel()
	manager.Close()

	if !called {
		t.Error("Expected handler to be called")
	}
}

func TestManager_RunAlreadyRunning(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Start manager
	go func() {
		_ = manager.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)

	// Try to run again
	err := manager.Run(ctx)
	if err == nil {
		t.Error("Expected error when manager is already running")
	}

	cancel()
	manager.Close()
}

func TestManager_Close(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)

	err := manager.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestManager_EnqueueJSONCtx(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)
	manager.HandleFunc("test-task", func(ctx context.Context, task tasks.Task) error {
		return nil
	})

	ctx := context.Background()
	payload := map[string]string{"key": "value"}
	id, err := manager.EnqueueJSONCtx(ctx, "test-task", payload)
	if err != nil {
		t.Fatalf("EnqueueJSONCtx failed: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty task ID")
	}
}

func TestTask(t *testing.T) {
	task := &Task{
		taskType: "test-type",
		payload:  []byte(`{"key":"value"}`),
	}

	if task.Type() != "test-type" {
		t.Errorf("Expected test-type, got %s", task.Type())
	}

	payload := task.Payload()
	if string(payload) != `{"key":"value"}` {
		t.Errorf("Expected payload, got %s", string(payload))
	}
}

func TestManager_Stats(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)

	processed := manager.processed.Load()
	failed := manager.failed.Load()

	if processed != 0 {
		t.Errorf("Expected processed=0, got %d", processed)
	}
	if failed != 0 {
		t.Errorf("Expected failed=0, got %d", failed)
	}
}
