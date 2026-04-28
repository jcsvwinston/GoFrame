package repositories

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrTagNotFound = errors.New("tags record not found")

type TagRecord struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListTagParams struct {
	Query string
}

type CreateTagParams struct {
	Name string
}

type UpdateTagParams struct {
	Name string
}

type TagRepository struct {
	mu     sync.RWMutex
	nextID uint
	items  map[uint]TagRecord
}

func NewTagRepository() *TagRepository {
	return &TagRepository{
		nextID: 1,
		items:  make(map[uint]TagRecord),
	}
}

func (r *TagRepository) List(_ context.Context, params ListTagParams) ([]TagRecord, error) {
	r.mu.RLock()
	records := make([]TagRecord, 0, len(r.items))
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

func (r *TagRepository) Get(_ context.Context, id uint) (TagRecord, error) {
	r.mu.RLock()
	record, ok := r.items[id]
	r.mu.RUnlock()
	if !ok {
		return TagRecord{}, ErrTagNotFound
	}
	return record, nil
}

func (r *TagRepository) Create(_ context.Context, params CreateTagParams) (TagRecord, error) {
	now := time.Now().UTC()

	r.mu.Lock()
	id := r.nextID
	r.nextID++
	record := TagRecord{
		ID:        id,
		Name:      params.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.items[id] = record
	r.mu.Unlock()

	return record, nil
}

func (r *TagRepository) Update(_ context.Context, id uint, params UpdateTagParams) (TagRecord, error) {
	r.mu.Lock()
	record, ok := r.items[id]
	if !ok {
		r.mu.Unlock()
		return TagRecord{}, ErrTagNotFound
	}

	record.Name = params.Name
	record.UpdatedAt = time.Now().UTC()
	r.items[id] = record
	r.mu.Unlock()

	return record, nil
}

func (r *TagRepository) Delete(_ context.Context, id uint) error {
	r.mu.Lock()
	if _, ok := r.items[id]; !ok {
		r.mu.Unlock()
		return ErrTagNotFound
	}
	delete(r.items, id)
	r.mu.Unlock()
	return nil
}
