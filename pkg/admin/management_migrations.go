package admin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	gfdb "github.com/jcsvwinston/GoFrame/pkg/db"
	gferrors "github.com/jcsvwinston/GoFrame/pkg/errors"
)

func (p *Panel) handleListMigrations(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "migration_view") {
		return
	}

	migrationsPath := p.migrationsPath()
	statuses, err := p.getMigrationStatus(migrationsPath)
	if err != nil {
		writeErr(w, fmt.Errorf("failed to list migrations: %w", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":    true,
		"path":       migrationsPath,
		"mode":       p.migrationMode(),
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

	migrator := p.migrationRuntime()
	if migrator == nil {
		writeErr(w, gferrors.BadRequest("database runtime is not configured for migrations"))
		return
	}

	migrationsPath := p.migrationsPath()
	before, err := p.getMigrationStatus(migrationsPath)
	if err != nil {
		writeErr(w, fmt.Errorf("failed to get migration status: %w", err))
		return
	}

	pendingBefore := countPendingMigrations(before)
	requestedSteps := req.Steps
	steps := requestedSteps
	if steps <= 0 || steps > pendingBefore {
		steps = pendingBefore
	}

	if steps > 0 {
		if requestedSteps <= 0 {
			err = migrator.Up()
		} else {
			err = migrator.Steps(steps)
		}
		if err != nil {
			writeErr(w, fmt.Errorf("failed to apply migrations: %w", err))
			return
		}
	}

	after, err := p.getMigrationStatus(migrationsPath)
	if err != nil {
		writeErr(w, fmt.Errorf("failed to refresh migration status: %w", err))
		return
	}

	appliedIDs := appliedMigrationIDs(before, after)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"applied":         len(appliedIDs),
		"applied_ids":     appliedIDs,
		"pending":         countPendingMigrations(after),
		"requested_steps": requestedSteps,
		"executed_steps":  steps,
		"mode":            "runtime",
		"migrations":      after,
	})
}

func (p *Panel) migrationsPath() string {
	if p == nil || strings.TrimSpace(p.config.MigrationsPath) == "" {
		return "migrations"
	}
	return strings.TrimSpace(p.config.MigrationsPath)
}

func (p *Panel) migrationMode() string {
	if p != nil && p.db != nil {
		return "runtime"
	}
	return "inspect-only"
}

func (p *Panel) migrationRuntime() *gfdb.Migrator {
	if p == nil || p.db == nil {
		return nil
	}
	return gfdb.NewMigrator(p.db, p.migrationsPath(), p.logger)
}

func (p *Panel) getMigrationStatus(migrationsPath string) ([]migrationStatusInfo, error) {
	if migrator := p.migrationRuntime(); migrator != nil {
		statuses, err := migrator.Status()
		if err != nil {
			return nil, err
		}
		return toMigrationStatusInfo(statuses), nil
	}
	return inspectMigrationFiles(migrationsPath)
}

type migrationStatusInfo struct {
	ID        string `json:"id"`
	HasUp     bool   `json:"has_up"`
	HasDown   bool   `json:"has_down"`
	Applied   bool   `json:"applied"`
	AppliedAt string `json:"applied_at,omitempty"`
}

func inspectMigrationFiles(migrationsPath string) ([]migrationStatusInfo, error) {
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

func toMigrationStatusInfo(statuses []gfdb.MigrationStatus) []migrationStatusInfo {
	rows := make([]migrationStatusInfo, 0, len(statuses))
	for _, status := range statuses {
		row := migrationStatusInfo{
			ID:      status.ID,
			HasUp:   status.HasUp,
			HasDown: status.HasDown,
			Applied: status.Applied,
		}
		if status.AppliedAt != nil {
			row.AppliedAt = status.AppliedAt.UTC().Format(time.RFC3339)
		}
		rows = append(rows, row)
	}
	return rows
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

func countPendingMigrations(statuses []migrationStatusInfo) int {
	total := 0
	for _, status := range statuses {
		if !status.Applied {
			total++
		}
	}
	return total
}

func appliedMigrationIDs(before, after []migrationStatusInfo) []string {
	beforeApplied := make(map[string]bool, len(before))
	for _, row := range before {
		beforeApplied[row.ID] = row.Applied
	}

	ids := make([]string, 0)
	for _, row := range after {
		if row.Applied && !beforeApplied[row.ID] {
			ids = append(ids, row.ID)
		}
	}
	sort.Strings(ids)
	return ids
}
