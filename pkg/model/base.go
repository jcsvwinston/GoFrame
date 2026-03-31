// Package model provides the model registry, metadata extraction, and generic
// CRUD operations for the GoFrame framework. It uses reflection to extract
// struct metadata at registration time and supports dual SQL runtimes.
package model

import (
	"time"
)

// BaseModel provides the standard fields that all GoFrame models should embed.
// It is the equivalent of Django's models.Model.
type BaseModel struct {
	ID        uint       `gorm:"primaryKey" db:"primaryKey" json:"id"`
	CreatedAt time.Time  `gorm:"autoCreateTime" db:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time  `gorm:"autoUpdateTime" db:"autoUpdateTime" json:"updated_at"`
	DeletedAt *time.Time `gorm:"index" db:"index" json:"-"`
}
