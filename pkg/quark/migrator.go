package quark

import (
	"context"
	"fmt"
	"reflect"
	"strings"

	"github.com/jcsvwinston/GoFrame/pkg/quark/internal/migrate"
)

// Migrate creates tables for the given models if they don't exist.
// This is a simplistic auto-migration tool for development.
// It uses the "db" and "pk" tags to generate CREATE TABLE statements.
// It also creates join tables for many-to-many relations.
func (c *Client) Migrate(ctx context.Context, models ...any) error {
	for _, model := range models {
		if err := c.createTable(ctx, model); err != nil {
			return err
		}
		if err := c.createJoinTables(ctx, model); err != nil {
			return err
		}
	}
	return nil
}

func (c *Client) createTable(ctx context.Context, model any) error {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return fmt.Errorf("model must be a struct, got %s", t.Kind())
	}

	meta := GetModelMetaByType(t)
	if meta == nil {
		return fmt.Errorf("failed to get metadata for %s", t.Name())
	}

	var columns []string
	for _, field := range meta.Fields {
		if field.Column == "" {
			continue
		}

		// For composite PKs, never mark individual columns as PRIMARY KEY —
		// we'll append a table-level constraint below instead.
		isPK := field.IsPK && !meta.HasCompositePK
		colDef := c.dialect.Quote(field.Column) + " " + migrate.SQLType(c.dialect.Name(), field.Type, isPK)
		columns = append(columns, colDef)
	}

	// Composite PK: append table-level PRIMARY KEY constraint
	if meta.HasCompositePK {
		pkCols := make([]string, len(meta.CompositePK))
		for i, pk := range meta.CompositePK {
			pkCols[i] = c.dialect.Quote(pk.Column)
		}
		columns = append(columns, fmt.Sprintf("PRIMARY KEY (%s)", strings.Join(pkCols, ", ")))
	}

	if len(columns) == 0 {
		return fmt.Errorf("no database columns found for model %s", t.Name())
	}

	var query string
	switch c.dialect.Name() {
	case "mysql", "postgres", "sqlite":
		query = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n);",
			c.dialect.Quote(meta.Table),
			strings.Join(columns, ",\n  "),
		)
	case "mssql":
		query = fmt.Sprintf(`IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = '%s') 
		CREATE TABLE %s (
			%s
		);`, meta.Table, c.dialect.Quote(meta.Table), strings.Join(columns, ",\n  "))
	case "oracle":
		query = fmt.Sprintf("CREATE TABLE %s (\n  %s\n)",
			c.dialect.Quote(meta.Table),
			strings.Join(columns, ",\n  "),
		)
	default:
		query = fmt.Sprintf("CREATE TABLE %s (\n  %s\n)",
			c.dialect.Quote(meta.Table),
			strings.Join(columns, ",\n  "),
		)
	}

	_, err := c.db.ExecContext(ctx, query)
	if err != nil {
		if c.dialect.Name() == "oracle" && strings.Contains(err.Error(), "ORA-00955") {
			return nil
		}
		return fmt.Errorf("failed to create table %s: %w", meta.Table, err)
	}

	return nil
}

// createJoinTables creates join tables for many-to-many relations.
func (c *Client) createJoinTables(ctx context.Context, model any) error {
	t := reflect.TypeOf(model)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	if t.Kind() != reflect.Struct {
		return nil
	}

	meta := GetModelMetaByType(t)
	if meta == nil {
		return nil
	}

	for _, rel := range meta.Relations {
		if rel.Type != "many_to_many" || rel.JoinTable == "" {
			continue
		}

		// Determine SQL types for FK columns (using int64 for simple auto-migration)
		thisFKType := migrate.SQLType(c.dialect.Name(), reflect.TypeOf(int64(0)), false)
		refFKType := migrate.SQLType(c.dialect.Name(), reflect.TypeOf(int64(0)), false)

		// Build join table columns
		columns := []string{
			fmt.Sprintf("%s %s", c.dialect.Quote(rel.JoinFK), thisFKType),
			fmt.Sprintf("%s %s", c.dialect.Quote(rel.JoinRefFK), refFKType),
		}

		// Create composite primary key
		pkConstraint := fmt.Sprintf("PRIMARY KEY (%s, %s)", c.dialect.Quote(rel.JoinFK), c.dialect.Quote(rel.JoinRefFK))
		columns = append(columns, pkConstraint)

		// Build CREATE TABLE query
		var query string
		switch c.dialect.Name() {
		case "mysql", "postgres", "sqlite":
			query = fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n  %s\n);",
				c.dialect.Quote(rel.JoinTable),
				strings.Join(columns, ",\n  "),
			)
		case "mssql":
			query = fmt.Sprintf(`IF NOT EXISTS (SELECT * FROM sys.tables WHERE name = '%s')
			CREATE TABLE %s (
				%s
			);`, rel.JoinTable, c.dialect.Quote(rel.JoinTable), strings.Join(columns, ",\n				"))
		case "oracle":
			query = fmt.Sprintf("CREATE TABLE %s (\n  %s\n)",
				c.dialect.Quote(rel.JoinTable),
				strings.Join(columns, ",\n  "),
			)
		default:
			query = fmt.Sprintf("CREATE TABLE %s (\n  %s\n)",
				c.dialect.Quote(rel.JoinTable),
				strings.Join(columns, ",\n  "),
			)
		}

		_, err := c.db.ExecContext(ctx, query)
		if err != nil {
			if c.dialect.Name() == "oracle" && strings.Contains(err.Error(), "ORA-00955") {
				continue
			}
			return fmt.Errorf("failed to create join table %s: %w", rel.JoinTable, err)
		}
	}

	return nil
}
