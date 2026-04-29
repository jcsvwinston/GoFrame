package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Alert struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AlertRepository struct {
	db *sql.DB
}

func NewAlertRepository(db *sql.DB) *AlertRepository {
	return &AlertRepository{db: db}
}

func (r *AlertRepository) List(ctx context.Context) ([]Alert, error) {
	query := fmt.Sprintf("SELECT id, created_at, updated_at, name FROM %s WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 100", "alerts")
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()
	items := []Alert{}
	for rows.Next() {
		var it Alert
		if err := rows.Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt, &it.Name); err != nil { return nil, err }
		items = append(items, it)
	}
	return items, nil
}

func (r *AlertRepository) Create(ctx context.Context, name string) (Alert, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf("INSERT INTO %s (created_at, updated_at, name) VALUES (?, ?, ?)", "alerts")
	res, err := r.db.ExecContext(ctx, query, now, now, name)
	if err != nil { return Alert{}, err }
	id, _ := res.LastInsertId()
	return Alert{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *AlertRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ?", "alerts")
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}
