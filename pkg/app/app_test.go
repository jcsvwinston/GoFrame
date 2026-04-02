package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func testAppConfig() *Config {
	return &Config{
		Host:                "127.0.0.1",
		Port:                0,
		ReadTimeout:         2 * time.Second,
		WriteTimeout:        2 * time.Second,
		IdleTimeout:         5 * time.Second,
		DatabaseURL:         "sqlite://:memory:",
		DatabaseMaxOpen:     1,
		DatabaseMaxIdle:     1,
		DatabaseMaxLifetime: time.Minute,
		LogLevel:            "error",
		LogFormat:           "text",
		AdminPrefix:         "/admin",
		AdminTitle:          "Test Admin",
	}
}

func TestAppNew_NilConfig(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
	if !errors.Is(err, ErrNilConfig) {
		t.Fatalf("expected ErrNilConfig, got: %v", err)
	}
}

func TestAppNew_InitializesCoreComponents(t *testing.T) {
	a, err := New(testAppConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer a.Shutdown(context.Background())

	if a.Config == nil || a.Logger == nil || a.Router == nil || a.DB == nil || a.Models == nil {
		t.Fatal("expected app core components to be initialized")
	}
	if a.Mailer == nil {
		t.Fatal("expected mailer to be initialized")
	}
	if a.Admin == nil {
		t.Fatal("expected admin panel to be initialized")
	}
	if err := a.DB.Health(context.Background()); err != nil {
		t.Fatalf("expected DB health to pass, got: %v", err)
	}
}

func TestAppNew_BunEngine_InitializesAdmin(t *testing.T) {
	cfg := testAppConfig()
	cfg.DatabaseEngine = "bun"

	a, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer a.Shutdown(context.Background())

	if a.DB == nil {
		t.Fatal("expected DB to be initialized")
	}
	if a.DB.Engine() != "bun" {
		t.Fatalf("expected db engine bun, got %s", a.DB.Engine())
	}
	if a.Admin == nil {
		t.Fatal("expected admin to be initialized when bun engine is selected")
	}
}

func TestAppRegisterModel(t *testing.T) {
	type User struct {
		ID    uint
		Email string
	}

	a, err := New(testAppConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer a.Shutdown(context.Background())

	if err := a.RegisterModel(&User{}); err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}
	if a.Models.Count() != 1 {
		t.Fatalf("expected 1 model registered, got %d", a.Models.Count())
	}
}

func TestAppShutdown_ReverseHookOrderAndErrorAggregation(t *testing.T) {
	a := &App{Config: testAppConfig()}
	var order []int

	a.OnShutdown(func(context.Context) error {
		order = append(order, 1)
		return nil
	})
	a.OnShutdown(func(context.Context) error {
		order = append(order, 2)
		return errors.New("hook two failed")
	})
	a.OnShutdown(func(context.Context) error {
		order = append(order, 3)
		return nil
	})

	err := a.Shutdown(context.Background())
	if err == nil {
		t.Fatal("expected aggregated shutdown error")
	}
	if !strings.Contains(err.Error(), "hook two failed") {
		t.Fatalf("expected hook error in shutdown error, got: %v", err)
	}

	got := fmt.Sprint(order)
	if got != "[3 2 1]" {
		t.Fatalf("expected reverse hook order [3 2 1], got %s", got)
	}
}

func TestAppMethods_NilReceiver(t *testing.T) {
	var a *App

	if err := a.Run(context.Background()); !errors.Is(err, ErrNilApp) {
		t.Fatalf("Run: expected ErrNilApp, got %v", err)
	}
	if err := a.Shutdown(context.Background()); !errors.Is(err, ErrNilApp) {
		t.Fatalf("Shutdown: expected ErrNilApp, got %v", err)
	}
	if err := a.MountAdmin(); !errors.Is(err, ErrNilApp) {
		t.Fatalf("MountAdmin: expected ErrNilApp, got %v", err)
	}
	if err := a.RegisterModel(&struct{ ID uint }{}); !errors.Is(err, ErrNilApp) {
		t.Fatalf("RegisterModel: expected ErrNilApp, got %v", err)
	}
}

func TestAppRun_NotInitialized(t *testing.T) {
	a := &App{}
	err := a.Run(context.Background())
	if !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("expected ErrNotInitialized, got %v", err)
	}
}

func TestAppRun_ContextCancel(t *testing.T) {
	a, err := New(testAppConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(120 * time.Millisecond)
		cancel()
	}()

	if err := a.Run(ctx); err != nil {
		t.Fatalf("Run should exit cleanly on context cancel: %v", err)
	}
}

func TestAppRun_InvalidAddress(t *testing.T) {
	cfg := testAppConfig()
	cfg.Port = -1 // produces an invalid host:port address

	a, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating app: %v", err)
	}
	defer a.Shutdown(context.Background())

	err = a.Run(context.Background())
	if err == nil {
		t.Fatal("expected run error for invalid server address")
	}
}

func TestAppNew_InvalidMailDriver(t *testing.T) {
	cfg := testAppConfig()
	cfg.MailDriver = "unknown-provider"

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for unknown mail driver")
	}
	if !strings.Contains(err.Error(), "unknown mail driver") {
		t.Fatalf("expected unknown mail driver error, got %v", err)
	}
}
