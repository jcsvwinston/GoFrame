package models

import "github.com/jcsvwinston/GoFrame/pkg/model"

type Article struct {
	model.BaseModel
	Title     string `db:"column:title;required" validate:"required,min=3" admin:"list,search"`
	Content   string `db:"column:content" admin:"list"`
	Published bool   `db:"column:published" admin:"list,filter"`
}

type Lead struct {
	model.BaseModel
	Name      string `db:"column:name;required" validate:"required,min=2" admin:"list,search"`
	Email     string `db:"column:email;required" validate:"required,email" admin:"list,search"`
	Company   string `db:"column:company" admin:"list,search"`
	WantsDemo bool   `db:"column:wants_demo" admin:"list,filter"`
}
