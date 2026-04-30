package controllers

import (
	"html/template"
	"net/http"

	"github.com/jcsvwinston/GoFrame/examples/mvc_api/internal/config"
	"github.com/jcsvwinston/GoFrame/examples/mvc_api/internal/dtos"
	"github.com/jcsvwinston/GoFrame/examples/mvc_api/internal/services"
	"github.com/jcsvwinston/GoFrame/pkg/app"
)

func HomePage(tpl *template.Template, cfg *app.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		adminPassword := ""
		if cfg != nil {
			adminPassword = cfg.AdminBootstrapPassword
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.ExecuteTemplate(w, "home.html", map[string]any{
			"Title":         "GoFrame Showcase",
			"AdminPassword": adminPassword,
			"DemoUser":      config.DemoAppUsername,
			"DemoPassword":  config.DemoAppPassword,
		})
	}
}

func PublishedArticlesPage(tpl *template.Template, svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		articles, err := svc.ListArticles(r.Context(), true, 24)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.ExecuteTemplate(w, "articles.html", map[string]any{
			"Title":    "Published Articles",
			"Articles": articles,
		})
	}
}

func LeadCapturePage(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		renderLeadCapturePage(w, tpl, http.StatusOK, dtos.LeadCapturePageData{
			Title:       "Request a Demo",
			FieldErrors: map[string]string{},
		})
	}
}

func LeadCaptureSubmit(a *app.App, tpl *template.Template, svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			renderLeadCapturePage(w, tpl, http.StatusBadRequest, dtos.LeadCapturePageData{
				Title:       "Request a Demo",
				FieldErrors: map[string]string{"form": "Invalid form data"},
			})
			return
		}

		in := dtos.CreateLeadInput{
			Name:      r.FormValue("name"),
			Email:     r.FormValue("email"),
			Company:   r.FormValue("company"),
			WantsDemo: r.FormValue("wants_demo") == "on" || r.FormValue("wants_demo") == "true",
		}

		_, err := svc.CreateLead(r.Context(), in)
		if err != nil {
			renderLeadCapturePage(w, tpl, http.StatusBadRequest, dtos.LeadCapturePageData{
				Title:       "Request a Demo",
				Form:        in,
				FieldErrors: map[string]string{"form": err.Error()},
			})
			return
		}

		renderLeadCapturePage(w, tpl, http.StatusOK, dtos.LeadCapturePageData{
			Title:     "Request a Demo",
			Submitted: true,
		})
	}
}

func renderLeadCapturePage(w http.ResponseWriter, tpl *template.Template, status int, data dtos.LeadCapturePageData) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = tpl.ExecuteTemplate(w, "contact.html", data)
}

func AppLoginPage(tpl *template.Template) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.ExecuteTemplate(w, "app_login.html", map[string]any{
			"Title": "Demo App Login",
		})
	}
}

func AppLoginPost(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			http.Redirect(w, r, "/app/login?error=invalid_form", http.StatusSeeOther)
			return
		}
		if a.Session != nil {
			_ = a.Session.RenewToken(r.Context())
			a.Session.Put(r.Context(), "app_user", config.DemoAppUsername)
		}
		http.Redirect(w, r, "/app/dashboard", http.StatusSeeOther)
	}
}

func AppLogout(a *app.App) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if a.Session != nil {
			_ = a.Session.Destroy(r.Context())
		}
		http.Redirect(w, r, "/app/login", http.StatusSeeOther)
	}
}

func AppDashboard(a *app.App, tpl *template.Template, svc *services.Services) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user := ""
		if a.Session != nil {
			user = a.Session.GetString(r.Context(), "app_user")
		}
		if user == "" {
			http.Redirect(w, r, "/app/login", http.StatusSeeOther)
			return
		}

		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_ = tpl.ExecuteTemplate(w, "app_dashboard.html", map[string]any{
			"Title":        "Demo Dashboard",
			"User":         user,
			"ArticleCount": svc.CountRows("articles"),
			"LeadCount":    svc.CountRows("leads"),
		})
	}
}
