package controllers

import (
	"net/http"
	"strings"
	"time"

	"github.com/jcsvwinston/GoFrame/examples/mvc_api/internal/dtos"
	"github.com/jcsvwinston/GoFrame/examples/mvc_api/internal/services"
	"github.com/jcsvwinston/GoFrame/pkg/app"
	"github.com/jcsvwinston/GoFrame/pkg/outbox"
	"github.com/jcsvwinston/GoFrame/pkg/router"
	"github.com/jcsvwinston/GoFrame/pkg/tasks"
	asynqprovider "github.com/jcsvwinston/GoFrame/pkg/tasks/providers/asynq"
)

func DemoRuntime(a *app.App, svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		outboxSnapshot := outbox.RuntimeSnapshot{Enabled: false, Reason: "outbox is not configured", Table: outbox.DefaultTableName}
		if svc.OutboxStore != nil {
			outboxSnapshot = svc.OutboxStore.Snapshot(r.Context())
		}

		jobsSnapshot := tasks.RuntimeSnapshot{Enabled: false, Reason: "redis_url is not configured"}
		if strings.TrimSpace(a.Config.RedisURL) != "" {
			jobsSnapshot = asynqprovider.InspectRuntime(a.Config.RedisURL)
		}

		previewMode, _ := a.Admin.FeatureFlag("articles_preview_mode")
		router.JSON(w, http.StatusOK, map[string]any{
			"name":                  "goframe-mvc-api-showcase",
			"admin_prefix":          a.Config.AdminPrefix,
			"openapi_path":          "/openapi.json",
			"feature_flags":         map[string]bool{"articles_preview_mode": previewMode},
			"outbox":                outboxSnapshot,
			"jobs":                  jobsSnapshot,
			"admin_cluster_enabled": a.Config.AdminClusterEnabled,
			"redis_configured":      strings.TrimSpace(a.Config.RedisURL) != "",
		})
	}
}

func EnqueueOutbox(a *app.App, svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svc.OutboxStore == nil {
			router.JSON(w, http.StatusBadRequest, map[string]any{
				"error": map[string]any{
					"code":    "OUTBOX_DISABLED",
					"message": "outbox store is not configured",
				},
			})
			return
		}
		msg, err := svc.OutboxStore.Enqueue(r.Context(), outbox.Entry{
			Topic: "demo.manual.trigger",
			Payload: map[string]any{
				"path":      r.URL.Path,
				"triggered": time.Now().UTC().Format(time.RFC3339),
			},
		})
		if err != nil {
			router.Error(w, err, a.Logger)
			return
		}
		router.Created(w, map[string]any{
			"data":     msg,
			"snapshot": svc.OutboxStore.Snapshot(r.Context()),
		})
	}
}

func DrainOutbox(a *app.App, svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svc.OutboxDispatcher == nil || svc.OutboxStore == nil {
			router.JSON(w, http.StatusBadRequest, map[string]any{
				"error": map[string]any{
					"code":    "OUTBOX_DISABLED",
					"message": "outbox runtime is not configured",
				},
			})
			return
		}
		result, err := svc.OutboxDispatcher.RunOnce(r.Context())
		if err != nil {
			router.Error(w, err, a.Logger)
			return
		}
		router.JSON(w, http.StatusOK, map[string]any{
			"result":   result,
			"snapshot": svc.OutboxStore.Snapshot(r.Context()),
		})
	}
}

func EnqueueTask(a *app.App, svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if svc.TaskManager == nil {
			router.JSON(w, http.StatusBadRequest, map[string]any{
				"error": map[string]any{
					"code":    "TASKS_DISABLED",
					"message": "set GOFRAME_EXAMPLE_REDIS_URL to enable task demo endpoints",
				},
			})
			return
		}

		policy := tasks.DefaultEnqueuePolicy()
		policy.Queue = "critical"
		policy.MaxRetry = 2
		id, err := svc.TaskManager.EnqueueJSONCtxWithPolicy(r.Context(), "demo.email.send", dtos.DemoTaskPayload{
			Kind:     "manual",
			Target:   "team@example.com",
			Source:   "/api/demo/tasks",
			QueuedAt: time.Now().UTC().Format(time.RFC3339),
		}, policy)
		if err != nil {
			router.Error(w, err, a.Logger)
			return
		}

		router.Created(w, map[string]any{
			"data": map[string]any{
				"id":    id,
				"queue": "critical",
				"type":  "demo.email.send",
			},
			"snapshot": asynqprovider.InspectRuntime(a.Config.RedisURL),
		})
	}
}
