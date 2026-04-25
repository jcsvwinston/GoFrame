package services

import (
	"context"
	"github.com/jcsvwinston/fleetmanager/internal/repositories"
)

type OrganizationService struct {
	repo *repositories.OrganizationRepository
}

func NewOrganizationService(repo *repositories.OrganizationRepository) *OrganizationService {
	return &OrganizationService{repo: repo}
}

func (s *OrganizationService) List(ctx context.Context) ([]repositories.Organization, error) {
	return s.repo.List(ctx)
}

func (s *OrganizationService) Create(ctx context.Context, name string) (repositories.Organization, error) {
	return s.repo.Create(ctx, name)
}

func (s *OrganizationService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
