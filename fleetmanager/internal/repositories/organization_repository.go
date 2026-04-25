package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type Organization struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type OrganizationRepository struct {
	db *sql.DB
}

func NewOrganizationRepository(db *sql.DB) *OrganizationRepository {
	return &OrganizationRepository{db: db}
}

func (r *OrganizationRepository) List(ctx context.Context) ([]Organization, error) {
	query := fmt.Sprintf("SELECT id, created_at, updated_at, name FROM %s WHERE deleted_at IS NULL ORDER BY id DESC LIMIT 100", "organizations")
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil { return nil, err }
	defer rows.Close()
	items := []Organization{}
	for rows.Next() {
		var it Organization
		if err := rows.Scan(&it.ID, &it.CreatedAt, &it.UpdatedAt, &it.Name); err != nil { return nil, err }
		items = append(items, it)
	}
	return items, nil
}

func (r *OrganizationRepository) Create(ctx context.Context, name string) (Organization, error) {
	now := time.Now().UTC()
	query := fmt.Sprintf("INSERT INTO %s (created_at, updated_at, name) VALUES (?, ?, ?)", "organizations")
	res, err := r.db.ExecContext(ctx, query, now, now, name)
	if err != nil { return Organization{}, err }
	id, _ := res.LastInsertId()
	return Organization{ID: id, Name: name, CreatedAt: now, UpdatedAt: now}, nil
}

func (r *OrganizationRepository) Delete(ctx context.Context, id int64) error {
	now := time.Now().UTC()
	query := fmt.Sprintf("UPDATE %s SET deleted_at = ? WHERE id = ?", "organizations")
	_, err := r.db.ExecContext(ctx, query, now, id)
	return err
}
