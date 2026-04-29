package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Asset struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type AssetRepository struct {
	db *sql.DB
}

func NewAssetRepository(db *sql.DB) *AssetRepository {
	return &AssetRepository{db: db}
}

func (r *AssetRepository) List(ctx context.Context) ([]Asset, error) {
	query := fmt.Sprintf("SELECT id, created_at, updated_at, name FROM %s WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 100", "assets")
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()
	items := []Asset{}
	for rows.Next() {
		var it Asset
		if err := rows.Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt, &it.Name); err != nil { return nil, err }
		items = append(items, it)
	}
	return items, nil
}

func (r *AssetRepository) Create(ctx context.Context, name string) (Asset, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf("INSERT INTO %s (created_at, updated_at, name) VALUES (?, ?, ?)", "assets")
	res, err := r.db.ExecContext(ctx, query, now, now, name)
	if err != nil { return Asset{}, err }
	id, _ := res.LastInsertId()
	return Asset{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *AssetRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ?", "assets")
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}
