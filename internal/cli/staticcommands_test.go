package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveStaticSources_DefaultsAndOutputExclusion(t *testing.T) {
	dir := t.TempDir()
	restore := withWorkingDir(t, dir)
	defer restore()

	if err := os.MkdirAll(filepath.Join("static"), 0o755); err != nil {
		t.Fatalf("mkdir static failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join("internal", "web", "static"), 0o755); err != nil {
		t.Fatalf("mkdir internal web static failed: %v", err)
	}
	if err := os.MkdirAll(filepath.Join("internal", "billing", "web", "static"), 0o755); err != nil {
		t.Fatalf("mkdir app static failed: %v", err)
	}

	sources, err := resolveStaticSources("", true, "static")
	if err != nil {
		t.Fatalf("resolveStaticSources failed: %v", err)
	}
	for _, source := range sources {
		if filepath.Clean(source) == "static" {
			t.Fatalf("expected output directory static to be excluded, got sources=%v", sources)
		}
	}
	got := strings.Join(sources, ",")
	if !strings.Contains(got, filepath.Join("internal", "web", "static")) {
		t.Fatalf("expected internal/web/static in sources, got %s", got)
	}
	if !strings.Contains(got, filepath.Join("internal", "billing", "web", "static")) {
		t.Fatalf("expected internal/billing/web/static in sources, got %s", got)
	}
}

func TestBuildCollectStaticPlan_DuplicateRelativePath(t *testing.T) {
	dir := t.TempDir()
	restore := withWorkingDir(t, dir)
	defer restore()

	writeTestFile(t, filepath.Join("internal", "web", "static", "app.css"), "a{}")
	writeTestFile(t, filepath.Join("internal", "billing", "web", "static", "app.css"), "b{}")

	plan, duplicates, err := buildCollectStaticPlan(
		[]string{
			filepath.Join("internal", "web", "static"),
			filepath.Join("internal", "billing", "web", "static"),
		},
		"collected",
	)
	if err != nil {
		t.Fatalf("buildCollectStaticPlan failed: %v", err)
	}
	if len(plan) != 1 {
		t.Fatalf("expected 1 planned file (first copy wins), got %d", len(plan))
	}
	if len(duplicates) != 1 {
		t.Fatalf("expected 1 duplicate, got %d", len(duplicates))
	}
	if duplicates[0].relativePath != "app.css" {
		t.Fatalf("unexpected duplicate relative path: %s", duplicates[0].relativePath)
	}
}

func TestFindStaticMatches_GlobAndFirst(t *testing.T) {
	dir := t.TempDir()
	restore := withWorkingDir(t, dir)
	defer restore()

	writeTestFile(t, filepath.Join("internal", "web", "static", "css", "app.css"), "x")
	writeTestFile(t, filepath.Join("internal", "web", "static", "css", "admin.css"), "y")
	writeTestFile(t, filepath.Join("internal", "shop", "web", "static", "css", "app.css"), "z")

	sources := []string{
		filepath.Join("internal", "web", "static"),
		filepath.Join("internal", "shop", "web", "static"),
	}

	first, err := findStaticMatches(sources, "css/app.css", true)
	if err != nil {
		t.Fatalf("findStaticMatches(first) failed: %v", err)
	}
	if len(first) != 1 {
		t.Fatalf("expected one first match, got %d", len(first))
	}
	if !strings.HasSuffix(filepath.ToSlash(first[0]), "internal/web/static/css/app.css") {
		t.Fatalf("unexpected first match: %s", first[0])
	}

	globbed, err := findStaticMatches(sources, "css/*.css", false)
	if err != nil {
		t.Fatalf("findStaticMatches(glob) failed: %v", err)
	}
	if len(globbed) != 3 {
		t.Fatalf("expected 3 globbed matches, got %d (%v)", len(globbed), globbed)
	}
}

func TestRunCollectStaticAndFindStatic(t *testing.T) {
	dir := t.TempDir()
	restore := withWorkingDir(t, dir)
	defer restore()

	cfgPath := filepath.Join(dir, "goframe.yaml")
	writeTestFile(t, cfgPath, "static_root: collected_static\n")
	writeTestFile(t, filepath.Join("internal", "web", "static", "site.css"), "body{}")
	writeTestFile(t, filepath.Join("internal", "web", "static", "js", "app.js"), "console.log('ok')")

	var out bytes.Buffer
	var errOut bytes.Buffer
	if err := runCollectStatic([]string{
		"--config", cfgPath,
		"--dry-run",
	}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runCollectStatic dry-run failed: %v", err)
	}
	if !strings.Contains(out.String(), "DRY-RUN\tCOLLECTSTATIC") {
		t.Fatalf("unexpected collectstatic dry-run output: %s", out.String())
	}

	out.Reset()
	errOut.Reset()
	if err := runCollectStatic([]string{
		"--config", cfgPath,
	}, strings.NewReader(""), &out, &errOut); err != nil {
		t.Fatalf("runCollectStatic failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join("collected_static", "site.css")); err != nil {
		t.Fatalf("expected collected static file: %v", err)
	}
	if _, err := os.Stat(filepath.Join("collected_static", "js", "app.js")); err != nil {
		t.Fatalf("expected collected nested static file: %v", err)
	}

	out.Reset()
	errOut.Reset()
	err := runFindStatic([]string{
		"--config", cfgPath,
		"js/app.js",
	}, strings.NewReader(""), &out, &errOut)
	if err != nil {
		t.Fatalf("runFindStatic failed: %v", err)
	}
	if !strings.Contains(out.String(), filepath.Join("internal", "web", "static", "js", "app.js")) {
		t.Fatalf("unexpected findstatic output: %s", out.String())
	}
}

func withWorkingDir(t *testing.T, target string) func() {
	t.Helper()
	previous, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd failed: %v", err)
	}
	if err := os.Chdir(target); err != nil {
		t.Fatalf("chdir failed: %v", err)
	}
	return func() {
		if err := os.Chdir(previous); err != nil {
			t.Fatalf("restore chdir failed: %v", err)
		}
	}
}
