package services

import (
	"context"
	"strings"

	"example.com/showcase_clean/internal/repositories"
)

type TagRecord struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ListTagInput struct {
	Query string
}

type CreateTagInput struct {
	Name string `json:"name" validate:"required"`
}

type UpdateTagInput struct {
	Name string `json:"name" validate:"required"`
}

type TagRepository interface {
	List(ctx context.Context, params repositories.ListTagParams) ([]repositories.TagRecord, error)
	Get(ctx context.Context, id uint) (repositories.TagRecord, error)
	Create(ctx context.Context, params repositories.CreateTagParams) (repositories.TagRecord, error)
	Update(ctx context.Context, id uint, params repositories.UpdateTagParams) (repositories.TagRecord, error)
	Delete(ctx context.Context, id uint) error
}

type TagService struct {
	repository TagRepository
}

func NewTagService(repository TagRepository) *TagService {
	return &TagService{repository: repository}
}

func (s *TagService) List(ctx context.Context, input ListTagInput) ([]TagRecord, error) {
	records, err := s.repository.List(ctx, repositories.ListTagParams{
		Query: strings.TrimSpace(input.Query),
	})
	if err != nil {
		return nil, err
	}

	items := make([]TagRecord, 0, len(records))
	for _, record := range records {
		items = append(items, mapTagRecord(record))
	}
	return items, nil
}

func (s *TagService) Get(ctx context.Context, id uint) (TagRecord, error) {
	record, err := s.repository.Get(ctx, id)
	if err != nil {
		return TagRecord{}, err
	}
	return mapTagRecord(record), nil
}

func (s *TagService) Create(ctx context.Context, input CreateTagInput) (TagRecord, error) {
	record, err := s.repository.Create(ctx, repositories.CreateTagParams{
		Name: strings.TrimSpace(input.Name),
	})
	if err != nil {
		return TagRecord{}, err
	}
	return mapTagRecord(record), nil
}

func (s *TagService) Update(ctx context.Context, id uint, input UpdateTagInput) (TagRecord, error) {
	record, err := s.repository.Update(ctx, id, repositories.UpdateTagParams{
		Name: strings.TrimSpace(input.Name),
	})
	if err != nil {
		return TagRecord{}, err
	}
	return mapTagRecord(record), nil
}

func (s *TagService) Delete(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}

func mapTagRecord(record repositories.TagRecord) TagRecord {
	return TagRecord{
		ID:   record.ID,
		Name: record.Name,
	}
}
