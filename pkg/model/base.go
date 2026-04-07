// Package model provides the model registry, metadata extraction, and generic
// CRUD operations for the GoFrame framework. It uses reflection to extract
// struct metadata at registration time and uses the native SQL runtime.
package model

import (
	"time"
)

// BaseModel provides the standard fields that all GoFrame models should embed.
// It is the equivalent of Django's models.Model.
type BaseModel struct {
	ID        uint       `db:"pk" json:"id"`
	CreatedAt time.Time  `db:"readonly" json:"created_at"`
	UpdatedAt time.Time  `db:"readonly" json:"updated_at"`
	DeletedAt *time.Time `db:"column:deleted_at" json:"-"`
}
