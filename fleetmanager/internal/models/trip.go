package models
import (
	"time"
	"github.com/jcsvwinston/GoFrame/pkg/model"
)
type Trip struct {
	model.BaseModel
	AssetID   int64      `db:"column:asset_id;required;index" admin:"list,filter" fk:"Asset"`
	DriverID  int64      `db:"column:driver_id;required;index" admin:"list,filter" fk:"Driver"`
	StartTime time.Time  `db:"column:start_time;required" admin:"list,filter"`
	EndTime   *time.Time `db:"column:end_time" admin:"list"`
	Distance  float64    `db:"column:distance" admin:"list"`
	DeletedAt *time.Time `db:"column:deleted_at" admin:"-"`
}
