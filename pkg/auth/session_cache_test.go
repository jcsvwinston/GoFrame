package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSessionCache_GetPut(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})
	cache := NewSessionCache(sm)

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-cache-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cache.Put(ctx, "key", "value", 0)
		value, ok := cache.Get(ctx, "key")
		if !ok {
			t.Fatal("expected key to exist")
		}
		if value != "value" {
			t.Fatalf("expected value 'value', got %q", value)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionCache_PutIntGetInt(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})
	cache := NewSessionCache(sm)

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-cache-int-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cache.PutInt(ctx, "count", 42, 0)
		value, ok := cache.GetInt(ctx, "count")
		if !ok {
			t.Fatal("expected key to exist")
		}
		if value != 42 {
			t.Fatalf("expected int 42, got %d", value)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionCache_PutBoolGetBool(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})
	cache := NewSessionCache(sm)

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-cache-bool-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cache.PutBool(ctx, "active", true, 0)
		value, ok := cache.GetBool(ctx, "active")
		if !ok {
			t.Fatal("expected key to exist")
		}
		if !value {
			t.Fatal("expected bool true")
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionCache_TTL(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})
	cache := NewSessionCache(sm)

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-cache-ttl-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Test with zero TTL (persists with session)
		cache.Put(ctx, "key", "value", 0)
		if _, ok := cache.Get(ctx, "key"); !ok {
			t.Fatal("expected key to exist with zero TTL")
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionCache_Remember(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})
	cache := NewSessionCache(sm)

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-cache-remember-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	callCount := 0
	fn := func() (string, error) {
		callCount++
		return "computed", nil
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// First call should execute function
		value, err := cache.Remember(ctx, "key", 0, fn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != "computed" {
			t.Fatalf("expected 'computed', got %q", value)
		}
		if callCount != 1 {
			t.Fatalf("expected function to be called once, got %d", callCount)
		}

		// Second call should use cached value
		value, err = cache.Remember(ctx, "key", 0, fn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != "computed" {
			t.Fatalf("expected 'computed', got %q", value)
		}
		if callCount != 1 {
			t.Fatalf("expected function to not be called again, got %d", callCount)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionCache_RememberInt(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})
	cache := NewSessionCache(sm)

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-cache-remember-int-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	callCount := 0
	fn := func() (int, error) {
		callCount++
		return 100, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		value, err := cache.RememberInt(ctx, "key", 0, fn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if value != 100 {
			t.Fatalf("expected 100, got %d", value)
		}
		if callCount != 1 {
			t.Fatalf("expected function to be called once, got %d", callCount)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionCache_RememberBool(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})
	cache := NewSessionCache(sm)

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-cache-remember-bool-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	callCount := 0
	fn := func() (bool, error) {
		callCount++
		return true, nil
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		value, err := cache.RememberBool(ctx, "key", 0, fn)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !value {
			t.Fatal("expected true")
		}
		if callCount != 1 {
			t.Fatalf("expected function to be called once, got %d", callCount)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionCache_Forget(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})
	cache := NewSessionCache(sm)

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-cache-forget-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cache.Put(ctx, "key", "value", 0)
		cache.Forget(ctx, "key")

		if _, ok := cache.Get(ctx, "key"); ok {
			t.Fatal("expected key to be removed")
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionCache_Has(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})
	cache := NewSessionCache(sm)

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-cache-has-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		if cache.Has(ctx, "key") {
			t.Fatal("expected key to not exist initially")
		}

		cache.Put(ctx, "key", "value", 0)
		if !cache.Has(ctx, "key") {
			t.Fatal("expected key to exist after put")
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionCache_Flush(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})
	cache := NewSessionCache(sm)

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-cache-flush-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		cache.Put(ctx, "key1", "value1", 0)
		cache.Put(ctx, "key2", "value2", 0)
		cache.Put(ctx, "key3", "value3", 0)

		cache.Flush(ctx)

		if cache.Has(ctx, "key1") {
			t.Fatal("expected key1 to be removed after flush")
		}
		if cache.Has(ctx, "key2") {
			t.Fatal("expected key2 to be removed after flush")
		}
		if cache.Has(ctx, "key3") {
			t.Fatal("expected key3 to be removed after flush")
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}
