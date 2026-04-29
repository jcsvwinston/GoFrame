package services

import (
	"context"
	"github.com/jcsvwinston/fleetmanager/internal/repositories"
)

type SensorService struct {
	repo *repositories.SensorRepository
}

func NewSensorService(repo *repositories.SensorRepository) *SensorService {
	return &SensorService{repo: repo}
}

func (s *SensorService) List(ctx context.Context) ([]repositories.Sensor, error) {
	return s.repo.List(ctx)
}

func (s *SensorService) Create(ctx context.Context, name string) (repositories.Sensor, error) {
	return s.repo.Create(ctx, name)
}

func (s *SensorService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
