package services

import (
	"context"
	"github.com/jcsvwinston/fleetmanager/internal/repositories"
)

type AssetService struct {
	repo *repositories.AssetRepository
}

func NewAssetService(repo *repositories.AssetRepository) *AssetService {
	return &AssetService{repo: repo}
}

func (s *AssetService) List(ctx context.Context) ([]repositories.Asset, error) {
	return s.repo.List(ctx)
}

func (s *AssetService) Create(ctx context.Context, name string) (repositories.Asset, error) {
	return s.repo.Create(ctx, name)
}

func (s *AssetService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
