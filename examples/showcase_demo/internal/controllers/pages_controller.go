package controllers

import (
	"bytes"
	"database/sql"
	"html/template"
	"net/http"
	"time"

	gfrender "github.com/jcsvwinston/GoFrame/pkg/router"
	"github.com/yuin/goldmark"
)

// markdownToHTML converts markdown to HTML
func markdownToHTML(content string) template.HTML {
	var buf bytes.Buffer
	if err := goldmark.Convert([]byte(content), &buf); err != nil {
		return template.HTML(content)
	}
	return template.HTML(buf.String())
}

// BlogPage renders the blog listing page
func BlogPage(tpl *template.Template, db *sql.DB) gfrender.Handler {
	return func(c *gfrender.Context) error {
		// Get articles with author and category info
		rows, err := db.Query(`
			SELECT a.id, a.title, a.slug, a.summary, a.published_at, a.view_count,
					au.name as author_name, au.avatar_url,
					c.name as category_name, c.color as category_color
			FROM articles a
			JOIN authors au ON a.author_id = au.id
			JOIN categories c ON a.category_id = c.id
			WHERE a.published = 1
			ORDER BY a.published_at DESC
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var articles []map[string]interface{}
		for rows.Next() {
			var id, viewCount int
			var title, slug, summary, authorName, avatar, categoryName, categoryColor string
			var publishedAt time.Time
			if err := rows.Scan(&id, &title, &slug, &summary, &publishedAt, &viewCount,
				&authorName, &avatar, &categoryName, &categoryColor); err != nil {
				continue
			}
			articles = append(articles, map[string]interface{}{
				"ID":            id,
				"Title":         title,
				"Slug":          slug,
				"Summary":       summary,
				"PublishedAt":   publishedAt.Format("Jan 2, 2006"),
				"ViewCount":     viewCount,
				"AuthorName":    authorName,
				"AvatarURL":     avatar,
				"CategoryName":  categoryName,
				"CategoryColor": categoryColor,
			})
		}

		return c.HTML(http.StatusOK, "blog.html", map[string]interface{}{
			"Title":    "Blog",
			"Articles": articles,
		})
	}
}

// ArticlePage renders a single article page
func ArticlePage(tpl *template.Template, db *sql.DB) gfrender.Handler {
	return func(c *gfrender.Context) error {
		slug := c.Request.PathValue("id")
		if slug == "" {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "slug required"})
		}

		// Increment view count
		db.Exec("UPDATE articles SET view_count = view_count + 1 WHERE slug = ?", slug)

		var article map[string]interface{}
		row := db.QueryRow(`
			SELECT a.id, a.title, a.content, a.published_at, a.view_count,
					au.name as author_name, au.bio, au.avatar_url,
					au.social_github, au.social_twitter,
					c.name as category_name, c.color as category_color
			FROM articles a
			JOIN authors au ON a.author_id = au.id
			JOIN categories c ON a.category_id = c.id
			WHERE a.slug = ? AND a.published = 1
		`, slug)

		var id, viewCount int
		var title, content, authorName, bio, avatar, github, twitter, categoryName, categoryColor string
		var publishedAt time.Time
		err := row.Scan(&id, &title, &content, &publishedAt, &viewCount,
			&authorName, &bio, &avatar, &github, &twitter, &categoryName, &categoryColor)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "article not found"})
		}

		article = map[string]interface{}{
			"ID":            id,
			"Title":         title,
			"Content":       markdownToHTML(content),
			"PublishedAt":   publishedAt.Format("January 2, 2006"),
			"ViewCount":     viewCount,
			"AuthorName":    authorName,
			"AuthorBio":     bio,
			"AuthorAvatar":  avatar,
			"AuthorGitHub":  github,
			"AuthorTwitter": twitter,
			"CategoryName":  categoryName,
			"CategoryColor": categoryColor,
		}

		// Get related articles
		relatedRows, _ := db.Query(`
			SELECT a.id, a.title, a.slug, a.summary, au.name as author_name
			FROM articles a
			JOIN authors au ON a.author_id = au.id
			WHERE a.published = 1 AND a.slug != ?
			ORDER BY RANDOM()
			LIMIT 3
		`, slug)
		defer relatedRows.Close()

		var related []map[string]string
		for relatedRows.Next() {
			var rid int
			var rtitle, rslug, rsummary, rauthor string
			if err := relatedRows.Scan(&rid, &rtitle, &rslug, &rsummary, &rauthor); err == nil {
				related = append(related, map[string]string{
					"Title":   rtitle,
					"Slug":    rslug,
					"Summary": rsummary,
					"Author":  rauthor,
				})
			}
		}

		return c.HTML(http.StatusOK, "article.html", map[string]interface{}{
			"Title":   title,
			"Article": article,
			"Related": related,
		})
	}
}

// CategoriesPage renders the categories listing page
func CategoriesPage(tpl *template.Template, db *sql.DB) gfrender.Handler {
	return func(c *gfrender.Context) error {
		rows, err := db.Query(`
			SELECT id, name, slug, description, color, icon, article_count
			FROM categories
			ORDER BY article_count DESC
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var categories []map[string]interface{}
		for rows.Next() {
			var id, count int
			var name, slug, desc, color, icon string
			if err := rows.Scan(&id, &name, &slug, &desc, &color, &icon, &count); err != nil {
				continue
			}
			categories = append(categories, map[string]interface{}{
				"ID":           id,
				"Name":         name,
				"Slug":         slug,
				"Description":  desc,
				"Color":        color,
				"Icon":         icon,
				"ArticleCount": count,
			})
		}

		return c.HTML(http.StatusOK, "categories.html", map[string]interface{}{
			"Title":      "Categorías",
			"Categories": categories,
		})
	}
}

// AboutPage renders the about page with team info
func AboutPage(tpl *template.Template, db *sql.DB) gfrender.Handler {
	return func(c *gfrender.Context) error {
		// Get team authors
		rows, err := db.Query(`
			SELECT id, name, email, bio, position, avatar_url, social_github, social_twitter
			FROM authors
			ORDER BY id
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var team []map[string]interface{}
		for rows.Next() {
			var id int
			var name, email, bio, position, avatar, github, twitter string
			if err := rows.Scan(&id, &name, &email, &bio, &position, &avatar, &github, &twitter); err != nil {
				continue
			}
			team = append(team, map[string]interface{}{
				"ID":       id,
				"Name":     name,
				"Email":    email,
				"Bio":      bio,
				"Position": position,
				"Avatar":   avatar,
				"GitHub":   github,
				"Twitter":  twitter,
			})
		}

		// Get stats
		var stats struct {
			Articles, Categories, Authors, Comments int
		}
		db.QueryRow("SELECT COUNT(*) FROM articles WHERE published = 1").Scan(&stats.Articles)
		db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&stats.Categories)
		db.QueryRow("SELECT COUNT(*) FROM authors").Scan(&stats.Authors)
		db.QueryRow("SELECT COUNT(*) FROM comments WHERE approved = 1").Scan(&stats.Comments)

		return c.HTML(http.StatusOK, "about.html", map[string]interface{}{
			"Title": "Acerca de",
			"Team":  team,
			"Stats": stats,
		})
	}
}

// ContactPage renders the contact form page
func ContactPage(tpl *template.Template) gfrender.Handler {
	return func(c *gfrender.Context) error {
		return c.HTML(http.StatusOK, "contact.html", map[string]interface{}{
			"Title":     "Contacto",
			"Submitted": false,
		})
	}
}

// ContactSubmit handles contact form submission
func ContactSubmit(db *sql.DB) gfrender.Handler {
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

		_, err := db.Exec(
			"INSERT INTO contacts (created_at, name, email, company, subject, message) VALUES (?, ?, ?, ?, ?, ?)",
			time.Now(), name, email, company, subject, message,
		)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.HTML(http.StatusOK, "contact.html", map[string]interface{}{
			"Title":     "Contacto",
			"Submitted": true,
			"Success":   "¡Gracias por tu mensaje! Te contactaremos pronto.",
		})
	}
}

// GetArticleBySlug returns a single article by slug (API)
func GetArticleBySlug(db *sql.DB) gfrender.Handler {
	return func(c *gfrender.Context) error {
		slug := c.Request.PathValue("slug")
		var article map[string]interface{}
		row := db.QueryRow(`
			SELECT a.id, a.title, a.slug, a.summary, a.content, a.published_at, a.view_count,
					au.name as author_name, c.name as category_name
			FROM articles a
			JOIN authors au ON a.author_id = au.id
			JOIN categories c ON a.category_id = c.id
			WHERE a.slug = ? AND a.published = 1
		`, slug)

		var id, viewCount int
		var title, rslug, summary, content, authorName, categoryName string
		var publishedAt time.Time
		err := row.Scan(&id, &title, &rslug, &summary, &content, &publishedAt, &viewCount, &authorName, &categoryName)
		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}

		article = map[string]interface{}{
			"id":           id,
			"title":        title,
			"slug":         rslug,
			"summary":      summary,
			"content":      content,
			"published_at": publishedAt,
			"view_count":   viewCount,
			"author":       authorName,
			"category":     categoryName,
		}

		return c.JSON(http.StatusOK, article)
	}
}

// ListCategoriesAPI returns all categories (API)
func ListCategoriesAPI(db *sql.DB) gfrender.Handler {
	return func(c *gfrender.Context) error {
		rows, err := db.Query(`
			SELECT id, name, slug, description, color, icon, article_count
			FROM categories
			ORDER BY name
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var categories []map[string]interface{}
		for rows.Next() {
			var id, count int
			var name, slug, desc, color, icon string
			if err := rows.Scan(&id, &name, &slug, &desc, &color, &icon, &count); err != nil {
				continue
			}
			categories = append(categories, map[string]interface{}{
				"id":            id,
				"name":          name,
				"slug":          slug,
				"description":   desc,
				"color":         color,
				"icon":          icon,
				"article_count": count,
			})
		}

		return c.JSON(http.StatusOK, categories)
	}
}

// ListAuthorsAPI returns all authors (API)
func ListAuthorsAPI(db *sql.DB) gfrender.Handler {
	return func(c *gfrender.Context) error {
		rows, err := db.Query(`
			SELECT id, name, email, bio, position, avatar_url, article_count, social_github, social_twitter
			FROM authors
			ORDER BY name
		`)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		defer rows.Close()

		var authors []map[string]interface{}
		for rows.Next() {
			var id, count int
			var name, email, bio, position, avatar, github, twitter string
			if err := rows.Scan(&id, &name, &email, &bio, &position, &avatar, &count, &github, &twitter); err != nil {
				continue
			}
			authors = append(authors, map[string]interface{}{
				"id":             id,
				"name":           name,
				"email":          email,
				"bio":            bio,
				"position":       position,
				"avatar_url":     avatar,
				"article_count":  count,
				"social_github":  github,
				"social_twitter": twitter,
			})
		}

		return c.JSON(http.StatusOK, authors)
	}
}

// GetStatsAPI returns statistics (API)
func GetStatsAPI(db *sql.DB) gfrender.Handler {
	return func(c *gfrender.Context) error {
		var stats struct {
			Articles   int `json:"articles"`
			Categories int `json:"categories"`
			Authors    int `json:"authors"`
			Tags       int `json:"tags"`
			Comments   int `json:"comments"`
			Views      int `json:"total_views"`
		}

		db.QueryRow("SELECT COUNT(*) FROM articles WHERE published = 1").Scan(&stats.Articles)
		db.QueryRow("SELECT COUNT(*) FROM categories").Scan(&stats.Categories)
		db.QueryRow("SELECT COUNT(*) FROM authors").Scan(&stats.Authors)
		db.QueryRow("SELECT COUNT(*) FROM tags").Scan(&stats.Tags)
		db.QueryRow("SELECT COUNT(*) FROM comments WHERE approved = 1").Scan(&stats.Comments)
		db.QueryRow("SELECT COALESCE(SUM(view_count), 0) FROM articles").Scan(&stats.Views)

		return c.JSON(http.StatusOK, stats)
	}
}
