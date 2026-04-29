package services

import (
	"context"
	"github.com/jcsvwinston/fleetmanager/internal/repositories"
)

type AlertService struct {
	repo *repositories.AlertRepository
}

func NewAlertService(repo *repositories.AlertRepository) *AlertService {
	return &AlertService{repo: repo}
}

func (s *AlertService) List(ctx context.Context) ([]repositories.Alert, error) {
	return s.repo.List(ctx)
}

func (s *AlertService) Create(ctx context.Context, name string) (repositories.Alert, error) {
	return s.repo.Create(ctx, name)
}

func (s *AlertService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
