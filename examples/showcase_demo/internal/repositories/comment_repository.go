package repositories

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrCommentNotFound = errors.New("comments record not found")

type CommentRecord struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListCommentParams struct {
	Query string
}

type CreateCommentParams struct {
	Name string
}

type UpdateCommentParams struct {
	Name string
}

type CommentRepository struct {
	mu     sync.RWMutex
	nextID uint
	items  map[uint]CommentRecord
}

func NewCommentRepository() *CommentRepository {
	return &CommentRepository{
		nextID: 1,
		items:  make(map[uint]CommentRecord),
	}
}

func (r *CommentRepository) List(_ context.Context, params ListCommentParams) ([]CommentRecord, error) {
	r.mu.RLock()
	records := make([]CommentRecord, 0, len(r.items))
	query := strings.ToLower(strings.TrimSpace(params.Query))
	for _, record := range r.items {
		if query != "" && !strings.Contains(strings.ToLower(record.Name), query) {
			continue
		}
		records = append(records, record)
	}
	r.mu.RUnlock()

	sort.Slice(records, func(i, j int) bool {
		return records[i].ID < records[j].ID
	})
	return records, nil
}

func (r *CommentRepository) Get(_ context.Context, id uint) (CommentRecord, error) {
	r.mu.RLock()
	record, ok := r.items[id]
	r.mu.RUnlock()
	if !ok {
		return CommentRecord{}, ErrCommentNotFound
	}
	return record, nil
}

func (r *CommentRepository) Create(_ context.Context, params CreateCommentParams) (CommentRecord, error) {
	now := time.Now().UTC()

	r.mu.Lock()
	id := r.nextID
	r.nextID++
	record := CommentRecord{
		ID:        id,
		Name:      params.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.items[id] = record
	r.mu.Unlock()

	return record, nil
}

func (r *CommentRepository) Update(_ context.Context, id uint, params UpdateCommentParams) (CommentRecord, error) {
	r.mu.Lock()
	record, ok := r.items[id]
	if !ok {
		r.mu.Unlock()
		return CommentRecord{}, ErrCommentNotFound
	}

	record.Name = params.Name
	record.UpdatedAt = time.Now().UTC()
	r.items[id] = record
	r.mu.Unlock()

	return record, nil
}

func (r *CommentRepository) Delete(_ context.Context, id uint) error {
	r.mu.Lock()
	if _, ok := r.items[id]; !ok {
		r.mu.Unlock()
		return ErrCommentNotFound
	}
	delete(r.items, id)
	r.mu.Unlock()
	return nil
}
