package config

import (
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jcsvwinston/nucleus/pkg/app"
)

const (
	DemoAppUsername = "demo"
	DemoAppPassword = "demo123456"
)

func DefaultConfig() *app.Config {
	cfg := &app.Config{
		Host:            "0.0.0.0",
		Port:            8090,
		DatabaseDefault: "default",
		Databases: map[string]app.DatabaseConfig{
			"default": {
				URL:         "sqlite://examples_mvc_api.db",
				MaxOpen:     10,
				MaxIdle:     5,
				MaxLifetime: 5 * time.Minute,
			},
		},
		AdminPrefix:            "/admin",
		AdminTitle:             "Nucleus Showcase Admin",
		AdminBootstrapUsername: "admin",
		AdminBootstrapEmail:    "admin@example.com",
		AdminBootstrapPassword: "supersecret123",
		LogLevel:               "info",
		LogFormat:              "text",
	}
	applyEnvOverrides(cfg)
	return cfg
}

func applyEnvOverrides(cfg *app.Config) {
	if cfg == nil {
		return
	}
	cfg.Port = getenvInt("NUCLEUS_EXAMPLE_PORT", cfg.Port)

	if dbURL := strings.TrimSpace(os.Getenv("NUCLEUS_EXAMPLE_DB_URL")); dbURL != "" {
		if cfg.Databases == nil {
			cfg.Databases = map[string]app.DatabaseConfig{}
		}
		dbCfg := cfg.Databases["default"]
		dbCfg.URL = dbURL
		cfg.Databases["default"] = dbCfg
	}

	if redisURL := strings.TrimSpace(os.Getenv("NUCLEUS_EXAMPLE_REDIS_URL")); redisURL != "" {
		cfg.RedisURL = redisURL
	}

	if sessionStore := strings.TrimSpace(os.Getenv("NUCLEUS_EXAMPLE_SESSION_STORE")); sessionStore != "" {
		cfg.SessionStore = strings.ToLower(sessionStore)
	}
	if sessionRedisURL := strings.TrimSpace(os.Getenv("NUCLEUS_EXAMPLE_SESSION_REDIS_URL")); sessionRedisURL != "" {
		cfg.SessionRedisURL = sessionRedisURL
	}

	cfg.AdminClusterEnabled = getenvBool("NUCLEUS_EXAMPLE_ADMIN_CLUSTER_ENABLED", cfg.AdminClusterEnabled)
	if clusterRedis := strings.TrimSpace(os.Getenv("NUCLEUS_EXAMPLE_ADMIN_CLUSTER_REDIS_URL")); clusterRedis != "" {
		cfg.AdminClusterRedisURL = clusterRedis
	}
	if clusterChannel := strings.TrimSpace(os.Getenv("NUCLEUS_EXAMPLE_ADMIN_CLUSTER_CHANNEL")); clusterChannel != "" {
		cfg.AdminClusterChannel = clusterChannel
	}
	if clusterNodeID := strings.TrimSpace(os.Getenv("NUCLEUS_EXAMPLE_ADMIN_CLUSTER_NODE_ID")); clusterNodeID != "" {
		cfg.AdminClusterNodeID = clusterNodeID
	}
	if clusterToken := strings.TrimSpace(os.Getenv("NUCLEUS_EXAMPLE_ADMIN_CLUSTER_TOKEN")); clusterToken != "" {
		cfg.AdminClusterToken = clusterToken
	}
	if traceURLTemplate := strings.TrimSpace(os.Getenv("NUCLEUS_EXAMPLE_ADMIN_TRACE_URL_TEMPLATE")); traceURLTemplate != "" {
		cfg.AdminTraceURLTemplate = traceURLTemplate
	}
	if otlpEndpoint := strings.TrimSpace(os.Getenv("NUCLEUS_EXAMPLE_OTLP_ENDPOINT")); otlpEndpoint != "" {
		cfg.OTLPEndpoint = otlpEndpoint
	}
	if adminTitle := strings.TrimSpace(os.Getenv("NUCLEUS_EXAMPLE_ADMIN_TITLE")); adminTitle != "" {
		cfg.AdminTitle = adminTitle
	}
	if bootstrapPassword := strings.TrimSpace(os.Getenv("NUCLEUS_EXAMPLE_ADMIN_BOOTSTRAP_PASSWORD")); bootstrapPassword != "" {
		cfg.AdminBootstrapPassword = bootstrapPassword
	}
}

func getenvInt(name string, fallback int) int {
	raw := strings.TrimSpace(os.Getenv(name))
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func getenvBool(name string, fallback bool) bool {
	raw := strings.TrimSpace(strings.ToLower(os.Getenv(name)))
	if raw == "" {
		return fallback
	}
	switch raw {
	case "1", "true", "yes", "y", "on":
		return true
	case "0", "false", "no", "n", "off":
		return false
	default:
		return fallback
	}
}
