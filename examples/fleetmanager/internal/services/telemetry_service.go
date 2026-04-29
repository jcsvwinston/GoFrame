package services

import (
	"context"
	"github.com/jcsvwinston/fleetmanager/internal/repositories"
)

type TelemetryService struct {
	repo *repositories.TelemetryRepository
}

func NewTelemetryService(repo *repositories.TelemetryRepository) *TelemetryService {
	return &TelemetryService{repo: repo}
}

func (s *TelemetryService) List(ctx context.Context) ([]repositories.Telemetry, error) {
	return s.repo.List(ctx)
}

func (s *TelemetryService) Create(ctx context.Context, name string) (repositories.Telemetry, error) {
	return s.repo.Create(ctx, name)
}

func (s *TelemetryService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
