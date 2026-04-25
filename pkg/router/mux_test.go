package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMux_NestedGroupMiddlewareInheritance(t *testing.T) {
	m := NewMux()

	outerMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Outer-MW", "1")
			next.ServeHTTP(w, r)
		})
	}

	innerMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Inner-MW", "1")
			next.ServeHTTP(w, r)
		})
	}

	m.Group(func(r *Mux) {
		r.Use(outerMW)
		r.Group(func(r *Mux) {
			r.Use(innerMW)
			r.Get("/nested", func(c *Context) error {
				return c.NoContent()
			})
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/nested", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if got := rec.Header().Get("X-Outer-MW"); got != "1" {
		t.Fatalf("expected outer middleware header, got %q", got)
	}
	if got := rec.Header().Get("X-Inner-MW"); got != "1" {
		t.Fatalf("expected inner middleware header, got %q", got)
	}
}

func TestMux_MountRoot(t *testing.T) {
	m := NewMux()
	sub := NewMux()
	sub.Get("/health", func(c *Context) error {
		return c.NoContent()
	})

	m.Mount("/", sub)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
}

func TestMux_GroupRouteInheritsGroupMiddleware(t *testing.T) {
	m := NewMux()

	groupMW := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Group-MW", "1")
			next.ServeHTTP(w, r)
		})
	}

	m.Group(func(r *Mux) {
		r.Use(groupMW)
		r.Route("/api", func(sub *Mux) {
			sub.Get("/ping", func(c *Context) error {
				return c.NoContent()
			})
		})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/ping", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", rec.Code)
	}
	if got := rec.Header().Get("X-Group-MW"); got != "1" {
		t.Fatalf("expected group middleware header, got %q", got)
	}
}

func TestMux_MountExactPrefixRedirectsToCanonicalSubtree(t *testing.T) {
	m := NewMux()
	sub := NewMux()
	sub.Get("/", func(c *Context) error {
		return c.NoContent()
	})

	m.Mount("/admin", sub)

	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	rec := httptest.NewRecorder()
	m.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected 307, got %d", rec.Code)
	}
	if got := rec.Header().Get("Location"); got != "/admin/" {
		t.Fatalf("expected redirect location /admin/, got %q", got)
	}
}
