package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Sensor struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SensorRepository struct {
	db *sql.DB
}

func NewSensorRepository(db *sql.DB) *SensorRepository {
	return &SensorRepository{db: db}
}

func (r *SensorRepository) List(ctx context.Context) ([]Sensor, error) {
	query := fmt.Sprintf("SELECT id, created_at, updated_at, name FROM %s WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 100", "sensors")
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()
	items := []Sensor{}
	for rows.Next() {
		var it Sensor
		if err := rows.Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt, &it.Name); err != nil { return nil, err }
		items = append(items, it)
	}
	return items, nil
}

func (r *SensorRepository) Create(ctx context.Context, name string) (Sensor, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf("INSERT INTO %s (created_at, updated_at, name) VALUES (?, ?, ?)", "sensors")
	res, err := r.db.ExecContext(ctx, query, now, now, name)
	if err != nil { return Sensor{}, err }
	id, _ := res.LastInsertId()
	return Sensor{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *SensorRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ?", "sensors")
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}
