package controllers

import (
	"net/http"

	"example.com/showcase_clean/internal/services"
	gfrender "github.com/jcsvwinston/GoFrame/pkg/router"
)

type createArticleInput struct {
	Title     string `json:"title" validate:"required,min=3"`
	Content   string `json:"content"`
	Published bool   `json:"published"`
}

func Health(c *gfrender.Context) error {
	return c.JSON(http.StatusOK, map[string]any{"status": "ok"})
}

func ListArticles(articleService *services.ArticleService) gfrender.Handler {
	return func(c *gfrender.Context) error {
		items, err := articleService.List(c.Request.Context(), services.ListArticleInput{
			Query: c.Query("q"),
		})
		if err != nil {
			return err
		}

		return c.JSON(http.StatusOK, map[string]any{
			"data":  items,
			"count": len(items),
		})
	}
}

func CreateArticle(articleService *services.ArticleService) gfrender.Handler {
	return func(c *gfrender.Context) error {
		var in createArticleInput
		if err := c.Bind(&in); err != nil {
			return err
		}

		item, err := articleService.Create(c.Request.Context(), services.CreateArticleInput{
			Title:     in.Title,
			Content:   in.Content,
			Published: in.Published,
		})
		if err != nil {
			return err
		}
		return c.JSON(http.StatusCreated, map[string]any{
			"data": item,
		})
	}
}
