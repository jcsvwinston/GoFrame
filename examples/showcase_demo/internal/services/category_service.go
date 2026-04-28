package services

import (
	"context"
	"strings"

	"example.com/showcase_clean/internal/repositories"
)

type CategoryRecord struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ListCategoryInput struct {
	Query string
}

type CreateCategoryInput struct {
	Name string `json:"name" validate:"required"`
}

type UpdateCategoryInput struct {
	Name string `json:"name" validate:"required"`
}

type CategoryRepository interface {
	List(ctx context.Context, params repositories.ListCategoryParams) ([]repositories.CategoryRecord, error)
	Get(ctx context.Context, id uint) (repositories.CategoryRecord, error)
	Create(ctx context.Context, params repositories.CreateCategoryParams) (repositories.CategoryRecord, error)
	Update(ctx context.Context, id uint, params repositories.UpdateCategoryParams) (repositories.CategoryRecord, error)
	Delete(ctx context.Context, id uint) error
}

type CategoryService struct {
	repository CategoryRepository
}

func NewCategoryService(repository CategoryRepository) *CategoryService {
	return &CategoryService{repository: repository}
}

func (s *CategoryService) List(ctx context.Context, input ListCategoryInput) ([]CategoryRecord, error) {
	records, err := s.repository.List(ctx, repositories.ListCategoryParams{
		Query: strings.TrimSpace(input.Query),
	})
	if err != nil {
		return nil, err
	}

	items := make([]CategoryRecord, 0, len(records))
	for _, record := range records {
		items = append(items, mapCategoryRecord(record))
	}
	return items, nil
}

func (s *CategoryService) Get(ctx context.Context, id uint) (CategoryRecord, error) {
	record, err := s.repository.Get(ctx, id)
	if err != nil {
		return CategoryRecord{}, err
	}
	return mapCategoryRecord(record), nil
}

func (s *CategoryService) Create(ctx context.Context, input CreateCategoryInput) (CategoryRecord, error) {
	record, err := s.repository.Create(ctx, repositories.CreateCategoryParams{
		Name: strings.TrimSpace(input.Name),
	})
	if err != nil {
		return CategoryRecord{}, err
	}
	return mapCategoryRecord(record), nil
}

func (s *CategoryService) Update(ctx context.Context, id uint, input UpdateCategoryInput) (CategoryRecord, error) {
	record, err := s.repository.Update(ctx, id, repositories.UpdateCategoryParams{
		Name: strings.TrimSpace(input.Name),
	})
	if err != nil {
		return CategoryRecord{}, err
	}
	return mapCategoryRecord(record), nil
}

func (s *CategoryService) Delete(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}

func mapCategoryRecord(record repositories.CategoryRecord) CategoryRecord {
	return CategoryRecord{
		ID:   record.ID,
		Name: record.Name,
	}
}
