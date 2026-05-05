package quarkdb

import (
	"context"
	"time"

	"example.com/showcase_clean/internal/models"
	"github.com/jcsvwinston/quark"
)

// Seed populates the database with initial data using Quark ORM
func (c *Client) Seed(ctx context.Context) error {
	// Check if already seeded
	count, err := quark.For[models.Article](ctx, c.Client).Count()
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	now := time.Now().UTC()

	// Create Authors
	authors := []models.Author{
		{Name: "María García", Email: "maria@example.com", Bio: "Full-stack developer con 10 años de experiencia.", Position: "Lead Developer", AvatarURL: "/static/images/authors/maria.jpg", SocialGitHub: "mariagarcia", SocialTwitter: "mariagarcia", CreatedAt: now, UpdatedAt: now},
		{Name: "Carlos Ruiz", Email: "carlos@example.com", Bio: "DevOps engineer especializado en Kubernetes.", Position: "DevOps Lead", AvatarURL: "/static/images/authors/carlos.jpg", SocialGitHub: "cruiz", SocialTwitter: "cruizdev", CreatedAt: now, UpdatedAt: now},
		{Name: "Ana Martínez", Email: "ana@example.com", Bio: "UX/UI Designer con enfoque en accesibilidad.", Position: "Lead Designer", AvatarURL: "/static/images/authors/ana.jpg", SocialGitHub: "anamtz", SocialTwitter: "anadesigns", CreatedAt: now, UpdatedAt: now},
		{Name: "Pedro Sánchez", Email: "pedro@example.com", Bio: "Backend developer especializado en APIs.", Position: "Senior Backend", AvatarURL: "/static/images/authors/pedro.jpg", SocialGitHub: "pedrosan", SocialTwitter: "pedrosan", CreatedAt: now, UpdatedAt: now},
	}

	authorIDs := make([]int64, len(authors))
	for i := range authors {
		if err := quark.For[models.Author](ctx, c.Client).Create(&authors[i]); err != nil {
			return err
		}
		authorIDs[i] = authors[i].ID
	}

	// Create Categories
	categories := []models.Category{
		{Name: "Tecnología", Slug: "tecnologia", Description: "Artículos sobre tecnología y desarrollo.", Color: "#3b82f6", Icon: "code", CreatedAt: now, UpdatedAt: now},
		{Name: "Tutoriales", Slug: "tutoriales", Description: "Guías paso a paso.", Color: "#10b981", Icon: "book-open", CreatedAt: now, UpdatedAt: now},
		{Name: "Opinión", Slug: "opinion", Description: "Reflexiones sobre el mundo tech.", Color: "#f59e0b", Icon: "message-circle", CreatedAt: now, UpdatedAt: now},
		{Name: "Noticias", Slug: "noticias", Description: "Últimas novedades.", Color: "#ef4444", Icon: "newspaper", CreatedAt: now, UpdatedAt: now},
		{Name: "Recursos", Slug: "recursos", Description: "Herramientas y librerías.", Color: "#8b5cf6", Icon: "package", CreatedAt: now, UpdatedAt: now},
	}

	categoryIDs := make([]int64, len(categories))
	for i := range categories {
		if err := quark.For[models.Category](ctx, c.Client).Create(&categories[i]); err != nil {
			return err
		}
		categoryIDs[i] = categories[i].ID
	}

	// Create Tags
	tags := []models.Tag{
		{Name: "Go", Slug: "go", Color: "#00add8", CreatedAt: now, UpdatedAt: now},
		{Name: "React", Slug: "react", Color: "#61dafb", CreatedAt: now, UpdatedAt: now},
		{Name: "TypeScript", Slug: "typescript", Color: "#3178c6", CreatedAt: now, UpdatedAt: now},
		{Name: "Docker", Slug: "docker", Color: "#2496ed", CreatedAt: now, UpdatedAt: now},
		{Name: "Kubernetes", Slug: "kubernetes", Color: "#326ce5", CreatedAt: now, UpdatedAt: now},
		{Name: "PostgreSQL", Slug: "postgresql", Color: "#336791", CreatedAt: now, UpdatedAt: now},
		{Name: "API", Slug: "api", Color: "#ff6b6b", CreatedAt: now, UpdatedAt: now},
		{Name: "Microservicios", Slug: "microservicios", Color: "#4ecdc4", CreatedAt: now, UpdatedAt: now},
		{Name: "Testing", Slug: "testing", Color: "#95e1d3", CreatedAt: now, UpdatedAt: now},
		{Name: "DevOps", Slug: "devops", Color: "#f38181", CreatedAt: now, UpdatedAt: now},
	}

	tagIDs := make([]int64, len(tags))
	for i := range tags {
		if err := quark.For[models.Tag](ctx, c.Client).Create(&tags[i]); err != nil {
			return err
		}
		tagIDs[i] = tags[i].ID
	}

	// Create Articles
	articles := []models.Article{
		{Title: "Introducción a GoFrame", Slug: "introduccion-goframe", Summary: "Descubre GoFrame, el framework MVC moderno.", Content: "GoFrame es un framework MVC completo para Go...", Published: true, PublishedAt: now, ViewCount: 1250, AuthorID: authorIDs[0], CategoryID: categoryIDs[0], CreatedAt: now, UpdatedAt: now},
		{Title: "Guía Completa de Docker", Slug: "guia-docker", Summary: "Aprende a containerizar tus aplicaciones.", Content: "Docker es una plataforma que permite desarrollar...", Published: true, PublishedAt: now, ViewCount: 890, AuthorID: authorIDs[1], CategoryID: categoryIDs[1], CreatedAt: now, UpdatedAt: now},
		{Title: "Patrones de Diseño en Go", Slug: "patrones-diseno", Summary: "Implementa patrones de forma idiomática.", Content: "Go favorece la composición sobre la herencia...", Published: true, PublishedAt: now, ViewCount: 650, AuthorID: authorIDs[0], CategoryID: categoryIDs[0], CreatedAt: now, UpdatedAt: now},
		{Title: "Kubernetes: Orquestación", Slug: "kubernetes", Summary: "Guía práctica para desplegar aplicaciones.", Content: "Kubernetes es una plataforma de orquestación...", Published: true, PublishedAt: now, ViewCount: 720, AuthorID: authorIDs[1], CategoryID: categoryIDs[0], CreatedAt: now, UpdatedAt: now},
		{Title: "PostgreSQL: Optimización", Slug: "postgresql", Summary: "Técnicas avanzadas de optimización.", Content: "Los índices son cruciales para el rendimiento...", Published: true, PublishedAt: now, ViewCount: 540, AuthorID: authorIDs[3], CategoryID: categoryIDs[0], CreatedAt: now, UpdatedAt: now},
		{Title: "Testing en Go", Slug: "testing-go", Summary: "Estrategias completas para testing.", Content: "Go tiene testing built-in...", Published: true, PublishedAt: now, ViewCount: 480, AuthorID: authorIDs[2], CategoryID: categoryIDs[1], CreatedAt: now, UpdatedAt: now},
		{Title: "API RESTful: Mejores Prácticas", Slug: "api-restful", Summary: "Diseña APIs robustas y documentadas.", Content: "Incluye la versión en la URL...", Published: true, PublishedAt: now, ViewCount: 620, AuthorID: authorIDs[0], CategoryID: categoryIDs[2], CreatedAt: now, UpdatedAt: now},
	}

	for i := range articles {
		if err := quark.For[models.Article](ctx, c.Client).Create(&articles[i]); err != nil {
			return err
		}
	}

	// Create Comments
	comments := []models.Comment{
		{ArticleID: articles[0].ID, AuthorName: "Juan Pérez", AuthorEmail: "juan@example.com", Content: "Excelente artículo, me ayudó mucho!", Approved: true, CreatedAt: now, UpdatedAt: now},
		{ArticleID: articles[0].ID, AuthorName: "Laura Gómez", AuthorEmail: "laura@example.com", Content: "¿Tienen planes para soportar GraphQL?", Approved: true, CreatedAt: now, UpdatedAt: now},
		{ArticleID: articles[1].ID, AuthorName: "Miguel Torres", AuthorEmail: "miguel@example.com", Content: "Docker ha cambiado mi workflow. Gran guía.", Approved: true, CreatedAt: now, UpdatedAt: now},
		{ArticleID: articles[2].ID, AuthorName: "Sofia López", AuthorEmail: "sofia@example.com", Content: "Los ejemplos son muy claros. Me encanta Go!", Approved: true, CreatedAt: now, UpdatedAt: now},
		{ArticleID: articles[3].ID, AuthorName: "Diego Martín", AuthorEmail: "diego@example.com", Content: "Kubernetes es complejo pero vale la pena.", Approved: true, CreatedAt: now, UpdatedAt: now},
	}

	for i := range comments {
		if err := quark.For[models.Comment](ctx, c.Client).Create(&comments[i]); err != nil {
			return err
		}
	}

	return nil
}
