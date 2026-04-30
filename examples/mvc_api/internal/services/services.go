package services

import (
	"context"
	"database/sql"
	"time"

	"github.com/jcsvwinston/GoFrame/examples/mvc_api/internal/dtos"
	"github.com/jcsvwinston/GoFrame/examples/mvc_api/internal/repositories"
	"github.com/jcsvwinston/GoFrame/pkg/app"
	"github.com/jcsvwinston/GoFrame/pkg/outbox"
	"github.com/jcsvwinston/GoFrame/pkg/tasks"
	asynqprovider "github.com/jcsvwinston/GoFrame/pkg/tasks/providers/asynq"
)

type Services struct {
	SQLDB            *sql.DB
	OutboxStore      *outbox.Store
	OutboxDispatcher *outbox.Dispatcher
	TaskManager      tasks.Manager
	Scheduler        *asynqprovider.Scheduler
	ArticleRepo      *repositories.ArticleRepository
	LeadRepo         *repositories.LeadRepository
}

func New(a *app.App) (*Services, error) {
	sqlDB, err := a.DB.SqlDB()
	if err != nil {
		return nil, err
	}

	svc := &Services{
		SQLDB:       sqlDB,
		ArticleRepo: repositories.NewArticleRepository(sqlDB),
		LeadRepo:    repositories.NewLeadRepository(sqlDB),
	}

	if err := svc.ensureSchema(); err != nil {
		return nil, err
	}
	if err := svc.ensureSeed(); err != nil {
		return nil, err
	}

	return svc, nil
}

func (s *Services) ensureSchema() error {
	stmts := []string{
		`CREATE TABLE IF NOT EXISTS articles (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			title TEXT NOT NULL,
			content TEXT,
			published BOOLEAN NOT NULL DEFAULT 0
		)`,
		`CREATE TABLE IF NOT EXISTS leads (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			created_at DATETIME,
			updated_at DATETIME,
			deleted_at DATETIME,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			company TEXT,
			wants_demo BOOLEAN NOT NULL DEFAULT 0
		)`,
	}
	for _, stmt := range stmts {
		if _, err := s.SQLDB.Exec(stmt); err != nil {
			return err
		}
	}
	return nil
}

func (s *Services) ensureSeed() error {
	var articleCount int
	if err := s.SQLDB.QueryRow(`SELECT COUNT(*) FROM articles`).Scan(&articleCount); err != nil {
		return err
	}
	if articleCount > 0 {
		return nil
	}

	now := time.Now().UTC()
	_, err := s.SQLDB.Exec(
		"INSERT INTO articles (created_at, updated_at, title, content, published) VALUES (?, ?, ?, ?, ?)",
		now, now, "Welcome to GoFrame", "This record is editable from /admin and visible via /api/articles.", true,
	)
	return err
}

func (s *Services) CountRows(table string) int {
	var count int
	if err := s.SQLDB.QueryRow("SELECT COUNT(*) FROM " + table).Scan(&count); err != nil {
		return 0
	}
	return count
}

func (s *Services) ListArticles(ctx context.Context, publishedOnly bool, limit int) ([]dtos.ArticleDTO, error) {
	return s.ArticleRepo.List(ctx, publishedOnly, limit)
}

func (s *Services) ListLeads(ctx context.Context, limit int) ([]dtos.LeadDTO, error) {
	return s.LeadRepo.List(ctx, limit)
}

func (s *Services) CreateLead(ctx context.Context, in dtos.CreateLeadInput) (dtos.LeadDTO, error) {
	return s.LeadRepo.Create(ctx, in)
}
