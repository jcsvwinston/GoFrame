package app

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/jcsvwinston/GoFrame/pkg/auth"
)

func testAppConfig() *Config {
	return &Config{
		Host:            "127.0.0.1",
		Port:            0,
		ReadTimeout:     2 * time.Second,
		WriteTimeout:    2 * time.Second,
		IdleTimeout:     5 * time.Second,
		DatabaseDefault: "default",
		Databases: map[string]DatabaseConfig{
			"default": {
				URL:         "sqlite://:memory:",
				MaxOpen:     1,
				MaxIdle:     1,
				MaxLifetime: time.Minute,
			},
		},
		LogLevel:    "error",
		LogFormat:   "text",
		AdminPrefix: "/admin",
		AdminTitle:  "Test Admin",
	}
}

func TestAppNew_NilConfig(t *testing.T) {
	_, err := New(nil)
	if err == nil {
		t.Fatal("expected error for nil config")
	}
	if !errors.Is(err, ErrNilConfig) {
		t.Fatalf("expected ErrNilConfig, got: %v", err)
	}
}

func TestAppNew_InitializesCoreComponents(t *testing.T) {
	a, err := New(testAppConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer a.Shutdown(context.Background())

	if a.Config == nil || a.Logger == nil || a.Router == nil || a.DB == nil || a.Models == nil {
		t.Fatal("expected app core components to be initialized")
	}
	if a.Mailer == nil {
		t.Fatal("expected mailer to be initialized")
	}
	if a.Session == nil {
		t.Fatal("expected session manager to be initialized")
	}
	if a.Admin == nil {
		t.Fatal("expected admin panel to be initialized")
	}
	if err := a.DB.Health(context.Background()); err != nil {
		t.Fatalf("expected DB health to pass, got: %v", err)
	}
}

func TestAppNew_SQLRuntime_InitializesAdmin(t *testing.T) {
	cfg := testAppConfig()

	a, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer a.Shutdown(context.Background())

	if a.DB == nil {
		t.Fatal("expected DB to be initialized")
	}
	if a.DB.Engine() != "sql" {
		t.Fatalf("expected db engine sql, got %s", a.DB.Engine())
	}
	if a.Admin == nil {
		t.Fatal("expected admin to be initialized when sql engine is selected")
	}
}

func TestAppRegisterModel(t *testing.T) {
	type User struct {
		ID    uint
		Email string
	}

	a, err := New(testAppConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer a.Shutdown(context.Background())

	if err := a.RegisterModel(&User{}); err != nil {
		t.Fatalf("RegisterModel failed: %v", err)
	}
	if a.Models.Count() != 1 {
		t.Fatalf("expected 1 model registered, got %d", a.Models.Count())
	}
}

func TestAppShutdown_ReverseHookOrderAndErrorAggregation(t *testing.T) {
	a := &App{Config: testAppConfig()}
	var order []int

	a.OnShutdown(func(context.Context) error {
		order = append(order, 1)
		return nil
	})
	a.OnShutdown(func(context.Context) error {
		order = append(order, 2)
		return errors.New("hook two failed")
	})
	a.OnShutdown(func(context.Context) error {
		order = append(order, 3)
		return nil
	})

	err := a.Shutdown(context.Background())
	if err == nil {
		t.Fatal("expected aggregated shutdown error")
	}
	if !strings.Contains(err.Error(), "hook two failed") {
		t.Fatalf("expected hook error in shutdown error, got: %v", err)
	}

	got := fmt.Sprint(order)
	if got != "[3 2 1]" {
		t.Fatalf("expected reverse hook order [3 2 1], got %s", got)
	}
}

func TestAppMethods_NilReceiver(t *testing.T) {
	var a *App

	if err := a.Run(context.Background()); !errors.Is(err, ErrNilApp) {
		t.Fatalf("Run: expected ErrNilApp, got %v", err)
	}
	if err := a.Shutdown(context.Background()); !errors.Is(err, ErrNilApp) {
		t.Fatalf("Shutdown: expected ErrNilApp, got %v", err)
	}
	if err := a.MountAdmin(); !errors.Is(err, ErrNilApp) {
		t.Fatalf("MountAdmin: expected ErrNilApp, got %v", err)
	}
	if err := a.RegisterModel(&struct{ ID uint }{}); !errors.Is(err, ErrNilApp) {
		t.Fatalf("RegisterModel: expected ErrNilApp, got %v", err)
	}
}

func TestAppRun_NotInitialized(t *testing.T) {
	a := &App{}
	err := a.Run(context.Background())
	if !errors.Is(err, ErrNotInitialized) {
		t.Fatalf("expected ErrNotInitialized, got %v", err)
	}
}

func TestAppRun_ContextCancel(t *testing.T) {
	a, err := New(testAppConfig())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(120 * time.Millisecond)
		cancel()
	}()

	if err := a.Run(ctx); err != nil {
		t.Fatalf("Run should exit cleanly on context cancel: %v", err)
	}
}

func TestAppRun_InvalidAddress(t *testing.T) {
	cfg := testAppConfig()
	cfg.Port = -1 // produces an invalid host:port address

	a, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating app: %v", err)
	}
	defer a.Shutdown(context.Background())

	err = a.Run(context.Background())
	if err == nil {
		t.Fatal("expected run error for invalid server address")
	}
}

func TestAppNew_InvalidMailDriver(t *testing.T) {
	cfg := testAppConfig()
	cfg.MailDriver = "unknown-provider"

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected error for unknown mail driver")
	}
	if !strings.Contains(err.Error(), "unknown mail driver") {
		t.Fatalf("expected unknown mail driver error, got %v", err)
	}
}

func TestAppNew_SessionStoreRedisRequiresURL(t *testing.T) {
	cfg := testAppConfig()
	cfg.SessionStore = "redis"
	cfg.RedisURL = ""
	cfg.SessionRedisURL = ""

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected redis session store config error")
	}
	if !strings.Contains(err.Error(), "session_store=redis requires") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppNew_UnsupportedSessionStore(t *testing.T) {
	cfg := testAppConfig()
	cfg.SessionStore = "unknown-store"

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected unsupported session store error")
	}
	if !strings.Contains(err.Error(), "unsupported session_store") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppNew_SQLSessionStorePersistsAcrossRequests(t *testing.T) {
	cfg := testAppConfig()
	cfg.SessionStore = "sql"
	cfg.SessionTable = "goframe_sessions"

	a, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer a.Shutdown(context.Background())

	a.Router.Get("/set", func(w http.ResponseWriter, r *http.Request) {
		a.Session.Put(r.Context(), "name", "alice")
		w.WriteHeader(http.StatusNoContent)
	})
	a.Router.Get("/get", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(a.Session.GetString(r.Context(), "name")))
	})

	setReq := httptest.NewRequest(http.MethodGet, "/set", nil)
	setRec := httptest.NewRecorder()
	a.Router.ServeHTTP(setRec, setReq)
	if setRec.Code != http.StatusNoContent {
		t.Fatalf("expected 204 from /set, got %d", setRec.Code)
	}

	var sessionCookie *http.Cookie
	for _, c := range setRec.Result().Cookies() {
		if c.Name == cfg.SessionCookieName || (cfg.SessionCookieName == "" && c.Name == "session") {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie to be set")
	}

	getReq := httptest.NewRequest(http.MethodGet, "/get", nil)
	getReq.AddCookie(sessionCookie)
	getRec := httptest.NewRecorder()
	a.Router.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusOK {
		t.Fatalf("expected 200 from /get, got %d", getRec.Code)
	}
	if strings.TrimSpace(getRec.Body.String()) != "alice" {
		t.Fatalf("expected persisted session value alice, got %q", getRec.Body.String())
	}

	sqlDB, err := a.DB.SqlDB()
	if err != nil {
		t.Fatalf("sql db handle: %v", err)
	}

	var count int
	if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM "goframe_sessions"`).Scan(&count); err != nil {
		t.Fatalf("count sessions failed: %v", err)
	}
	if count < 1 {
		t.Fatalf("expected at least 1 persisted session row, got %d", count)
	}
}

func TestAppNew_AdminBootstrapMode_AllowsAccessWithoutUsers(t *testing.T) {
	a, err := New(testAppConfig())
	if err != nil {
		t.Fatalf("unexpected error creating app: %v", err)
	}
	defer a.Shutdown(context.Background())

	req := httptest.NewRequest(http.MethodGet, "/admin/api/models", nil)
	rec := httptest.NewRecorder()
	a.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected bootstrap admin access without users to return 200, got %d", rec.Code)
	}
}

func TestAppNew_AdminRequiresLoginAfterCreateUser(t *testing.T) {
	a, err := New(testAppConfig())
	if err != nil {
		t.Fatalf("unexpected error creating app: %v", err)
	}
	defer a.Shutdown(context.Background())

	sqlDB, err := a.DB.SqlDB()
	if err != nil {
		t.Fatalf("sql db handle: %v", err)
	}
	if err := seedAdminUser(sqlDB, "admin", "admin@example.com", "supersecret123"); err != nil {
		t.Fatalf("seed admin user failed: %v", err)
	}

	protectedReq := httptest.NewRequest(http.MethodGet, "/admin/api/models", nil)
	protectedRec := httptest.NewRecorder()
	a.Router.ServeHTTP(protectedRec, protectedReq)
	if protectedRec.Code != http.StatusFound {
		t.Fatalf("expected unauthenticated admin access to redirect, got %d", protectedRec.Code)
	}
	if loc := protectedRec.Header().Get("Location"); loc != "/admin/login" {
		t.Fatalf("expected redirect to /admin/login, got %q", loc)
	}

	form := url.Values{
		"username": {"admin"},
		"password": {"supersecret123"},
		"next":     {"/admin/"},
	}
	loginReq := httptest.NewRequest(http.MethodPost, "/admin/login", strings.NewReader(form.Encode()))
	loginReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRec := httptest.NewRecorder()
	a.Router.ServeHTTP(loginRec, loginReq)
	if loginRec.Code != http.StatusSeeOther {
		t.Fatalf("expected login to return 303, got %d body=%s", loginRec.Code, loginRec.Body.String())
	}
	if loc := loginRec.Header().Get("Location"); loc != "/admin/" {
		t.Fatalf("expected login redirect to /admin/, got %q", loc)
	}

	cookies := loginRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected login to set session cookie")
	}

	authReq := httptest.NewRequest(http.MethodGet, "/admin/api/models", nil)
	for _, c := range cookies {
		authReq.AddCookie(c)
	}
	authRec := httptest.NewRecorder()
	a.Router.ServeHTTP(authRec, authReq)
	if authRec.Code != http.StatusOK {
		t.Fatalf("expected authenticated admin access to return 200, got %d body=%s", authRec.Code, authRec.Body.String())
	}
}

func TestAppNew_OpensMultipleDatabaseAliases(t *testing.T) {
	cfg := testAppConfig()
	cfg.DatabaseDefault = "primary"
	cfg.Databases = map[string]DatabaseConfig{
		"primary": {
			URL: "sqlite://:memory:",
		},
		"analytics": {
			URL: "sqlite://:memory:",
		},
	}

	a, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer a.Shutdown(context.Background())

	if got := a.DefaultDatabaseAlias(); got != "primary" {
		t.Fatalf("expected default alias primary, got %s", got)
	}
	if len(a.DBs) != 2 {
		t.Fatalf("expected 2 database aliases, got %d", len(a.DBs))
	}
	analytics, err := a.Database("analytics")
	if err != nil {
		t.Fatalf("resolve analytics db: %v", err)
	}
	if analytics == nil {
		t.Fatal("expected analytics db handle")
	}
	if _, err := a.Database("missing"); !errors.Is(err, ErrDatabaseAliasNotFound) {
		t.Fatalf("expected ErrDatabaseAliasNotFound, got %v", err)
	}
}

func seedAdminUser(sqlDB *sql.DB, username, email, password string) error {
	if sqlDB == nil {
		return fmt.Errorf("nil sql db")
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if _, err := sqlDB.Exec(`
CREATE TABLE IF NOT EXISTS goframe_admin_users (
	id VARCHAR(64) PRIMARY KEY,
	username VARCHAR(191) NOT NULL UNIQUE,
	email VARCHAR(191) NOT NULL UNIQUE,
	password_hash TEXT NOT NULL,
	is_superuser INTEGER NOT NULL DEFAULT 0,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
)`); err != nil {
		return fmt.Errorf("create admin users table: %w", err)
	}

	_, err = sqlDB.Exec(
		`INSERT INTO goframe_admin_users (id, username, email, password_hash, is_superuser, created_at, updated_at) VALUES (?, ?, ?, ?, 1, datetime('now'), datetime('now'))`,
		"u_test_admin",
		username,
		email,
		hash,
	)
	if err != nil {
		return fmt.Errorf("insert admin user: %w", err)
	}
	return nil
}

func TestAppDatabaseForRequest_UsesTenantDatabaseAlias(t *testing.T) {
	cfg := testAppConfig()
	cfg.DatabaseDefault = "default"
	cfg.Databases = map[string]DatabaseConfig{
		"default":      {URL: "sqlite://:memory:"},
		"tenant_acme":  {URL: "sqlite://:memory:"},
		"tenant_omega": {URL: "sqlite://:memory:"},
	}
	cfg.MultiSite = MultiSiteConfig{
		Enabled:     true,
		DefaultSite: "main",
		Sites: map[string]SiteConfig{
			"main": {
				Hosts:                       []string{"*.site.com"},
				Database:                    "default",
				TenantDatabaseAliasTemplate: "tenant_%s",
			},
		},
	}
	cfg.MultiTenant = MultiTenantConfig{
		Enabled:  true,
		Resolver: "subdomain",
		Tenants:  map[string]TenantConfig{},
	}

	a, err := New(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer a.Shutdown(context.Background())

	a.Router.Get("/scope", func(w http.ResponseWriter, r *http.Request) {
		scope, ok := RequestScopeFromContext(r.Context())
		if !ok {
			http.Error(w, "scope missing", http.StatusInternalServerError)
			return
		}
		if _, err := a.DatabaseForRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		_, _ = w.Write([]byte(scope.Site + "|" + scope.Tenant + "|" + scope.DatabaseAlias))
	})

	req := httptest.NewRequest(http.MethodGet, "/scope", nil)
	req.Host = "acme.site.com"
	rec := httptest.NewRecorder()
	a.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rec.Code, rec.Body.String())
	}
	if strings.TrimSpace(rec.Body.String()) != "main|acme|tenant_acme" {
		t.Fatalf("unexpected scope payload: %s", rec.Body.String())
	}

	missingReq := httptest.NewRequest(http.MethodGet, "/scope", nil)
	missingReq.Host = "unknown.site.com"
	missingRec := httptest.NewRecorder()
	a.Router.ServeHTTP(missingRec, missingReq)
	if missingRec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for missing tenant alias, got %d", missingRec.Code)
	}
	if !strings.Contains(missingRec.Body.String(), "database alias not found") {
		t.Fatalf("expected missing alias error, got %s", missingRec.Body.String())
	}
}

func TestAppNew_MultiTenantIsolationRejectsSharedDatabaseAlias(t *testing.T) {
	cfg := testAppConfig()
	cfg.DatabaseDefault = "default"
	cfg.Databases = map[string]DatabaseConfig{
		"default": {URL: "sqlite://:memory:"},
		"shared":  {URL: "sqlite://:memory:"},
	}
	cfg.MultiSite = MultiSiteConfig{
		Enabled:     true,
		DefaultSite: "main",
		Sites: map[string]SiteConfig{
			"main": {Database: "default"},
		},
	}
	cfg.MultiTenant = MultiTenantConfig{
		Enabled:           true,
		RequireIsolatedDB: true,
		Tenants: map[string]TenantConfig{
			"tenant_a": {Site: "main", Database: "shared"},
			"tenant_b": {Site: "main", Database: "shared"},
		},
	}

	_, err := New(cfg)
	if err == nil {
		t.Fatal("expected multitenant isolation validation error")
	}
	if !strings.Contains(err.Error(), "share database alias") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAppNew_TenantIsolationRequiresTenantAwareTemplate(t *testing.T) {
	cfg := testAppConfig()
	cfg.DatabaseDefault = "default"
	cfg.Databases = map[string]DatabaseConfig{
		"default": {URL: "sqlite://:memory:"},
	}
	cfg.MultiSite = MultiSiteConfig{
		Enabled:     true,
		DefaultSite: "main",
		Sites: map[string]SiteConfig{
			"main": {
				Hosts:    []string{"*.site.com"},
				Database: "default",
			},
		},
	}
	cfg.MultiTenant = MultiTenantConfig{
		Enabled:               true,
		Resolver:              "subdomain",
		RequireIsolatedDB:     true,
		DatabaseAliasTemplate: "tenant_shared",
	}

	a, err := New(cfg)
	if err == nil {
		_ = a.Shutdown(context.Background())
		t.Fatal("expected New to fail when no tenant-isolated template is provided")
	}
	if !strings.Contains(err.Error(), "database_alias_template") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDetectDatabaseDialect(t *testing.T) {
	cases := []struct {
		raw  string
		want string
	}{
		{raw: "sqlite://app.db", want: "sqlite"},
		{raw: "postgres://localhost/db", want: "postgres"},
		{raw: "postgresql://localhost/db", want: "postgres"},
		{raw: "mysql://localhost/db", want: "mysql"},
		{raw: "sqlserver://localhost/db", want: "sqlserver"},
		{raw: "mssql://localhost/db", want: "sqlserver"},
		{raw: "oracle://localhost/db", want: "oracle"},
		{raw: "custom://localhost/db", want: "unknown"},
	}
	for _, tc := range cases {
		if got := detectDatabaseDialect(tc.raw); got != tc.want {
			t.Fatalf("detectDatabaseDialect(%q)=%q want=%q", tc.raw, got, tc.want)
		}
	}
}
