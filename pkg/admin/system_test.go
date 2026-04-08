package admin

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jcsvwinston/GoFrame/pkg/db"
)

func TestBuildSystemEnvironmentRowsMasksSensitiveValues(t *testing.T) {
	rows := buildSystemEnvironmentRows([]string{
		"APP_MODE=dev",
		"API_TOKEN=secret-value",
		"SIGNING_KEY=abc",
	})
	if len(rows) != 3 {
		t.Fatalf("expected three env rows, got %d", len(rows))
	}

	index := map[string]systemEnvVar{}
	for _, row := range rows {
		index[row.Name] = row
	}

	if got := index["APP_MODE"].Value; got != "dev" {
		t.Fatalf("expected APP_MODE=dev, got %q", got)
	}
	if !index["API_TOKEN"].Masked || index["API_TOKEN"].Value != "***" {
		t.Fatalf("expected API_TOKEN masked, got %#v", index["API_TOKEN"])
	}
	if !index["SIGNING_KEY"].Masked || index["SIGNING_KEY"].Value != "***" {
		t.Fatalf("expected SIGNING_KEY masked, got %#v", index["SIGNING_KEY"])
	}
}

func TestGatherGoroutineStateCounts(t *testing.T) {
	rows := gatherGoroutineStateCounts()
	if len(rows) == 0 {
		t.Fatalf("expected at least one state row")
	}
	total := 0
	for _, row := range rows {
		if row.Count <= 0 {
			t.Fatalf("expected positive goroutine count row, got %#v", row)
		}
		total += row.Count
	}
	if total <= 0 {
		t.Fatalf("expected total goroutine count > 0")
	}
}

func TestPanelSystemSnapshotEndpoint(t *testing.T) {
	panel, cleanup := setupPanelForTest(t, db.EngineSQL)
	defer cleanup()

	panel.config.Databases = []DatabaseRuntimeInfo{
		{Alias: "default", Engine: "sql", Dialect: "sqlite", IsDefault: true},
	}
	panel.config.DatabaseHandles = map[string]*db.DB{"default": panel.db}
	panel.bootEnv = buildSystemEnvironmentRows([]string{
		"APP_ENV=test",
		"SERVICE_TOKEN=token-value",
		"DB_PASSWORD=pass-value",
	})

	srv := httptest.NewServer(panel.Handler())
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/system/snapshot?env_limit=50")
	if err != nil {
		t.Fatalf("snapshot request failed: %v", err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.StatusCode)
	}

	var payload systemSnapshotResponse
	if err := json.NewDecoder(res.Body).Decode(&payload); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if !payload.Enabled {
		t.Fatalf("expected enabled response")
	}
	if payload.Goroutines.Count <= 0 {
		t.Fatalf("expected goroutines count > 0")
	}
	if len(payload.Databases) == 0 {
		t.Fatalf("expected at least one db pool row")
	}
	if len(payload.Environment) < 2 {
		t.Fatalf("expected environment rows")
	}

	env := map[string]systemEnvVar{}
	for _, row := range payload.Environment {
		env[row.Name] = row
	}
	if env["SERVICE_TOKEN"].Value != "***" || !env["SERVICE_TOKEN"].Masked {
		t.Fatalf("expected SERVICE_TOKEN masked, got %#v", env["SERVICE_TOKEN"])
	}
	if env["DB_PASSWORD"].Value != "***" || !env["DB_PASSWORD"].Masked {
		t.Fatalf("expected DB_PASSWORD masked, got %#v", env["DB_PASSWORD"])
	}
	if env["APP_ENV"].Masked {
		t.Fatalf("expected APP_ENV not masked, got %#v", env["APP_ENV"])
	}
}
