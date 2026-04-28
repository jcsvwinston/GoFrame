package repositories

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrCategoryNotFound = errors.New("categories record not found")

type CategoryRecord struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListCategoryParams struct {
	Query string
}

type CreateCategoryParams struct {
	Name string
}

type UpdateCategoryParams struct {
	Name string
}

type CategoryRepository struct {
	mu     sync.RWMutex
	nextID uint
	items  map[uint]CategoryRecord
}

func NewCategoryRepository() *CategoryRepository {
	return &CategoryRepository{
		nextID: 1,
		items:  make(map[uint]CategoryRecord),
	}
}

func (r *CategoryRepository) List(_ context.Context, params ListCategoryParams) ([]CategoryRecord, error) {
	r.mu.RLock()
	records := make([]CategoryRecord, 0, len(r.items))
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

func (r *CategoryRepository) Get(_ context.Context, id uint) (CategoryRecord, error) {
	r.mu.RLock()
	record, ok := r.items[id]
	r.mu.RUnlock()
	if !ok {
		return CategoryRecord{}, ErrCategoryNotFound
	}
	return record, nil
}

func (r *CategoryRepository) Create(_ context.Context, params CreateCategoryParams) (CategoryRecord, error) {
	now := time.Now().UTC()

	r.mu.Lock()
	id := r.nextID
	r.nextID++
	record := CategoryRecord{
		ID:        id,
		Name:      params.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.items[id] = record
	r.mu.Unlock()

	return record, nil
}

func (r *CategoryRepository) Update(_ context.Context, id uint, params UpdateCategoryParams) (CategoryRecord, error) {
	r.mu.Lock()
	record, ok := r.items[id]
	if !ok {
		r.mu.Unlock()
		return CategoryRecord{}, ErrCategoryNotFound
	}

	record.Name = params.Name
	record.UpdatedAt = time.Now().UTC()
	r.items[id] = record
	r.mu.Unlock()

	return record, nil
}

func (r *CategoryRepository) Delete(_ context.Context, id uint) error {
	r.mu.Lock()
	if _, ok := r.items[id]; !ok {
		r.mu.Unlock()
		return ErrCategoryNotFound
	}
	delete(r.items, id)
	r.mu.Unlock()
	return nil
}
