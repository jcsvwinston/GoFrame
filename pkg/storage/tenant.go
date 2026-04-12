package storage

import (
	"context"
	"io"
	"strings"
	"time"
)

// TenantKey is the context key for tenant ID.
type TenantKey struct{}

// TenantStore wraps a Store and automatically prefixes all keys with the
// current tenant ID (extracted from context). This provides automatic
// tenant isolation at the storage level without requiring application code changes.
//
// Key transformation example:
//
//	// App code:
//	store.Put(ctx, "uploads/invoice.pdf", ...)
//
//	// Actual S3 key (if tenant "acme" is in context):
//	"acme/uploads/invoice.pdf"
//
// If no tenant is in context, keys are stored without prefix.
type TenantStore struct {
	store  Store
	getter func(context.Context) string
}

// NewTenantStore creates a tenant-prefixing wrapper.
// The getter function extracts the tenant ID from context.
// Pass nil for getter to disable tenant prefixing.
func NewTenantStore(store Store, getter func(context.Context) string) *TenantStore {
	return &TenantStore{
		store:  store,
		getter: getter,
	}
}

func (t *TenantStore) resolveTenant(ctx context.Context) string {
	if t.getter == nil {
		return ""
	}
	return t.getter(ctx)
}

func (t *TenantStore) prefixKey(ctx context.Context, key string) string {
	tenant := t.resolveTenant(ctx)
	if tenant == "" {
		return key
	}
	// Ensure tenant prefix doesn't double-slash
	tenant = strings.TrimRight(tenant, "/")
	key = strings.TrimLeft(key, "/")
	return tenant + "/" + key
}

func (t *TenantStore) Put(ctx context.Context, key string, reader io.Reader, opts PutOptions) (ObjectInfo, error) {
	if opts.TenantPrefix == "" {
		key = t.prefixKey(ctx, key)
	} else if opts.TenantPrefix != "" {
		// Explicit override: prepend the custom prefix
		key = strings.TrimRight(opts.TenantPrefix, "/") + "/" + strings.TrimLeft(key, "/")
	}
	return t.store.Put(ctx, key, reader, opts)
}

func (t *TenantStore) Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error) {
	key = t.prefixKey(ctx, key)
	return t.store.Get(ctx, key)
}

func (t *TenantStore) Delete(ctx context.Context, key string) error {
	key = t.prefixKey(ctx, key)
	return t.store.Delete(ctx, key)
}

func (t *TenantStore) Exists(ctx context.Context, key string) (bool, error) {
	key = t.prefixKey(ctx, key)
	return t.store.Exists(ctx, key)
}

func (t *TenantStore) List(ctx context.Context, opts ListOptions) (ListResult, error) {
	if tenant := t.resolveTenant(ctx); tenant != "" {
		tenant = strings.TrimRight(tenant, "/")
		if opts.Prefix == "" {
			opts.Prefix = tenant + "/"
		} else {
			opts.Prefix = tenant + "/" + strings.TrimLeft(opts.Prefix, "/")
		}
	}
	return t.store.List(ctx, opts)
}

func (t *TenantStore) PublicURL(ctx context.Context, key string, opts URLConfig) (string, error) {
	key = t.prefixKey(ctx, key)
	return t.store.PublicURL(ctx, key, opts)
}

func (t *TenantStore) SignedURL(ctx context.Context, key string, expires time.Duration, opts URLConfig) (string, error) {
	key = t.prefixKey(ctx, key)
	return t.store.SignedURL(ctx, key, expires, opts)
}

func (t *TenantStore) Copy(ctx context.Context, srcKey, dstKey string) (ObjectInfo, error) {
	srcKey = t.prefixKey(ctx, srcKey)
	dstKey = t.prefixKey(ctx, dstKey)
	return t.store.Copy(ctx, srcKey, dstKey)
}

func (t *TenantStore) Close() error {
	return t.store.Close()
}

// Unwrap returns the underlying store (for type assertions to provider-specific features).
func (t *TenantStore) Unwrap() Store {
	return t.store
}

// UnwrapIfCleaner returns the underlying store for cleanup configuration.
func (t *TenantStore) UnwrapIfCleaner() Store {
	return t.store
}
