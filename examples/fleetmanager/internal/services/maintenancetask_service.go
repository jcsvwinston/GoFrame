package services

import (
	"context"
	"github.com/jcsvwinston/fleetmanager/internal/repositories"
)

type MaintenanceTaskService struct {
	repo *repositories.MaintenanceTaskRepository
}

func NewMaintenanceTaskService(repo *repositories.MaintenanceTaskRepository) *MaintenanceTaskService {
	return &MaintenanceTaskService{repo: repo}
}

func (s *MaintenanceTaskService) List(ctx context.Context) ([]repositories.MaintenanceTask, error) {
	return s.repo.List(ctx)
}

func (s *MaintenanceTaskService) Create(ctx context.Context, name string) (repositories.MaintenanceTask, error) {
	return s.repo.Create(ctx, name)
}

func (s *MaintenanceTaskService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
