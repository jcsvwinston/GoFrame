package db

import (
	"database/sql"
	"fmt"

	"github.com/jcsvwinston/GoFrame/pkg/quark"
	"github.com/spf13/viper"
)

func GetQuarkClient() (*quark.Client, error) {
	driver := viper.GetString("database.default.driver")
	dsn := viper.GetString("database.default.dsn")

	if driver == "" || dsn == "" {
		return nil, fmt.Errorf("database configuration missing")
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	return quark.New(db, quark.WithLimits(quark.Limits{AllowRawQueries: true}))
}

func GetAdminQuarkClient() (*quark.Client, error) {
	driver := viper.GetString("database.admin.driver")
	dsn := viper.GetString("database.admin.dsn")

	if driver == "" || dsn == "" {
		return nil, fmt.Errorf("admin database configuration missing")
	}

	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, err
	}

	return quark.New(db, quark.WithLimits(quark.Limits{AllowRawQueries: true}))
}
