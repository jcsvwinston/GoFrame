package models
import (
	"time"
	"github.com/jcsvwinston/GoFrame/pkg/model"
)
type Telemetry struct {
	model.BaseModel
	SensorID  int64      `db:"column:sensor_id;required;index" admin:"list,filter" fk:"Sensor"`
	Value     float64    `db:"column:value;required" admin:"list"`
	Unit      string     `db:"column:unit" admin:"list"`
	Timestamp time.Time  `db:"column:timestamp;required;index" admin:"list,filter"`
	DeletedAt *time.Time `db:"column:deleted_at" admin:"-"`
}
