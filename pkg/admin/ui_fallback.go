package admin

import (
	"embed"
	"fmt"
	"html"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const adminUIDirEnv = "NUCLEUS_ADMIN_UI_DIR"

//go:embed ui_fallback/*
var fallbackUIFS embed.FS

func adminUIContentFS() fs.FS {
	if uiFS, ok := adminUIBuildFS(); ok {
		return uiFS
	}
	fsys, err := fs.Sub(fallbackUIFS, "ui_fallback")
	if err != nil {
		return os.DirFS(".")
	}
	return fsys
}

func adminUIBuildFS() (fs.FS, bool) {
	if dir := strings.TrimSpace(os.Getenv(adminUIDirEnv)); dir != "" {
		if adminUIBuildDirUsable(dir) {
			return os.DirFS(dir), true
		}
		return nil, false
	}
	for _, dir := range adminUIBuildDirCandidates() {
		if adminUIBuildDirUsable(dir) {
			return os.DirFS(dir), true
		}
	}
	return nil, false
}

func adminUIBuildDirCandidates() []string {
	cwd, err := os.Getwd()
	if err != nil || cwd == "" {
		return nil
	}

	seen := map[string]struct{}{}
	var dirs []string
	add := func(dir string) {
		cleaned := filepath.Clean(dir)
		if _, ok := seen[cleaned]; ok {
			return
		}
		seen[cleaned] = struct{}{}
		dirs = append(dirs, cleaned)
	}

	for dir := cwd; ; dir = filepath.Dir(dir) {
		add(filepath.Join(dir, "pkg", "admin", "ui", "dist"))
		add(filepath.Join(dir, "ui", "dist"))
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
	}
	return dirs
}

func adminUIBuildDirUsable(dir string) bool {
	if strings.TrimSpace(dir) == "" {
		return false
	}
	info, err := os.Stat(filepath.Join(dir, "index.html"))
	return err == nil && !info.IsDir()
}

func injectAdminPrefix(content []byte, prefix string) []byte {
	adminPrefix := NormalizePrefix(prefix)
	injection := fmt.Sprintf(`<head><meta name="nucleus-admin-prefix" content="%s">`, html.EscapeString(adminPrefix))
	contentStr := string(content)
	if strings.Contains(contentStr, "<head>") {
		return []byte(strings.Replace(contentStr, "<head>", injection, 1))
	}
	return []byte(injection + "\n" + contentStr)
}
