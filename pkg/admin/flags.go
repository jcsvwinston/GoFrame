package admin

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	gferrors "github.com/jcsvwinston/GoFrame/pkg/errors"
)

type featureFlagStore struct {
	mu    sync.RWMutex
	flags map[string]featureFlagState
}

type featureFlagState struct {
	Name      string `json:"name"`
	Enabled   bool   `json:"enabled"`
	UpdatedAt string `json:"updated_at,omitempty"`
	UpdatedBy string `json:"updated_by,omitempty"`
}

func newFeatureFlagStore(initial map[string]bool) *featureFlagStore {
	store := &featureFlagStore{
		flags: make(map[string]featureFlagState),
	}
	now := time.Now().UTC().Format(time.RFC3339)
	for name, enabled := range initial {
		normalized := normalizeFeatureFlagName(name)
		if normalized == "" {
			continue
		}
		store.flags[normalized] = featureFlagState{
			Name:      normalized,
			Enabled:   enabled,
			UpdatedAt: now,
			UpdatedBy: "bootstrap",
		}
	}
	return store
}

func normalizeFeatureFlagName(name string) string {
	normalized := strings.ToLower(strings.TrimSpace(name))
	normalized = strings.ReplaceAll(normalized, " ", "_")
	return normalized
}

func (s *featureFlagStore) list() []featureFlagState {
	if s == nil {
		return []featureFlagState{}
	}
	s.mu.RLock()
	rows := make([]featureFlagState, 0, len(s.flags))
	for _, item := range s.flags {
		rows = append(rows, item)
	}
	s.mu.RUnlock()
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].Name < rows[j].Name
	})
	return rows
}

func (s *featureFlagStore) get(name string) (featureFlagState, bool) {
	if s == nil {
		return featureFlagState{}, false
	}
	key := normalizeFeatureFlagName(name)
	if key == "" {
		return featureFlagState{}, false
	}
	s.mu.RLock()
	row, ok := s.flags[key]
	s.mu.RUnlock()
	return row, ok
}

func (s *featureFlagStore) set(name string, enabled bool, actor string) featureFlagState {
	key := normalizeFeatureFlagName(name)
	row := featureFlagState{
		Name:      key,
		Enabled:   enabled,
		UpdatedAt: time.Now().UTC().Format(time.RFC3339),
		UpdatedBy: strings.TrimSpace(actor),
	}
	if row.UpdatedBy == "" {
		row.UpdatedBy = "admin"
	}
	if s == nil || key == "" {
		return row
	}
	s.mu.Lock()
	s.flags[key] = row
	s.mu.Unlock()
	return row
}

// FeatureFlag returns one in-memory feature flag value.
func (p *Panel) FeatureFlag(name string) (enabled bool, ok bool) {
	if p == nil || p.flags == nil {
		return false, false
	}
	row, exists := p.flags.get(name)
	return row.Enabled, exists
}

// SetFeatureFlag upserts one in-memory feature flag value.
func (p *Panel) SetFeatureFlag(name string, enabled bool) {
	if p == nil || p.flags == nil {
		return
	}
	p.flags.set(name, enabled, "runtime")
}

func (p *Panel) handleListSystemFlags(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "system_pulse") {
		return
	}
	rows := []featureFlagState{}
	if p != nil && p.flags != nil {
		rows = p.flags.list()
	}
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled": true,
		"count":   len(rows),
		"flags":   rows,
	})
}

func (p *Panel) handleSetSystemFlag(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "feature_flags_write") {
		return
	}
	if p == nil || p.flags == nil {
		writeErr(w, gferrors.BadRequest("feature flags store is not available"))
		return
	}

	name := normalizeFeatureFlagName(r.PathValue("name"))
	if name == "" {
		writeErr(w, gferrors.BadRequest("feature flag name is required"))
		return
	}

	var payload struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeErr(w, gferrors.BadRequest("invalid JSON"))
		return
	}

	actor := "admin"
	if p.config.Auth != nil {
		if user, err := p.authenticatedUser(r); err == nil && user != nil {
			if trimmed := strings.TrimSpace(user.ID); trimmed != "" {
				actor = trimmed
			}
		}
	}
	row := p.flags.set(name, payload.Enabled, actor)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"updated": true,
		"flag":    row,
	})
}
