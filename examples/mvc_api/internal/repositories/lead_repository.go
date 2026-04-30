package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/jcsvwinston/GoFrame/examples/mvc_api/internal/dtos"
)

type LeadRepository struct {
	db *sql.DB
}

func NewLeadRepository(db *sql.DB) *LeadRepository {
	return &LeadRepository{db: db}
}

func (r *LeadRepository) List(ctx context.Context, limit int) ([]dtos.LeadDTO, error) {
	if limit <= 0 {
		limit = 100
	}
	query := `SELECT id, name, email, company, wants_demo, created_at, updated_at FROM leads ORDER BY created_at DESC LIMIT ?`
	rows, err := r.db.QueryContext(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []dtos.LeadDTO
	for rows.Next() {
		var item dtos.LeadDTO
		if err := rows.Scan(&item.ID, &item.Name, &item.Email, &item.Company, &item.WantsDemo, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *LeadRepository) Create(ctx context.Context, in dtos.CreateLeadInput) (dtos.LeadDTO, error) {
	now := time.Now().UTC()
	res, err := r.db.ExecContext(
		ctx,
		`INSERT INTO leads (created_at, updated_at, name, email, company, wants_demo) VALUES (?, ?, ?, ?, ?, ?)`,
		now, now, in.Name, in.Email, in.Company, in.WantsDemo,
	)
	if err != nil {
		return dtos.LeadDTO{}, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return dtos.LeadDTO{}, err
	}

	return dtos.LeadDTO{
		ID:        id,
		Name:      in.Name,
		Email:     in.Email,
		Company:   in.Company,
		WantsDemo: in.WantsDemo,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func (r *LeadRepository) Count(ctx context.Context) int {
	var count int
	if err := r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM leads").Scan(&count); err != nil {
		return 0
	}
	return count
}
