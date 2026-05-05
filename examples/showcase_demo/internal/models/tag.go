package models

import "time"

// Tag represents a content tag
type Tag struct {
	ID           int64     `db:"id" pk:"true"`
	Name         string    `db:"name" validate:"required" admin:"list,search"`
	Slug         string    `db:"slug"`
	Color        string    `db:"color"`
	ArticleCount int       `db:"article_count"`
	CreatedAt    time.Time `db:"created_at"`
	UpdatedAt    time.Time `db:"updated_at"`
}
