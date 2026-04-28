package services

import (
	"context"
	"strings"

	"example.com/showcase_clean/internal/repositories"
)

type AuthorRecord struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ListAuthorInput struct {
	Query string
}

type CreateAuthorInput struct {
	Name string `json:"name" validate:"required"`
}

type UpdateAuthorInput struct {
	Name string `json:"name" validate:"required"`
}

type AuthorRepository interface {
	List(ctx context.Context, params repositories.ListAuthorParams) ([]repositories.AuthorRecord, error)
	Get(ctx context.Context, id uint) (repositories.AuthorRecord, error)
	Create(ctx context.Context, params repositories.CreateAuthorParams) (repositories.AuthorRecord, error)
	Update(ctx context.Context, id uint, params repositories.UpdateAuthorParams) (repositories.AuthorRecord, error)
	Delete(ctx context.Context, id uint) error
}

type AuthorService struct {
	repository AuthorRepository
}

func NewAuthorService(repository AuthorRepository) *AuthorService {
	return &AuthorService{repository: repository}
}

func (s *AuthorService) List(ctx context.Context, input ListAuthorInput) ([]AuthorRecord, error) {
	records, err := s.repository.List(ctx, repositories.ListAuthorParams{
		Query: strings.TrimSpace(input.Query),
	})
	if err != nil {
		return nil, err
	}

	items := make([]AuthorRecord, 0, len(records))
	for _, record := range records {
		items = append(items, mapAuthorRecord(record))
	}
	return items, nil
}

func (s *AuthorService) Get(ctx context.Context, id uint) (AuthorRecord, error) {
	record, err := s.repository.Get(ctx, id)
	if err != nil {
		return AuthorRecord{}, err
	}
	return mapAuthorRecord(record), nil
}

func (s *AuthorService) Create(ctx context.Context, input CreateAuthorInput) (AuthorRecord, error) {
	record, err := s.repository.Create(ctx, repositories.CreateAuthorParams{
		Name: strings.TrimSpace(input.Name),
	})
	if err != nil {
		return AuthorRecord{}, err
	}
	return mapAuthorRecord(record), nil
}

func (s *AuthorService) Update(ctx context.Context, id uint, input UpdateAuthorInput) (AuthorRecord, error) {
	record, err := s.repository.Update(ctx, id, repositories.UpdateAuthorParams{
		Name: strings.TrimSpace(input.Name),
	})
	if err != nil {
		return AuthorRecord{}, err
	}
	return mapAuthorRecord(record), nil
}

func (s *AuthorService) Delete(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}

func mapAuthorRecord(record repositories.AuthorRecord) AuthorRecord {
	return AuthorRecord{
		ID:   record.ID,
		Name: record.Name,
	}
}
