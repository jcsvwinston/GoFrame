package db

import (
	"database/sql"
	"fmt"
	"strings"
)

type TableInfo struct {
	Name    string
	Columns []ColumnInfo
}

type ColumnInfo struct {
	Name       string
	Type       string
	IsNullable bool
	IsPK       bool
	IsAuto     bool
	Default    sql.NullString
}

func GetTableInfo(db *sql.DB, dialect, tableName string) (*TableInfo, error) {
	switch dialect {
	case "postgres", "postgresql":
		return getPostgresTableInfo(db, tableName)
	case "mysql":
		return getMySQLTableInfo(db, tableName)
	case "sqlite":
		return getSQLiteTableInfo(db, tableName)
	default:
		return nil, fmt.Errorf("unsupported dialect for introspection: %s", dialect)
	}
}

func getPostgresTableInfo(db *sql.DB, tableName string) (*TableInfo, error) {
	query := `
		SELECT 
			column_name, 
			data_type, 
			is_nullable,
			column_default,
			(SELECT count(*) FROM information_schema.table_constraints tc 
			 JOIN information_schema.key_column_usage kcu ON tc.constraint_name = kcu.constraint_name
			 WHERE tc.table_name = $1 AND kcu.column_name = cols.column_name AND tc.constraint_type = 'PRIMARY KEY') as is_pk
		FROM information_schema.columns cols
		WHERE table_name = $1
		ORDER BY ordinal_position
	`
	rows, err := db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	info := &TableInfo{Name: tableName}
	for rows.Next() {
		var col ColumnInfo
		var isNullable string
		var isPK int
		if err := rows.Scan(&col.Name, &col.Type, &isNullable, &col.Default, &isPK); err != nil {
			return nil, err
		}
		col.IsNullable = isNullable == "YES"
		col.IsPK = isPK > 0
		col.IsAuto = col.Default.Valid && strings.Contains(col.Default.String, "nextval")
		info.Columns = append(info.Columns, col)
	}
	return info, nil
}

func getMySQLTableInfo(db *sql.DB, tableName string) (*TableInfo, error) {
	query := "DESCRIBE " + tableName
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	info := &TableInfo{Name: tableName}
	for rows.Next() {
		var field, typ, null, key, extra string
		var def sql.NullString
		if err := rows.Scan(&field, &typ, &null, &key, &def, &extra); err != nil {
			return nil, err
		}
		info.Columns = append(info.Columns, ColumnInfo{
			Name:       field,
			Type:       typ,
			IsNullable: null == "YES",
			IsPK:       key == "PRI",
			IsAuto:     strings.Contains(extra, "auto_increment"),
			Default:    def,
		})
	}
	return info, nil
}

func getSQLiteTableInfo(db *sql.DB, tableName string) (*TableInfo, error) {
	query := fmt.Sprintf("PRAGMA table_info(%s)", tableName)
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	info := &TableInfo{Name: tableName}
	for rows.Next() {
		var cid int
		var name, typ string
		var notnull, pk int
		var dfltValue sql.NullString
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dfltValue, &pk); err != nil {
			return nil, err
		}
		info.Columns = append(info.Columns, ColumnInfo{
			Name:       name,
			Type:       typ,
			IsNullable: notnull == 0,
			IsPK:       pk > 0,
			IsAuto:     pk > 0 && strings.Contains(strings.ToUpper(typ), "INTEGER"), // Simple heuristic for SQLite
			Default:    dfltValue,
		})
	}
	return info, nil
}
