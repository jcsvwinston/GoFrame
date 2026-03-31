package model

import (
	"context"

	"github.com/uptrace/bun"
	"gorm.io/gorm"
)

const (
	HookEngineGORM = "gorm"
	HookEngineBun  = "bun"
)

// HookContext exposes runtime information to lifecycle hooks in an
// engine-agnostic way.
type HookContext struct {
	Context context.Context
	Engine  string
	GORM    *gorm.DB
	Bun     bun.IDB
}

// HookFunc is the signature for model lifecycle hooks.
type HookFunc func(ctx HookContext, entity interface{}) error

func newGORMHookContext(ctx context.Context, db *gorm.DB) HookContext {
	return HookContext{
		Context: ctx,
		Engine:  HookEngineGORM,
		GORM:    db,
	}
}

func newBunHookContext(ctx context.Context, db bun.IDB) HookContext {
	return HookContext{
		Context: ctx,
		Engine:  HookEngineBun,
		Bun:     db,
	}
}
