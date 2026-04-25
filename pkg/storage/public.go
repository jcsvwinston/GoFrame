package storage

import (
	"context"
	"fmt"
	"github.com/jcsvwinston/GoFrame/pkg/router"
	"io"
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
	Get(string, ...router.Handler)
}, publicPath string, storagePrefix string) {
	storagePrefix = normalizeKey(storagePrefix)
	publicPath = strings.TrimRight(publicPath, "/")

	mux.Get(publicPath+"/{filepath...}", func(c *router.Context) error {
		filepath := c.Param("filepath")
		key := storagePrefix + "/" + filepath

		reader, info, err := m.store.Get(c.Request.Context(), key)
		if err != nil {
			return err
		}
		defer reader.Close()

		// Set cache headers for public content
		c.Writer.Header().Set("Content-Type", info.ContentType)
		c.Writer.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		c.Writer.Header().Set("X-Content-Type-Options", "nosniff")

		// Stream the file directly
		c.Writer.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size))
		_, err = io.Copy(c.Writer, reader)
		return err
	})
}

// MountAll registers HTTP handlers for ALL configured public paths.
func (m *PublicMapper) MountAll(mux interface {
	Get(string, ...router.Handler)
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
