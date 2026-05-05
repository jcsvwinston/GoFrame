package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"

	"example.com/showcase_clean/internal/controllers"
	"example.com/showcase_clean/internal/models"
	"example.com/showcase_clean/internal/quarkdb"
	gfrender "github.com/jcsvwinston/GoFrame/pkg/router"
)

func main() {
	ctx := context.Background()

	// 1. Initialize Quark ORM Client
	client, err := quarkdb.NewClient("app.db")
	if err != nil {
		log.Fatalf("Failed to initialize Quark client: %v", err)
	}
	defer client.Close()

	// 2. Auto-Migrate using Quark ORM
	fmt.Println("🚀 Running Quark ORM migrations...")
	if err := client.Migrate(ctx,
		&models.Author{},
		&models.Category{},
		&models.Tag{},
		&models.Article{},
		&models.Comment{},
	); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}
	fmt.Println("✅ Migrations completed")

	// 3. Seed data using Quark ORM
	fmt.Println("🌱 Seeding data with Quark ORM...")
	if err := client.Seed(ctx); err != nil {
		log.Printf("Seeding failed: %v", err)
	} else {
		fmt.Println("✅ Seeding completed")
	}

	// 4. Setup Router with Quark-powered Controllers
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	router := gfrender.New(logger)

	// Static Files
	router.Static("/static", "./internal/web/static")

	// Public Web Pages - Using Quark-powered controllers
	router.Get("/", controllers.HomePageQuark(client))
	router.Get("/blog", controllers.BlogPageQuark(client))
	router.Get("/blog/{slug}", controllers.ArticlePageQuark(client))
	router.Get("/categories", controllers.CategoriesPageQuark(client))
	router.Get("/about", controllers.AboutPageQuark(client))
	router.Get("/contact", controllers.ContactPageQuark(nil))
	router.Post("/contact", controllers.ContactSubmitQuark(client))

	// API Endpoints - Using Quark ORM
	router.Get("/api/health", controllers.Health)
	router.Get("/api/stats", controllers.GetStatsAPIQuark(client))
	router.Get("/api/categories", controllers.ListCategoriesAPIQuark(client))
	router.Get("/api/authors", controllers.ListAuthorsAPIQuark(client))
	router.Get("/api/articles/{slug}", controllers.GetArticleBySlugQuark(client))

	// 5. Start Server
	fmt.Println("🌐 Showcase Demo with Quark ORM is ready!")
	fmt.Println("   URL: http://localhost:8080")
	fmt.Println("   Using: github.com/jcsvwinston/quark ORM")

	if err := http.ListenAndServe(":8080", router); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
