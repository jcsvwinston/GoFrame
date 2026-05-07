package controllers

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"example.com/showcase_clean/internal/models"
	"example.com/showcase_clean/internal/quarkdb"
	gfrender "github.com/jcsvwinston/GoFrame/pkg/router"
	"github.com/jcsvwinston/quark"
)

// DataTablesRequest represents the request structure from DataTables
type DataTablesRequest struct {
	Draw    int    `json:"draw"`
	Start   int    `json:"start"`
	Length  int    `json:"length"`
	Search  Search `json:"search"`
	Order   []Order
	Columns []Column
}

type Search struct {
	Value string `json:"value"`
	Regex bool   `json:"regex"`
}

type Order struct {
	Column int    `json:"column"`
	Dir    string `json:"dir"`
}

type Column struct {
	Data       string `json:"data"`
	Name       string `json:"name"`
	Searchable bool   `json:"searchable"`
	Orderable  bool   `json:"orderable"`
	Search     Search `json:"search"`
}

// DataTablesResponse represents the response structure for DataTables
type DataTablesResponse struct {
	Draw            int                      `json:"draw"`
	RecordsTotal    int64                    `json:"recordsTotal"`
	RecordsFiltered int64                    `json:"recordsFiltered"`
	Data            []map[string]interface{} `json:"data"`
	Error           string                   `json:"error,omitempty"`
}

// APIDataTableArticles handles server-side processing for articles DataTable
func APIDataTableArticles(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		var req DataTablesRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		// Default values
		if req.Length == 0 {
			req.Length = 10
		}
		if req.Length == -1 {
			req.Length = 100 // Max limit
		}

		// Build query
		query := quark.For[models.Article](ctx, client.Client)

		// Apply search filter
		if req.Search.Value != "" {
			query = query.Where("title", "LIKE", "%"+req.Search.Value+"%")
		}

		// Get total records
		totalRecords, _ := quark.For[models.Article](ctx, client.Client).Count()

		// Get filtered records count
		filteredRecords, _ := query.Count()

		// Apply ordering
		if len(req.Order) > 0 {
			order := req.Order[0]
			dir := "ASC"
			if order.Dir == "desc" {
				dir = "DESC"
			}
			query = query.OrderBy("published_at", dir)
		} else {
			query = query.OrderBy("published_at", "DESC")
		}

		// Apply pagination
		query = query.Limit(req.Length)
		query = query.Offset(req.Start)

		// Get data
		articles, err := query.List()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Collect unique author IDs and category IDs from articles
		authorIDs := make([]int64, 0)
		categoryIDs := make([]int64, 0)
		authorIDSet := make(map[int64]bool)
		categoryIDSet := make(map[int64]bool)

		for _, article := range articles {
			if !authorIDSet[article.AuthorID] {
				authorIDs = append(authorIDs, article.AuthorID)
				authorIDSet[article.AuthorID] = true
			}
			if !categoryIDSet[article.CategoryID] {
				categoryIDs = append(categoryIDs, article.CategoryID)
				categoryIDSet[article.CategoryID] = true
			}
		}

		// Get only the authors we need
		authorMap := make(map[int64]string)
		for _, authorID := range authorIDs {
			authors, _ := quark.For[models.Author](ctx, client.Client).
				Where("id", "=", authorID).
				Limit(1).
				List()
			if len(authors) > 0 {
				authorMap[authorID] = authors[0].Name
			}
		}

		// Get only the categories we need
		categoryMap := make(map[int64]string)
		for _, categoryID := range categoryIDs {
			categories, _ := quark.For[models.Category](ctx, client.Client).
				Where("id", "=", categoryID).
				Limit(1).
				List()
			if len(categories) > 0 {
				categoryMap[categoryID] = categories[0].Name
			}
		}

		// Transform data for DataTables
		data := make([]map[string]interface{}, len(articles))
		for i, article := range articles {
			data[i] = map[string]interface{}{
				"id":           article.ID,
				"title":        article.Title,
				"slug":         article.Slug,
				"author":       authorMap[article.AuthorID],
				"category":     categoryMap[article.CategoryID],
				"published":    article.Published,
				"published_at": article.PublishedAt.Format("2006-01-02 15:04"),
				"view_count":   article.ViewCount,
				"created_at":   article.CreatedAt.Format("2006-01-02 15:04"),
			}
		}

		response := DataTablesResponse{
			Draw:            req.Draw,
			RecordsTotal:    totalRecords,
			RecordsFiltered: filteredRecords,
			Data:            data,
		}

		return c.JSON(http.StatusOK, response)
	}
}

// APIDataTableAuthors handles server-side processing for authors DataTable
func APIDataTableAuthors(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		var req DataTablesRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if req.Length == 0 {
			req.Length = 10
		}
		if req.Length == -1 {
			req.Length = 100
		}

		query := quark.For[models.Author](ctx, client.Client)

		if req.Search.Value != "" {
			query = query.Where("name", "LIKE", "%"+req.Search.Value+"%")
		}

		totalRecords, _ := quark.For[models.Author](ctx, client.Client).Count()
		filteredRecords, _ := query.Count()

		query = query.OrderBy("created_at", "DESC")
		query = query.Limit(req.Length)
		query = query.Offset(req.Start)

		authors, err := query.List()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		data := make([]map[string]interface{}, len(authors))
		for i, author := range authors {
			data[i] = map[string]interface{}{
				"id":            author.ID,
				"name":          author.Name,
				"email":         author.Email,
				"position":      author.Position,
				"article_count": author.ArticleCount,
				"social_github": author.SocialGitHub,
				"created_at":    author.CreatedAt.Format("2006-01-02 15:04"),
			}
		}

		response := DataTablesResponse{
			Draw:            req.Draw,
			RecordsTotal:    totalRecords,
			RecordsFiltered: filteredRecords,
			Data:            data,
		}

		return c.JSON(http.StatusOK, response)
	}
}

// APIDataTableComments handles server-side processing for comments DataTable
func APIDataTableComments(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		var req DataTablesRequest
		if err := c.Bind(&req); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
		}

		if req.Length == 0 {
			req.Length = 10
		}
		if req.Length == -1 {
			req.Length = 100
		}

		query := quark.For[models.Comment](ctx, client.Client)

		if req.Search.Value != "" {
			query = query.Where("content", "LIKE", "%"+req.Search.Value+"%")
		}

		totalRecords, _ := quark.For[models.Comment](ctx, client.Client).Count()
		filteredRecords, _ := query.Count()

		query = query.OrderBy("created_at", "DESC")
		query = query.Limit(req.Length)
		query = query.Offset(req.Start)

		comments, err := query.List()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		data := make([]map[string]interface{}, len(comments))
		for i, comment := range comments {
			data[i] = map[string]interface{}{
				"id":           comment.ID,
				"article_id":   comment.ArticleID,
				"author_name":  comment.AuthorName,
				"author_email": comment.AuthorEmail,
				"content":      comment.Content,
				"approved":     comment.Approved,
				"created_at":   comment.CreatedAt.Format("2006-01-02 15:04"),
			}
		}

		response := DataTablesResponse{
			Draw:            req.Draw,
			RecordsTotal:    totalRecords,
			RecordsFiltered: filteredRecords,
			Data:            data,
		}

		return c.JSON(http.StatusOK, response)
	}
}

// QuarkPlaygroundPage renders the Quark ORM playground page
func QuarkPlaygroundPage(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// Get stats for the playground
		articleCount, _ := quark.For[models.Article](ctx, client.Client).Count()
		authorCount, _ := quark.For[models.Author](ctx, client.Client).Count()
		categoryCount, _ := quark.For[models.Category](ctx, client.Client).Count()
		tagCount, _ := quark.For[models.Tag](ctx, client.Client).Count()
		commentCount, _ := quark.For[models.Comment](ctx, client.Client).Count()

		stats := map[string]interface{}{
			"Articles":   articleCount,
			"Authors":    authorCount,
			"Categories": categoryCount,
			"Tags":       tagCount,
			"Comments":   commentCount,
		}

		return c.HTML(http.StatusOK, "quark_playground.html", map[string]interface{}{
			"Title": "Quark ORM Playground",
			"Stats": stats,
		})
	}
}

// QuarkDocsPage renders the Quark ORM documentation page
func QuarkDocsPage(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		return c.HTML(http.StatusOK, "quark_docs.html", map[string]interface{}{
			"Title": "Quark ORM Documentation",
		})
	}
}

// APIQuarkFirst demonstrates First() - get first record
func APIQuarkFirst(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// First() - get first published article
		article, err := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			OrderBy("published_at", "DESC").
			First()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"method":      "First()",
			"description": "Get the first record from the query",
			"result":      article,
		})
	}
}

// APIQuarkFind demonstrates Find() - find by ID
func APIQuarkFind(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()
		id := c.Request.PathValue("id")

		// Find() - find by ID
		article, err := quark.For[models.Article](ctx, client.Client).
			Where("id", "=", id).
			First()

		if err != nil {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "not found"})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"method":      "Find()",
			"description": "Find a record by ID",
			"result":      article,
		})
	}
}

// APIQuarkDelete demonstrates Delete() - delete records
func APIQuarkDelete(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()
		id := c.Request.PathValue("id")

		// First find the article
		articles, err := quark.For[models.Article](ctx, client.Client).
			Where("id", "=", id).
			List()

		if err != nil || len(articles) == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "article not found"})
		}

		// Delete() - delete by passing the article
		_, err = quark.For[models.Article](ctx, client.Client).Delete(&articles[0])

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"method":      "Delete()",
			"description": "Delete a record by ID",
			"result":      "deleted successfully",
		})
	}
}

// APIQuarkComplex demonstrates complex queries with Or/And
func APIQuarkComplex(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// Complex query with multiple conditions
		articles, err := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			Where("view_count", ">", 500).
			OrderBy("published_at", "DESC").
			Limit(5).
			List()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"method":      "Where() chained",
			"description": "Complex query with multiple Where conditions",
			"query":       "WHERE published = true AND view_count > 500 ORDER BY published_at DESC LIMIT 5",
			"result":      articles,
		})
	}
}

// APIQuarkSearch demonstrates Like() for text search
func APIQuarkSearch(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()
		query := c.Query("q")
		if query == "" {
			query = "Go"
		}

		// Like() - text search
		articles, err := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			Where("title", "LIKE", "%"+query+"%").
			OrderBy("published_at", "DESC").
			Limit(10).
			List()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"method":      "Where() with LIKE",
			"description": "Text search using LIKE operator",
			"query":       "WHERE published = true AND title LIKE '%" + query + "%'",
			"result":      articles,
		})
	}
}

// APIQuarkRange demonstrates Between() for range queries
func APIQuarkRange(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// Get articles from last 30 days
		thirtyDaysAgo := time.Now().AddDate(0, 0, -30)

		articles, err := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			Where("published_at", ">=", thirtyDaysAgo.Format("2006-01-02")).
			OrderBy("published_at", "DESC").
			List()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"method":      "Where() with date range",
			"description": "Query articles published in the last 30 days",
			"query":       "WHERE published = true AND published_at >= '" + thirtyDaysAgo.Format("2006-01-02") + "'",
			"result":      articles,
		})
	}
}

// APIQuarkAggregations demonstrates Sum, Avg, Min, Max
func APIQuarkAggregations(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// Count
		totalArticles, _ := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			Count()

		// Get all articles to calculate aggregations manually
		articles, _ := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			List()

		// Calculate aggregations
		totalViews := 0
		minViews := 0
		maxViews := 0
		if len(articles) > 0 {
			minViews = articles[0].ViewCount
			maxViews = articles[0].ViewCount
		}

		for _, art := range articles {
			totalViews += art.ViewCount
			if art.ViewCount < minViews {
				minViews = art.ViewCount
			}
			if art.ViewCount > maxViews {
				maxViews = art.ViewCount
			}
		}

		avgViews := 0
		if len(articles) > 0 {
			avgViews = totalViews / len(articles)
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"method":      "Aggregations",
			"description": "Count, Sum, Avg, Min, Max operations",
			"results": map[string]interface{}{
				"count":     totalArticles,
				"sum_views": totalViews,
				"avg_views": avgViews,
				"min_views": minViews,
				"max_views": maxViews,
			},
		})
	}
}

// APIQuarkTransaction demonstrates transactions
func APIQuarkTransaction(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// Create a test article and comment in a transaction
		// Note: This is a simplified example. Real transactions would use Quark's transaction API
		// if available, or manual transaction management with the underlying sql.DB

		now := time.Now().UTC()
		testArticle := models.Article{
			Title:       "Transaction Test Article",
			Slug:        "transaction-test-" + now.Format("20060102150405"),
			Summary:     "This article was created as part of a transaction test",
			Content:     "Transaction test content",
			Published:   false,
			PublishedAt: now,
			ViewCount:   0,
			AuthorID:    1,
			CategoryID:  1,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		// Create article
		if err := quark.For[models.Article](ctx, client.Client).Create(&testArticle); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create article: " + err.Error()})
		}

		// Create comment for the article
		testComment := models.Comment{
			ArticleID:   testArticle.ID,
			AuthorName:  "Transaction Test",
			AuthorEmail: "test@example.com",
			Content:     "This comment was created in the same transaction",
			Approved:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := quark.For[models.Comment](ctx, client.Client).Create(&testComment); err != nil {
			// In a real transaction, we would rollback here
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create comment: " + err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"method":      "Transaction",
			"description": "Create article and comment in a transaction (simplified - note: real transaction support depends on Quark's API)",
			"result": map[string]interface{}{
				"article_id": testArticle.ID,
				"comment_id": testComment.ID,
				"message":    "Created successfully (transaction simulation)",
			},
		})
	}
}

// APIQuarkPaginated demonstrates pagination with Offset and Limit
func APIQuarkPaginated(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		page := 1
		pageSize := 5

		// Get page from query
		if pageStr := c.Query("page"); pageStr != "" {
			if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
				page = p
			}
		}

		// Calculate offset
		offset := (page - 1) * pageSize

		// Get total count
		total, _ := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			Count()

		// Get paginated results
		articles, err := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			OrderBy("published_at", "DESC").
			Offset(offset).
			Limit(pageSize).
			List()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		// Calculate pagination metadata
		totalPages := (int(total) + pageSize - 1) / pageSize
		if totalPages == 0 {
			totalPages = 1
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"method":      "Pagination with Offset and Limit",
			"description": "Paginated results with Offset and Limit",
			"pagination": map[string]interface{}{
				"page":        page,
				"page_size":   pageSize,
				"total":       total,
				"total_pages": totalPages,
				"has_next":    page < totalPages,
				"has_prev":    page > 1,
			},
			"result": articles,
		})
	}
}

// APIQuarkRelations demonstrates relationships between models
func APIQuarkRelations(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// Get an article with its author and category
		articles, err := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			OrderBy("published_at", "DESC").
			Limit(1).
			List()

		if err != nil || len(articles) == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "no articles found"})
		}

		article := articles[0]

		// Get author
		authors, _ := quark.For[models.Author](ctx, client.Client).
			Where("id", "=", article.AuthorID).
			List()

		// Get category
		categories, _ := quark.For[models.Category](ctx, client.Client).
			Where("id", "=", article.CategoryID).
			List()

		// Get comments for this article
		comments, _ := quark.For[models.Comment](ctx, client.Client).
			Where("article_id", "=", article.ID).
			Where("approved", "=", true).
			List()

		result := map[string]interface{}{
			"article":  article,
			"author":   map[string]interface{}{},
			"category": map[string]interface{}{},
			"comments": comments,
		}

		if len(authors) > 0 {
			result["author"] = authors[0]
		}

		if len(categories) > 0 {
			result["category"] = categories[0]
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"method":      "Relations",
			"description": "Demonstrate relationships between Article, Author, Category, and Comment models",
			"result":      result,
		})
	}
}

// APIQuarkUpdateWhere demonstrates Update with Where clause
func APIQuarkUpdateWhere(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// Increment view count for all published articles
		// This demonstrates a bulk update operation
		articles, err := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			Limit(1).
			List()

		if err != nil || len(articles) == 0 {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "no articles found"})
		}

		article := articles[0]
		article.ViewCount++
		article.UpdatedAt = time.Now().UTC()

		_, err = quark.For[models.Article](ctx, client.Client).Update(&article)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"method":      "Update()",
			"description": "Update a single record",
			"result": map[string]interface{}{
				"article_id": article.ID,
				"view_count": article.ViewCount,
				"message":    "Updated successfully",
			},
		})
	}
}

// APIQuarkIn demonstrates In() operator
func APIQuarkIn(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		// Get articles from specific categories
		categoryIDs := []int64{1, 2, 3}

		articles, err := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			Where("category_id", "IN", categoryIDs).
			OrderBy("published_at", "DESC").
			Limit(10).
			List()

		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"method":      "Where() with IN",
			"description": "Query articles from specific categories using IN operator",
			"query":       "WHERE published = true AND category_id IN (1, 2, 3)",
			"result":      articles,
		})
	}
}

// APIQuarkAllFeatures demonstrates all Quark features in one response
func APIQuarkAllFeatures(client *quarkdb.Client) gfrender.Handler {
	return func(c *gfrender.Context) error {
		ctx := context.Background()

		features := make(map[string]interface{})

		// 1. Count
		count, _ := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			Count()
		features["count"] = map[string]interface{}{
			"description": "Count records",
			"result":      count,
		}

		// 2. First
		first, _ := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			OrderBy("published_at", "DESC").
			First()
		features["first"] = map[string]interface{}{
			"description": "Get first record",
			"result":      first,
		}

		// 3. List with Where and Limit
		list, _ := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			Where("view_count", ">", 0).
			OrderBy("published_at", "DESC").
			Limit(3).
			List()
		features["list"] = map[string]interface{}{
			"description": "List with Where and Limit",
			"result":      list,
		}

		// 4. OrderBy
		orderBy, _ := quark.For[models.Article](ctx, client.Client).
			Where("published", "=", true).
			OrderBy("view_count", "DESC").
			Limit(3).
			List()
		features["order_by"] = map[string]interface{}{
			"description": "Order by view_count DESC",
			"result":      orderBy,
		}

		return c.JSON(http.StatusOK, map[string]interface{}{
			"title":    "Quark ORM - All Features",
			"features": features,
		})
	}
}
