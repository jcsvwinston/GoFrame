package repositories

import (
	"context"
	"errors"
	"sort"
	"strings"
	"sync"
	"time"
)

var ErrAuthorNotFound = errors.New("authors record not found")

type AuthorRecord struct {
	ID        uint      `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListAuthorParams struct {
	Query string
}

type CreateAuthorParams struct {
	Name string
}

type UpdateAuthorParams struct {
	Name string
}

type AuthorRepository struct {
	mu     sync.RWMutex
	nextID uint
	items  map[uint]AuthorRecord
}

func NewAuthorRepository() *AuthorRepository {
	return &AuthorRepository{
		nextID: 1,
		items:  make(map[uint]AuthorRecord),
	}
}

func (r *AuthorRepository) List(_ context.Context, params ListAuthorParams) ([]AuthorRecord, error) {
	r.mu.RLock()
	records := make([]AuthorRecord, 0, len(r.items))
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

func (r *AuthorRepository) Get(_ context.Context, id uint) (AuthorRecord, error) {
	r.mu.RLock()
	record, ok := r.items[id]
	r.mu.RUnlock()
	if !ok {
		return AuthorRecord{}, ErrAuthorNotFound
	}
	return record, nil
}

func (r *AuthorRepository) Create(_ context.Context, params CreateAuthorParams) (AuthorRecord, error) {
	now := time.Now().UTC()

	r.mu.Lock()
	id := r.nextID
	r.nextID++
	record := AuthorRecord{
		ID:        id,
		Name:      params.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.items[id] = record
	r.mu.Unlock()

	return record, nil
}

func (r *AuthorRepository) Update(_ context.Context, id uint, params UpdateAuthorParams) (AuthorRecord, error) {
	r.mu.Lock()
	record, ok := r.items[id]
	if !ok {
		r.mu.Unlock()
		return AuthorRecord{}, ErrAuthorNotFound
	}

	record.Name = params.Name
	record.UpdatedAt = time.Now().UTC()
	r.items[id] = record
	r.mu.Unlock()

	return record, nil
}

func (r *AuthorRepository) Delete(_ context.Context, id uint) error {
	r.mu.Lock()
	if _, ok := r.items[id]; !ok {
		r.mu.Unlock()
		return ErrAuthorNotFound
	}
	delete(r.items, id)
	r.mu.Unlock()
	return nil
}
