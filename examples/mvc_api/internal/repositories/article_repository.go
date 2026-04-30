package repositories

import (
	"context"
	"database/sql"

	"github.com/jcsvwinston/GoFrame/examples/mvc_api/internal/dtos"
)

type ArticleRepository struct {
	db *sql.DB
}

func NewArticleRepository(db *sql.DB) *ArticleRepository {
	return &ArticleRepository{db: db}
}

func (r *ArticleRepository) List(ctx context.Context, publishedOnly bool, limit int) ([]dtos.ArticleDTO, error) {
	query := `SELECT id, title, content, published, created_at, updated_at FROM articles`
	args := make([]any, 0, 1)
	if publishedOnly {
		query += ` WHERE published = 1`
	}
	query += ` ORDER BY created_at DESC`
	if limit > 0 {
		query += ` LIMIT ?`
		args = append(args, limit)
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []dtos.ArticleDTO
	for rows.Next() {
		var item dtos.ArticleDTO
		if err := rows.Scan(&item.ID, &item.Title, &item.Content, &item.Published, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
