package main

import (
	"context"
	"log"

	"example.com/showcase_clean/internal/repositories"
	"example.com/showcase_clean/internal/services"
	projecttasks "example.com/showcase_clean/internal/tasks"
	"github.com/jcsvwinston/GoFrame/pkg/app"
	gftasks "github.com/jcsvwinston/GoFrame/pkg/tasks"
	asynqprovider "github.com/jcsvwinston/GoFrame/pkg/tasks/providers/asynq"
)

func main() {
	cfg, err := app.LoadConfig("goframe.yaml")
	if err != nil {
		log.Fatal(err)
	}
	if cfg.RedisURL == "" {
		log.Fatal("redis_url is required to run worker")
	}

	a, err := app.New(cfg)
	if err != nil {
		log.Fatal(err)
	}

	sqlDB, err := a.DB.SqlDB()
	if err != nil {
		log.Fatal(err)
	}

	articleRepository := repositories.NewArticleRepository(sqlDB)
	articleService := services.NewArticleService(articleRepository)

	manager, err := asynqprovider.NewManager(gftasks.Config{
		RedisURL:    cfg.RedisURL,
		Concurrency: 10,
		Queues:      map[string]int{"default": 1},
	}, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer manager.Close()

	if err := projecttasks.Register(manager, articleService); err != nil {
		log.Fatal(err)
	}

	log.Println("Worker listening for background tasks")
	if err := manager.Run(context.Background()); err != nil {
		log.Fatal(err)
	}
}
