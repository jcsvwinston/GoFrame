package services

import (
	"context"
	"github.com/jcsvwinston/fleetmanager/internal/repositories"
)

type FleetService struct {
	repo *repositories.FleetRepository
}

func NewFleetService(repo *repositories.FleetRepository) *FleetService {
	return &FleetService{repo: repo}
}

func (s *FleetService) List(ctx context.Context) ([]repositories.Fleet, error) {
	return s.repo.List(ctx)
}

func (s *FleetService) Create(ctx context.Context, name string) (repositories.Fleet, error) {
	return s.repo.Create(ctx, name)
}

func (s *FleetService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
