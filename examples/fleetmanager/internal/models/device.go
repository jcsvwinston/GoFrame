package models
import (
	"time"
	"github.com/jcsvwinston/GoFrame/pkg/model"
)
type Device struct {
	model.BaseModel
	FleetID   int64      `db:"column:fleet_id;required;index" admin:"list,filter" fk:"Fleet"`
	Name      string     `db:"column:name;required;index" admin:"list,search,filter"`
	Serial    string     `db:"column:serial;required;unique" admin:"list,search"`
	DeletedAt *time.Time `db:"column:deleted_at" admin:"-"`
}
