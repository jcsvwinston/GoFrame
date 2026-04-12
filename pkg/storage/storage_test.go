package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLocalStore_PutAndGet(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalStore(LocalConfig{Path: dir})
	if err != nil {
		t.Fatalf("NewLocalStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	content := strings.NewReader("hello world")
	info, err := store.Put(ctx, "test/file.txt", content, PutOptions{
		ContentType: "text/plain",
	})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}

	if info.Key != "test/file.txt" {
		t.Errorf("expected key 'test/file.txt', got %q", info.Key)
	}
	if info.Size != 11 {
		t.Errorf("expected size 11, got %d", info.Size)
	}

	reader, gotInfo, err := store.Get(ctx, "test/file.txt")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	defer reader.Close()

	if gotInfo.Key != "test/file.txt" {
		t.Errorf("expected key 'test/file.txt', got %q", gotInfo.Key)
	}
}

func TestLocalStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalStore(LocalConfig{Path: dir})
	if err != nil {
		t.Fatalf("NewLocalStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	store.Put(ctx, "test/to-delete.txt", strings.NewReader("data"), PutOptions{})

	exists, _ := store.Exists(ctx, "test/to-delete.txt")
	if !exists {
		t.Fatal("expected file to exist after Put")
	}

	if err := store.Delete(ctx, "test/to-delete.txt"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	exists, _ = store.Exists(ctx, "test/to-delete.txt")
	if exists {
		t.Error("expected file to be deleted")
	}
}

func TestLocalStore_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalStore(LocalConfig{Path: dir})
	if err != nil {
		t.Fatalf("NewLocalStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	_, _, err = store.Get(ctx, "nonexistent/file.txt")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
	if _, ok := err.(ErrNotFound); !ok {
		t.Errorf("expected ErrNotFound, got %T: %v", err, err)
	}
}

func TestLocalStore_List(t *testing.T) {
	dir := t.TempDir()
	store, err := NewLocalStore(LocalConfig{Path: dir})
	if err != nil {
		t.Fatalf("NewLocalStore: %v", err)
	}
	defer store.Close()

	ctx := context.Background()
	for i := 0; i < 3; i++ {
		store.Put(ctx, fmt.Sprintf("test/file%d.txt", i), strings.NewReader("data"), PutOptions{})
	}

	result, err := store.List(ctx, ListOptions{Prefix: "test/"})
	if err != nil {
		t.Fatalf("List: %v", err)
	}

	if len(result.Objects) != 3 {
		t.Errorf("expected 3 objects, got %d", len(result.Objects))
	}
}

func TestTenantStore_PrefixesKeys(t *testing.T) {
	dir := t.TempDir()
	baseStore, _ := NewLocalStore(LocalConfig{Path: dir})
	defer baseStore.Close()

	tenantGetter := func(ctx context.Context) string {
		if tenant, ok := ctx.Value(TenantKey{}).(string); ok {
			return tenant
		}
		return ""
	}

	store := NewTenantStore(baseStore, tenantGetter)

	ctx := context.WithValue(context.Background(), TenantKey{}, "acme")
	content := strings.NewReader("tenant data")
	info, err := store.Put(ctx, "uploads/doc.pdf", content, PutOptions{})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}

	if info.Key != "acme/uploads/doc.pdf" {
		t.Errorf("expected key 'acme/uploads/doc.pdf', got %q", info.Key)
	}

	// Verify the file exists with the prefixed path
	fullPath := filepath.Join(dir, "acme/uploads/doc.pdf")
	if _, err := os.Stat(fullPath); err != nil {
		t.Errorf("expected file at %q, got error: %v", fullPath, err)
	}
}

func TestTenantStore_NoTenant_NoPrefix(t *testing.T) {
	dir := t.TempDir()
	baseStore, _ := NewLocalStore(LocalConfig{Path: dir})
	defer baseStore.Close()

	store := NewTenantStore(baseStore, nil)

	ctx := context.Background()
	info, err := store.Put(ctx, "uploads/doc.pdf", strings.NewReader("data"), PutOptions{})
	if err != nil {
		t.Fatalf("Put: %v", err)
	}

	if info.Key != "uploads/doc.pdf" {
		t.Errorf("expected key 'uploads/doc.pdf', got %q", info.Key)
	}
}

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/foo/bar", "foo/bar"},
		{"foo//bar", "foo/bar"},
		{"foo\\\\bar", "foo/bar"},
		{"", ""},
	}

	for _, tt := range tests {
		got := normalizeKey(tt.input)
		if got != tt.expected {
			t.Errorf("normalizeKey(%q) = %q, want %q", tt.input, got, tt.expected)
		}
	}
}

func TestConfig_Validate(t *testing.T) {
	// Valid local config
	cfg := Config{
		Provider: ProviderLocal,
		Local:    LocalConfig{Path: t.TempDir()},
	}
	if err := cfg.Validate(); err != nil {
		t.Errorf("expected valid config, got error: %v", err)
	}

	// Invalid provider
	cfg2 := Config{Provider: "invalid"}
	if err := cfg2.Validate(); err == nil {
		t.Error("expected error for invalid provider")
	}

	// Missing S3 bucket
	cfg3 := Config{Provider: ProviderS3}
	if err := cfg3.Validate(); err == nil {
		t.Error("expected error for missing S3 bucket")
	}
}

func TestCredentialSource_Resolve(t *testing.T) {
	// 1. Direct value
	cs := CredentialSource{Value: "my-secret-key"}
	val, err := cs.Resolve()
	if err != nil {
		t.Fatalf("Resolve: %v", err)
	}
	if val != "my-secret-key" {
		t.Errorf("expected 'my-secret-key', got %q", val)
	}

	// 2. Environment variable
	t.Setenv("TEST_STORAGE_SECRET", "env-secret-value")
	cs2 := CredentialSource{EnvVar: "TEST_STORAGE_SECRET"}
	val2, err := cs2.Resolve()
	if err != nil {
		t.Fatalf("Resolve env var: %v", err)
	}
	if val2 != "env-secret-value" {
		t.Errorf("expected 'env-secret-value', got %q", val2)
	}

	// 3. Missing env var
	cs3 := CredentialSource{EnvVar: "NONEXISTENT_VAR_XYZ"}
	_, err = cs3.Resolve()
	if err == nil {
		t.Error("expected error for missing env var")
	}

	// 4. File
	tmpFile := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(tmpFile, []byte("file-secret\n"), 0600); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	cs4 := CredentialSource{File: tmpFile}
	val4, err := cs4.Resolve()
	if err != nil {
		t.Fatalf("Resolve file: %v", err)
	}
	if val4 != "file-secret" {
		t.Errorf("expected 'file-secret', got %q", val4)
	}

	// 5. Missing file
	cs5 := CredentialSource{File: "/nonexistent/path/secret"}
	_, err = cs5.Resolve()
	if err == nil {
		t.Error("expected error for missing file")
	}

	// 6. Empty file
	emptyFile := filepath.Join(t.TempDir(), "empty")
	if err := os.WriteFile(emptyFile, []byte(""), 0600); err != nil {
		t.Fatalf("write empty file: %v", err)
	}
	cs6 := CredentialSource{File: emptyFile}
	_, err = cs6.Resolve()
	if err == nil {
		t.Error("expected error for empty file")
	}

	// 7. nil source
	var cs7 *CredentialSource
	val7, err := cs7.Resolve()
	if err != nil {
		t.Fatalf("nil Resolve: %v", err)
	}
	if val7 != "" {
		t.Errorf("expected empty string for nil, got %q", val7)
	}

	// 8. Priority: Value > EnvVar
	t.Setenv("TEST_STORAGE_SECRET2", "env-value")
	cs8 := CredentialSource{Value: "direct-value", EnvVar: "TEST_STORAGE_SECRET2"}
	val8, err := cs8.Resolve()
	if err != nil {
		t.Fatalf("priority Resolve: %v", err)
	}
	if val8 != "direct-value" {
		t.Errorf("expected 'direct-value' (priority), got %q", val8)
	}

	// 9. Priority: EnvVar > File
	cs9 := CredentialSource{EnvVar: "TEST_STORAGE_SECRET2", File: tmpFile}
	val9, err := cs9.Resolve()
	if err != nil {
		t.Fatalf("priority2 Resolve: %v", err)
	}
	if val9 != "env-value" {
		t.Errorf("expected 'env-value' (env var priority), got %q", val9)
	}
}
