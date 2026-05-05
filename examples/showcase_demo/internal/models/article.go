package models

import "time"

// Article represents a blog article in the showcase demo
type Article struct {
	ID          int64     `db:"id" pk:"true"`
	Title       string    `db:"title" validate:"required,min=3" admin:"list,search"`
	Slug        string    `db:"slug" validate:"required"`
	Summary     string    `db:"summary"`
	Content     string    `db:"content" admin:"list"`
	Published   bool      `db:"published" admin:"list,filter"`
	PublishedAt time.Time `db:"published_at"`
	ViewCount   int       `db:"view_count"`
	AuthorID    int64     `db:"author_id"`
	CategoryID  int64     `db:"category_id"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
