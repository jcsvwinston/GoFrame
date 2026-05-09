package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestParseMailPluginDriver(t *testing.T) {
	driver, ok := parseMailPluginDriver("nucleus-mail-sendgrid")
	if !ok || driver != "sendgrid" {
		t.Fatalf("unexpected parse result: ok=%v driver=%q", ok, driver)
	}

	if runtime.GOOS == "windows" {
		driver, ok = parseMailPluginDriver("nucleus-mail-mailgun.exe")
		if !ok || driver != "mailgun" {
			t.Fatalf("unexpected windows parse result: ok=%v driver=%q", ok, driver)
		}
	}

	if _, ok := parseMailPluginDriver("nucleus-other-sendgrid"); ok {
		t.Fatal("expected invalid plugin name to be rejected")
	}
}

func TestDiscoverExternalMailPlugins(t *testing.T) {
	dir := t.TempDir()
	pluginName := "nucleus-mail-mailgun"
	if runtime.GOOS == "windows" {
		pluginName += ".exe"
	}
	pluginPath := filepath.Join(dir, pluginName)
	if err := os.WriteFile(pluginPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write plugin failed: %v", err)
	}

	previousPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", dir+string(os.PathListSeparator)+previousPath); err != nil {
		t.Fatalf("set PATH failed: %v", err)
	}
	defer func() {
		_ = os.Setenv("PATH", previousPath)
	}()

	plugins, err := discoverExternalMailPlugins()
	if err != nil {
		t.Fatalf("discoverExternalMailPlugins failed: %v", err)
	}
	got, ok := plugins["mailgun"]
	if !ok {
		t.Fatalf("expected mailgun plugin detected, got: %v", plugins)
	}
	if filepath.Clean(got) != filepath.Clean(pluginPath) {
		t.Fatalf("unexpected plugin path: got=%s want=%s", got, pluginPath)
	}
}

func TestRunMailProviders(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "nucleus.yml")
	if err := os.WriteFile(cfgPath, []byte("mail_driver: sendgrid\n"), 0o644); err != nil {
		t.Fatalf("write config failed: %v", err)
	}

	pluginName := "nucleus-mail-mailgun"
	if runtime.GOOS == "windows" {
		pluginName += ".exe"
	}
	pluginPath := filepath.Join(dir, pluginName)
	if err := os.WriteFile(pluginPath, []byte("#!/bin/sh\nexit 0\n"), 0o755); err != nil {
		t.Fatalf("write plugin failed: %v", err)
	}

	previousPath := os.Getenv("PATH")
	if err := os.Setenv("PATH", dir+string(os.PathListSeparator)+previousPath); err != nil {
		t.Fatalf("set PATH failed: %v", err)
	}
	defer func() {
		_ = os.Setenv("PATH", previousPath)
	}()

	var out bytes.Buffer
	var errOut bytes.Buffer
	err := runMailProviders([]string{"--config", cfgPath}, strings.NewReader(""), &out, &errOut)
	if err != nil {
		t.Fatalf("runMailProviders failed: %v (stderr=%s)", err, errOut.String())
	}
	text := out.String()
	if !strings.Contains(text, "Active driver: sendgrid") {
		t.Fatalf("missing active driver in output: %s", text)
	}
	if !strings.Contains(text, "sendgrid") {
		t.Fatalf("missing sendgrid provider in output: %s", text)
	}
	if !strings.Contains(text, "mailgun") {
		t.Fatalf("missing external mailgun plugin in output: %s", text)
	}
}
