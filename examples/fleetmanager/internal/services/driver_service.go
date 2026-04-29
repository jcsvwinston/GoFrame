package services

import (
	"context"
	"github.com/jcsvwinston/fleetmanager/internal/repositories"
)

type DriverService struct {
	repo *repositories.DriverRepository
}

func NewDriverService(repo *repositories.DriverRepository) *DriverService {
	return &DriverService{repo: repo}
}

func (s *DriverService) List(ctx context.Context) ([]repositories.Driver, error) {
	return s.repo.List(ctx)
}

func (s *DriverService) Create(ctx context.Context, name string) (repositories.Driver, error) {
	return s.repo.Create(ctx, name)
}

func (s *DriverService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
