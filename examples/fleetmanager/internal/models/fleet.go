package models
import (
	"time"
	"github.com/jcsvwinston/nucleus/pkg/model"
)
type Fleet struct {
	model.BaseModel
	OrganizationID int64      `db:"column:organization_id;required;index" admin:"list,filter" fk:"Organization"`
	Name           string     `db:"column:name;required;index" admin:"list,search,filter"`
	DeletedAt      *time.Time `db:"column:deleted_at" admin:"-"`
}
