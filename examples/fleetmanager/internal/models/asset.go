package models
import (
	"time"
	"github.com/jcsvwinston/nucleus/pkg/model"
)
type Asset struct {
	model.BaseModel
	FleetID   int64      `db:"column:fleet_id;required;index" admin:"list,filter" fk:"Fleet"`
	Name      string     `db:"column:name;required;index" admin:"list,search"`
	Type      string     `db:"column:type;required;index" admin:"list,filter" choices:"truck,van,car,trailer"`
	VIN       string     `db:"column:vin;required;unique" admin:"list,search"`
	DeletedAt *time.Time `db:"column:deleted_at" admin:"-"`
}
