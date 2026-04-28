package services

import (
	"context"
	"strings"

	"example.com/showcase_clean/internal/repositories"
)

type CommentRecord struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type ListCommentInput struct {
	Query string
}

type CreateCommentInput struct {
	Name string `json:"name" validate:"required"`
}

type UpdateCommentInput struct {
	Name string `json:"name" validate:"required"`
}

type CommentRepository interface {
	List(ctx context.Context, params repositories.ListCommentParams) ([]repositories.CommentRecord, error)
	Get(ctx context.Context, id uint) (repositories.CommentRecord, error)
	Create(ctx context.Context, params repositories.CreateCommentParams) (repositories.CommentRecord, error)
	Update(ctx context.Context, id uint, params repositories.UpdateCommentParams) (repositories.CommentRecord, error)
	Delete(ctx context.Context, id uint) error
}

type CommentService struct {
	repository CommentRepository
}

func NewCommentService(repository CommentRepository) *CommentService {
	return &CommentService{repository: repository}
}

func (s *CommentService) List(ctx context.Context, input ListCommentInput) ([]CommentRecord, error) {
	records, err := s.repository.List(ctx, repositories.ListCommentParams{
		Query: strings.TrimSpace(input.Query),
	})
	if err != nil {
		return nil, err
	}

	items := make([]CommentRecord, 0, len(records))
	for _, record := range records {
		items = append(items, mapCommentRecord(record))
	}
	return items, nil
}

func (s *CommentService) Get(ctx context.Context, id uint) (CommentRecord, error) {
	record, err := s.repository.Get(ctx, id)
	if err != nil {
		return CommentRecord{}, err
	}
	return mapCommentRecord(record), nil
}

func (s *CommentService) Create(ctx context.Context, input CreateCommentInput) (CommentRecord, error) {
	record, err := s.repository.Create(ctx, repositories.CreateCommentParams{
		Name: strings.TrimSpace(input.Name),
	})
	if err != nil {
		return CommentRecord{}, err
	}
	return mapCommentRecord(record), nil
}

func (s *CommentService) Update(ctx context.Context, id uint, input UpdateCommentInput) (CommentRecord, error) {
	record, err := s.repository.Update(ctx, id, repositories.UpdateCommentParams{
		Name: strings.TrimSpace(input.Name),
	})
	if err != nil {
		return CommentRecord{}, err
	}
	return mapCommentRecord(record), nil
}

func (s *CommentService) Delete(ctx context.Context, id uint) error {
	return s.repository.Delete(ctx, id)
}

func mapCommentRecord(record repositories.CommentRecord) CommentRecord {
	return CommentRecord{
		ID:   record.ID,
		Name: record.Name,
	}
}
