package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Geofence struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GeofenceRepository struct {
	db *sql.DB
}

func NewGeofenceRepository(db *sql.DB) *GeofenceRepository {
	return &GeofenceRepository{db: db}
}

func (r *GeofenceRepository) List(ctx context.Context) ([]Geofence, error) {
	query := fmt.Sprintf("SELECT id, created_at, updated_at, name FROM %s WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 100", "geofences")
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()
	items := []Geofence{}
	for rows.Next() {
		var it Geofence
		if err := rows.Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt, &it.Name); err != nil { return nil, err }
		items = append(items, it)
	}
	return items, nil
}

func (r *GeofenceRepository) Create(ctx context.Context, name string) (Geofence, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf("INSERT INTO %s (created_at, updated_at, name) VALUES (?, ?, ?)", "geofences")
	res, err := r.db.ExecContext(ctx, query, now, now, name)
	if err != nil { return Geofence{}, err }
	id, _ := res.LastInsertId()
	return Geofence{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *GeofenceRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ?", "geofences")
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}
