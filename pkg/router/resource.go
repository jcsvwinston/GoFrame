package router

import (
	"strings"
)

// ResourceHandlers groups CRUD handlers for one REST resource.
// Nil handlers are skipped.
type ResourceHandlers struct {
	List     Handler
	Create   Handler
	Retrieve Handler
	Update   Handler
	Delete   Handler
}

// Resource registers a conventional REST route set for one resource prefix:
// - GET    /<resource>/        -> List
// - POST   /<resource>/        -> Create
// - GET    /<resource>/{id}    -> Retrieve
// - PUT    /<resource>/{id}    -> Update
// - DELETE /<resource>/{id}    -> Delete
//
// Example:
//
//	r.Resource("/users", router.ResourceHandlers{ ... })
func (m *Mux) Resource(pattern string, handlers ResourceHandlers) {
	if m == nil {
		return
	}

	base := normalizeResourcePattern(pattern)
	m.Route(base, func(r *Mux) {
		if handlers.List != nil {
			r.Get("/", handlers.List)
		}
		if handlers.Create != nil {
			r.Post("/", handlers.Create)
		}
		if handlers.Retrieve != nil {
			r.Get("/{id}", handlers.Retrieve)
		}
		if handlers.Update != nil {
			r.Put("/{id}", handlers.Update)
		}
		if handlers.Delete != nil {
			r.Delete("/{id}", handlers.Delete)
		}
	})
}

func normalizeResourcePattern(raw string) string {
	pattern := strings.TrimSpace(raw)
	if pattern == "" {
		return "/"
	}
	if !strings.HasPrefix(pattern, "/") {
		pattern = "/" + pattern
	}
	if len(pattern) > 1 {
		pattern = strings.TrimRight(pattern, "/")
	}
	if pattern == "" {
		return "/"
	}
	return pattern
}
