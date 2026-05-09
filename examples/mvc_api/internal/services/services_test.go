package services

import (
	"context"
	"testing"

	"github.com/jcsvwinston/nucleus/pkg/app"
)

func TestServices_EnsureSchema(t *testing.T) {
	cfg := &app.Config{
		DatabaseDefault: "default",
		Databases: map[string]app.DatabaseConfig{
			"default": {
				URL: "sqlite://:memory:",
			},
		},
		LogLevel: "error",
	}

	a, err := app.New(cfg)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}
	defer a.Shutdown(context.Background())

	svc, err := New(a)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// Verify tables exist by counting rows (should not error)
	count := svc.CountRows("articles")
	if count < 0 {
		t.Errorf("CountRows returned negative: %d", count)
	}
}

func TestServices_CountRows(t *testing.T) {
	cfg := &app.Config{
		DatabaseDefault: "default",
		Databases: map[string]app.DatabaseConfig{
			"default": {
				URL: "sqlite://:memory:",
			},
		},
		LogLevel: "error",
	}

	a, err := app.New(cfg)
	if err != nil {
		t.Fatalf("failed to create app: %v", err)
	}
	defer a.Shutdown(context.Background())

	svc, err := New(a)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	// The seed should have created at least 1 article
	count := svc.CountRows("articles")
	if count != 1 {
		t.Errorf("CountRows() = %d, want 1 (seed data)", count)
	}
}
