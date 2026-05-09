package memoryprovider

import (
	"context"
	"testing"
	"time"

	"github.com/jcsvwinston/nucleus/pkg/tasks"
)

func TestNewScheduler(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)
	cfg := SchedulerConfig{
		Manager:  manager,
		Location: time.UTC,
	}

	scheduler, err := NewScheduler(cfg)
	if err != nil {
		t.Fatalf("NewScheduler failed: %v", err)
	}
	if scheduler == nil {
		t.Fatal("Expected non-nil scheduler")
	}
	if scheduler.manager != manager {
		t.Error("Expected manager to be set")
	}
}

func TestNewSchedulerNilManager(t *testing.T) {
	cfg := SchedulerConfig{
		Manager: nil,
	}

	_, err := NewScheduler(cfg)
	if err == nil {
		t.Error("Expected error for nil manager")
	}
}

func TestNewSchedulerDefaultLocation(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)
	cfg := SchedulerConfig{
		Manager: manager,
	}

	scheduler, err := NewScheduler(cfg)
	if err != nil {
		t.Fatalf("NewScheduler failed: %v", err)
	}
	if scheduler == nil {
		t.Fatal("Expected non-nil scheduler")
	}
}

func TestScheduler_RegisterJSON(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)
	manager.HandleFunc("test-task", func(ctx context.Context, task tasks.Task) error {
		return nil
	})

	scheduler, _ := NewScheduler(SchedulerConfig{Manager: manager})

	payload := map[string]string{"key": "value"}
	policy := tasks.DefaultEnqueuePolicy()

	id, err := scheduler.RegisterJSON("@every 1h", "test-task", payload, policy)
	if err != nil {
		t.Fatalf("RegisterJSON failed: %v", err)
	}
	if id == "" {
		t.Error("Expected non-empty entry ID")
	}
}

func TestScheduler_RegisterJSONInvalidSpec(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)
	scheduler, _ := NewScheduler(SchedulerConfig{Manager: manager})

	payload := map[string]string{"key": "value"}
	policy := tasks.DefaultEnqueuePolicy()

	_, err := scheduler.RegisterJSON("invalid-cron-spec", "test-task", payload, policy)
	if err == nil {
		t.Error("Expected error for invalid cron spec")
	}
}

func TestScheduler_Unregister(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)
	manager.HandleFunc("test-task", func(ctx context.Context, task tasks.Task) error {
		return nil
	})

	scheduler, _ := NewScheduler(SchedulerConfig{Manager: manager})

	payload := map[string]string{"key": "value"}
	policy := tasks.DefaultEnqueuePolicy()

	id, _ := scheduler.RegisterJSON("@every 1h", "test-task", payload, policy)

	err := scheduler.Unregister(id)
	if err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}
}

func TestScheduler_UnregisterNonExistent(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)
	scheduler, _ := NewScheduler(SchedulerConfig{Manager: manager})

	err := scheduler.Unregister("non-existent-id")
	if err != nil {
		t.Fatalf("Unregister failed: %v", err)
	}
}

func TestScheduler_Start(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)
	scheduler, _ := NewScheduler(SchedulerConfig{Manager: manager})

	err := scheduler.Start()
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	scheduler.Close()
}

func TestScheduler_Close(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)
	scheduler, _ := NewScheduler(SchedulerConfig{Manager: manager})

	err := scheduler.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}

func TestScheduler_NilScheduler(t *testing.T) {
	var scheduler *Scheduler

	_, err := scheduler.RegisterJSON("@every 1h", "test-task", nil, tasks.DefaultEnqueuePolicy())
	if err != ErrNilScheduler {
		t.Errorf("Expected ErrNilScheduler, got %v", err)
	}

	err = scheduler.Unregister("id")
	if err != ErrNilScheduler {
		t.Errorf("Expected ErrNilScheduler, got %v", err)
	}

	err = scheduler.Start()
	if err != ErrNilScheduler {
		t.Errorf("Expected ErrNilScheduler, got %v", err)
	}

	err = scheduler.Close()
	if err != ErrNilScheduler {
		t.Errorf("Expected ErrNilScheduler, got %v", err)
	}
}

func TestSchedulerConfig(t *testing.T) {
	manager, _ := NewManager(tasks.Config{}, nil)
	loc := time.FixedZone("UTC-8", -8*60*60)

	cfg := SchedulerConfig{
		Manager:  manager,
		Location: loc,
	}

	scheduler, err := NewScheduler(cfg)
	if err != nil {
		t.Fatalf("NewScheduler failed: %v", err)
	}
	if scheduler == nil {
		t.Fatal("Expected non-nil scheduler")
	}
}
