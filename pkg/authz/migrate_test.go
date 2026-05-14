package authz

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func writePolicy(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "rbac.csv")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
	return path
}

func readPolicy(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read policy: %v", err)
	}
	return string(b)
}

func TestMigrateCSVPolicyFile_UpgradesThreeFieldPolicies(t *testing.T) {
	input := strings.Join([]string{
		"# comment line preserved",
		"",
		"p, alice, /data1, read",
		"p, bob, /data2, write",
		"g, alice, admin",
		"",
	}, "\n")

	path := writePolicy(t, input)

	report, err := MigrateCSVPolicyFile(path, "")
	if err != nil {
		t.Fatalf("MigrateCSVPolicyFile: %v", err)
	}
	if !report.Changed {
		t.Fatal("expected Changed=true after upgrading 3-field policies")
	}
	if got, want := report.PolicyLinesUpgraded, 2; got != want {
		t.Errorf("PolicyLinesUpgraded = %d, want %d", got, want)
	}
	if got, want := report.PolicyLinesAlreadyMigrated, 0; got != want {
		t.Errorf("PolicyLinesAlreadyMigrated = %d, want %d", got, want)
	}
	if got, want := report.GroupingLinesPreserved, 1; got != want {
		t.Errorf("GroupingLinesPreserved = %d, want %d", got, want)
	}
	if got, want := report.BlankOrCommentLines, 2; got != want {
		t.Errorf("BlankOrCommentLines = %d, want %d", got, want)
	}

	out := readPolicy(t, path)
	for _, expected := range []string{
		"# comment line preserved",
		"p, alice, /data1, read, allow",
		"p, bob, /data2, write, allow",
		"g, alice, admin",
	} {
		if !strings.Contains(out, expected) {
			t.Errorf("output missing line %q\nfull output:\n%s", expected, out)
		}
	}
}

func TestMigrateCSVPolicyFile_IsIdempotent(t *testing.T) {
	already := "p, alice, /data, read, allow\np, bob, /admin, write, deny\n"
	path := writePolicy(t, already)

	report, err := MigrateCSVPolicyFile(path, "allow")
	if err != nil {
		t.Fatalf("first run: %v", err)
	}
	if report.Changed {
		t.Error("Changed=true for an already-migrated file")
	}
	if got, want := report.PolicyLinesAlreadyMigrated, 2; got != want {
		t.Errorf("PolicyLinesAlreadyMigrated = %d, want %d", got, want)
	}
	if report.PolicyLinesUpgraded != 0 {
		t.Errorf("PolicyLinesUpgraded = %d, want 0", report.PolicyLinesUpgraded)
	}

	// File content unchanged.
	if got := readPolicy(t, path); got != already {
		t.Errorf("file mutated on idempotent run:\n got: %q\nwant: %q", got, already)
	}

	// Second run is also a no-op.
	report2, err := MigrateCSVPolicyFile(path, "allow")
	if err != nil {
		t.Fatalf("second run: %v", err)
	}
	if report2.Changed {
		t.Error("second run reported Changed=true")
	}
}

func TestMigrateCSVPolicyFile_DefaultDeny(t *testing.T) {
	input := "p, blocked, /sensitive, read\n"
	path := writePolicy(t, input)

	report, err := MigrateCSVPolicyFile(path, "deny")
	if err != nil {
		t.Fatalf("MigrateCSVPolicyFile: %v", err)
	}
	if !report.Changed {
		t.Fatal("expected Changed=true")
	}

	out := readPolicy(t, path)
	if !strings.Contains(out, "p, blocked, /sensitive, read, deny") {
		t.Errorf("expected deny effect, got:\n%s", out)
	}
}

func TestMigrateCSVPolicyFile_RejectsInvalidEffect(t *testing.T) {
	path := writePolicy(t, "p, x, y, z\n")

	if _, err := MigrateCSVPolicyFile(path, "maybe"); err == nil {
		t.Fatal("expected error for invalid effect, got nil")
	}
	// File must not have been rewritten.
	if got := readPolicy(t, path); got != "p, x, y, z\n" {
		t.Errorf("file mutated on invalid input: %q", got)
	}
}

func TestMigrateCSVPolicyFile_MixedFile(t *testing.T) {
	input := strings.Join([]string{
		"# header comment",
		"p, alice, /data, read",
		"p, bob, /admin, write, deny",
		"g, alice, admin",
		"g2, alice, finance",
		"p2, alice, /finance, *",
		"",
	}, "\n")

	path := writePolicy(t, input)
	report, err := MigrateCSVPolicyFile(path, "allow")
	if err != nil {
		t.Fatalf("MigrateCSVPolicyFile: %v", err)
	}

	if got, want := report.PolicyLinesUpgraded, 2; got != want {
		t.Errorf("PolicyLinesUpgraded = %d, want %d (p alice + p2 alice)", got, want)
	}
	if got, want := report.PolicyLinesAlreadyMigrated, 1; got != want {
		t.Errorf("PolicyLinesAlreadyMigrated = %d, want %d (p bob)", got, want)
	}
	if got, want := report.GroupingLinesPreserved, 2; got != want {
		t.Errorf("GroupingLinesPreserved = %d, want %d (g + g2)", got, want)
	}

	out := readPolicy(t, path)
	if !strings.Contains(out, "p, alice, /data, read, allow") {
		t.Error("p alice not upgraded")
	}
	if !strings.Contains(out, "p, bob, /admin, write, deny") {
		t.Error("p bob (already migrated) not preserved")
	}
	if !strings.Contains(out, "p2, alice, /finance, *, allow") {
		t.Error("p2 alice not upgraded")
	}
	if strings.Contains(out, "g, alice, admin, allow") {
		t.Error("grouping policy was incorrectly rewritten")
	}
}

func TestMigrateCSVPolicyFile_LoadableByEnforcer(t *testing.T) {
	input := "p, alice, /data, read\np, bob, /data, write\n"
	path := writePolicy(t, input)

	if _, err := MigrateCSVPolicyFile(path, ""); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	e, err := New(slog.Default(), path)
	if err != nil {
		t.Fatalf("New with migrated policy file: %v", err)
	}

	if !e.Can("alice", "/data", "read") {
		t.Error("alice should be allowed to read /data after migration")
	}
	if !e.Can("bob", "/data", "write") {
		t.Error("bob should be allowed to write /data after migration")
	}
	if e.Can("alice", "/data", "delete") {
		t.Error("alice should not be allowed to delete /data")
	}
}

func TestMigrateCSVPolicyFile_PreservesBlankLinesAndComments(t *testing.T) {
	input := "# top comment\n\np, alice, /data, read\n\n# trailing comment\n"
	path := writePolicy(t, input)

	if _, err := MigrateCSVPolicyFile(path, "allow"); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	out := readPolicy(t, path)
	if !strings.HasPrefix(out, "# top comment\n\n") {
		t.Errorf("leading comment + blank not preserved:\n%s", out)
	}
	if !strings.Contains(out, "\n# trailing comment\n") {
		t.Errorf("trailing comment not preserved:\n%s", out)
	}
}
