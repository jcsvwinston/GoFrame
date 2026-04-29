package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Trip struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TripRepository struct {
	db *sql.DB
}

func NewTripRepository(db *sql.DB) *TripRepository {
	return &TripRepository{db: db}
}

func (r *TripRepository) List(ctx context.Context) ([]Trip, error) {
	query := fmt.Sprintf("SELECT id, created_at, updated_at, name FROM %s WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 100", "trips")
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()
	items := []Trip{}
	for rows.Next() {
		var it Trip
		if err := rows.Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt, &it.Name); err != nil { return nil, err }
		items = append(items, it)
	}
	return items, nil
}

func (r *TripRepository) Create(ctx context.Context, name string) (Trip, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf("INSERT INTO %s (created_at, updated_at, name) VALUES (?, ?, ?)", "trips")
	res, err := r.db.ExecContext(ctx, query, now, now, name)
	if err != nil { return Trip{}, err }
	id, _ := res.LastInsertId()
	return Trip{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *TripRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ?", "trips")
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}
