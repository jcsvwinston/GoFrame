package nucleus

import (
	"testing"
)

func TestNew(t *testing.T) {
	builder := New()
	if builder == nil {
		t.Fatal("New() returned nil")
	}
	if builder.config.Port == 0 {
		t.Error("Expected default port to be set")
	}
	if builder.logger == nil {
		t.Error("Expected logger to be initialized")
	}
}

func TestAppBuilder_Port(t *testing.T) {
	builder := New()
	result := builder.Port(8080)
	if result == nil {
		t.Fatal("Port() returned nil")
	}
	if result.config.Port != 8080 {
		t.Errorf("Expected port 8080, got %d", result.config.Port)
	}
}

func TestAppBuilder_Host(t *testing.T) {
	builder := New()
	result := builder.Host("localhost")
	if result == nil {
		t.Fatal("Host() returned nil")
	}
	if result.config.Host != "localhost" {
		t.Errorf("Expected host localhost, got %s", result.config.Host)
	}
}

func TestAppBuilder_SQLite(t *testing.T) {
	builder := New()
	result := builder.SQLite(":memory:")
	if result == nil {
		t.Fatal("SQLite() returned nil")
	}
	dbConfig, ok := result.config.Databases["default"]
	if !ok {
		t.Fatal("SQLite() did not set default database")
	}
	if dbConfig.URL != "sqlite://:memory:" {
		t.Errorf("Expected sqlite://:memory:, got %s", dbConfig.URL)
	}
}

func TestAppBuilder_Postgres(t *testing.T) {
	builder := New()
	result := builder.Postgres("postgres://localhost:5432/db")
	if result == nil {
		t.Fatal("Postgres() returned nil")
	}
	dbConfig, ok := result.config.Databases["default"]
	if !ok {
		t.Fatal("Postgres() did not set default database")
	}
	if dbConfig.URL != "postgres://localhost:5432/db" {
		t.Errorf("Expected postgres://localhost:5432/db, got %s", dbConfig.URL)
	}
}

func TestAppBuilder_MySQL(t *testing.T) {
	builder := New()
	result := builder.MySQL("mysql://localhost:3306/db")
	if result == nil {
		t.Fatal("MySQL() returned nil")
	}
	dbConfig, ok := result.config.Databases["default"]
	if !ok {
		t.Fatal("MySQL() did not set default database")
	}
	if dbConfig.URL != "mysql://localhost:3306/db" {
		t.Errorf("Expected mysql://localhost:3306/db, got %s", dbConfig.URL)
	}
}

func TestAppBuilder_WithAdmin(t *testing.T) {
	builder := New()
	result := builder.WithAdmin("/admin")
	if result == nil {
		t.Fatal("WithAdmin() returned nil")
	}
	if result.config.AdminPrefix != "/admin" {
		t.Errorf("Expected /admin, got %s", result.config.AdminPrefix)
	}
}

func TestAppBuilder_Templates(t *testing.T) {
	builder := New()
	result := builder.Templates("./templates")
	if result == nil {
		t.Fatal("Templates() returned nil")
	}
	if result.config.TemplatesDir != "./templates" {
		t.Errorf("Expected ./templates, got %s", result.config.TemplatesDir)
	}
}

func TestAppBuilder_Static(t *testing.T) {
	builder := New()
	result := builder.Static("./static")
	if result == nil {
		t.Fatal("Static() returned nil")
	}
	if result.config.StaticRoot != "./static" {
		t.Errorf("Expected ./static, got %s", result.config.StaticRoot)
	}
	if result.config.StaticPrefix != "/static/" {
		t.Errorf("Expected /static/, got %s", result.config.StaticPrefix)
	}
}

func TestAppBuilder_Provide(t *testing.T) {
	builder := New()
	svc := "test-service"
	result := builder.Provide(svc)
	if result == nil {
		t.Fatal("Provide() returned nil")
	}
	if len(result.providers) != 1 {
		t.Errorf("Expected 1 provider, got %d", len(result.providers))
	}
}

func TestAppBuilder_Model(t *testing.T) {
	builder := New()
	model := struct{ Name string }{}
	result := builder.Model(model)
	if result == nil {
		t.Fatal("Model() returned nil")
	}
	if len(result.models) != 1 {
		t.Errorf("Expected 1 model, got %d", len(result.models))
	}
}

func TestAppBuilder_AutoMigrate(t *testing.T) {
	builder := New()
	result := builder.AutoMigrate()
	if result == nil {
		t.Fatal("AutoMigrate() returned nil")
	}
}

func TestAppBuilder_WithConfig(t *testing.T) {
	builder := New()
	result := builder.WithConfigAny(func(cfg interface{}) {
		// This test just verifies the chaining works
		// The actual type is *app.Config, but we can't use it directly without circular imports
	})
	if result == nil {
		t.Fatal("WithConfigAny() returned nil")
	}
}

func TestAppBuilder_Config(t *testing.T) {
	builder := New()
	cfg := builder.Config()
	if cfg == nil {
		t.Fatal("Config() returned nil")
	}
}

func TestAppBuilder_Logger(t *testing.T) {
	builder := New()
	logger := builder.Logger()
	if logger == nil {
		t.Fatal("Logger() returned nil")
	}
}

func TestCorsAllowAll(t *testing.T) {
	cfg := CorsAllowAll()
	if !cfg.AllowAll {
		t.Error("Expected AllowAll to be true")
	}
}

func TestSPAConfig(t *testing.T) {
	cfg := SPAConfig{
		IndexFile: "index.html",
		APIPrefix: "/api",
	}
	if cfg.IndexFile != "index.html" {
		t.Errorf("Expected index.html, got %s", cfg.IndexFile)
	}
	if cfg.APIPrefix != "/api" {
		t.Errorf("Expected /api, got %s", cfg.APIPrefix)
	}
}
