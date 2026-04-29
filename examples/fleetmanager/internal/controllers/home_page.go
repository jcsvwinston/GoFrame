package controllers

import (
	"html/template"
	"net/http"

	gfrender "github.com/jcsvwinston/GoFrame/pkg/router"
)

func HomePage(tpl *template.Template) gfrender.Handler {
	return func(c *gfrender.Context) error {
		return c.HTML(http.StatusOK, "home.html", map[string]any{
			"Title": "GoFrame Starter",
		})
	}
}
