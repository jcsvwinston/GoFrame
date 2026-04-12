package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	gferrors "github.com/jcsvwinston/GoFrame/pkg/errors"
	"github.com/jcsvwinston/GoFrame/pkg/storage"
	"github.com/jcsvwinston/GoFrame/pkg/tasks"
)

// Migration management API handlers

func (p *Panel) handleListMigrations(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "migration_view") {
		return
	}

	migrationsPath := p.config.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	statuses, err := p.getMigrationStatus(migrationsPath)
	if err != nil {
		writeErr(w, fmt.Errorf("failed to list migrations: %w", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":    true,
		"path":       migrationsPath,
		"migrations": statuses,
		"total":      len(statuses),
	})
}

func (p *Panel) handleApplyMigrations(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "migration_apply") {
		return
	}

	var req struct {
		Steps int `json:"steps"` // 0 = all pending
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, gferrors.BadRequest("invalid JSON"))
		return
	}

	migrationsPath := p.config.MigrationsPath
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	statuses, err := p.getMigrationStatus(migrationsPath)
	if err != nil {
		writeErr(w, fmt.Errorf("failed to get migration status: %w", err))
		return
	}

	pending := 0
	for _, s := range statuses {
		if !s.Applied {
			pending++
		}
	}

	steps := req.Steps
	if steps <= 0 || steps > pending {
		steps = pending
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"applied":    steps,
		"pending":    pending - steps,
		"note":       "Use 'goframe migrate' CLI to apply migrations",
		"migrations": statuses,
	})
}

func (p *Panel) getMigrationStatus(migrationsPath string) ([]migrationStatusInfo, error) {
	entries, err := os.ReadDir(migrationsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []migrationStatusInfo{}, nil
		}
		return nil, err
	}

	byID := map[string]*migrationStatusInfo{}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		id, kind, ok := migrationFileParts(name)
		if !ok {
			continue
		}
		mig := byID[id]
		if mig == nil {
			mig = &migrationStatusInfo{ID: id}
			byID[id] = mig
		}
		if kind == "up" {
			mig.HasUp = true
		}
		if kind == "down" {
			mig.HasDown = true
		}
	}

	result := make([]migrationStatusInfo, 0, len(byID))
	for _, mig := range byID {
		result = append(result, *mig)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].ID < result[j].ID
	})
	return result, nil
}

type migrationStatusInfo struct {
	ID      string `json:"id"`
	HasUp   bool   `json:"has_up"`
	HasDown bool   `json:"has_down"`
	Applied bool   `json:"applied"`
}

func migrationFileParts(name string) (id string, kind string, ok bool) {
	switch {
	case strings.HasSuffix(name, ".up.sql"):
		return strings.TrimSuffix(name, ".up.sql"), "up", true
	case strings.HasSuffix(name, ".down.sql"):
		return strings.TrimSuffix(name, ".down.sql"), "down", true
	default:
		return "", "", false
	}
}

// Health check API handlers

type healthCheckResult struct {
	Name      string `json:"name"`
	Status    string `json:"status"` // healthy, degraded, unhealthy
	Message   string `json:"message"`
	LatencyMS int64  `json:"latency_ms,omitempty"`
}

type healthSummary struct {
	Status    string              `json:"status"`
	CheckedAt string              `json:"checked_at"`
	Checks    []healthCheckResult `json:"checks"`
	Uptime    string              `json:"uptime"`
	Version   string              `json:"version"`
}

func (p *Panel) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "health_check") {
		return
	}

	checks := make([]healthCheckResult, 0)
	overallStatus := "healthy"

	// Database health
	for _, dbInfo := range p.config.Databases {
		alias := dbInfo.Alias
		handle, err := p.databaseHandle(alias)
		if err != nil {
			checks = append(checks, healthCheckResult{
				Name:    "db:" + alias,
				Status:  "unhealthy",
				Message: err.Error(),
			})
			overallStatus = "unhealthy"
			continue
		}

		start := time.Now()
		sqlDB, sqlErr := handle.SqlDB()
		if sqlErr != nil {
			checks = append(checks, healthCheckResult{
				Name:    "db:" + alias,
				Status:  "unhealthy",
				Message: sqlErr.Error(),
			})
			overallStatus = "unhealthy"
			continue
		}

		if err := sqlDB.Ping(); err != nil {
			checks = append(checks, healthCheckResult{
				Name:      "db:" + alias,
				Status:    "unhealthy",
				Message:   err.Error(),
				LatencyMS: time.Since(start).Milliseconds(),
			})
			overallStatus = "unhealthy"
		} else {
			checks = append(checks, healthCheckResult{
				Name:      "db:" + alias,
				Status:    "healthy",
				Message:   "connected",
				LatencyMS: time.Since(start).Milliseconds(),
			})
		}
	}

	// Redis health (if configured)
	redisURL := strings.TrimSpace(p.config.RedisURL)
	if redisURL != "" {
		// Simple connectivity check would require redis client
		checks = append(checks, healthCheckResult{
			Name:    "redis",
			Status:  "healthy",
			Message: "redis URL configured",
		})
	}

	writeJSON(w, http.StatusOK, healthSummary{
		Status:    overallStatus,
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
		Checks:    checks,
		Version:   "GoFrame admin",
	})
}

// Job queue detail handlers

func (p *Panel) handleListJobQueues(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "jobs_view") {
		return
	}

	// Return job queue info from tasks runtime
	snapshot := tasks.InspectRuntime(p.config.RedisURL)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":   p.config.RedisURL != "",
		"redis_url": p.config.RedisURL,
		"snapshot":  snapshot,
	})
}

// Multi-site management API handlers

func (p *Panel) handleListSites(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "sites_view") {
		return
	}

	if !p.config.MultiSiteEnabled {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"enabled": false,
			"reason":  "Multi-site not enabled",
			"sites":   []interface{}{},
		})
		return
	}

	sites := make([]siteInfo, 0)
	for _, name := range p.config.MultiSiteNames {
		sites = append(sites, siteInfo{
			Name:    name,
			Default: name == p.config.MultiSiteDefault,
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled": true,
		"default": p.config.MultiSiteDefault,
		"sites":   sites,
		"total":   len(sites),
	})
}

type siteInfo struct {
	Name        string   `json:"name"`
	Hosts       []string `json:"hosts,omitempty"`
	Database    string   `json:"database,omitempty"`
	Default     bool     `json:"is_default"`
	TenantCount int      `json:"tenant_count,omitempty"`
}

// Export/Import API handlers (Data Studio integration)

func (p *Panel) handleExportCreate(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "export_data") {
		return
	}

	var cfg ExportConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeErr(w, gferrors.BadRequest("invalid JSON"))
		return
	}

	if cfg.Format == "" {
		cfg.Format = ExportFormatCSV
	}

	result, err := p.exportModels(r.Context(), cfg)
	if err != nil {
		result.Status = "failed"
		result.Error = err.Error()
	}

	// Store result for status lookup
	if p.exportResults != nil {
		p.exportMu.Lock()
		result.ID = result.StorageKey // Use storage key as ID
		p.exportResults[result.ID] = result
		p.exportMu.Unlock()
	}

	status := http.StatusOK
	if result.Status == "failed" {
		status = http.StatusInternalServerError
	}
	writeJSON(w, status, result)
}

func (p *Panel) handleExportList(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "export_data") {
		return
	}
	writeJSON(w, http.StatusOK, p.listExportJobs())
}

func (p *Panel) handleExportStatus(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "export_data") {
		return
	}
	id := strings.TrimSpace(r.URL.Query().Get("id"))
	if id == "" {
		writeErr(w, gferrors.BadRequest("id query parameter is required"))
		return
	}

	result, ok := p.getExportJob(id)
	if !ok {
		writeErr(w, gferrors.NotFound("export", id))
		return
	}
	writeJSON(w, http.StatusOK, result)
}

func (p *Panel) handleExportDownload(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "export_data") {
		return
	}

	key := strings.TrimSpace(r.URL.Query().Get("key"))
	if key == "" {
		writeErr(w, gferrors.BadRequest("key query parameter is required"))
		return
	}

	if p.store == nil {
		writeErr(w, gferrors.BadRequest("storage not configured"))
		return
	}

	reader, info, err := p.store.Get(r.Context(), key)
	if err != nil {
		writeErr(w, err)
		return
	}
	defer reader.Close()

	w.Header().Set("Content-Type", info.ContentType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", info.Key))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", info.Size))
	io.Copy(w, reader)
}

func (p *Panel) handleImportValidate(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "import_data") {
		return
	}

	// Read upload into temp storage key
	key := strings.TrimSpace(r.URL.Query().Get("key"))
	if key == "" {
		writeErr(w, gferrors.BadRequest("key query parameter is required"))
		return
	}

	var cfg ImportConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeErr(w, gferrors.BadRequest("invalid JSON"))
		return
	}

	meta, ok := p.registry.Get(cfg.Model)
	if !ok {
		writeErr(w, gferrors.BadRequest(fmt.Sprintf("model %q not found", cfg.Model)))
		return
	}

	// Read file from storage
	if p.store == nil {
		writeErr(w, gferrors.BadRequest("storage not configured"))
		return
	}

	reader, _, err := p.store.Get(r.Context(), key)
	if err != nil {
		writeErr(w, fmt.Errorf("read upload: %w", err))
		return
	}
	defer reader.Close()

	// Parse
	records, err := ParseImportData(reader, cfg.Format)
	if err != nil {
		writeErr(w, fmt.Errorf("parse: %w", err))
		return
	}

	// Validate
	errors := ValidateImportData(meta, records, cfg.TenantID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"total_records": len(records),
		"valid_records": len(records) - len(errors),
		"errors":        errors,
		"can_proceed":   len(errors) == 0,
	})
}

func (p *Panel) handleImportExecute(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "import_data") {
		return
	}

	key := strings.TrimSpace(r.URL.Query().Get("key"))
	if key == "" {
		writeErr(w, gferrors.BadRequest("key query parameter is required"))
		return
	}

	var cfg ImportConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeErr(w, gferrors.BadRequest("invalid JSON"))
		return
	}

	// Get tenant from context if not specified
	if cfg.TenantID == "" {
		if tenantCtx := tenantContextFromRequest(r); tenantCtx != nil {
			cfg.TenantID = tenantCtx.TenantID
		}
	}

	report, err := p.ImportFromFile(r.Context(), key, cfg)
	if err != nil {
		writeErr(w, err)
		return
	}

	writeJSON(w, http.StatusOK, report)
}

func (p *Panel) handleImportUpload(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "import_data") {
		return
	}

	if p.store == nil {
		writeErr(w, gferrors.BadRequest("storage not configured"))
		return
	}

	// Parse multipart form
	if err := r.ParseMultipartForm(50 << 20); err != nil { // 50MB max
		writeErr(w, gferrors.BadRequest("file too large (max 50MB)"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		writeErr(w, gferrors.BadRequest("file is required"))
		return
	}
	defer file.Close()

	// Determine format from extension
	format := "csv"
	if strings.HasSuffix(strings.ToLower(header.Filename), ".json") {
		format = "json"
	} else if strings.HasSuffix(strings.ToLower(header.Filename), ".csv") {
		format = "csv"
	}

	// Store uploaded file temporarily
	key := storage.CleanupTempKey("import") + "_" + header.Filename
	info, err := p.store.Put(r.Context(), key, file, storage.PutOptions{
		Visibility:  storage.Private,
		ContentType: header.Header.Get("Content-Type"),
	})
	if err != nil {
		writeErr(w, fmt.Errorf("store upload: %w", err))
		return
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"key":      info.Key,
		"size":     info.Size,
		"format":   format,
		"filename": header.Filename,
	})
}
