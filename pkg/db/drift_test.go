package db

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/jcsvwinston/nucleus/pkg/observe"
)

func TestMigrator_Drift_EmptyWhenAllFilesPresent(t *testing.T) {
	d := newTestDB(t)
	dir := t.TempDir()
	writeMigrationPair(t, dir, "000001_create_items",
		"CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT NOT NULL);",
		"DROP TABLE IF EXISTS items;",
	)

	m := NewMigrator(d, dir, observe.NewLogger("error", "text"))
	if err := m.Up(); err != nil {
		t.Fatalf("Up: %v", err)
	}

	drift, err := m.Drift()
	if err != nil {
		t.Fatalf("Drift: %v", err)
	}
	if len(drift) != 0 {
		t.Fatalf("expected no drift when files match applied state, got %+v", drift)
	}
}

func TestMigrator_Drift_FlagsAppliedWithMissingUpFile(t *testing.T) {
	d := newTestDB(t)
	dir := t.TempDir()
	writeMigrationPair(t, dir, "000001_create_items",
		"CREATE TABLE items (id INTEGER PRIMARY KEY, name TEXT NOT NULL);",
		"DROP TABLE IF EXISTS items;",
	)
	writeMigrationPair(t, dir, "000002_create_audit",
		"CREATE TABLE audit_logs (id INTEGER PRIMARY KEY, message TEXT NOT NULL);",
		"DROP TABLE IF EXISTS audit_logs;",
	)

	m := NewMigrator(d, dir, observe.NewLogger("error", "text"))
	if err := m.Up(); err != nil {
		t.Fatalf("Up: %v", err)
	}

	// Simulate an operator deleting a migration file after it was applied.
	if err := os.Remove(filepath.Join(dir, "000002_create_audit.up.sql")); err != nil {
		t.Fatalf("remove up file: %v", err)
	}
	if err := os.Remove(filepath.Join(dir, "000002_create_audit.down.sql")); err != nil {
		t.Fatalf("remove down file: %v", err)
	}

	drift, err := m.Drift()
	if err != nil {
		t.Fatalf("Drift: %v", err)
	}
	if len(drift) != 1 {
		t.Fatalf("expected exactly one drift entry, got %d (%+v)", len(drift), drift)
	}
	got := drift[0]
	if got.ID != "000002_create_audit" {
		t.Fatalf("unexpected drift ID: %q", got.ID)
	}
	if got.Kind != DriftKindMissingUpFile {
		t.Fatalf("unexpected drift kind: %q", got.Kind)
	}
	if got.AppliedAt.IsZero() {
		t.Fatalf("expected applied_at to be set")
	}
}

func TestMigrator_Drift_ResultsSortedByID(t *testing.T) {
	d := newTestDB(t)
	dir := t.TempDir()
	// Apply three migrations.
	writeMigrationPair(t, dir, "000001_a", "CREATE TABLE a (id INTEGER);", "DROP TABLE IF EXISTS a;")
	writeMigrationPair(t, dir, "000002_b", "CREATE TABLE b (id INTEGER);", "DROP TABLE IF EXISTS b;")
	writeMigrationPair(t, dir, "000003_c", "CREATE TABLE c (id INTEGER);", "DROP TABLE IF EXISTS c;")
	m := NewMigrator(d, dir, observe.NewLogger("error", "text"))
	if err := m.Up(); err != nil {
		t.Fatalf("Up: %v", err)
	}
	// Delete the middle and the last to produce two drift rows.
	for _, name := range []string{"000003_c.up.sql", "000003_c.down.sql", "000002_b.up.sql", "000002_b.down.sql"} {
		_ = os.Remove(filepath.Join(dir, name))
	}

	drift, err := m.Drift()
	if err != nil {
		t.Fatalf("Drift: %v", err)
	}
	if len(drift) != 2 {
		t.Fatalf("expected 2 drift entries, got %d", len(drift))
	}
	if drift[0].ID != "000002_b" || drift[1].ID != "000003_c" {
		t.Fatalf("drift entries not sorted by ID: %+v", drift)
	}
}
