package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"example.com/showcase_clean/internal/repositories"
	"example.com/showcase_clean/internal/services"
	"github.com/jcsvwinston/GoFrame/pkg/router"
)

func TestCategoryHandler_CRUDLifecycle(t *testing.T) {
	repository := repositories.NewCategoryRepository()
	service := services.NewCategoryService(repository)
	h := NewCategoryHandler(service)
	r := router.NewMux()
	h.Mount(r)

	createRec := performCategoryRequest(t, r, http.MethodPost, "/categories/", map[string]any{"name": "Books"})
	if createRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, createRec.Code)
	}

	createBody := decodeCategoryJSON(t, createRec.Body.Bytes())
	createData, ok := createBody["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected create response data object, got %T", createBody["data"])
	}

	resourceID, ok := createData["id"].(float64)
	if !ok || resourceID <= 0 {
		t.Fatalf("expected created record id, got %v", createData["id"])
	}
	if got := createData["name"]; got != "Books" {
		t.Fatalf("expected created name %q, got %v", "Books", got)
	}

	secondCreateRec := performCategoryRequest(t, r, http.MethodPost, "/categories/", map[string]any{"name": "Games"})
	if secondCreateRec.Code != http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, secondCreateRec.Code)
	}

	listRec := performCategoryRequest(t, r, http.MethodGet, "/categories/", nil)
	if listRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, listRec.Code)
	}
	listBody := decodeCategoryJSON(t, listRec.Body.Bytes())
	if got := int(listBody["count"].(float64)); got != 2 {
		t.Fatalf("expected list count 2, got %d", got)
	}

	filteredRec := performCategoryRequest(t, r, http.MethodGet, "/categories/?q=book", nil)
	if filteredRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, filteredRec.Code)
	}
	filteredBody := decodeCategoryJSON(t, filteredRec.Body.Bytes())
	if got := int(filteredBody["count"].(float64)); got != 1 {
		t.Fatalf("expected filtered count 1, got %d", got)
	}

	resourcePath := fmt.Sprintf("/categories/%d", int(resourceID))
	getRec := performCategoryRequest(t, r, http.MethodGet, resourcePath, nil)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, getRec.Code)
	}

	updateRec := performCategoryRequest(t, r, http.MethodPut, resourcePath, map[string]any{"name": "Novels"})
	if updateRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updateRec.Code)
	}
	updateBody := decodeCategoryJSON(t, updateRec.Body.Bytes())
	updateData, ok := updateBody["data"].(map[string]any)
	if !ok {
		t.Fatalf("expected update response data object, got %T", updateBody["data"])
	}
	if got := updateData["name"]; got != "Novels" {
		t.Fatalf("expected updated name %q, got %v", "Novels", got)
	}

	updatedFilteredRec := performCategoryRequest(t, r, http.MethodGet, "/categories/?q=nov", nil)
	if updatedFilteredRec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, updatedFilteredRec.Code)
	}
	updatedFilteredBody := decodeCategoryJSON(t, updatedFilteredRec.Body.Bytes())
	if got := int(updatedFilteredBody["count"].(float64)); got != 1 {
		t.Fatalf("expected filtered count 1 after update, got %d", got)
	}

	deleteRec := performCategoryRequest(t, r, http.MethodDelete, resourcePath, nil)
	if deleteRec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, deleteRec.Code)
	}

	finalListRec := performCategoryRequest(t, r, http.MethodGet, "/categories/", nil)
	finalListBody := decodeCategoryJSON(t, finalListRec.Body.Bytes())
	if got := int(finalListBody["count"].(float64)); got != 1 {
		t.Fatalf("expected list count 1 after delete, got %d", got)
	}

	badIDRec := performCategoryRequest(t, r, http.MethodGet, "/categories/not-a-number", nil)
	assertStructuredErrorResponse(t, badIDRec, http.StatusBadRequest, "BAD_REQUEST")

	missingRec := performCategoryRequest(t, r, http.MethodGet, resourcePath, nil)
	assertStructuredErrorResponse(t, missingRec, http.StatusNotFound, "NOT_FOUND")
}

func TestCategoryHandler_RejectsInvalidPayload(t *testing.T) {
	repository := repositories.NewCategoryRepository()
	service := services.NewCategoryService(repository)
	h := NewCategoryHandler(service)
	r := router.NewMux()
	h.Mount(r)

	rec := performCategoryRequest(t, r, http.MethodPost, "/categories/", map[string]any{"name": "  "})
	assertStructuredErrorResponse(t, rec, http.StatusBadRequest, "BAD_REQUEST")
}

func performCategoryRequest(t *testing.T, handler http.Handler, method, path string, payload map[string]any) *httptest.ResponseRecorder {
	t.Helper()

	var body bytes.Buffer
	if payload != nil {
		if err := json.NewEncoder(&body).Encode(payload); err != nil {
			t.Fatalf("encode request body failed: %v", err)
		}
	}

	req := httptest.NewRequest(method, path, &body)
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

func decodeCategoryJSON(t *testing.T, raw []byte) map[string]any {
	t.Helper()

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		t.Fatalf("decode response failed: %v raw=%s", err, string(raw))
	}
	return payload
}

func assertStructuredErrorResponse(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("expected status %d, got %d body=%s", status, rec.Code, rec.Body.String())
	}

	body := decodeCategoryJSON(t, rec.Body.Bytes())
	errorBody, ok := body["error"].(map[string]any)
	if !ok {
		t.Fatalf("expected structured error body, got %#v", body)
	}
	if got := errorBody["code"]; got != code {
		t.Fatalf("expected error code %q, got %v", code, got)
	}
	if message, ok := errorBody["message"].(string); !ok || message == "" {
		t.Fatalf("expected non-empty error message, got %#v", errorBody)
	}
}
