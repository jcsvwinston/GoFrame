package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/jcsvwinston/GoFrame/pkg/app"
)

func TestExampleMVCAPIAdmin_Smoke(t *testing.T) {
	a, cleanup := newExampleTestApp(t)
	defer cleanup()

	respHome := mustGET(t, a.Router, "/")
	if respHome.StatusCode != http.StatusOK {
		t.Fatalf("home status=%d", respHome.StatusCode)
	}
	bodyHome := mustReadBody(t, respHome)
	if !strings.Contains(bodyHome, "GoFrame MVC") || !strings.Contains(bodyHome, "API Example") {
		t.Fatalf("home body does not contain title: %s", bodyHome)
	}

	respHealth := mustGET(t, a.Router, "/api/health")
	if respHealth.StatusCode != http.StatusOK {
		t.Fatalf("health status=%d", respHealth.StatusCode)
	}
	var health map[string]any
	mustDecodeJSON(t, respHealth.Body, &health)
	if health["status"] != "ok" {
		t.Fatalf("unexpected health payload: %#v", health)
	}

	respArticles := mustGET(t, a.Router, "/api/articles")
	if respArticles.StatusCode != http.StatusOK {
		t.Fatalf("articles status=%d", respArticles.StatusCode)
	}
	var listBefore struct {
		Items []articleDTO `json:"items"`
		Total int          `json:"total"`
	}
	mustDecodeJSON(t, respArticles.Body, &listBefore)
	if listBefore.Total < 1 || len(listBefore.Items) < 1 {
		t.Fatalf("expected seeded data in /api/articles, got total=%d len=%d", listBefore.Total, len(listBefore.Items))
	}

	payload := map[string]any{
		"title":     "E2E Smoke Article",
		"content":   "Created from smoke test",
		"published": true,
	}
	body, _ := json.Marshal(payload)
	createRes := mustRequest(t, a.Router, http.MethodPost, "/api/articles", bytes.NewReader(body), map[string]string{
		"Content-Type": "application/json",
	})
	if createRes.StatusCode != http.StatusCreated {
		raw := mustReadBody(t, createRes)
		t.Fatalf("create status=%d body=%s", createRes.StatusCode, raw)
	}
	var created map[string]any
	mustDecodeJSON(t, createRes.Body, &created)
	if created["id"] == nil {
		t.Fatalf("create response missing id: %#v", created)
	}

	respAfter := mustGET(t, a.Router, "/api/articles")
	if respAfter.StatusCode != http.StatusOK {
		t.Fatalf("articles (after create) status=%d", respAfter.StatusCode)
	}
	var listAfter struct {
		Items []articleDTO `json:"items"`
		Total int          `json:"total"`
	}
	mustDecodeJSON(t, respAfter.Body, &listAfter)
	if listAfter.Total <= listBefore.Total {
		t.Fatalf("expected total to increase after create (before=%d after=%d)", listBefore.Total, listAfter.Total)
	}
	if !containsArticleTitle(listAfter.Items, "E2E Smoke Article") {
		t.Fatalf("created article not found in list: %#v", listAfter.Items)
	}

	// Live feature-flag demo: published_only (default false for preview mode).
	unpublishedPayload := map[string]any{
		"title":     "Draft Preview Article",
		"content":   "Should appear only when preview mode is enabled",
		"published": false,
	}
	unpublishedBody, _ := json.Marshal(unpublishedPayload)
	unpublishedRes := mustRequest(t, a.Router, http.MethodPost, "/api/articles", bytes.NewReader(unpublishedBody), map[string]string{
		"Content-Type": "application/json",
	})
	if unpublishedRes.StatusCode != http.StatusCreated {
		raw := mustReadBody(t, unpublishedRes)
		t.Fatalf("create unpublished status=%d body=%s", unpublishedRes.StatusCode, raw)
	}

	respLiveFlagDefault := mustGET(t, a.Router, "/api/articles/live-flag")
	if respLiveFlagDefault.StatusCode != http.StatusOK {
		t.Fatalf("live-flag default status=%d", respLiveFlagDefault.StatusCode)
	}
	var liveFlagDefault struct {
		FeatureFlag string       `json:"feature_flag"`
		Enabled     bool         `json:"enabled"`
		Mode        string       `json:"mode"`
		Items       []articleDTO `json:"items"`
	}
	mustDecodeJSON(t, respLiveFlagDefault.Body, &liveFlagDefault)
	if liveFlagDefault.FeatureFlag != "articles_preview_mode" {
		t.Fatalf("unexpected feature_flag: %q", liveFlagDefault.FeatureFlag)
	}
	if liveFlagDefault.Enabled {
		t.Fatalf("expected default preview mode disabled")
	}
	if liveFlagDefault.Mode != "published_only" {
		t.Fatalf("unexpected mode when disabled: %q", liveFlagDefault.Mode)
	}
	if containsArticleTitle(liveFlagDefault.Items, "Draft Preview Article") {
		t.Fatalf("draft article should not be visible with preview mode disabled")
	}

	a.Admin.SetFeatureFlag("articles_preview_mode", true)
	respLiveFlagEnabled := mustGET(t, a.Router, "/api/articles/live-flag")
	if respLiveFlagEnabled.StatusCode != http.StatusOK {
		t.Fatalf("live-flag enabled status=%d", respLiveFlagEnabled.StatusCode)
	}
	var liveFlagEnabled struct {
		Enabled bool         `json:"enabled"`
		Mode    string       `json:"mode"`
		Items   []articleDTO `json:"items"`
	}
	mustDecodeJSON(t, respLiveFlagEnabled.Body, &liveFlagEnabled)
	if !liveFlagEnabled.Enabled {
		t.Fatalf("expected preview mode enabled")
	}
	if liveFlagEnabled.Mode != "preview_all" {
		t.Fatalf("unexpected mode when enabled: %q", liveFlagEnabled.Mode)
	}
	if !containsArticleTitle(liveFlagEnabled.Items, "Draft Preview Article") {
		t.Fatalf("draft article should be visible with preview mode enabled")
	}

	respAdmin := mustGET(t, a.Router, "/admin/")
	if respAdmin.StatusCode != http.StatusFound {
		t.Fatalf("admin index status=%d", respAdmin.StatusCode)
	}
	if got := respAdmin.Header.Get("Location"); !strings.HasPrefix(got, "/admin/login?next=") {
		t.Fatalf("expected redirect to /admin/login with next parameter, got %q", got)
	}

	adminCookies := mustAdminLogin(t, a.Router, "/admin/login", "admin", "supersecret123")

	respAdmin = mustRequestWithCookies(t, a.Router, http.MethodGet, "/admin/", nil, nil, adminCookies)
	if respAdmin.StatusCode != http.StatusOK {
		t.Fatalf("admin index authenticated status=%d", respAdmin.StatusCode)
	}
	bodyAdmin := mustReadBody(t, respAdmin)
	if !strings.Contains(bodyAdmin, `content="/admin"`) {
		t.Fatalf("admin index missing injected prefix metadata")
	}
	if !strings.Contains(bodyAdmin, `./assets/`) {
		t.Fatalf("admin index missing Vite asset references")
	}

	respAdminModels := mustRequestWithCookies(t, a.Router, http.MethodGet, "/admin/api/models", nil, nil, adminCookies)
	if respAdminModels.StatusCode != http.StatusOK {
		t.Fatalf("admin models status=%d", respAdminModels.StatusCode)
	}
	var adminModels struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	mustDecodeJSON(t, respAdminModels.Body, &adminModels)
	if !containsModel(adminModels.Models, "Article") {
		t.Fatalf("Article model not found in admin models payload: %#v", adminModels.Models)
	}

	assetPath := regexp.MustCompile(`\./assets/[^"]+\.js`).FindString(bodyAdmin)
	if assetPath == "" {
		t.Fatalf("admin index missing javascript asset path")
	}

	respComponents := mustRequestWithCookies(t, a.Router, http.MethodGet, "/admin/"+strings.TrimPrefix(assetPath, "./"), nil, nil, adminCookies)
	if respComponents.StatusCode != http.StatusOK {
		t.Fatalf("vite asset status=%d", respComponents.StatusCode)
	}
	bodyComponents := mustReadBody(t, respComponents)
	if !strings.Contains(bodyComponents, "createRoot") {
		t.Fatalf("vite asset missing expected bundle content")
	}
}

func TestExampleMVCAPI_Minimal_Smoke(t *testing.T) {
	a, cleanup := newExampleTestApp(t)
	defer cleanup()

	respHome := mustGET(t, a.Router, "/")
	if respHome.StatusCode != http.StatusOK {
		t.Fatalf("home status=%d", respHome.StatusCode)
	}
	bodyHome := mustReadBody(t, respHome)
	if !strings.Contains(bodyHome, "GoFrame MVC") || !strings.Contains(bodyHome, "API Example") {
		t.Fatalf("home body does not contain title: %s", bodyHome)
	}

	respHealth := mustGET(t, a.Router, "/api/health")
	if respHealth.StatusCode != http.StatusOK {
		t.Fatalf("health status=%d", respHealth.StatusCode)
	}
	var health map[string]any
	mustDecodeJSON(t, respHealth.Body, &health)
	if health["status"] != "ok" {
		t.Fatalf("unexpected health payload: %#v", health)
	}

	respArticles := mustGET(t, a.Router, "/api/articles")
	if respArticles.StatusCode != http.StatusOK {
		t.Fatalf("articles status=%d", respArticles.StatusCode)
	}
	var listBefore struct {
		Items []articleDTO `json:"items"`
		Total int          `json:"total"`
	}
	mustDecodeJSON(t, respArticles.Body, &listBefore)
	if listBefore.Total < 1 || len(listBefore.Items) < 1 {
		t.Fatalf("expected seeded data in /api/articles, got total=%d len=%d", listBefore.Total, len(listBefore.Items))
	}

	payload := map[string]any{
		"title":     "Minimal Smoke Article",
		"content":   "Created from minimal smoke test",
		"published": true,
	}
	body, _ := json.Marshal(payload)
	createRes := mustRequest(t, a.Router, http.MethodPost, "/api/articles", bytes.NewReader(body), map[string]string{
		"Content-Type": "application/json",
	})
	if createRes.StatusCode != http.StatusCreated {
		raw := mustReadBody(t, createRes)
		t.Fatalf("create status=%d body=%s", createRes.StatusCode, raw)
	}
	var created map[string]any
	mustDecodeJSON(t, createRes.Body, &created)
	if created["id"] == nil {
		t.Fatalf("create response missing id: %#v", created)
	}
}

func newExampleTestApp(t *testing.T) (*app.App, func()) {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "example_test.db")
	cfg := defaultExampleConfig()
	cfg.Databases["default"] = app.DatabaseConfig{
		URL:         "sqlite://" + dbPath,
		MaxOpen:     10,
		MaxIdle:     5,
		MaxLifetime: 5 * time.Minute,
	}
	cfg.Port = 0
	cfg.LogLevel = "error"
	cfg.LogFormat = "text"
	cfg.AdminBootstrapUsername = "admin"
	cfg.AdminBootstrapEmail = "admin@example.com"
	cfg.AdminBootstrapPassword = "supersecret123"

	a, err := newExampleApp(cfg)
	if err != nil {
		t.Fatalf("newExampleApp failed: %v", err)
	}

	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_ = a.Shutdown(ctx)
	}

	return a, cleanup
}

func mustGET(t *testing.T, handler http.Handler, url string) *http.Response {
	t.Helper()

	return mustRequest(t, handler, http.MethodGet, url, nil, nil)
}

func mustRequest(t *testing.T, handler http.Handler, method, url string, body io.Reader, headers map[string]string) *http.Response {
	t.Helper()

	req := httptest.NewRequest(method, url, body)
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec.Result()
}

func mustRequestWithCookies(
	t *testing.T,
	handler http.Handler,
	method string,
	url string,
	body io.Reader,
	headers map[string]string,
	cookies []*http.Cookie,
) *http.Response {
	t.Helper()

	req := httptest.NewRequest(method, url, body)
	for _, c := range cookies {
		req.AddCookie(c)
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec.Result()
}

func mustAdminLogin(t *testing.T, handler http.Handler, loginPath, username, password string) []*http.Cookie {
	t.Helper()

	form := url.Values{
		"username": {username},
		"password": {password},
		"next":     {"/admin/"},
	}
	res := mustRequest(t, handler, http.MethodPost, loginPath, strings.NewReader(form.Encode()), map[string]string{
		"Content-Type": "application/x-www-form-urlencoded",
	})
	if res.StatusCode != http.StatusSeeOther {
		body := mustReadBody(t, res)
		t.Fatalf("admin login status=%d body=%s", res.StatusCode, body)
	}
	cookies := res.Cookies()
	if len(cookies) == 0 {
		t.Fatal("admin login did not set any session cookie")
	}
	return cookies
}

func mustReadBody(t *testing.T, res *http.Response) string {
	t.Helper()
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("read body failed: %v", err)
	}
	return string(raw)
}

func mustDecodeJSON(t *testing.T, r io.ReadCloser, out any) {
	t.Helper()
	defer r.Close()
	if err := json.NewDecoder(r).Decode(out); err != nil {
		t.Fatalf("decode json failed: %v", err)
	}
}

func containsArticleTitle(items []articleDTO, title string) bool {
	for _, it := range items {
		if it.Title == title {
			return true
		}
	}
	return false
}

func containsModel(models []struct {
	Name string `json:"name"`
}, name string) bool {
	for _, m := range models {
		if m.Name == name {
			return true
		}
	}
	return false
}
