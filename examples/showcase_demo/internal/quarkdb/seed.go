package quarkdb

import (
	"context"
	"fmt"
	"math/rand"
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

	c.logger.Info("seed_start", "message", "Starting seed with large dataset")

	now := time.Now().UTC()

	// Create 50 Authors
	firstNames := []string{"María", "Carlos", "Ana", "Pedro", "Laura", "Miguel", "Sofia", "Diego", "Elena", "Ricardo", "Carmen", "Javier", "Isabel", "Fernando", "Patricia", "Luis", "Marta", "Antonio", "Adriana", "Francisco", "Lucía", "Daniel", "Paula", "Ángel", "Beatriz", "Sergio", "Natalia", "Roberto", "Cristina", "Alejandro", "Valentina", "David", "Sara", "Raúl", "Julia", "Manuel", "Clara", "Iván", "Andrea", "Jorge", "Silvia", "Víctor", "Carla", "Eduardo", "Mónica", "Óscar", "Nuria", "Adrián", "Teresa", "Pablo"}
	lastNames := []string{"García", "Ruiz", "Martínez", "Sánchez", "López", "Gómez", "Fernández", "Pérez", "González", "Rodríguez", "Morales", "Torres", "Jiménez", "Navarro", "Ramos", "Serrano", "Blanco", "Castro", "Ortega", "Alonso", "Vargas", "Castillo", "Cruz", "Mendoza", "Herrera", "Flores", "Guerrero", "Romero", "Vargas", "Molina", "Peña", "Cabrera", "Rosales", "Reyes", "Soto", "Paredes", "Ríos", "Cortés", "Méndez", "Ávila", "Vega", "Lara", "Santana", "Miranda", "Benítez", "Campos", "Cordero", "Pineda"}

	positions := []string{"Lead Developer", "Senior Backend", "Frontend Developer", "DevOps Engineer", "Full-stack Developer", "Tech Lead", "Software Architect", "Data Engineer", "Security Engineer", "Mobile Developer"}

	authorIDs := make([]int64, 50)
	for i := 0; i < 50; i++ {
		firstName := firstNames[rand.Intn(len(firstNames))]
		lastName := lastNames[rand.Intn(len(lastNames))]
		position := positions[rand.Intn(len(positions))]

		author := models.Author{
			Name:          fmt.Sprintf("%s %s", firstName, lastName),
			Email:         fmt.Sprintf("%s.%s%d@example.com", firstName, lastName, i),
			Bio:           fmt.Sprintf("%s con %d años de experiencia en desarrollo de software.", position, rand.Intn(15)+1),
			Position:      position,
			AvatarURL:     fmt.Sprintf("/static/images/authors/author%d.jpg", i),
			SocialGitHub:  fmt.Sprintf("%s%s%d", firstName, lastName, i),
			SocialTwitter: fmt.Sprintf("@%s%s", firstName, lastName),
			ArticleCount:  rand.Intn(50),
			CreatedAt:     now,
			UpdatedAt:     now,
		}

		if err := quark.For[models.Author](ctx, c.Client).Create(&author); err != nil {
			return err
		}
		authorIDs[i] = author.ID
	}

	c.logger.Info("seed_authors_complete", "count", 50)

	// Create 20 Categories
	categoryNames := []string{"Tecnología", "Tutoriales", "Opinión", "Noticias", "Recursos", "DevOps", "Backend", "Frontend", "Cloud", "Seguridad", "Data Science", "Machine Learning", "Mobile", "Testing", "Arquitectura", "Performance", "Microservicios", "Databases", "API Design", "Open Source"}
	categorySlugs := []string{"tecnologia", "tutoriales", "opinion", "noticias", "recursos", "devops", "backend", "frontend", "cloud", "seguridad", "data-science", "machine-learning", "mobile", "testing", "arquitectura", "performance", "microservicios", "databases", "api-design", "open-source"}
	colors := []string{"#3b82f6", "#10b981", "#f59e0b", "#ef4444", "#8b5cf6", "#06b6d4", "#ec4899", "#84cc16", "#f97316", "#6366f1", "#14b8a6", "#a855f7", "#0ea5e9", "#22c55e", "#e11d48", "#8b5cf6", "#f43f5e", "#64748b", "#0d9488", "#7c3aed"}
	icons := []string{"code", "book-open", "message-circle", "newspaper", "package", "server", "database", "layout", "cloud", "shield", "brain", "cpu", "smartphone", "check-circle", "layers", "zap", "grid", "hard-drive", "api", "github"}

	categoryIDs := make([]int64, 20)
	for i := 0; i < 20; i++ {
		category := models.Category{
			Name:        categoryNames[i],
			Slug:        categorySlugs[i],
			Description: fmt.Sprintf("Artículos sobre %s y desarrollo.", categoryNames[i]),
			Color:       colors[i],
			Icon:        icons[i],
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if err := quark.For[models.Category](ctx, c.Client).Create(&category); err != nil {
			return err
		}
		categoryIDs[i] = category.ID
	}

	c.logger.Info("seed_categories_complete", "count", 20)

	// Create 30 Tags
	tagNames := []string{"Go", "React", "TypeScript", "Docker", "Kubernetes", "PostgreSQL", "API", "Microservicios", "Testing", "DevOps", "GraphQL", "Redis", "MongoDB", "AWS", "GCP", "Azure", "CI/CD", "Terraform", "Prometheus", "Grafana", "Linux", "Git", "Jenkins", "Nginx", "RabbitMQ", "Kafka", "Elasticsearch", "Logstash", "Kibana", "Vue.js"}
	tagSlugs := []string{"go", "react", "typescript", "docker", "kubernetes", "postgresql", "api", "microservicios", "testing", "devops", "graphql", "redis", "mongodb", "aws", "gcp", "azure", "ci-cd", "terraform", "prometheus", "grafana", "linux", "git", "jenkins", "nginx", "rabbitmq", "kafka", "elasticsearch", "logstash", "kibana", "vuejs"}
	tagColors := []string{"#00add8", "#61dafb", "#3178c6", "#2496ed", "#326ce5", "#336791", "#ff6b6b", "#4ecdc4", "#95e1d3", "#f38181", "#e10098", "#dc3812", "#4db33d", "#ff9900", "#4285f4", "#0089d6", "#ee2725", "#7b42bc", "#e6522c", "#f46800", "#fcc624", "#f05032", "#d24739", "#009639", "#ff6600", "#23a9f2", "#fec514", "#d63031", "#61dafb", "#42b883"}

	tagIDs := make([]int64, 30)
	for i := 0; i < 30; i++ {
		tag := models.Tag{
			Name:      tagNames[i],
			Slug:      tagSlugs[i],
			Color:     tagColors[i],
			CreatedAt: now,
			UpdatedAt: now,
		}

		if err := quark.For[models.Tag](ctx, c.Client).Create(&tag); err != nil {
			return err
		}
		tagIDs[i] = tag.ID
	}

	c.logger.Info("seed_tags_complete", "count", 30)

	// Create 1000 Articles with realistic data
	articleTitles := []string{
		"Introducción a GoFrame", "Guía Completa de Docker", "Patrones de Diseño en Go", "Kubernetes: Orquestación", "PostgreSQL: Optimización",
		"Testing en Go", "API RESTful: Mejores Prácticas", "Microservicios con Go", "React Hooks Avanzados", "TypeScript Tips",
		"CI/CD con Jenkins", "Terraform para Infraestructura", "Monitoring con Prometheus", "Grafana Dashboards", "Redis Caching Strategies",
		"MongoDB vs PostgreSQL", "AWS Lambda Functions", "GCP Cloud Run", "Azure Container Instances", "Vue.js Composition API",
		"GraphQL Schema Design", "RabbitMQ Message Queues", "Kafka Event Streaming", "Elasticsearch Search", "Logstash Pipelines",
		"Kibana Visualizations", "Linux Performance Tuning", "Git Best Practices", "Nginx Configuration", "Security Headers",
		"OAuth 2.0 Implementation", "JWT Authentication", "Rate Limiting Strategies", "Database Sharding", "Connection Pooling",
		"SQL Index Optimization", "Query Performance Analysis", "Slow Query Logging", "Database Backup Strategies", "Disaster Recovery",
		"Container Security", "Image Hardening", "Vulnerability Scanning", "Secrets Management", "Zero Trust Architecture",
		"Network Security", "DDoS Protection", "WAF Configuration", "SSL/TLS Best Practices", "Certificate Management",
		"Load Balancing", "CDN Optimization", "Edge Computing", "Serverless Architecture", "Event-Driven Design",
		"CQRS Pattern", "Event Sourcing", "Saga Pattern", "Circuit Breaker", "Bulkhead Pattern",
		"Retry Mechanisms", "Timeout Handling", "Graceful Degradation", "Circuit Testing", "Chaos Engineering",
	}

	articleContents := []string{
		"GoFrame es un framework MVC completo para Go que proporciona una estructura sólida para aplicaciones web modernas. Con su arquitectura modular y sistema de rutas potente, GoFrame facilita el desarrollo de aplicaciones escalables y mantenibles.",
		"Docker ha revolucionado la forma en que desarrollamos y desplegamos aplicaciones. Esta guía completa cubre desde los conceptos básicos hasta técnicas avanzadas de optimización de imágenes y orquestación de contenedores.",
		"Go ofrece un enfoque único para los patrones de diseño debido a su simplicidad y composición. Este artículo explora cómo implementar patrones comunes de forma idiomática en Go.",
		"Kubernetes es la plataforma de orquestación de contenedores más popular. Aprende a desplegar aplicaciones Go en Kubernetes con Deployments, Services, ConfigMaps y Secrets.",
		"PostgreSQL es una base de datos poderosa. Descubre técnicas avanzadas de optimización de consultas, uso eficiente de índices y configuración del pool de conexiones.",
	}

	for i := 0; i < 1000; i++ {
		titleIndex := i % len(articleTitles)
		slugIndex := i % len(tagSlugs)
		title := fmt.Sprintf("%s - Parte %d", articleTitles[titleIndex], i/len(articleTitles)+1)
		slug := fmt.Sprintf("%s-parte-%d", tagSlugs[slugIndex], i/len(articleTitles)+1)

		// Random date in last 2 years
		daysAgo := rand.Intn(730)
		publishedAt := now.AddDate(0, 0, -daysAgo)

		article := models.Article{
			Title:       title,
			Slug:        slug,
			Summary:     fmt.Sprintf("Este artículo explora %s en profundidad, cubriendo conceptos clave y mejores prácticas.", articleTitles[titleIndex]),
			Content:     articleContents[rand.Intn(len(articleContents))],
			Published:   rand.Intn(10) > 2, // 70% published
			PublishedAt: publishedAt,
			ViewCount:   rand.Intn(10000),
			AuthorID:    authorIDs[rand.Intn(len(authorIDs))],
			CategoryID:  categoryIDs[rand.Intn(len(categoryIDs))],
			CreatedAt:   publishedAt,
			UpdatedAt:   publishedAt,
		}

		if err := quark.For[models.Article](ctx, c.Client).Create(&article); err != nil {
			return err
		}

		if i > 0 && i%100 == 0 {
			c.logger.Info("seed_articles_progress", "count", i)
		}
	}

	c.logger.Info("seed_articles_complete", "count", 1000)

	// Create 5000 Comments
	commentContents := []string{
		"Excelente artículo, me ayudó mucho a entender el tema!",
		"¿Tienen planes para cubrir más sobre esto en futuros artículos?",
		"Los ejemplos son muy claros. ¡Gracias por compartir!",
		"Me gustaría ver más contenido sobre este tema específico.",
		"Esta es exactamente la información que estaba buscando.",
		"Gran explicación, pero ¿podrías profundizar más en X?",
		"Me encanta la claridad de tus explicaciones. Sigue así!",
		"¿Recomiendas algún recurso adicional para aprender más?",
		"Este artículo me ahorró horas de investigación. ¡Gracias!",
		"¿Podrías hacer un tutorial más avanzado sobre esto?",
	}

	for i := 0; i < 5000; i++ {
		// Get a random article ID (articles 100-999 to ensure they exist)
		articleID := int64(rand.Intn(900) + 100)

		firstName := firstNames[rand.Intn(len(firstNames))]
		lastName := lastNames[rand.Intn(len(lastNames))]

		comment := models.Comment{
			ArticleID:   articleID,
			AuthorName:  fmt.Sprintf("%s %s", firstName, lastName),
			AuthorEmail: fmt.Sprintf("%s.%s%d@example.com", firstName, lastName, i),
			Content:     commentContents[rand.Intn(len(commentContents))],
			Approved:    rand.Intn(10) > 3, // 70% approved
			CreatedAt:   now.AddDate(0, 0, -rand.Intn(365)),
			UpdatedAt:   now.AddDate(0, 0, -rand.Intn(365)),
		}

		if err := quark.For[models.Comment](ctx, c.Client).Create(&comment); err != nil {
			return err
		}

		if i > 0 && i%500 == 0 {
			c.logger.Info("seed_comments_progress", "count", i)
		}
	}

	c.logger.Info("seed_comments_complete", "count", 5000)
	c.logger.Info("seed_complete", "message", "Seed completed successfully")

	return nil
}
