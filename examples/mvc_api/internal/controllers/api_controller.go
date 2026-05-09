package controllers

import (
	"net/http"
	"time"

	"github.com/jcsvwinston/nucleus/examples/mvc_api/internal/dtos"
	"github.com/jcsvwinston/nucleus/examples/mvc_api/internal/services"
	"github.com/jcsvwinston/nucleus/pkg/app"
	"github.com/jcsvwinston/nucleus/pkg/router"
)

func Health() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		router.JSON(w, http.StatusOK, map[string]any{"status": "ok", "service": "nucleus-mvc-api"})
	}
}

func ListArticles(svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := svc.ListArticles(r.Context(), false, 100)
		if err != nil {
			router.Error(w, r, err, nil)
			return
		}
		router.JSON(w, http.StatusOK, map[string]any{
			"data":  items,
			"count": len(items),
		})
	}
}

func ListArticlesLiveFlag(a *app.App, svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		previewMode, _ := a.Admin.FeatureFlag("articles_preview_mode")
		items, err := svc.ListArticles(r.Context(), !previewMode, 100)
		if err != nil {
			router.Error(w, r, err, nil)
			return
		}
		router.JSON(w, http.StatusOK, map[string]any{
			"feature_flag": "articles_preview_mode",
			"enabled":      previewMode,
			"mode":         map[bool]string{true: "preview", false: "published-only"}[previewMode],
			"data":         items,
			"count":        len(items),
		})
	}
}

func CreateArticle(svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in dtos.CreateArticleInput
		if err := router.Bind(r, &in); err != nil {
			router.Error(w, r, err, nil)
			return
		}

		if in.Title == "" {
			router.JSON(w, http.StatusBadRequest, map[string]any{
				"error": map[string]any{
					"code":    "VALIDATION_ERROR",
					"message": "title is required",
				},
			})
			return
		}

		now := time.Now().UTC()
		res, err := svc.SQLDB.ExecContext(
			r.Context(),
			`INSERT INTO articles (created_at, updated_at, title, content, published) VALUES (?, ?, ?, ?, ?)`,
			now, now, in.Title, in.Content, in.Published,
		)
		if err != nil {
			router.Error(w, r, err, nil)
			return
		}
		id, _ := res.LastInsertId()

		router.Created(w, map[string]any{
			"data": map[string]any{
				"id":         id,
				"title":      in.Title,
				"content":    in.Content,
				"published":  in.Published,
				"created_at": now,
			},
		})
	}
}

func ListLeads(svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := svc.ListLeads(r.Context(), 100)
		if err != nil {
			router.Error(w, r, err, nil)
			return
		}
		router.JSON(w, http.StatusOK, map[string]any{
			"data":  items,
			"count": len(items),
		})
	}
}

func CreateLead(svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in dtos.CreateLeadInput
		if err := router.Bind(r, &in); err != nil {
			router.Error(w, r, err, nil)
			return
		}

		if in.Name == "" || in.Email == "" {
			status := http.StatusBadRequest
			fields := map[string]string{}
			if in.Name == "" {
				fields["name"] = "required"
			}
			if in.Email == "" {
				fields["email"] = "required"
			}
			router.JSON(w, status, map[string]any{
				"error": map[string]any{
					"code":    "VALIDATION_ERROR",
					"message": "validation failed",
					"fields":  fields,
				},
			})
			return
		}

		item, err := svc.CreateLead(r.Context(), in)
		if err != nil {
			router.Error(w, r, err, nil)
			return
		}
		router.Created(w, map[string]any{"data": item})
	}
}
