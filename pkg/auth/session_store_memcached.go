package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
)

const defaultSessionMemcachedPrefix = "nucleus:sessions:"

// MemcachedSessionStore persists sessions in Memcached with key TTL.
type MemcachedSessionStore struct {
	client    *memcache.Client
	keyPrefix string
}

// NewMemcachedSessionStore creates a Memcached-backed session store from an existing client.
func NewMemcachedSessionStore(client *memcache.Client, keyPrefix string) (*MemcachedSessionStore, error) {
	if client == nil {
		return nil, fmt.Errorf("new memcached session store: nil memcached client")
	}
	if keyPrefix == "" {
		keyPrefix = defaultSessionMemcachedPrefix
	}

	return &MemcachedSessionStore{
		client:    client,
		keyPrefix: keyPrefix,
	}, nil
}

// NewMemcachedSessionStoreFromServers creates a memcache.Client and a Memcached session store.
func NewMemcachedSessionStoreFromServers(servers []string, keyPrefix string) (*MemcachedSessionStore, *memcache.Client, error) {
	client := memcache.New(servers...)
	store, err := NewMemcachedSessionStore(client, keyPrefix)
	if err != nil {
		return nil, nil, err
	}

	return store, client, nil
}

// Delete removes the session token from the store.
func (s *MemcachedSessionStore) Delete(token string) error {
	return s.DeleteCtx(context.Background(), token)
}

// Find retrieves the session payload for token.
func (s *MemcachedSessionStore) Find(token string) ([]byte, bool, error) {
	return s.FindCtx(context.Background(), token)
}

// Commit stores the session payload for token with absolute expiry.
func (s *MemcachedSessionStore) Commit(token string, b []byte, expiry time.Time) error {
	return s.CommitCtx(context.Background(), token, b, expiry)
}

// All returns all active sessions visible from the configured key prefix.
// Note: Memcached doesn't support key listing, so this returns an empty map.
// Use Redis or SQL if you need to list all sessions.
func (s *MemcachedSessionStore) All() (map[string][]byte, error) {
	return s.AllCtx(context.Background())
}

// DeleteCtx removes the session token from Memcached.
func (s *MemcachedSessionStore) DeleteCtx(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.client.Delete(s.key(token)); err != nil && err != memcache.ErrCacheMiss {
		return fmt.Errorf("memcached session delete: %w", err)
	}
	return nil
}

// FindCtx retrieves the session payload from Memcached.
func (s *MemcachedSessionStore) FindCtx(ctx context.Context, token string) ([]byte, bool, error) {
	if token == "" {
		return nil, false, nil
	}
	item, err := s.client.Get(s.key(token))
	if err == memcache.ErrCacheMiss {
		return nil, false, nil
	}
	if err != nil {
		return nil, false, fmt.Errorf("memcached session find: %w", err)
	}

	result := make([]byte, len(item.Value))
	copy(result, item.Value)
	return result, true, nil
}

// CommitCtx stores the session payload in Memcached with a TTL derived from expiry.
func (s *MemcachedSessionStore) CommitCtx(ctx context.Context, token string, b []byte, expiry time.Time) error {
	if token == "" {
		return fmt.Errorf("memcached session commit: empty token")
	}
	if expiry.IsZero() {
		return fmt.Errorf("memcached session commit: zero expiry")
	}

	ttl := time.Until(expiry)
	if ttl <= 0 {
		return s.DeleteCtx(ctx, token)
	}

	item := &memcache.Item{
		Key:        s.key(token),
		Value:      b,
		Expiration: int32(ttl.Seconds()),
	}
	if err := s.client.Set(item); err != nil {
		return fmt.Errorf("memcached session commit: %w", err)
	}
	return nil
}

// AllCtx returns all active sessions.
// Note: Memcached doesn't support key listing, so this returns an empty map.
// Use Redis or SQL if you need to list all sessions.
func (s *MemcachedSessionStore) AllCtx(ctx context.Context) (map[string][]byte, error) {
	return map[string][]byte{}, nil
}

func (s *MemcachedSessionStore) key(token string) string {
	return s.keyPrefix + token
}
