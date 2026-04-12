package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GCSStore implements the Store interface using Google Cloud Storage.
type GCSStore struct {
	client       *storage.Client
	bucket       string
	publicBucket string
}

// NewGCSStore creates a GCS client using the provided configuration.
// If cfg.CredentialsSource is configured, it resolves the credentials
// and uses them to authenticate. If empty, uses Application Default Credentials (ADC).
func NewGCSStore(cfg GCSConfig) (*GCSStore, error) {
	ctx := context.Background()

	var opts []option.ClientOption

	// Resolve credentials if configured
	credSource := cfg.CredentialsSource
	if credSource.Value != "" || credSource.EnvVar != "" || credSource.File != "" {
		creds, err := credSource.Resolve()
		if err != nil {
			return nil, fmt.Errorf("storage: gcs resolve credentials: %w", err)
		}
		if creds != "" {
			// Determine if this is a file path (JSON key file) or inline JSON
			if _, err := os.Stat(creds); err == nil {
				// It's a file path
				opts = append(opts, option.WithCredentialsFile(creds))
			} else {
				// Treat as inline JSON credentials
				opts = append(opts, option.WithCredentialsJSON([]byte(creds)))
			}
		}
	}
	// If no credentials configured, ADC is used automatically by the GCS client.

	client, err := storage.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("storage: gcs create client: %w", err)
	}

	return &GCSStore{
		client:       client,
		bucket:       cfg.Bucket,
		publicBucket: cfg.PublicBucket,
	}, nil
}

// Put uploads a file from an io.Reader to GCS.
func (s *GCSStore) Put(ctx context.Context, key string, reader io.Reader, opts PutOptions) (ObjectInfo, error) {
	bucketName := s.bucket
	if opts.Visibility == Public && s.publicBucket != "" {
		bucketName = s.publicBucket
	}

	obj := s.client.Bucket(bucketName).Object(key)
	w := obj.NewWriter(ctx)

	if opts.ContentType != "" {
		w.ContentType = opts.ContentType
	}
	if len(opts.Metadata) > 0 {
		w.Metadata = opts.Metadata
	}

	if _, err := io.Copy(w, reader); err != nil {
		w.Close()
		return ObjectInfo{}, fmt.Errorf("storage: gcs put %q: %w", key, err)
	}

	if err := w.Close(); err != nil {
		return ObjectInfo{}, fmt.Errorf("storage: gcs put %q close writer: %w", key, err)
	}

	attrs := w.Attrs()
	return ObjectInfo{
		Key:         attrs.Name,
		Size:        attrs.Size,
		ContentType: attrs.ContentType,
		Visibility:  opts.Visibility,
		Metadata:    attrs.Metadata,
		UpdatedAt:   attrs.Updated,
	}, nil
}

// Get retrieves a file by key from GCS.
func (s *GCSStore) Get(ctx context.Context, key string) (io.ReadCloser, ObjectInfo, error) {
	// Try the primary bucket first
	obj := s.client.Bucket(s.bucket).Object(key)
	attrs, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, ObjectInfo{}, ErrNotFound(key)
		}
		return nil, ObjectInfo{}, fmt.Errorf("storage: gcs get %q attrs: %w", key, err)
	}

	r, err := obj.NewReader(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return nil, ObjectInfo{}, ErrNotFound(key)
		}
		return nil, ObjectInfo{}, fmt.Errorf("storage: gcs get %q new reader: %w", key, err)
	}

	info := ObjectInfo{
		Key:         attrs.Name,
		Size:        attrs.Size,
		ContentType: attrs.ContentType,
		Visibility:  Private,
		Metadata:    attrs.Metadata,
		UpdatedAt:   attrs.Updated,
	}

	return r, info, nil
}

// Delete removes an object by key from GCS. Idempotent: no error if key doesn't exist.
func (s *GCSStore) Delete(ctx context.Context, key string) error {
	obj := s.client.Bucket(s.bucket).Object(key)
	if err := obj.Delete(ctx); err != nil {
		if err == storage.ErrObjectNotExist {
			return nil
		}
		return fmt.Errorf("storage: gcs delete %q: %w", key, err)
	}
	return nil
}

// Exists checks if a key exists in GCS.
func (s *GCSStore) Exists(ctx context.Context, key string) (bool, error) {
	obj := s.client.Bucket(s.bucket).Object(key)
	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return false, nil
		}
		return false, fmt.Errorf("storage: gcs exists %q: %w", key, err)
	}
	return true, nil
}

// List returns objects with the given prefix from GCS.
func (s *GCSStore) List(ctx context.Context, opts ListOptions) (ListResult, error) {
	query := &storage.Query{
		Prefix:    opts.Prefix,
		Delimiter: opts.Delimiter,
	}
	if opts.Marker != "" {
		query.StartOffset = opts.Marker
	}

	it := s.client.Bucket(s.bucket).Objects(ctx, query)

	var result ListResult
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return result, fmt.Errorf("storage: gcs list: %w", err)
		}

		if attrs.Prefix != "" {
			// This is a common prefix (directory-like entry)
			result.CommonPrefixes = append(result.CommonPrefixes, attrs.Prefix)
		} else {
			result.Objects = append(result.Objects, ObjectInfo{
				Key:         attrs.Name,
				Size:        attrs.Size,
				ContentType: attrs.ContentType,
				Visibility:  Private,
				Metadata:    attrs.Metadata,
				UpdatedAt:   attrs.Updated,
			})
		}

		if opts.Limit > 0 && len(result.Objects) >= opts.Limit {
			result.Truncated = true
			if len(result.Objects) > 0 {
				result.NextMarker = result.Objects[len(result.Objects)-1].Key
			}
			break
		}
	}

	return result, nil
}

// PublicURL returns a publicly accessible URL for a key.
// If cfg.PublicBucket is set and the object is in that bucket, returns
// the direct GCS URL. Otherwise returns empty string.
func (s *GCSStore) PublicURL(ctx context.Context, key string, opts URLConfig) (string, error) {
	if s.publicBucket == "" {
		return "", nil
	}
	// Check if the object exists in the public bucket
	obj := s.client.Bucket(s.publicBucket).Object(key)
	_, err := obj.Attrs(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return "", nil
		}
		return "", fmt.Errorf("storage: gcs public url %q: %w", key, err)
	}

	escapedKey := url.PathEscape(key)
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.publicBucket, escapedKey), nil
}

// SignedURL returns a time-limited URL for accessing a private object.
// Uses 24h expiry by default if expires is zero.
func (s *GCSStore) SignedURL(ctx context.Context, key string, expires time.Duration, opts URLConfig) (string, error) {
	if expires <= 0 {
		expires = 24 * time.Hour
	}

	bucketName := s.bucket
	// Check if object might be in public bucket
	if s.publicBucket != "" {
		obj := s.client.Bucket(s.publicBucket).Object(key)
		if _, err := obj.Attrs(ctx); err == nil {
			bucketName = s.publicBucket
		}
	}

	urlOpts := &storage.SignedURLOptions{
		Scheme:  storage.SigningSchemeV4,
		Method:  "GET",
		Expires: time.Now().Add(expires),
	}

	if opts.ContentType != "" {
		urlOpts.ContentType = opts.ContentType
	}

	signedURL, err := storage.SignedURL(bucketName, key, urlOpts)
	if err != nil {
		return "", fmt.Errorf("storage: gcs signed url %q: %w", key, err)
	}

	return signedURL, nil
}

// Copy copies an object from srcKey to dstKey within the same bucket.
func (s *GCSStore) Copy(ctx context.Context, srcKey, dstKey string) (ObjectInfo, error) {
	src := s.client.Bucket(s.bucket).Object(srcKey)
	dst := s.client.Bucket(s.bucket).Object(dstKey)

	copier := dst.CopierFrom(src)
	attrs, err := copier.Run(ctx)
	if err != nil {
		if err == storage.ErrObjectNotExist {
			return ObjectInfo{}, ErrNotFound(srcKey)
		}
		return ObjectInfo{}, fmt.Errorf("storage: gcs copy %q to %q: %w", srcKey, dstKey, err)
	}

	return ObjectInfo{
		Key:         attrs.Name,
		Size:        attrs.Size,
		ContentType: attrs.ContentType,
		Visibility:  Private,
		Metadata:    attrs.Metadata,
		UpdatedAt:   attrs.Updated,
	}, nil
}

// Close releases the GCS client resources.
func (s *GCSStore) Close() error {
	if err := s.client.Close(); err != nil {
		return fmt.Errorf("storage: gcs close client: %w", err)
	}
	return nil
}

// Ensure GCSStore implements the Store interface at compile time.
var _ Store = (*GCSStore)(nil)
