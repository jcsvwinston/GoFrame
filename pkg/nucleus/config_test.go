package nucleus

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// writeTempConfig writes content to a temp file with the given
// extension and returns the path. Cleanup happens via t.TempDir().
func writeTempConfig(t *testing.T, ext, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "config"+ext)
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}

func TestLoadFromFile_HappyPathYAML(t *testing.T) {
	t.Parallel()

	yamlBody := `
host: 0.0.0.0
port: 9090
log_level: warn
`
	path := writeTempConfig(t, ".yaml", yamlBody)
	cfg, err := loadFromFile(path)
	if err != nil {
		t.Fatalf("loadFromFile: %v", err)
	}
	if cfg.Host != "0.0.0.0" {
		t.Errorf("Host: got %q want %q", cfg.Host, "0.0.0.0")
	}
	if cfg.Port != 9090 {
		t.Errorf("Port: got %d want %d", cfg.Port, 9090)
	}
	if cfg.LogLevel != "warn" {
		t.Errorf("LogLevel: got %q want %q", cfg.LogLevel, "warn")
	}
}

func TestLoadFromFile_PreservesDefaultsForUnsetKeys(t *testing.T) {
	t.Parallel()

	// Body sets only Port; every other field should come from
	// app.DefaultConfig() (struct defaults applied first).
	path := writeTempConfig(t, ".yaml", "port: 1234\n")
	cfg, err := loadFromFile(path)
	if err != nil {
		t.Fatalf("loadFromFile: %v", err)
	}
	if cfg.Port != 1234 {
		t.Errorf("Port: got %d want 1234", cfg.Port)
	}
	// LogLevel is part of app.DefaultConfig and must survive the load.
	if cfg.LogLevel == "" {
		t.Error("LogLevel was reset to zero value; defaults not applied")
	}
}

func TestLoadFromFile_RejectsUnsupportedExtension(t *testing.T) {
	t.Parallel()

	path := writeTempConfig(t, ".ini", "[server]\nport = 80\n")
	_, err := loadFromFile(path)
	if err == nil {
		t.Fatal("expected an error for .ini extension")
	}
	if !errors.Is(err, ErrUnsupportedConfigFormat) {
		t.Errorf("want ErrUnsupportedConfigFormat, got %v", err)
	}
}

func TestLoadFromFile_TOMLAndJSONReportPhase2b(t *testing.T) {
	t.Parallel()

	for _, ext := range []string{".toml", ".json"} {
		path := writeTempConfig(t, ext, "port = 80\n")
		_, err := loadFromFile(path)
		if !errors.Is(err, ErrUnsupportedConfigFormat) {
			t.Errorf("ext=%s want ErrUnsupportedConfigFormat, got %v", ext, err)
		}
		if !strings.Contains(err.Error(), "Phase 2b") {
			t.Errorf("ext=%s error should reference Phase 2b, got %q", ext, err.Error())
		}
	}
}

func TestLoadFromFile_FileTooLarge(t *testing.T) {
	t.Parallel()

	// Build a YAML body larger than MaxConfigFileBytes. Use a single
	// long scalar to avoid producing valid YAML by accident — the
	// cap is enforced BEFORE the parser ever runs.
	big := strings.Repeat("# padding line that takes some bytes per line\n", (MaxConfigFileBytes/40)+1)
	path := writeTempConfig(t, ".yaml", big)
	_, err := loadFromFile(path)
	if !errors.Is(err, ErrConfigFileTooLarge) {
		t.Fatalf("want ErrConfigFileTooLarge, got %v", err)
	}
	if !strings.Contains(err.Error(), "cap=") {
		t.Errorf("error should mention the cap, got %q", err.Error())
	}
}

func TestLoadFromFile_FileAtCapBoundaryIsAccepted(t *testing.T) {
	t.Parallel()

	// Exactly at the cap should succeed. Use a body sized so the
	// final file is at or just below MaxConfigFileBytes. The body
	// must remain parseable, so we keep one valid key + padding
	// comment lines.
	header := "port: 8080\n"
	want := MaxConfigFileBytes
	pad := want - len(header)
	if pad < 0 {
		t.Skip("MaxConfigFileBytes is smaller than the header; skipping boundary test")
	}
	body := header + strings.Repeat("#", pad)
	if len(body) > MaxConfigFileBytes {
		body = body[:MaxConfigFileBytes]
	}
	path := writeTempConfig(t, ".yaml", body)
	cfg, err := loadFromFile(path)
	if err != nil {
		t.Fatalf("expected boundary-sized file to load, got %v", err)
	}
	if cfg.Port != 8080 {
		t.Errorf("Port: got %d want 8080", cfg.Port)
	}
}

func TestLoadFromFile_StrictUnknownKey(t *testing.T) {
	t.Parallel()

	// `prot` is a likely typo for `port`. Strict mode rejects it.
	path := writeTempConfig(t, ".yaml", "prot: 80\n")
	_, err := loadFromFile(path)
	if !errors.Is(err, ErrUnknownConfigKeys) {
		t.Fatalf("want ErrUnknownConfigKeys, got %v", err)
	}
}

func TestLoadFromFile_DidYouMeanHint(t *testing.T) {
	t.Parallel()

	// `loging_level` is one insertion away from `log_level`. The
	// hint should surface.
	path := writeTempConfig(t, ".yaml", "loging_level: warn\n")
	_, err := loadFromFile(path)
	if !errors.Is(err, ErrUnknownConfigKeys) {
		t.Fatalf("want ErrUnknownConfigKeys, got %v", err)
	}
	if !strings.Contains(err.Error(), "did you mean") {
		t.Errorf("error should include a did-you-mean hint, got %q", err.Error())
	}
}

func TestLoadFromFile_MissingFile(t *testing.T) {
	t.Parallel()

	_, err := loadFromFile("/nonexistent/path/nucleus.yaml")
	if err == nil {
		t.Fatal("expected an error for missing file")
	}
	if errors.Is(err, ErrUnknownConfigKeys) || errors.Is(err, ErrConfigFileTooLarge) {
		t.Errorf("missing-file error should not be wrapped as a config-content error, got %v", err)
	}
}

func TestLoadFromFile_MalformedYAML(t *testing.T) {
	t.Parallel()

	path := writeTempConfig(t, ".yaml", "port: : bad\n  - mixed: types\n")
	_, err := loadFromFile(path)
	if err == nil {
		t.Fatal("expected a parse error for malformed YAML")
	}
}

func TestLoadFromFile_EmptyPath(t *testing.T) {
	t.Parallel()

	_, err := loadFromFile("")
	if err == nil {
		t.Fatal("expected an error for empty path")
	}
}

func TestAppBuilder_FromConfigFile_Happy(t *testing.T) {
	t.Parallel()

	path := writeTempConfig(t, ".yaml", "port: 7777\n")
	a, err := New().FromConfigFile(path).Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if a.Port != 7777 {
		t.Errorf("Port: got %d want 7777", a.Port)
	}
	if a.Modules == nil {
		t.Error("Modules map should be non-nil after Build")
	}
}

func TestAppBuilder_FromConfigFile_PreservesPriorMount(t *testing.T) {
	t.Parallel()

	// Mount BEFORE FromConfigFile and confirm the file load does not
	// drop the registered module.
	mod := Module[struct{}]{Name: "articles", Prefix: "/articles"}.Build()
	path := writeTempConfig(t, ".yaml", "port: 7777\n")
	a, err := New().Mount(mod).FromConfigFile(path).Build()
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, ok := a.Modules["articles"]; !ok {
		t.Error("Modules registered before FromConfigFile were dropped by the loader")
	}
}

func TestLevenshtein_Basics(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"port", "port", 0},
		{"prot", "port", 2},  // two transpositions (substitution-based distance)
		{"port", "ports", 1}, // one insertion
		{"port", "", 4},
		{"", "port", 4},
	} {
		got := levenshtein(tc.a, tc.b)
		if got != tc.want {
			t.Errorf("levenshtein(%q, %q): got %d want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestKeyMatchesAny_Wildcards(t *testing.T) {
	t.Parallel()

	patterns := compileKeyPatterns([]string{
		"port",
		"databases.*.url",
		"jwt_keys.*.kid",
	})
	cases := []struct {
		key  string
		want bool
	}{
		{"port", true},
		{"databases.default.url", true},
		{"databases.analytics.url", true},
		{"databases.default.user", false}, // *.user not in patterns
		{"jwt_keys.signing.kid", true},
		{"jwt_keys.signing.algorithm", false},
		{"unknown", false},
	}
	for _, tc := range cases {
		got := keyMatchesAny(tc.key, patterns)
		if got != tc.want {
			t.Errorf("keyMatchesAny(%q): got %v want %v", tc.key, got, tc.want)
		}
	}
}
