package openapi

import (
	"net/http"
)

// DocumentProvider returns the current OpenAPI document to serve.
type DocumentProvider func() *Document

// Handler builds an http.Handler that serves the document as JSON.
func Handler(provider DocumentProvider) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		doc := (*Document)(nil)
		if provider != nil {
			doc = provider()
		}
		if doc == nil {
			http.Error(w, "openapi: document provider returned nil", http.StatusServiceUnavailable)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		if err := WriteJSON(w, doc); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

// HandlerFunc adapts a provider into a GET-friendly handler func.
func HandlerFunc(provider DocumentProvider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		Handler(provider).ServeHTTP(w, r)
	}
}
