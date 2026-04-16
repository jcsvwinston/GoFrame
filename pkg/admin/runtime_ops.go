package admin

import (
	"context"
	"fmt"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/jcsvwinston/GoFrame/pkg/storage"
	"github.com/redis/go-redis/v9"
)

const (
	adminStorageBrowseRoot = "uploads"
	adminRedisDialTimeout  = 500 * time.Millisecond
)

type redisRuntimeSnapshot struct {
	Enabled   bool   `json:"enabled"`
	RedisURL  string `json:"redis_url,omitempty"`
	Status    string `json:"status"`
	Message   string `json:"message"`
	LatencyMS int64  `json:"latency_ms,omitempty"`
	KeyCount  int64  `json:"key_count,omitempty"`
}

type redisFlushResult struct {
	RedisURL       string `json:"redis_url,omitempty"`
	Status         string `json:"status"`
	Message        string `json:"message"`
	LatencyMS      int64  `json:"latency_ms,omitempty"`
	KeyCountBefore int64  `json:"key_count_before,omitempty"`
	KeyCountAfter  int64  `json:"key_count_after,omitempty"`
}

func inspectRedisRuntime(ctx context.Context, redisURL string) redisRuntimeSnapshot {
	redisURL = strings.TrimSpace(redisURL)
	if redisURL == "" {
		return redisRuntimeSnapshot{
			Enabled: false,
			Status:  "disabled",
			Message: "redis url is not configured",
		}
	}

	client, err := newAdminRedisClient(redisURL)
	if err != nil {
		return redisRuntimeSnapshot{
			Enabled:  true,
			RedisURL: redisURL,
			Status:   "unhealthy",
			Message:  err.Error(),
		}
	}
	defer client.Close()

	start := time.Now()
	if err := client.Ping(ctx).Err(); err != nil {
		return redisRuntimeSnapshot{
			Enabled:   true,
			RedisURL:  redisURL,
			Status:    "unhealthy",
			Message:   err.Error(),
			LatencyMS: time.Since(start).Milliseconds(),
		}
	}

	keyCount, err := client.DBSize(ctx).Result()
	if err != nil {
		return redisRuntimeSnapshot{
			Enabled:   true,
			RedisURL:  redisURL,
			Status:    "degraded",
			Message:   fmt.Sprintf("connected but failed to inspect cache size: %v", err),
			LatencyMS: time.Since(start).Milliseconds(),
		}
	}

	return redisRuntimeSnapshot{
		Enabled:   true,
		RedisURL:  redisURL,
		Status:    "healthy",
		Message:   "connected",
		LatencyMS: time.Since(start).Milliseconds(),
		KeyCount:  keyCount,
	}
}

func flushRedisRuntime(ctx context.Context, redisURL string) (redisFlushResult, error) {
	redisURL = strings.TrimSpace(redisURL)
	if redisURL == "" {
		return redisFlushResult{}, fmt.Errorf("redis url is not configured")
	}

	client, err := newAdminRedisClient(redisURL)
	if err != nil {
		return redisFlushResult{}, err
	}
	defer client.Close()

	start := time.Now()
	before, err := client.DBSize(ctx).Result()
	if err != nil {
		return redisFlushResult{}, err
	}
	if err := client.FlushDB(ctx).Err(); err != nil {
		return redisFlushResult{}, err
	}
	after, err := client.DBSize(ctx).Result()
	if err != nil {
		return redisFlushResult{}, err
	}

	return redisFlushResult{
		RedisURL:       redisURL,
		Status:         "healthy",
		Message:        "cache flushed",
		LatencyMS:      time.Since(start).Milliseconds(),
		KeyCountBefore: before,
		KeyCountAfter:  after,
	}, nil
}

func newAdminRedisClient(redisURL string) (*redis.Client, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis url: %w", err)
	}
	options.DialTimeout = adminRedisDialTimeout
	options.ReadTimeout = adminRedisDialTimeout
	options.WriteTimeout = adminRedisDialTimeout
	return redis.NewClient(options), nil
}

func normalizeStorageBrowsePath(raw string) (string, error) {
	browsePath := strings.TrimSpace(raw)
	if browsePath == "" || browsePath == "/" {
		return adminStorageBrowseRoot, nil
	}

	normalized := path.Clean("/" + strings.ReplaceAll(browsePath, "\\", "/"))
	normalized = strings.TrimPrefix(normalized, "/")
	if normalized == "." || normalized == "" {
		return adminStorageBrowseRoot, nil
	}
	if normalized != adminStorageBrowseRoot && !strings.HasPrefix(normalized, adminStorageBrowseRoot+"/") {
		return "", fmt.Errorf("access denied: path outside storage root")
	}
	return normalized, nil
}

func listConfiguredStorage(ctx context.Context, store storage.Store, browsePath string) ([]storageFileInfo, error) {
	prefix := browsePath
	if prefix != "" && !strings.HasSuffix(prefix, "/") {
		prefix += "/"
	}

	result, err := store.List(ctx, storage.ListOptions{
		Prefix:    prefix,
		Delimiter: "/",
		Limit:     1000,
	})
	if err != nil {
		return nil, err
	}

	files := make([]storageFileInfo, 0, len(result.Objects)+len(result.CommonPrefixes))
	for _, dir := range result.CommonPrefixes {
		dirPath := strings.TrimSuffix(dir, "/")
		files = append(files, storageFileInfo{
			Name:  path.Base(dirPath),
			Path:  dirPath,
			IsDir: true,
		})
	}
	for _, object := range result.Objects {
		objectPath := strings.TrimSuffix(object.Key, "/")
		files = append(files, storageFileInfo{
			Name:    path.Base(objectPath),
			Path:    objectPath,
			Size:    object.Size,
			IsDir:   false,
			ModTime: object.UpdatedAt,
		})
	}

	sortStorageEntries(files)
	return files, nil
}

func sortStorageEntries(files []storageFileInfo) {
	sort.Slice(files, func(i, j int) bool {
		if files[i].IsDir != files[j].IsDir {
			return files[i].IsDir
		}
		return files[i].Name < files[j].Name
	})
}
