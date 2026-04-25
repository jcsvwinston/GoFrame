package models
import (
	"time"
	"github.com/jcsvwinston/GoFrame/pkg/model"
)
type Geofence struct {
	model.BaseModel
	FleetID   int64      `db:"column:fleet_id;required;index" admin:"list,filter" fk:"Fleet"`
	Name      string     `db:"column:name;required" admin:"list,search"`
	Area      string     `db:"column:area;required" admin:"-"` // WKT or GeoJSON
	DeletedAt *time.Time `db:"column:deleted_at" admin:"-"`
}
