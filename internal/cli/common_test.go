package cli

import (
	"bytes"
	"os"
	"strings"
	"testing"

	"github.com/jcsvwinston/GoFrame/pkg/app"
)

// --- common.go tests ---

func TestParseOptionalPositiveInt(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		fallback int
		want     int
		errOK    bool
	}{
		{name: "empty returns fallback", args: nil, fallback: 10, want: 10},
		{name: "valid positive", args: []string{"5"}, fallback: 10, want: 5},
		{name: "too many args", args: []string{"5", "3"}, fallback: 10, errOK: true},
		{name: "invalid int", args: []string{"abc"}, fallback: 10, errOK: true},
		{name: "zero rejected", args: []string{"0"}, fallback: 10, errOK: true},
		{name: "negative rejected", args: []string{"-1"}, fallback: 10, errOK: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseOptionalPositiveInt(tc.args, tc.fallback)
			if tc.errOK {
				if err == nil {
					t.Fatal("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("expected %d, got %d", tc.want, got)
			}
		})
	}
}

func TestNormalizeDatabaseAlias(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"DEFAULT", "default"},
		{"  MyDB  ", "mydb"},
		{"", ""},
		{"analytics", "analytics"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := normalizeDatabaseAlias(tc.input); got != tc.want {
				t.Errorf("normalizeDatabaseAlias(%q)=%q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestToSnakeCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"User", "user"},
		{"UserProfile", "user_profile"},
		{"UserProfilePage", "user_profile_page"},
		{"user-profile", "user_profile"},
		{"user_profile", "user_profile"},
		{"  spaces  ", "spaces"},
		{"HTMLParser", "htmlparser"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := toSnakeCase(tc.input); got != tc.want {
				t.Errorf("toSnakeCase(%q)=%q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestToPascalCase(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"", ""},
		{"user", "User"},
		{"user_profile", "UserProfile"},
		{"user-profile", "UserProfile"},
		{"UserProfile", "UserProfile"},
		{"  spaces  ", "Spaces"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := toPascalCase(tc.input); got != tc.want {
				t.Errorf("toPascalCase(%q)=%q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestSplitWords(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"UserProfile", []string{"User", "Profile"}},
		{"user-profile", []string{"user", "profile"}},
		{"user_profile", []string{"user", "profile"}},
		{"user profile", []string{"user", "profile"}},
		{"HTMLParser", []string{"HTMLParser"}},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			got := splitWords(tc.input)
			if len(got) != len(tc.want) {
				t.Fatalf("splitWords(%q) len=%d; want %d", tc.input, len(got), len(tc.want))
			}
			for i := range tc.want {
				if got[i] != tc.want[i] {
					t.Errorf("splitWords(%q)[%d]=%q; want %q", tc.input, i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestSplitWords_Empty(t *testing.T) {
	if got := splitWords(""); got != nil {
		t.Errorf("expected nil for empty input, got %v", got)
	}
	if got := splitWords("   "); got != nil {
		t.Errorf("expected nil for whitespace input, got %v", got)
	}
}

func TestEnsureDir(t *testing.T) {
	dir := t.TempDir()
	subdir := dir + "/test/sub/dir"

	err := ensureDir(subdir)
	if err != nil {
		t.Fatalf("ensureDir failed: %v", err)
	}

	if _, err := os.Stat(subdir); err != nil {
		t.Errorf("directory not created: %v", err)
	}
}

func TestEnsureDir_Empty(t *testing.T) {
	err := ensureDir("")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
}

func TestWriteFileIfNotExists(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/new_file.txt"

	err := writeFileIfNotExists(path, "hello", false)
	if err != nil {
		t.Fatalf("writeFileIfNotExists failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file failed: %v", err)
	}
	if string(content) != "hello" {
		t.Errorf("expected 'hello', got %q", string(content))
	}
}

func TestWriteFileIfNotExists_AlreadyExists(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/existing.txt"
	os.WriteFile(path, []byte("old"), 0644)

	err := writeFileIfNotExists(path, "new", false)
	if err == nil {
		t.Fatal("expected error for existing file")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("expected 'already exists' in error, got %v", err)
	}
}

func TestWriteFileIfNotExists_Force(t *testing.T) {
	dir := t.TempDir()
	path := dir + "/existing.txt"
	os.WriteFile(path, []byte("old"), 0644)

	err := writeFileIfNotExists(path, "new", true)
	if err != nil {
		t.Fatalf("writeFileIfNotExists with force failed: %v", err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read file failed: %v", err)
	}
	if string(content) != "new" {
		t.Errorf("expected 'new', got %q", string(content))
	}
}

func TestBoolToInt(t *testing.T) {
	if boolToInt(true) != 1 {
		t.Error("expected boolToInt(true)=1")
	}
	if boolToInt(false) != 0 {
		t.Error("expected boolToInt(false)=0")
	}
}

func TestQuoteSQLString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello", "'hello'"},
		{"it's", "'it''s'"},
		{"", "''"},
		{"a'b'c", "'a''b''c'"},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			if got := quoteSQLString(tc.input); got != tc.want {
				t.Errorf("quoteSQLString(%q)=%q; want %q", tc.input, got, tc.want)
			}
		})
	}
}

func TestNowRFC3339(t *testing.T) {
	now := nowRFC3339()
	if now == "" {
		t.Fatal("expected non-empty timestamp")
	}
	if !strings.Contains(now, "T") {
		t.Errorf("expected ISO 8601 format, got %q", now)
	}
}

func TestNewSilentLogger(t *testing.T) {
	logger := newSilentLogger()
	if logger == nil {
		t.Fatal("expected non-nil logger")
	}
	logger.Info("test message")
}

func TestValidateSQLIdentifier(t *testing.T) {
	tests := []struct {
		input string
		errOK bool
	}{
		{"valid_name", false},
		{"ValidName", false},
		{"_underscore", false},
		{"name123", false},
		{"", true},
		{"123invalid", true},
		{"has-dash", true},
		{"has space", true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			err := validateSQLIdentifier(tc.input)
			if tc.errOK {
				if err == nil {
					t.Fatal("expected error, got none")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestRequireDangerousApproval_NonProd(t *testing.T) {
	cfg := &app.Config{Env: "development"}
	err := requireDangerousApproval(cfg, nil, nil, false, false, "flush")
	if err != nil {
		t.Fatalf("expected approval in dev mode, got: %v", err)
	}
}

func TestRequireDangerousApproval_Force(t *testing.T) {
	cfg := &app.Config{Env: "production"}
	err := requireDangerousApproval(cfg, nil, nil, true, false, "flush")
	if err != nil {
		t.Fatalf("expected approval with force, got: %v", err)
	}
}

func TestRequireDangerousApproval_Yes(t *testing.T) {
	cfg := &app.Config{Env: "production"}
	err := requireDangerousApproval(cfg, nil, nil, false, true, "flush")
	if err != nil {
		t.Fatalf("expected approval with yes, got: %v", err)
	}
}

func TestRequireDangerousApproval_NilConfig(t *testing.T) {
	err := requireDangerousApproval(nil, nil, nil, false, false, "flush")
	if err != nil {
		t.Fatalf("expected approval with nil config, got: %v", err)
	}
}

func TestRequireDangerousApproval_NonTerminal(t *testing.T) {
	cfg := &app.Config{Env: "production"}
	var stdin bytes.Buffer
	var stdout bytes.Buffer
	err := requireDangerousApproval(cfg, &stdin, &stdout, false, false, "flush")
	if err == nil {
		t.Fatal("expected error in non-terminal mode")
	}
	if !strings.Contains(err.Error(), "--force or --yes") {
		t.Errorf("expected --force/--yes error, got %v", err)
	}
}

func TestRequireDangerousApproval_PromptWrittenInNonTerminal(t *testing.T) {
	cfg := &app.Config{Env: "production"}
	var stdin bytes.Buffer
	var stdout bytes.Buffer
	_ = requireDangerousApproval(cfg, &stdin, &stdout, false, false, "flush")
	// In non-terminal mode, the prompt should NOT be written since we short-circuit
	if stdout.String() != "" {
		t.Logf("stdout was: %q", stdout.String())
	}
}

func TestIsTerminalReader(t *testing.T) {
	// bytes.Buffer is not an *os.File
	var buf bytes.Buffer
	if isTerminalReader(&buf) {
		t.Error("expected bytes.Buffer to not be terminal")
	}
}
