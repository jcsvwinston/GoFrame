package models

import "time"

// Author represents a content author
type Author struct {
	ID            int64     `db:"id" pk:"true"`
	Name          string    `db:"name" validate:"required" admin:"list,search"`
	Email         string    `db:"email"`
	Bio           string    `db:"bio"`
	Position      string    `db:"position"`
	AvatarURL     string    `db:"avatar_url"`
	SocialGitHub  string    `db:"social_github"`
	SocialTwitter string    `db:"social_twitter"`
	ArticleCount  int       `db:"article_count"`
	CreatedAt     time.Time `db:"created_at"`
	UpdatedAt     time.Time `db:"updated_at"`
}
