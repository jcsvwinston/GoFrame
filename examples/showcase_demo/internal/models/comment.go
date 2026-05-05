package models

import "time"

// Comment represents an article comment
type Comment struct {
	ID          int64     `db:"id" pk:"true"`
	ArticleID   int64     `db:"article_id"`
	AuthorName  string    `db:"author_name"`
	AuthorEmail string    `db:"author_email"`
	Content     string    `db:"content"`
	Approved    bool      `db:"approved"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
