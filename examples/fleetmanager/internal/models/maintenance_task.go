package models
import (
	"time"
	"github.com/jcsvwinston/nucleus/pkg/model"
)
type MaintenanceTask struct {
	model.BaseModel
	DeviceID    int64      `db:"column:device_id;required;index" admin:"list,filter" fk:"Device"`
	Description string     `db:"column:description;required" admin:"list,search"`
	Status      string     `db:"column:status;required;index" admin:"list,filter" choices:"pending,in_progress,completed"`
	DueDate     time.Time  `db:"column:due_date;required" admin:"list,filter"`
	DeletedAt   *time.Time `db:"column:deleted_at" admin:"-"`
}
