package quarkdb

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	"github.com/jcsvwinston/quark"
	_ "modernc.org/sqlite"
)

// Client wraps the Quark ORM client
type Client struct {
	*quark.Client
	db     *sql.DB
	logger *slog.Logger
}

// NewClient creates a new Quark client with SQLite
func NewClient(dbPath string, logger *slog.Logger) (*Client, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	client, err := quark.New(db, quark.WithDialect(quark.SQLite()))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create quark client: %w", err)
	}

	return &Client{
		Client: client,
		db:     db,
		logger: logger,
	}, nil
}

// Close closes the database connection
func (c *Client) Close() error {
	return c.db.Close()
}

// DB returns the underlying sql.DB for compatibility
func (c *Client) DB() *sql.DB {
	return c.db
}

// Migrate runs auto-migration for the given models
func (c *Client) Migrate(ctx context.Context, models ...any) error {
	c.logger.Info("quark_migration_start", "models", len(models))
	start := time.Now()
	err := c.Client.Migrate(ctx, models...)
	duration := time.Since(start)
	c.logger.Info("quark_migration_complete", "duration_ms", float64(duration.Nanoseconds())/1e6, "error", err)
	return err
}

// Logger returns the logger
func (c *Client) Logger() *slog.Logger {
	return c.logger
}
