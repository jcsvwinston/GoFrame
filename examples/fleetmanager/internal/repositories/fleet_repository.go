package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Fleet struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type FleetRepository struct {
	db *sql.DB
}

func NewFleetRepository(db *sql.DB) *FleetRepository {
	return &FleetRepository{db: db}
}

func (r *FleetRepository) List(ctx context.Context) ([]Fleet, error) {
	query := fmt.Sprintf("SELECT id, created_at, updated_at, name FROM %s WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 100", "fleets")
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()
	items := []Fleet{}
	for rows.Next() {
		var it Fleet
		if err := rows.Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt, &it.Name); err != nil { return nil, err }
		items = append(items, it)
	}
	return items, nil
}

func (r *FleetRepository) Create(ctx context.Context, name string) (Fleet, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf("INSERT INTO %s (created_at, updated_at, name) VALUES (?, ?, ?)", "fleets")
	res, err := r.db.ExecContext(ctx, query, now, now, name)
	if err != nil { return Fleet{}, err }
	id, _ := res.LastInsertId()
	return Fleet{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *FleetRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ?", "fleets")
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}
