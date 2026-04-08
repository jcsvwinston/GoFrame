package router

import (
	"net/http"
	"strings"
	"sync"
)

// Middleware is a function that wraps an http.Handler with additional behavior.
type Middleware = func(http.Handler) http.Handler

// RouteEntry represents a registered route for introspection via Walk.
type RouteEntry struct {
	Method      string
	Pattern     string
	Middlewares int
}

// Mux wraps http.ServeMux with convenience methods for route registration,
// middleware chaining, grouping, and sub-router mounting. It serves as a
// drop-in replacement for chi.Router using only the Go standard library
// (requires Go 1.22+ for method-aware patterns and path value extraction).
type Mux struct {
	mux         *http.ServeMux
	middlewares []Middleware
	handler     http.Handler // cached: middleware chain wrapping mux
	routes      []RouteEntry
	mu          sync.RWMutex
	isGroup     bool // true when this Mux is a Group scope sharing parent's ServeMux
}

// NewMux creates a new Mux backed by a fresh http.ServeMux.
func NewMux() *Mux {
	smux := http.NewServeMux()
	return &Mux{
		mux:     smux,
		handler: smux,
	}
}

// ---------------------------------------------------------------------------
// Middleware
// ---------------------------------------------------------------------------

// Use appends one or more middlewares to the Mux's middleware stack.
// For top-level Mux instances, middlewares are applied via ServeHTTP to all
// requests. For Group scopes, middlewares wrap individual handlers at
// registration time.
func (m *Mux) Use(mws ...Middleware) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.middlewares = append(m.middlewares, mws...)
	if !m.isGroup {
		m.rebuildHandler()
	}
}

// rebuildHandler recomputes the cached handler chain. Must be called under lock.
func (m *Mux) rebuildHandler() {
	h := http.Handler(m.mux)
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		h = m.middlewares[i](h)
	}
	m.handler = h
}

// applyGroupMiddlewares wraps a handler with the middleware stack of a Group
// scope. For non-group muxes the handler is returned unchanged.
func (m *Mux) applyGroupMiddlewares(h http.Handler) http.Handler {
	if !m.isGroup || len(m.middlewares) == 0 {
		return h
	}
	for i := len(m.middlewares) - 1; i >= 0; i-- {
		h = m.middlewares[i](h)
	}
	return h
}

// ---------------------------------------------------------------------------
// Route registration
// ---------------------------------------------------------------------------

// Get registers a handler for GET requests matching pattern.
func (m *Mux) Get(pattern string, h http.HandlerFunc) {
	m.handle("GET", pattern, h)
}

// Post registers a handler for POST requests matching pattern.
func (m *Mux) Post(pattern string, h http.HandlerFunc) {
	m.handle("POST", pattern, h)
}

// Put registers a handler for PUT requests matching pattern.
func (m *Mux) Put(pattern string, h http.HandlerFunc) {
	m.handle("PUT", pattern, h)
}

// Patch registers a handler for PATCH requests matching pattern.
func (m *Mux) Patch(pattern string, h http.HandlerFunc) {
	m.handle("PATCH", pattern, h)
}

// Delete registers a handler for DELETE requests matching pattern.
func (m *Mux) Delete(pattern string, h http.HandlerFunc) {
	m.handle("DELETE", pattern, h)
}

// Handle registers a handler for all HTTP methods matching pattern.
func (m *Mux) Handle(pattern string, h http.Handler) {
	m.handle("", pattern, h)
}

// HandleFunc registers a HandlerFunc for all HTTP methods matching pattern.
func (m *Mux) HandleFunc(pattern string, h http.HandlerFunc) {
	m.handle("", pattern, h)
}

func (m *Mux) handle(method, pattern string, h http.Handler) {
	// In Group scopes, wrap handlers so only routes inside the group are
	// affected by group-local middleware.
	h = m.applyGroupMiddlewares(h)

	p := pattern
	// In ServeMux (Go 1.22+), a trailing slash denotes a subtree match.
	// To preserve Chi's exact-match semantics and avoid pattern conflicts
	// (e.g., "GET /" vs Mount "/admin/"), we force exact matches.
	if strings.HasSuffix(p, "/") && !strings.HasSuffix(p, "{$}") {
		p = p + "{$}"
	}

	if method != "" {
		p = method + " " + p
	}
	m.mux.Handle(p, h)

	m.mu.Lock()
	m.routes = append(m.routes, RouteEntry{
		Method:      method,
		Pattern:     pattern,
		Middlewares: len(m.middlewares),
	})
	m.mu.Unlock()
}

// ---------------------------------------------------------------------------
// Grouping and sub-routing
// ---------------------------------------------------------------------------

// Group creates an inline scope that shares the parent's ServeMux but
// maintains its own middleware stack. Middlewares added via Use inside the
// group only apply to routes registered within that group.
func (m *Mux) Group(fn func(sub *Mux)) {
	sub := &Mux{
		mux:     m.mux,
		isGroup: true,
	}
	// Nested Group scopes inherit parent group middlewares.
	if m.isGroup && len(m.middlewares) > 0 {
		sub.middlewares = append(sub.middlewares, m.middlewares...)
	}
	fn(sub)

	m.mu.Lock()
	m.routes = append(m.routes, sub.routes...)
	m.mu.Unlock()
}

// Route creates a sub-router mounted under the given pattern prefix. The sub-
// router has its own middleware stack and its own route namespace.
func (m *Mux) Route(pattern string, fn func(sub *Mux)) {
	sub := NewMux()
	fn(sub)
	m.Mount(pattern, sub)
}

// Mount registers handler under the given pattern prefix. Requests matching
// the prefix are forwarded to handler with the prefix stripped. If pattern
// does not end with "/", a trailing slash is appended so that the ServeMux
// treats it as a subtree pattern.
func (m *Mux) Mount(pattern string, handler http.Handler) {
	cleanPattern := strings.TrimSpace(pattern)
	if cleanPattern == "" || cleanPattern == "/" {
		// Mounting at root should not strip prefix and must avoid invalid ""
		// patterns in net/http.ServeMux.
		m.mux.Handle("/", m.applyGroupMiddlewares(handler))

		m.mu.Lock()
		m.routes = append(m.routes, RouteEntry{
			Method:      "*",
			Pattern:     "/*",
			Middlewares: len(m.middlewares),
		})
		m.mu.Unlock()
		return
	}
	if !strings.HasPrefix(cleanPattern, "/") {
		cleanPattern = "/" + cleanPattern
	}
	cleanPattern = strings.TrimRight(cleanPattern, "/")

	mounted := http.StripPrefix(cleanPattern, handler)
	mounted = m.applyGroupMiddlewares(mounted)

	// Register subtree handler (with trailing slash).
	m.mux.Handle(cleanPattern+"/", mounted)

	// Register exact match without trailing slash and redirect to canonical
	// subtree path ("/admin" -> "/admin/").
	var exact http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, cleanPattern+"/", http.StatusTemporaryRedirect)
	})
	exact = m.applyGroupMiddlewares(exact)
	m.mux.Handle(cleanPattern, exact)

	m.mu.Lock()
	m.routes = append(m.routes, RouteEntry{
		Method:      "*",
		Pattern:     cleanPattern + "/*",
		Middlewares: len(m.middlewares),
	})
	m.mu.Unlock()
}

// ---------------------------------------------------------------------------
// http.Handler implementation
// ---------------------------------------------------------------------------

// ServeHTTP dispatches the request through the middleware chain and into the
// underlying ServeMux.
func (m *Mux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.mu.RLock()
	h := m.handler
	m.mu.RUnlock()
	h.ServeHTTP(w, r)
}

// ---------------------------------------------------------------------------
// Introspection
// ---------------------------------------------------------------------------

// Walk iterates over all registered routes, calling fn for each one. The
// signature is compatible with the chi.Walk callback API so that callers
// can migrate without code changes beyond the call site.
func (m *Mux) Walk(fn func(method string, route string, handler http.Handler, middlewares ...func(http.Handler) http.Handler) error) error {
	m.mu.RLock()
	snapshot := make([]RouteEntry, len(m.routes))
	copy(snapshot, m.routes)
	m.mu.RUnlock()

	for _, re := range snapshot {
		dummyMWs := make([]func(http.Handler) http.Handler, re.Middlewares)
		if err := fn(re.Method, re.Pattern, nil, dummyMWs...); err != nil {
			return err
		}
	}
	return nil
}
