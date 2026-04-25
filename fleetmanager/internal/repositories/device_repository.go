package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Device struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DeviceRepository struct {
	db *sql.DB
}

func NewDeviceRepository(db *sql.DB) *DeviceRepository {
	return &DeviceRepository{db: db}
}

func (r *DeviceRepository) List(ctx context.Context) ([]Device, error) {
	query := fmt.Sprintf("SELECT id, created_at, updated_at, name FROM %s WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 100", "devices")
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()
	items := []Device{}
	for rows.Next() {
		var it Device
		if err := rows.Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt, &it.Name); err != nil { return nil, err }
		items = append(items, it)
	}
	return items, nil
}

func (r *DeviceRepository) Create(ctx context.Context, name string) (Device, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf("INSERT INTO %s (created_at, updated_at, name) VALUES (?, ?, ?)", "devices")
	res, err := r.db.ExecContext(ctx, query, now, now, name)
	if err != nil { return Device{}, err }
	id, _ := res.LastInsertId()
	return Device{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *DeviceRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ?", "devices")
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}
