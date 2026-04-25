package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Telemetry struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type TelemetryRepository struct {
	db *sql.DB
}

func NewTelemetryRepository(db *sql.DB) *TelemetryRepository {
	return &TelemetryRepository{db: db}
}

func (r *TelemetryRepository) List(ctx context.Context) ([]Telemetry, error) {
	query := fmt.Sprintf("SELECT id, created_at, updated_at, name FROM %s WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 100", "telemetries")
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()
	items := []Telemetry{}
	for rows.Next() {
		var it Telemetry
		if err := rows.Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt, &it.Name); err != nil { return nil, err }
		items = append(items, it)
	}
	return items, nil
}

func (r *TelemetryRepository) Create(ctx context.Context, name string) (Telemetry, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf("INSERT INTO %s (created_at, updated_at, name) VALUES (?, ?, ?)", "telemetries")
	res, err := r.db.ExecContext(ctx, query, now, now, name)
	if err != nil { return Telemetry{}, err }
	id, _ := res.LastInsertId()
	return Telemetry{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *TelemetryRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ?", "telemetries")
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}
