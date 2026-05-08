package outbox

import (
	"context"
	"log/slog"
	"testing"
	"time"
)

func TestNewManagedOutbox(t *testing.T) {
	t.Run("nil database", func(t *testing.T) {
		_, err := NewManagedOutbox(ManagedConfig{})
		if err == nil {
			t.Error("expected error for nil database")
		}
	})

	t.Run("valid config", func(t *testing.T) {
		db := openOutboxTestDB(t)
		managed, err := NewManagedOutbox(ManagedConfig{
			DB:        db,
			Flavor:    FlavorSQLite,
			TableName: "goframe_outbox",
			Logger:    slog.Default(),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if managed == nil {
			t.Error("expected managed outbox to be created")
		}
		if managed.store == nil {
			t.Error("expected store to be initialized")
		}
		if managed.dispatcher == nil {
			t.Error("expected dispatcher to be initialized")
		}
		if managed.registry == nil {
			t.Error("expected registry to be initialized")
		}
		if managed.router == nil {
			t.Error("expected router to be initialized")
		}
	})

	t.Run("nil logger uses default", func(t *testing.T) {
		db := openOutboxTestDB(t)
		managed, err := NewManagedOutbox(ManagedConfig{
			DB:     db,
			Flavor: FlavorSQLite,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if managed.logger == nil {
			t.Error("expected default logger to be set")
		}
	})
}

func TestManagedOutbox_RegisterBridge(t *testing.T) {
	db := openOutboxTestDB(t)
	managed, _ := NewManagedOutbox(ManagedConfig{
		DB:     db,
		Flavor: FlavorSQLite,
	})

	bridge := &testBridge{name: "test"}
	err := managed.RegisterBridge(bridge)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	bridges := managed.registry.List()
	if len(bridges) != 1 {
		t.Errorf("expected 1 bridge, got %d", len(bridges))
	}
}

func TestManagedOutbox_AddRoute(t *testing.T) {
	db := openOutboxTestDB(t)
	managed, _ := NewManagedOutbox(ManagedConfig{
		DB:     db,
		Flavor: FlavorSQLite,
	})

	managed.AddRoute("billing.*", "kafka", "webhook")
	// Should not panic
}

func TestManagedOutbox_Enqueue(t *testing.T) {
	db := openOutboxTestDB(t)
	managed, _ := NewManagedOutbox(ManagedConfig{
		DB:     db,
		Flavor: FlavorSQLite,
	})

	msg, err := managed.Enqueue(context.Background(), Entry{
		Topic:   "test.topic",
		Payload: map[string]any{"key": "value"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if msg.ID == "" {
		t.Error("expected message ID to be set")
	}
	if msg.Status != StatusPending {
		t.Errorf("expected pending status, got %s", msg.Status)
	}
}

func TestManagedOutbox_EnqueueTx(t *testing.T) {
	db := openOutboxTestDB(t)
	managed, _ := NewManagedOutbox(ManagedConfig{
		DB:     db,
		Flavor: FlavorSQLite,
	})

	tx, _ := db.BeginTx(context.Background(), nil)
	msg, err := managed.EnqueueTx(context.Background(), tx, Entry{
		Topic:   "test.topic",
		Payload: map[string]any{"key": "value"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit failed: %v", err)
	}

	if msg.ID == "" {
		t.Error("expected message ID to be set")
	}
}

func TestManagedOutbox_Snapshot(t *testing.T) {
	db := openOutboxTestDB(t)
	managed, _ := NewManagedOutbox(ManagedConfig{
		DB:     db,
		Flavor: FlavorSQLite,
	})

	snapshot := managed.Snapshot(context.Background())
	if !snapshot.Enabled {
		t.Error("expected snapshot to be enabled")
	}
}

func TestManagedOutbox_StartStop(t *testing.T) {
	db := openOutboxTestDB(t)
	managed, _ := NewManagedOutbox(ManagedConfig{
		DB:           db,
		Flavor:       FlavorSQLite,
		LeaseOwner:   "test-node",
		PollInterval: 10 * time.Millisecond,
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err := managed.Start(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Starting again should fail
	err = managed.Start(ctx)
	if err == nil {
		t.Error("expected error when starting already running managed outbox")
	}

	// Stop should succeed
	err = managed.Stop(ctx)
	if err != nil {
		t.Fatalf("unexpected error stopping: %v", err)
	}

	// Stopping again should succeed (idempotent)
	err = managed.Stop(ctx)
	if err != nil {
		t.Fatalf("unexpected error stopping again: %v", err)
	}
}

func TestManagedOutbox_Store(t *testing.T) {
	db := openOutboxTestDB(t)
	managed, _ := NewManagedOutbox(ManagedConfig{
		DB:     db,
		Flavor: FlavorSQLite,
	})

	store := managed.Store()
	if store == nil {
		t.Error("expected store to be returned")
	}
}

func TestManagedOutbox_Registry(t *testing.T) {
	db := openOutboxTestDB(t)
	managed, _ := NewManagedOutbox(ManagedConfig{
		DB:     db,
		Flavor: FlavorSQLite,
	})

	registry := managed.Registry()
	if registry == nil {
		t.Error("expected registry to be returned")
	}
}

func TestManagedOutbox_Router(t *testing.T) {
	db := openOutboxTestDB(t)
	managed, _ := NewManagedOutbox(ManagedConfig{
		DB:     db,
		Flavor: FlavorSQLite,
	})

	router := managed.Router()
	if router == nil {
		t.Error("expected router to be returned")
	}
}

// testBridge is a minimal bridge implementation for testing
type testBridge struct {
	name string
}

func (b *testBridge) Name() string {
	return b.name
}

func (b *testBridge) Send(ctx context.Context, msg Message) error {
	return nil
}

func (b *testBridge) Healthy(ctx context.Context) error {
	return nil
}

func (b *testBridge) Close() error {
	return nil
}
