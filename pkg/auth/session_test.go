package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewSessionManager(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: 24 * time.Hour,
		Secure:   false,
	})
	if sm == nil {
		t.Fatal("expected non-nil session manager")
	}
	if sm.SCS() == nil {
		t.Fatal("expected non-nil underlying SCS")
	}
}

func TestNewSessionManager_Defaults(t *testing.T) {
	sm := NewSessionManager(SessionConfig{})
	if sm.SCS().Lifetime != 72*time.Hour {
		t.Errorf("expected 72h default lifetime, got %v", sm.SCS().Lifetime)
	}
}

func TestSessionManager_Middleware(t *testing.T) {
	sm := NewSessionManager(SessionConfig{})
	mw := sm.Middleware()
	if mw == nil {
		t.Fatal("expected non-nil middleware")
	}
}

func TestNewSessionManager_CookieSettings(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Secure:     true,
		Path:       "/app",
		Domain:     "example.com",
		CookieName: "nucleus_session",
		SameSite:   "strict",
	})

	if !sm.SCS().Cookie.Secure {
		t.Fatal("expected secure cookie")
	}
	if sm.SCS().Cookie.Path != "/app" {
		t.Fatalf("expected /app cookie path, got %q", sm.SCS().Cookie.Path)
	}
	if sm.SCS().Cookie.Domain != "example.com" {
		t.Fatalf("expected cookie domain example.com, got %q", sm.SCS().Cookie.Domain)
	}
	if sm.SCS().Cookie.Name != "nucleus_session" {
		t.Fatalf("expected cookie name nucleus_session, got %q", sm.SCS().Cookie.Name)
	}
	if sm.SCS().Cookie.SameSite != http.SameSiteStrictMode {
		t.Fatalf("expected strict same-site, got %v", sm.SCS().Cookie.SameSite)
	}
}

func TestNewSessionManager_InvalidSameSiteFallsBackToLax(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		SameSite: "invalid-value",
	})

	if sm.SCS().Cookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected lax same-site fallback, got %v", sm.SCS().Cookie.SameSite)
	}
}

func TestSessionManager_FlashData(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-flash-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	var ctx context.Context
	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx = r.Context()
		sm.Flash(ctx, "status", "success")
		if got := sm.GetFlash(ctx, "status"); got != "success" {
			t.Fatalf("expected flash value 'success', got %q", got)
		}

		sm.FlashInt(ctx, "count", 42)
		if got := sm.GetFlashInt(ctx, "count"); got != 42 {
			t.Fatalf("expected flash int 42, got %d", got)
		}

		sm.FlashBool(ctx, "active", true)
		if got := sm.GetFlashBool(ctx, "active"); !got {
			t.Fatalf("expected flash bool true, got %v", got)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionManager_Pull(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-pull-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sm.Put(ctx, "key", "value")
		if got := sm.Pull(ctx, "key"); got != "value" {
			t.Fatalf("expected pulled value 'value', got %q", got)
		}
		if sm.Exists(ctx, "key") {
			t.Fatal("expected key to be removed after pull")
		}

		sm.PutInt(ctx, "count", 100)
		if got := sm.PullInt(ctx, "count"); got != 100 {
			t.Fatalf("expected pulled int 100, got %d", got)
		}

		sm.PutBool(ctx, "flag", true)
		if got := sm.PullBool(ctx, "flag"); !got {
			t.Fatalf("expected pulled bool true, got %v", got)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionManager_ForgetMultiple(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-forget-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sm.Put(ctx, "key1", "value1")
		sm.Put(ctx, "key2", "value2")
		sm.Put(ctx, "key3", "value3")

		sm.Forget(ctx, []string{"key1", "key2"})

		if sm.Exists(ctx, "key1") {
			t.Fatal("expected key1 to be removed")
		}
		if sm.Exists(ctx, "key2") {
			t.Fatal("expected key2 to be removed")
		}
		if !sm.Exists(ctx, "key3") {
			t.Fatal("expected key3 to still exist")
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionManager_Reflash(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-reflash-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sm.Flash(ctx, "message", "hello")
		sm.Reflash(ctx)
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionManager_Keep(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-keep-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sm.Flash(ctx, "username", "john")
		sm.Flash(ctx, "email", "john@example.com")

		sm.Keep(ctx, []string{"username"})

		// Username should be in old namespace, email should not
		// This is a basic test - in practice, you'd need to check the old namespace
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}

func TestSessionManager_Now(t *testing.T) {
	sm := NewSessionManager(SessionConfig{
		Lifetime: time.Hour,
	})

	deadline := time.Now().UTC().Add(time.Hour)
	payload, err := sm.SCS().Codec.Encode(deadline, map[string]interface{}{})
	if err != nil {
		t.Fatalf("encode session payload: %v", err)
	}
	token := "test-now-token"
	if err := sm.SCS().Store.Commit(token, payload, deadline); err != nil {
		t.Fatalf("commit seed payload: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(&http.Cookie{Name: sm.SCS().Cookie.Name, Value: token})

	handler := sm.Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		sm.Now(ctx, "temp", "temporary")
		if got := sm.GetFlash(ctx, "temp"); got != "temporary" {
			t.Fatalf("expected now value 'temporary', got %q", got)
		}
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
}
