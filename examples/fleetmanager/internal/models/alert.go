package models
import (
	"time"
	"github.com/jcsvwinston/GoFrame/pkg/model"
)
type Alert struct {
	model.BaseModel
	FleetID   int64      `db:"column:fleet_id;index" admin:"list,filter" fk:"Fleet"`
	Type      string     `db:"column:type;required;index" admin:"list,filter" choices:"speeding,geofence,maintenance,sensor"`
	Severity  string     `db:"column:severity;required;index" admin:"list,filter" choices:"low,medium,high,critical"`
	Message   string     `db:"column:message;required" admin:"list,search"`
	Resolved  bool       `db:"column:resolved;required;index" admin:"list,filter"`
	DeletedAt *time.Time `db:"column:deleted_at" admin:"-"`
}
