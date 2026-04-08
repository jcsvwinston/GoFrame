package router

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMux_ResourceRegistersRESTRoutes(t *testing.T) {
	mux := NewMux()
	mux.Resource("/users", ResourceHandlers{
		List: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("list"))
		},
		Create: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("create"))
		},
		Retrieve: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("get:" + r.PathValue("id")))
		},
		Update: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("update:" + r.PathValue("id")))
		},
		Delete: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("delete:" + r.PathValue("id")))
		},
	})

	cases := []struct {
		method string
		path   string
		code   int
		body   string
	}{
		{method: http.MethodGet, path: "/users/", code: http.StatusOK, body: "list"},
		{method: http.MethodPost, path: "/users/", code: http.StatusCreated, body: "create"},
		{method: http.MethodGet, path: "/users/7", code: http.StatusOK, body: "get:7"},
		{method: http.MethodPut, path: "/users/7", code: http.StatusOK, body: "update:7"},
		{method: http.MethodDelete, path: "/users/7", code: http.StatusOK, body: "delete:7"},
	}

	for _, tc := range cases {
		req := httptest.NewRequest(tc.method, tc.path, nil)
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, req)

		if rec.Code != tc.code {
			t.Fatalf("%s %s expected status %d, got %d", tc.method, tc.path, tc.code, rec.Code)
		}
		if rec.Body.String() != tc.body {
			t.Fatalf("%s %s expected body %q, got %q", tc.method, tc.path, tc.body, rec.Body.String())
		}
	}
}

func TestMux_ResourceSkipsNilHandlers(t *testing.T) {
	mux := NewMux()
	mux.Resource("projects", ResourceHandlers{
		List: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("list"))
		},
	})

	getReq := httptest.NewRequest(http.MethodGet, "/projects/", nil)
	getRec := httptest.NewRecorder()
	mux.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("GET /projects/ expected status 200, got %d", getRec.Code)
	}

	postReq := httptest.NewRequest(http.MethodPost, "/projects/", nil)
	postRec := httptest.NewRecorder()
	mux.ServeHTTP(postRec, postReq)
	if postRec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /projects/ expected status 405 when handler is nil, got %d", postRec.Code)
	}
}

func TestMux_ResourceCanonicalRedirect(t *testing.T) {
	mux := NewMux()
	mux.Resource("/reports", ResourceHandlers{
		List: func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/reports", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusTemporaryRedirect {
		t.Fatalf("expected status 307, got %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); loc != "/reports/" {
		t.Fatalf("expected redirect to /reports/, got %q", loc)
	}
}
