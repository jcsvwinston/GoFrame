package models
import (
	"time"
	"github.com/jcsvwinston/nucleus/pkg/model"
)
type Sensor struct {
	model.BaseModel
	DeviceID  int64      `db:"column:device_id;required;index" admin:"list,filter" fk:"Device"`
	Type      string     `db:"column:type;required;index" admin:"list,filter" choices:"gps,temp,fuel,engine"`
	Name      string     `db:"column:name;required" admin:"list,search"`
	DeletedAt *time.Time `db:"column:deleted_at" admin:"-"`
}
