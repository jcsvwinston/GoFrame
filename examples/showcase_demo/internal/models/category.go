package models

import "time"

// Category represents an article category
type Category struct {
	ID           int64     `db:"id" pk:"true"`
	Name         string    `db:"name" validate:"required" admin:"list,search"`
	Slug         string    `db:"slug"`
	Description  string    `db:"description"`
	Color        string    `db:"color"`
	Icon         string    `db:"icon"`
	ArticleCount int       `db:"article_count"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
