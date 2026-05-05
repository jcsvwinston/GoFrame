package quarkdb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/jcsvwinston/quark"
	_ "modernc.org/sqlite"
)

// Client wraps the Quark ORM client
type Client struct {
	*quark.Client
	db *sql.DB
}

// NewClient creates a new Quark client with SQLite
func NewClient(dbPath string) (*Client, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	client, err := quark.New(db, quark.WithDialect(quark.SQLite()))
	if err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to create quark client: %w", err)
	}

	return &Client{
		Client: client,
		db:     db,
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
	return c.Client.Migrate(ctx, models...)
}
