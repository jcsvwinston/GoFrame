package services

import (
	"context"
	"github.com/jcsvwinston/fleetmanager/internal/repositories"
)

type TripService struct {
	repo *repositories.TripRepository
}

func NewTripService(repo *repositories.TripRepository) *TripService {
	return &TripService{repo: repo}
}

func (s *TripService) List(ctx context.Context) ([]repositories.Trip, error) {
	return s.repo.List(ctx)
}

func (s *TripService) Create(ctx context.Context, name string) (repositories.Trip, error) {
	return s.repo.Create(ctx, name)
}

func (s *TripService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
