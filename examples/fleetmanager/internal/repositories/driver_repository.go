package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Driver struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DriverRepository struct {
	db *sql.DB
}

func NewDriverRepository(db *sql.DB) *DriverRepository {
	return &DriverRepository{db: db}
}

func (r *DriverRepository) List(ctx context.Context) ([]Driver, error) {
	query := fmt.Sprintf("SELECT id, created_at, updated_at, name FROM %s WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 100", "drivers")
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()
	items := []Driver{}
	for rows.Next() {
		var it Driver
		if err := rows.Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt, &it.Name); err != nil { return nil, err }
		items = append(items, it)
	}
	return items, nil
}

func (r *DriverRepository) Create(ctx context.Context, name string) (Driver, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf("INSERT INTO %s (created_at, updated_at, name) VALUES (?, ?, ?)", "drivers")
	res, err := r.db.ExecContext(ctx, query, now, now, name)
	if err != nil { return Driver{}, err }
	id, _ := res.LastInsertId()
	return Driver{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *DriverRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ?", "drivers")
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}
