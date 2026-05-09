package nucleus

import (
	"net/http"
	"testing"
)

func TestAppBuilder_Get(t *testing.T) {
	builder := New()
	handler := func(c *Context) error {
		return c.String(200, "hello")
	}
	result := builder.Get("/test", handler)
	if result == nil {
		t.Fatal("Get() returned nil")
	}
	if len(result.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(result.routes))
	}
	if result.routes[0].method != "GET" {
		t.Errorf("Expected GET method, got %s", result.routes[0].method)
	}
	if result.routes[0].pattern != "/test" {
		t.Errorf("Expected /test pattern, got %s", result.routes[0].pattern)
	}
}

func TestAppBuilder_Post(t *testing.T) {
	builder := New()
	handler := func(c *Context) error {
		return c.String(200, "hello")
	}
	result := builder.Post("/test", handler)
	if result == nil {
		t.Fatal("Post() returned nil")
	}
	if len(result.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(result.routes))
	}
	if result.routes[0].method != "POST" {
		t.Errorf("Expected POST method, got %s", result.routes[0].method)
	}
}

func TestAppBuilder_Put(t *testing.T) {
	builder := New()
	handler := func(c *Context) error {
		return c.String(200, "hello")
	}
	result := builder.Put("/test", handler)
	if result == nil {
		t.Fatal("Put() returned nil")
	}
	if len(result.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(result.routes))
	}
	if result.routes[0].method != "PUT" {
		t.Errorf("Expected PUT method, got %s", result.routes[0].method)
	}
}

func TestAppBuilder_Delete(t *testing.T) {
	builder := New()
	handler := func(c *Context) error {
		return c.String(200, "hello")
	}
	result := builder.Delete("/test", handler)
	if result == nil {
		t.Fatal("Delete() returned nil")
	}
	if len(result.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(result.routes))
	}
	if result.routes[0].method != "DELETE" {
		t.Errorf("Expected DELETE method, got %s", result.routes[0].method)
	}
}

func TestAppBuilder_Group(t *testing.T) {
	builder := New()
	mw := func(next http.Handler) http.Handler {
		return next
	}
	group := builder.Group("/api", mw)
	if group == nil {
		t.Fatal("Group() returned nil")
	}
	if group.prefix != "/api" {
		t.Errorf("Expected /api prefix, got %s", group.prefix)
	}
	if len(group.middlewares) != 1 {
		t.Errorf("Expected 1 middleware, got %d", len(group.middlewares))
	}
}

func TestRouterGroup_Get(t *testing.T) {
	builder := New()
	group := builder.Group("/api")
	handler := func(c *Context) error {
		return c.String(200, "hello")
	}
	result := group.Get("/test", handler)
	if result == nil {
		t.Fatal("RouterGroup.Get() returned nil")
	}
	if len(builder.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(builder.routes))
	}
	if builder.routes[0].pattern != "/api/test" {
		t.Errorf("Expected /api/test pattern, got %s", builder.routes[0].pattern)
	}
}

func TestRouterGroup_Post(t *testing.T) {
	builder := New()
	group := builder.Group("/api")
	handler := func(c *Context) error {
		return c.String(200, "hello")
	}
	result := group.Post("/test", handler)
	if result == nil {
		t.Fatal("RouterGroup.Post() returned nil")
	}
	if len(builder.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(builder.routes))
	}
	if builder.routes[0].method != "POST" {
		t.Errorf("Expected POST method, got %s", builder.routes[0].method)
	}
}

func TestRouterGroup_Put(t *testing.T) {
	builder := New()
	group := builder.Group("/api")
	handler := func(c *Context) error {
		return c.String(200, "hello")
	}
	result := group.Put("/test", handler)
	if result == nil {
		t.Fatal("RouterGroup.Put() returned nil")
	}
	if len(builder.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(builder.routes))
	}
	if builder.routes[0].method != "PUT" {
		t.Errorf("Expected PUT method, got %s", builder.routes[0].method)
	}
}

func TestRouterGroup_Delete(t *testing.T) {
	builder := New()
	group := builder.Group("/api")
	handler := func(c *Context) error {
		return c.String(200, "hello")
	}
	result := group.Delete("/test", handler)
	if result == nil {
		t.Fatal("RouterGroup.Delete() returned nil")
	}
	if len(builder.routes) != 1 {
		t.Errorf("Expected 1 route, got %d", len(builder.routes))
	}
	if builder.routes[0].method != "DELETE" {
		t.Errorf("Expected DELETE method, got %s", builder.routes[0].method)
	}
}

func TestAppBuilder_Resource(t *testing.T) {
	builder := New()
	resource := &mockResource{}
	result := builder.Resource("/users", resource)
	if result == nil {
		t.Fatal("Resource() returned nil")
	}
	// Should register 5 routes: List, Create, Get, Update, Delete
	if len(builder.routes) != 5 {
		t.Errorf("Expected 5 routes, got %d", len(builder.routes))
	}
}

func TestAppBuilder_Use(t *testing.T) {
	builder := New()
	handler := func(c *Context) error {
		return c.String(200, "hello")
	}
	builder.Get("/test", handler)

	mw := func(next http.Handler) http.Handler {
		return next
	}
	result := builder.Use(mw)
	if result == nil {
		t.Fatal("Use() returned nil")
	}
	// Use() adds middleware to the router, not to individual routes
	// This test just verifies chaining works
}

func TestAppBuilder_SPA(t *testing.T) {
	builder := New()
	cfg := SPAConfig{
		IndexFile: "index.html",
		APIPrefix: "/api",
	}
	result := builder.SPA("./dist", cfg)
	if result == nil {
		t.Fatal("SPA() returned nil")
	}
	if !result.spaEnabled {
		t.Error("Expected spaEnabled to be true")
	}
	if result.spaDir != "./dist" {
		t.Errorf("Expected ./dist, got %s", result.spaDir)
	}
}

func TestAppBuilder_Cors(t *testing.T) {
	builder := New()
	cfg := CorsAllowAll()
	result := builder.Cors(cfg)
	if result == nil {
		t.Fatal("Cors() returned nil")
	}
	// CORS middleware is added via router.Use(), we just verify chaining works
}

// mockResource implements Resource interface for testing
type mockResource struct{}

func (m *mockResource) List(c *Context) error {
	return c.String(200, "list")
}

func (m *mockResource) Create(c *Context) error {
	return c.String(200, "create")
}

func (m *mockResource) Get(c *Context) error {
	return c.String(200, "get")
}

func (m *mockResource) Update(c *Context) error {
	return c.String(200, "update")
}

func (m *mockResource) Delete(c *Context) error {
	return c.String(200, "delete")
}
