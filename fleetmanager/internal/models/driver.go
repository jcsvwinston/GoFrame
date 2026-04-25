package models
import (
	"time"
	"github.com/jcsvwinston/GoFrame/pkg/model"
)
type Driver struct {
	model.BaseModel
	Name      string     `db:"column:name;required;index" admin:"list,search"`
	LicenseNo string     `db:"column:license_no;required;unique" admin:"list,search"`
	Phone     string     `db:"column:phone" admin:"list"`
	DeletedAt *time.Time `db:"column:deleted_at" admin:"-"`
}
