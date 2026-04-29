package models
import (
	"time"
	"github.com/jcsvwinston/GoFrame/pkg/model"
)
type Organization struct {
	model.BaseModel
	Name      string     `db:"column:name;required;index" admin:"list,search,filter"`
	Slug      string     `db:"column:slug;required;unique" admin:"list,search"`
	DeletedAt *time.Time `db:"column:deleted_at" admin:"-"`
}
