package controllers

import (
	"context"
	"net/http"

	"example.com/showcase_clean/internal/models"
	"example.com/showcase_clean/internal/quarkdb"
	gfrender "github.com/jcsvwinston/GoFrame/pkg/router"
	"github.com/jcsvwinston/quark"
)

// HomePageQuark renders the home page using Quark ORM
func HomePageQuark(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// Get latest 3 published articles using Quark ORM
		articles, err := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			OrderBy("published_at", "DESC").
			Limit(3).
			List()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Enrich articles with category info
		var enriched []map[string]interface{}
		for _, art := range articles {
			// Get category for this article
			cats, _ := quark.For[models.Category](ctx, client.Client).
				Where("id", "=", art.CategoryID).
				List()
			catName := ""
			catColor := "#666"
			if len(cats) > 0 {
				catName = cats[0].Name
				catColor = cats[0].Color
			}

			enriched = append(enriched, map[string]interface{}{
				"Title":         art.Title,
				"Slug":          art.Slug,
				"Summary":       art.Summary,
				"PublishedAt":   art.PublishedAt.Format("Jan 2, 2006"),
				"CategoryName":  catName,
				"CategoryColor": catColor,
			})
		}

		return c.HTML(http.StatusOK, "home.html", map[string]any{
			"Title":    "GoFrame Showcase",
			"Articles": enriched,
		})
	}
}

// BlogPageQuark renders the blog listing page using Quark ORM
func BlogPageQuark(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// Get all published articles using Quark ORM
		articles, err := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			OrderBy("published_at", "DESC").
			List()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Enrich with author and category info
		var enriched []map[string]interface{}
		for _, art := range articles {
			// Get author
			authors, _ := quark.For[models.Author](ctx, client.Client).
				Where("id", "=", art.AuthorID).
				List()
			authorName := ""
			authorAvatar := ""
			if len(authors) > 0 {
				authorName = authors[0].Name
				authorAvatar = authors[0].AvatarURL
			}

			// Get category
			cats, _ := quark.For[models.Category](ctx, client.Client).
				Where("id", "=", art.CategoryID).
				List()
			catName := ""
			catColor := "#666"
			if len(cats) > 0 {
				catName = cats[0].Name
				catColor = cats[0].Color
			}

			enriched = append(enriched, map[string]interface{}{
				"ID":            art.ID,
				"Title":         art.Title,
				"Slug":          art.Slug,
				"Summary":       art.Summary,
				"PublishedAt":   art.PublishedAt.Format("Jan 2, 2006"),
				"ViewCount":     art.ViewCount,
				"AuthorName":    authorName,
				"AvatarURL":     authorAvatar,
				"CategoryName":  catName,
				"CategoryColor": catColor,
			})
		}

		return c.HTML(http.StatusOK, "blog.html", map[string]interface{}{
			"Title":    "Blog",
			"Articles": enriched,
		})
	}
}

// ArticlePageQuark renders a single article page using Quark ORM
func ArticlePageQuark(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()
		slug := c.Request.PathValue("slug")
		if slug == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "slug required"})
		}

		// Get article by slug using Quark ORM
		articles, err := quark.For[models.Article](ctx, client.Client).
			Where("slug", "=", slug).
			Where("published", "=", true).
			List()

		if err != nil || len(articles) == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "article not found"})
		}

		art := articles[0]

		// Increment view count
		art.ViewCount++
		quark.For[models.Article](ctx, client.Client).Update(&art)

		// Get author
		authors, _ := quark.For[models.Author](ctx, client.Client).
			Where("id", "=", art.AuthorID).
			List()
		var author models.Author
		if len(authors) > 0 {
			author = authors[0]
		}

		// Get category
		cats, _ := quark.For[models.Category](ctx, client.Client).
			Where("id", "=", art.CategoryID).
			List()
		var cat models.Category
		if len(cats) > 0 {
			cat = cats[0]
		}

		article := map[string]interface{}{
			"ID":            art.ID,
			"Title":         art.Title,
			"Content":       markdownToHTML(art.Content),
			"PublishedAt":   art.PublishedAt.Format("January 2, 2006"),
			"ViewCount":     art.ViewCount,
			"AuthorName":    author.Name,
			"AuthorBio":     author.Bio,
			"AuthorAvatar":  author.AvatarURL,
			"AuthorGitHub":  author.SocialGitHub,
			"AuthorTwitter": author.SocialTwitter,
			"CategoryName":  cat.Name,
			"CategoryColor": cat.Color,
		}

		// Get related articles (random other articles)
		relatedArts, _ := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			Where("id", "!=", art.ID).
			Limit(3).
			List()

		var related []map[string]string
		for _, ra := range relatedArts {
			// Get author for related
			raAuthors, _ := quark.For[models.Author](ctx, client.Client).
				Where("id", "=", ra.AuthorID).
				List()
			raAuthor := ""
			if len(raAuthors) > 0 {
				raAuthor = raAuthors[0].Name
			}

			related = append(related, map[string]string{
				"Title":   ra.Title,
				"Slug":    ra.Slug,
				"Summary": ra.Summary,
				"Author":  raAuthor,
			})
		}

		return c.HTML(http.StatusOK, "article.html", map[string]interface{}{
			"Title":   art.Title,
			"Article": article,
			"Related": related,
		})
	}
}

// CategoriesPageQuark renders the categories listing page using Quark ORM
func CategoriesPageQuark(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// Get all categories using Quark ORM
		categories, err := quark.For[models.Category](ctx, client.Client).
			OrderBy("article_count", "DESC").
			List()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		var enriched []map[string]interface{}
		for _, cat := range categories {
			enriched = append(enriched, map[string]interface{}{
				"ID":           cat.ID,
				"Name":         cat.Name,
				"Slug":         cat.Slug,
				"Description":  cat.Description,
				"Color":        cat.Color,
				"Icon":         cat.Icon,
				"ArticleCount": cat.ArticleCount,
			})
		}

		return c.HTML(http.StatusOK, "categories.html", map[string]interface{}{
			"Title":      "Categorías",
			"Categories": enriched,
		})
	}
}

// ContactPageQuark renders the contact form page
func ContactPageQuark(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		return c.HTML(http.StatusOK, "contact.html", map[string]interface{}{
			"Title":     "Contacto",
			"Submitted": false,
		})
	}
}

// AboutPageQuark renders the about page with team info using Quark ORM
func AboutPageQuark(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// Get all authors using Quark ORM
		authors, err := quark.For[models.Author](ctx, client.Client).
			OrderBy("id", "ASC").
			List()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		var team []map[string]interface{}
		for _, a := range authors {
			team = append(team, map[string]interface{}{
				"ID":       a.ID,
				"Name":     a.Name,
				"Email":    a.Email,
				"Bio":      a.Bio,
				"Position": a.Position,
				"Avatar":   a.AvatarURL,
				"GitHub":   a.SocialGitHub,
				"Twitter":  a.SocialTwitter,
			})
		}

		// Get stats using Quark ORM count
		articleCount, _ := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			Count()
		categoryCount, _ := quark.For[models.Category](ctx, client.Client).Count()
		authorCount, _ := quark.For[models.Author](ctx, client.Client).Count()
		commentCount, _ := quark.For[models.Comment](ctx, client.Client).
			Where("approved", "=", true).
			Count()

		stats := struct {
			Articles   int
			Categories int
			Authors    int
			Comments   int
		}{
			Articles:   int(articleCount),
			Categories: int(categoryCount),
			Authors:    int(authorCount),
			Comments:   int(commentCount),
		}

		return c.HTML(http.StatusOK, "about.html", map[string]interface{}{
			"Title": "Acerca de",
			"Team":  team,
			"Stats": stats,
		})
	}
}

// ContactSubmitQuark handles contact form submission using Quark ORM
func ContactSubmitQuark(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		if err := c.Request.ParseForm(); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid form"})
		}

		name := c.Request.FormValue("name")
		email := c.Request.FormValue("email")
		company := c.Request.FormValue("company")
		subject := c.Request.FormValue("subject")
		message := c.Request.FormValue("message")

		if name == "" || email == "" || message == "" {
			return c.HTML(http.StatusOK, "contact.html", map[string]interface{}{
				"Title":   "Contacto",
				"Error":   "Por favor complete todos los campos obligatorios",
				"Name":    name,
				"Email":   email,
				"Company": company,
				"Subject": subject,
				"Message": message,
			})
		}

		// Note: Contact storage would require a Contact model
		// For now, just show success

		return c.HTML(http.StatusOK, "contact.html", map[string]interface{}{
			"Title":     "Contacto",
			"Submitted": true,
			"Success":   "¡Gracias por tu mensaje! Te contactaremos pronto.",
		})
	}
}

// ListCategoriesAPIQuark returns all categories using Quark ORM (API)
func ListCategoriesAPIQuark(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		categories, err := quark.For[models.Category](ctx, client.Client).
			OrderBy("name", "ASC").
			List()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		var result []map[string]interface{}
		for _, cat := range categories {
			result = append(result, map[string]interface{}{
				"id":            cat.ID,
				"name":          cat.Name,
				"slug":          cat.Slug,
				"description":   cat.Description,
				"color":         cat.Color,
				"icon":          cat.Icon,
				"article_count": cat.ArticleCount,
			})
		}

		return c.JSON(http.StatusOK, result)
	}
}

// ListAuthorsAPIQuark returns all authors using Quark ORM (API)
func ListAuthorsAPIQuark(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		authors, err := quark.For[models.Author](ctx, client.Client).
			OrderBy("name", "ASC").
			List()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		var result []map[string]interface{}
		for _, a := range authors {
			result = append(result, map[string]interface{}{
				"id":             a.ID,
				"name":           a.Name,
				"email":          a.Email,
				"bio":            a.Bio,
				"position":       a.Position,
				"avatar_url":     a.AvatarURL,
				"article_count":  a.ArticleCount,
				"social_github":  a.SocialGitHub,
				"social_twitter": a.SocialTwitter,
			})
		}

		return c.JSON(http.StatusOK, result)
	}
}

// GetArticleBySlugQuark returns a single article by slug using Quark ORM (API)
func GetArticleBySlugQuark(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()
		slug := c.Request.PathValue("slug")

		articles, err := quark.For[models.Article](ctx, client.Client).
			Where("slug", "=", slug).
			Where("published", "=", true).
			List()

		if err != nil || len(articles) == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}

		art := articles[0]

		// Get author
		authors, _ := quark.For[models.Author](ctx, client.Client).
			Where("id", "=", art.AuthorID).
			List()
		authorName := ""
		if len(authors) > 0 {
			authorName = authors[0].Name
		}

		// Get category
		cats, _ := quark.For[models.Category](ctx, client.Client).
			Where("id", "=", art.CategoryID).
			List()
		catName := ""
		if len(cats) > 0 {
			catName = cats[0].Name
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"id":           art.ID,
			"title":        art.Title,
			"slug":         art.Slug,
			"summary":      art.Summary,
			"content":      art.Content,
			"published_at": art.PublishedAt,
			"view_count":   art.ViewCount,
			"author":       authorName,
			"category":     catName,
		})
	}
}

// GetStatsAPIQuark returns statistics using Quark ORM (API)
func GetStatsAPIQuark(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		articleCount, _ := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			Count()
		categoryCount, _ := quark.For[models.Category](ctx, client.Client).Count()
		authorCount, _ := quark.For[models.Author](ctx, client.Client).Count()
		tagCount, _ := quark.For[models.Tag](ctx, client.Client).Count()
		commentCount, _ := quark.For[models.Comment](ctx, client.Client).
			Where("approved", "=", true).
			Count()

		// Calculate total views
		articles, _ := quark.For[models.Article](ctx, client.Client).List()
		totalViews := 0
		for _, art := range articles {
			totalViews += art.ViewCount
		}

		stats := struct {
			Articles   int `json:"articles"`
			Categories int `json:"categories"`
			Authors    int `json:"authors"`
			Tags       int `json:"tags"`
			Comments   int `json:"comments"`
			Views      int `json:"total_views"`
		}{
			Articles:   int(articleCount),
			Categories: int(categoryCount),
			Authors:    int(authorCount),
			Tags:       int(tagCount),
			Comments:   int(commentCount),
			Views:      totalViews,
		}

		return c.JSON(http.StatusOK, stats)
	}
}
