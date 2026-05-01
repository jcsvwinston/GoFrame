package quark

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"time"
)

// Client is the main entry point for quark ORM operations.
// It wraps a database connection and provides type-safe query building.
type Client struct {
	db         *sql.DB
	dialect    Dialect
	logger     *slog.Logger
	guard      *SQLGuard
	observers  []QueryObserver
	middleware []Middleware
	limits     Limits
}

// New creates a new quark Client with the given database connection and options.
//
// Example:
//
//	db, err := sql.Open("postgres", "postgres://user:pass@localhost/db")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	client, err := quark.New(db,
//	    quark.WithDialect(quark.PostgreSQL()),
//	    quark.WithLogger(slog.Default()),
//	)
func New(db *sql.DB, opts ...Option) (*Client, error) {
	if db == nil {
		return nil, fmt.Errorf("%w: db cannot be nil", ErrConnection)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnection, err)
	}

	c := &Client{
		db:         db,
		logger:     slog.Default(),
		guard:      NewSQLGuard(),
		observers:  make([]QueryObserver, 0),
		middleware: make([]Middleware, 0),
		limits:     DefaultLimits(),
	}

	// Auto-detect dialect if not specified
	if c.dialect == nil {
		driver := reflect.TypeOf(db.Driver()).String()
		// Try to extract driver name from type string like "*pq.Driver" or "*stdlib.Driver"
		dialect, err := detectDialectFromDriver(driver, db)
		if err != nil {
			c.logger.Warn("could not auto-detect dialect, defaulting to generic",
				"driver", driver,
				"error", err)
			// Default to PostgreSQL as most common
			c.dialect = PostgreSQL()
		} else {
			c.dialect = dialect
		}
	}

	// Apply options
	for _, opt := range opts {
		opt(c)
	}

	c.logger.Info("quark client initialized",
		"dialect", c.dialect.Name(),
		"max_results", c.limits.MaxResults,
	)

	return c, nil
}

// For creates a Query builder for the given model type.
// This is the primary entry point for type-safe database operations.
//
// Example:
//
//	type User struct {
//	    ID   int64  `db:"id"`
//	    Name string `db:"name"`
//	}
//
//	user, err := quark.For[User](ctx, client).Find(1)
//	users, err := quark.For[User](ctx, client).Where("active", "=", true).List()
func For[T any](ctx context.Context, client *Client) *Query[T] {
	meta := GetModelMeta[T]()

	return &Query[T]{
		ctx:     ctx,
		client:  client,
		dialect: client.dialect,
		guard:   client.guard,
		table:   meta.Table,
		pk:      meta.PK,
		exec:    client.db,
		meta:    meta,
	}
}

// tableNameFromType derives table name from generic type T.
func tableNameFromType[T any]() string {
	var zero T
	t := reflect.TypeOf(zero)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// Convert struct name to snake_case plural
	name := t.Name()
	if name == "" {
		return ""
	}
	return toSnakeCase(pluralize(name))
}

// simple pluralization: add 's' (can be improved)
func pluralize(s string) string {
	if strings.HasSuffix(s, "s") || strings.HasSuffix(s, "x") ||
		strings.HasSuffix(s, "ch") || strings.HasSuffix(s, "sh") {
		return s + "es"
	}
	if strings.HasSuffix(s, "y") && len(s) > 1 && !isVowel(s[len(s)-2]) {
		return s[:len(s)-1] + "ies"
	}
	return s + "s"
}

func isVowel(c byte) bool {
	return c == 'a' || c == 'e' || c == 'i' || c == 'o' || c == 'u' ||
		c == 'A' || c == 'E' || c == 'I' || c == 'O' || c == 'U'
}

// toSnakeCase converts CamelCase to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteByte('_')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}

// RawQuery executes a raw SQL query with the given arguments.
// By default, this requires placeholders to prevent SQL injection.
// Enable with WithLimits(Limits{AllowRawQueries: true}).
func (c *Client) RawQuery(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if !c.limits.AllowRawQueries {
		return nil, fmt.Errorf("%w: raw queries are disabled by default, enable with WithLimits", ErrInvalidQuery)
	}

	if err := c.guard.ValidateRawQuery(query, true); err != nil {
		return nil, err
	}

	start := time.Now()
	rows, err := c.db.QueryContext(ctx, query, args...)
	duration := time.Since(start)

	// Notify observers
	event := QueryEvent{
		SQL:       query,
		Args:      args,
		Duration:  duration,
		Error:     err,
		Operation: "RAW",
	}
	for _, obs := range c.observers {
		obs.ObserveQuery(event)
	}

	return rows, err
}

// Close closes the underlying database connection.
func (c *Client) Close() error {
	return c.db.Close()
}

// DB returns the underlying *sql.DB for advanced operations.
func (c *Client) DB() *sql.DB {
	return c.db
}

// Dialect returns the dialect being used.
func (c *Client) Dialect() Dialect {
	return c.dialect
}

// detectDialectFromDriver attempts to detect the dialect from the driver type.
func detectDialectFromDriver(driverType string, db *sql.DB) (Dialect, error) {
	// Try to get the driver name from the db
	// This is heuristic-based
	switch {
	case containsAny(driverType, "pgx", "pq.", "postgres"):
		return PostgreSQL(), nil
	case containsAny(driverType, "mysql", "mariadb"):
		return MySQL(), nil
	case containsAny(driverType, "sqlite", "modernc"):
		return SQLite(), nil
	default:
		return nil, fmt.Errorf("could not detect dialect from driver: %s", driverType)
	}
}

func containsAny(s string, substrs ...string) bool {
	lower := strings.ToLower(s)
	for _, sub := range substrs {
		if strings.Contains(lower, sub) {
			return true
		}
	}
	return false
}

