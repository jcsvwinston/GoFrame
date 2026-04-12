package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// PublicMapper handles the mapping between public URL paths and storage keys.
// It is used both for generating public URLs and for mounting HTTP handlers.
type PublicMapper struct {
	store         Store
	publicPaths   map[string]string // "/media" -> "storage/public/media/"
	publicURLBase string            // "https://cdn.example.com"
}

// NewPublicMapper creates a public path mapper.
func NewPublicMapper(store Store, publicPaths map[string]string, publicURLBase string) *PublicMapper {
	return &PublicMapper{
		store:         store,
		publicPaths:   publicPaths,
		publicURLBase: strings.TrimRight(publicURLBase, "/"),
	}
}

// PublicURL constructs a public URL for a storage key.
// It checks if the key matches any configured public path and returns
// the corresponding public URL. Returns empty string if no mapping exists.
func (m *PublicMapper) PublicURL(ctx context.Context, key string, opts URLConfig) (string, error) {
	key = normalizeKey(key)

	// Check if this key matches any public path
	for publicPath, storagePrefix := range m.publicPaths {
		storagePrefix = normalizeKey(storagePrefix)
		if strings.HasPrefix(key, storagePrefix) {
			relativePath := strings.TrimPrefix(key, storagePrefix)
			relativePath = strings.TrimLeft(relativePath, "/")
			return m.publicURLBase + publicPath + "/" + relativePath, nil
		}
	}

	// Fall back to signed URL for private objects
	if m.store != nil {
		return m.store.SignedURL(ctx, key, opts.Expires, opts)
	}
	return "", nil
}

// Mount registers HTTP handlers for all configured public paths.
// Requests to /media/* will be served directly from the storage backend,
// bypassing the need for signed URLs.
//
// Example:
//
//	app.Storage.Public().Mount(router, "/media", "storage/public/media/")
//	// GET /media/blog/hero.png -> serves storage/public/media/blog/hero.png from storage
func (m *PublicMapper) Mount(mux interface {
	Get(string, http.HandlerFunc)
}, publicPath string, storagePrefix string) {
	storagePrefix = normalizeKey(storagePrefix)
	publicPath = strings.TrimRight(publicPath, "/")

	mux.Get(publicPath+"/{filepath...}", func(w http.ResponseWriter, r *http.Request) {
		filepath := r.PathValue("filepath")
		key := storagePrefix + "/" + filepath

		reader, info, err := m.store.Get(r.Context(), key)
		if err != nil {
			if _, ok := err.(ErrNotFound); ok {
				http.NotFound(w, r)
				return
			}
			http.Error(w, "storage error", http.StatusInternalServerError)
			return
		}
		defer reader.Close()

		// Set cache headers for public content
		w.Header().Set("Content-Type", info.ContentType)
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Stream the file directly
		w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size))
		io.Copy(w, reader)
	})
}

// MountAll registers HTTP handlers for ALL configured public paths.
func (m *PublicMapper) MountAll(mux interface {
	Get(string, http.HandlerFunc)
}) {
	for publicPath, storagePrefix := range m.publicPaths {
		m.Mount(mux, publicPath, storagePrefix)
	}
}

// IsPublicKey checks if a key falls under any public path mapping.
func (m *PublicMapper) IsPublicKey(key string) bool {
	key = normalizeKey(key)
	for _, storagePrefix := range m.publicPaths {
		if strings.HasPrefix(key, normalizeKey(storagePrefix)) {
			return true
		}
	}
	return false
}
