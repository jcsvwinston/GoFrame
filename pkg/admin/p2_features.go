package admin

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	gferrors "github.com/jcsvwinston/GoFrame/pkg/errors"
)

// Deployment detection API handlers

type deploymentInfo struct {
	Runtime      string            `json:"runtime"` // standalone, docker, kubernetes
	Host         string            `json:"host"`
	Pod          string            `json:"pod,omitempty"`
	Instance     string            `json:"instance,omitempty"`
	NodeID       string            `json:"node_id"`
	Environment  string            `json:"environment"`
	StartedAt    string            `json:"started_at"`
	Uptime       string            `json:"uptime"`
	GoVersion    string            `json:"go_version"`
	GOOS         string            `json:"go_os"`
	GOARCH       string            `json:"go_arch"`
	GOMAXPROCS   int               `json:"gomaxprocs"`
	ClusterMode  bool              `json:"cluster_mode"`
	ClusterNodes []clusterNodeInfo `json:"cluster_nodes,omitempty"`
}

type clusterNodeInfo struct {
	NodeID   string `json:"node_id"`
	LastSeen string `json:"last_seen"`
	Status   string `json:"status"`
}

func (p *Panel) handleDeploymentInfo(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "deployment_view") {
		return
	}

	runtimeLabel := classifyRuntime(p.config.SessionRuntime)

	info := deploymentInfo{
		Runtime:     runtimeLabel,
		Host:        strings.TrimSpace(p.config.SessionRuntime.Host),
		Pod:         strings.TrimSpace(p.config.SessionRuntime.Pod),
		Instance:    strings.TrimSpace(p.config.SessionRuntime.Instance),
		NodeID:      p.liveNodeID(),
		Environment: strings.TrimSpace(p.config.Environment),
		GoVersion:   runtime.Version(),
		GOOS:        runtime.GOOS,
		GOARCH:      runtime.GOARCH,
		GOMAXPROCS:  runtime.GOMAXPROCS(0),
		ClusterMode: p.config.LiveClusterEnabled,
	}

	// Add cluster nodes if enabled
	if p.config.LiveClusterEnabled {
		info.ClusterNodes = p.getClusterNodes()
	}

	writeJSON(w, http.StatusOK, info)
}

func (p *Panel) getClusterNodes() []clusterNodeInfo {
	if p.live == nil {
		return nil
	}

	nodes := p.live.nodes.active(liveNodeOnlineWindow)
	result := make([]clusterNodeInfo, 0, len(nodes))
	for id, lastSeen := range nodes {
		status := "online"
		if time.Since(lastSeen) > liveNodeDegradedWindow {
			status = "degraded"
		}
		result = append(result, clusterNodeInfo{
			NodeID:   id,
			LastSeen: lastSeen.Format(time.RFC3339),
			Status:   status,
		})
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].NodeID < result[j].NodeID
	})
	return result
}

// Cache management API handlers

func (p *Panel) handleCacheStats(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "cache_view") {
		return
	}

	snapshot := inspectRedisRuntime(r.Context(), p.config.RedisURL)
	if !snapshot.Enabled {
		writeJSON(w, http.StatusOK, map[string]interface{}{
			"enabled": false,
			"reason":  snapshot.Message,
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled":    true,
		"redis_url":  snapshot.RedisURL,
		"status":     snapshot.Status,
		"message":    snapshot.Message,
		"latency_ms": snapshot.LatencyMS,
		"key_count":  snapshot.KeyCount,
	})
}

func (p *Panel) handleFlushCache(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "cache_manage") {
		return
	}

	result, err := flushRedisRuntime(r.Context(), p.config.RedisURL)
	if err != nil {
		writeErr(w, gferrors.BadRequest(err.Error()))
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"flushed":          true,
		"redis_url":        result.RedisURL,
		"status":           result.Status,
		"message":          result.Message,
		"latency_ms":       result.LatencyMS,
		"key_count_before": result.KeyCountBefore,
		"key_count_after":  result.KeyCountAfter,
	})
}

// File storage browser API handlers

type storageFileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	IsDir   bool      `json:"is_dir"`
	ModTime time.Time `json:"mod_time"`
}

func (p *Panel) handleListStorage(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "storage_view") {
		return
	}

	storagePath, err := normalizeStorageBrowsePath(r.URL.Query().Get("path"))
	if err != nil {
		writeErr(w, gferrors.Forbidden(err.Error()))
		return
	}

	if p.store != nil {
		files, err := listConfiguredStorage(r.Context(), p.store, storagePath)
		if err != nil {
			writeErr(w, err)
			return
		}

		writeJSON(w, http.StatusOK, map[string]interface{}{
			"backend": "store",
			"path":    storagePath,
			"files":   files,
			"total":   len(files),
		})
		return
	}

	relativePath := strings.TrimPrefix(storagePath, adminStorageBrowseRoot)
	relativePath = strings.TrimPrefix(relativePath, "/")

	absRoot, err := filepath.Abs(adminStorageBrowseRoot)
	if err != nil {
		writeErr(w, gferrors.InternalError("resolve storage root"))
		return
	}
	absPath := filepath.Join(absRoot, filepath.FromSlash(relativePath))

	entries, err := os.ReadDir(absPath)
	if err != nil {
		writeErr(w, gferrors.NotFound("storage path", storagePath))
		return
	}

	files := make([]storageFileInfo, 0, len(entries))
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}
		files = append(files, storageFileInfo{
			Name:    entry.Name(),
			Path:    filepath.Join(storagePath, entry.Name()),
			Size:    info.Size(),
			IsDir:   entry.IsDir(),
			ModTime: info.ModTime(),
		})
	}

	sortStorageEntries(files)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"backend": "filesystem",
		"path":    storagePath,
		"files":   files,
		"total":   len(files),
	})
}

// Email stats API handlers

func (p *Panel) handleEmailStats(w http.ResponseWriter, r *http.Request) {
	if !p.authorizeAction(w, r, "*", "email_view") {
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"enabled": true,
		"note":    "Email queue stats would require mail driver integration",
	})
}
