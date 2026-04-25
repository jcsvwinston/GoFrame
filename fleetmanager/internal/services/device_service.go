package services

import (
	"context"
	"github.com/jcsvwinston/fleetmanager/internal/repositories"
)

type DeviceService struct {
	repo *repositories.DeviceRepository
}

func NewDeviceService(repo *repositories.DeviceRepository) *DeviceService {
	return &DeviceService{repo: repo}
}

func (s *DeviceService) List(ctx context.Context) ([]repositories.Device, error) {
	return s.repo.List(ctx)
}

func (s *DeviceService) Create(ctx context.Context, name string) (repositories.Device, error) {
	return s.repo.Create(ctx, name)
}

func (s *DeviceService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
