package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type MaintenanceTask struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MaintenanceTaskRepository struct {
	db *sql.DB
}

func NewMaintenanceTaskRepository(db *sql.DB) *MaintenanceTaskRepository {
	return &MaintenanceTaskRepository{db: db}
}

func (r *MaintenanceTaskRepository) List(ctx context.Context) ([]MaintenanceTask, error) {
	query := fmt.Sprintf("SELECT id, created_at, updated_at, name FROM %s WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 100", "maintenancetasks")
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()
	items := []MaintenanceTask{}
	for rows.Next() {
		var it MaintenanceTask
		if err := rows.Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt, &it.Name); err != nil { return nil, err }
		items = append(items, it)
	}
	return items, nil
}

func (r *MaintenanceTaskRepository) Create(ctx context.Context, name string) (MaintenanceTask, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf("INSERT INTO %s (created_at, updated_at, name) VALUES (?, ?, ?)", "maintenancetasks")
	res, err := r.db.ExecContext(ctx, query, now, now, name)
	if err != nil { return MaintenanceTask{}, err }
	id, _ := res.LastInsertId()
	return MaintenanceTask{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *MaintenanceTaskRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ?", "maintenancetasks")
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}
