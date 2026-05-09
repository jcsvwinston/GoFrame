package nucleus

import (
	"net/http"

	routerpkg "github.com/jcsvwinston/nucleus/pkg/router"
)

// Handler is the simplified handler function signature
type Handler func(*Context) error

// Middleware is the middleware function signature
type Middleware func(http.Handler) http.Handler

// Get registers a GET route
func (b *AppBuilder) Get(path string, handlers ...Handler) *AppBuilder {
	routerHandlers := make([]routerpkg.Handler, len(handlers))
	for i, h := range handlers {
		routerHandlers[i] = adaptHandler(h)
	}
	b.routes = append(b.routes, routeDef{
		method:   "GET",
		pattern:  path,
		handlers: routerHandlers,
	})
	return b
}

// Post registers a POST route
func (b *AppBuilder) Post(path string, handlers ...Handler) *AppBuilder {
	routerHandlers := make([]routerpkg.Handler, len(handlers))
	for i, h := range handlers {
		routerHandlers[i] = adaptHandler(h)
	}
	b.routes = append(b.routes, routeDef{
		method:   "POST",
		pattern:  path,
		handlers: routerHandlers,
	})
	return b
}

// Put registers a PUT route
func (b *AppBuilder) Put(path string, handlers ...Handler) *AppBuilder {
	routerHandlers := make([]routerpkg.Handler, len(handlers))
	for i, h := range handlers {
		routerHandlers[i] = adaptHandler(h)
	}
	b.routes = append(b.routes, routeDef{
		method:   "PUT",
		pattern:  path,
		handlers: routerHandlers,
	})
	return b
}

// Delete registers a DELETE route
func (b *AppBuilder) Delete(path string, handlers ...Handler) *AppBuilder {
	routerHandlers := make([]routerpkg.Handler, len(handlers))
	for i, h := range handlers {
		routerHandlers[i] = adaptHandler(h)
	}
	b.routes = append(b.routes, routeDef{
		method:   "DELETE",
		pattern:  path,
		handlers: routerHandlers,
	})
	return b
}

// Group creates a route group with prefix and middleware
func (b *AppBuilder) Group(prefix string, mws ...Middleware) *RouterGroup {
	group := &RouterGroup{
		builder:     b,
		prefix:      prefix,
		middlewares: mws,
	}
	return group
}

// RouterGroup represents a group of routes with common prefix and middleware
type RouterGroup struct {
	builder     *AppBuilder
	prefix      string
	middlewares []Middleware
}

// Get registers a GET route in the group
func (g *RouterGroup) Get(path string, handlers ...Handler) *RouterGroup {
	fullPath := g.prefix + path
	routerHandlers := make([]routerpkg.Handler, len(handlers))
	for i, h := range handlers {
		routerHandlers[i] = adaptHandler(h)
	}

	middlewares := make([]func(http.Handler) http.Handler, len(g.middlewares))
	for i, m := range g.middlewares {
		middlewares[i] = m
	}

	g.builder.routes = append(g.builder.routes, routeDef{
		method:      "GET",
		pattern:     fullPath,
		handlers:    routerHandlers,
		middlewares: middlewares,
	})
	return g
}

// Post registers a POST route in the group
func (g *RouterGroup) Post(path string, handlers ...Handler) *RouterGroup {
	fullPath := g.prefix + path
	routerHandlers := make([]routerpkg.Handler, len(handlers))
	for i, h := range handlers {
		routerHandlers[i] = adaptHandler(h)
	}

	middlewares := make([]func(http.Handler) http.Handler, len(g.middlewares))
	for i, m := range g.middlewares {
		middlewares[i] = m
	}

	g.builder.routes = append(g.builder.routes, routeDef{
		method:      "POST",
		pattern:     fullPath,
		handlers:    routerHandlers,
		middlewares: middlewares,
	})
	return g
}

// Put registers a PUT route in the group
func (g *RouterGroup) Put(path string, handlers ...Handler) *RouterGroup {
	fullPath := g.prefix + path
	routerHandlers := make([]routerpkg.Handler, len(handlers))
	for i, h := range handlers {
		routerHandlers[i] = adaptHandler(h)
	}

	middlewares := make([]func(http.Handler) http.Handler, len(g.middlewares))
	for i, m := range g.middlewares {
		middlewares[i] = m
	}

	g.builder.routes = append(g.builder.routes, routeDef{
		method:      "PUT",
		pattern:     fullPath,
		handlers:    routerHandlers,
		middlewares: middlewares,
	})
	return g
}

// Delete registers a DELETE route in the group
func (g *RouterGroup) Delete(path string, handlers ...Handler) *RouterGroup {
	fullPath := g.prefix + path
	routerHandlers := make([]routerpkg.Handler, len(handlers))
	for i, h := range handlers {
		routerHandlers[i] = adaptHandler(h)
	}

	middlewares := make([]func(http.Handler) http.Handler, len(g.middlewares))
	for i, m := range g.middlewares {
		middlewares[i] = m
	}

	g.builder.routes = append(g.builder.routes, routeDef{
		method:      "DELETE",
		pattern:     fullPath,
		handlers:    routerHandlers,
		middlewares: middlewares,
	})
	return g
}

// adaptHandler converts a simplified Handler to router.Handler
func adaptHandler(h Handler) routerpkg.Handler {
	return func(c *routerpkg.Context) error {
		return h(&Context{Context: c})
	}
}

// Resource registers RESTful routes for a resource
type Resource interface {
	List(c *Context) error
	Create(c *Context) error
	Get(c *Context) error
	Update(c *Context) error
	Delete(c *Context) error
}

// Resource registers RESTful routes
func (b *AppBuilder) Resource(path string, r Resource) *AppBuilder {
	b.Get(path, r.List)
	b.Post(path, r.Create)
	b.Get(path+"/:id", r.Get)
	b.Put(path+"/:id", r.Update)
	b.Delete(path+"/:id", r.Delete)
	return b
}

// Use adds middleware to the application
func (b *AppBuilder) Use(mws ...Middleware) *AppBuilder {
	for _, route := range b.routes {
		for _, m := range mws {
			route.middlewares = append(route.middlewares, m)
		}
	}
	return b
}
