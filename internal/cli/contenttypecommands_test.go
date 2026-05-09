package cli

import (
	"testing"
)

func TestParseNormalizedNameSet(t *testing.T) {
	set, err := parseNormalizedNameSet(" users,Orders, users ")
	if err != nil {
		t.Fatalf("parseNormalizedNameSet failed: %v", err)
	}
	if len(set) != 2 {
		t.Fatalf("expected 2 unique names, got %d", len(set))
	}
	if _, ok := set["users"]; !ok {
		t.Fatal("expected users in set")
	}
	if _, ok := set["orders"]; !ok {
		t.Fatal("expected orders in set")
	}
}

func TestBuildRemoveStaleContentTypeStatements(t *testing.T) {
	stmts := buildRemoveStaleContentTypeStatements(dbFlavorSQLite, "nucleus_content_types", "model", []string{"ghost"})
	if len(stmts) != 1 {
		t.Fatalf("expected one statement, got %d", len(stmts))
	}
	want := `DELETE FROM "nucleus_content_types" WHERE LOWER(TRIM("model")) = 'ghost'`
	if stmts[0] != want {
		t.Fatalf("unexpected statement:\n got: %s\nwant: %s", stmts[0], want)
	}
}
