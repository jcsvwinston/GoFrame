package fluent

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	routerpkg "github.com/jcsvwinston/GoFrame/pkg/router"
)

func TestContext_Query(t *testing.T) {
	req := httptest.NewRequest("GET", "/test?key=value", nil)
	ctx := &routerpkg.Context{
		Request: req,
	}
	fc := &Context{Context: ctx}
	
	result := fc.Query("key")
	if result != "value" {
		t.Errorf("Expected value, got %s", result)
	}
}

func TestContext_JSON(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := &routerpkg.Context{
		Request: req,
		Writer:  w,
	}
	fc := &Context{Context: ctx}
	
	err := fc.JSON(200, map[string]string{"message": "hello"})
	if err != nil {
		t.Fatalf("JSON() returned error: %v", err)
	}
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestContext_String(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := &routerpkg.Context{
		Request: req,
		Writer:  w,
	}
	fc := &Context{Context: ctx}
	
	err := fc.String(200, "hello")
	if err != nil {
		t.Fatalf("String() returned error: %v", err)
	}
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
	if !strings.Contains(w.Body.String(), "hello") {
		t.Errorf("Expected body to contain 'hello', got %s", w.Body.String())
	}
}

func TestContext_HTML(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := &routerpkg.Context{
		Request: req,
		Writer:  w,
	}
	fc := &Context{Context: ctx}
	
	err := fc.HTML(200, "<html><body>hello</body></html>")
	if err != nil {
		t.Fatalf("HTML() returned error: %v", err)
	}
	if w.Code != 200 {
		t.Errorf("Expected status 200, got %d", w.Code)
	}
}

func TestContext_Status(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := &routerpkg.Context{
		Request: req,
		Writer:  w,
	}
	fc := &Context{Context: ctx}
	
	fc.Status(404)
	if w.Code != 404 {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

func TestContext_NoContent(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := &routerpkg.Context{
		Request: req,
		Writer:  w,
	}
	fc := &Context{Context: ctx}
	
	err := fc.NoContent()
	if err != nil {
		t.Fatalf("NoContent() returned error: %v", err)
	}
	if w.Code != http.StatusNoContent {
		t.Errorf("Expected status 204, got %d", w.Code)
	}
}

func TestContext_Redirect(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	ctx := &routerpkg.Context{
		Request: req,
		Writer:  w,
	}
	fc := &Context{Context: ctx}
	
	err := fc.Redirect(302, "/new-location")
	if err != nil {
		t.Fatalf("Redirect() returned error: %v", err)
	}
	if w.Code != 302 {
		t.Errorf("Expected status 302, got %d", w.Code)
	}
}

func TestContext_Set(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	ctx := &routerpkg.Context{
		Request: req,
	}
	fc := &Context{Context: ctx}
	
	fc.Set("key", "value")
	// Just verify it doesn't panic - actual storage is in router.Context
}

func TestContext_Get(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	ctx := &routerpkg.Context{
		Request: req,
	}
	fc := &Context{Context: ctx}
	
	result := fc.Get("nonexistent")
	if result != nil {
		t.Errorf("Expected nil for nonexistent key, got %v", result)
	}
}
