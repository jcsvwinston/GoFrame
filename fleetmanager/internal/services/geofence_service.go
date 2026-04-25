package services

import (
	"context"
	"github.com/jcsvwinston/fleetmanager/internal/repositories"
)

type GeofenceService struct {
	repo *repositories.GeofenceRepository
}

func NewGeofenceService(repo *repositories.GeofenceRepository) *GeofenceService {
	return &GeofenceService{repo: repo}
}

func (s *GeofenceService) List(ctx context.Context) ([]repositories.Geofence, error) {
	return s.repo.List(ctx)
}

func (s *GeofenceService) Create(ctx context.Context, name string) (repositories.Geofence, error) {
	return s.repo.Create(ctx, name)
}

func (s *GeofenceService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
