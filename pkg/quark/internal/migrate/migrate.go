// Package migrate provides internal utilities for database schema migrations.
package migrate

import (
	"reflect"
)

// SQLType maps Go types to SQL types for the given dialect name.
func SQLType(dialectName string, t reflect.Type, isPK bool) string {
	if isPK {
		switch dialectName {
		case "sqlite":
			return "INTEGER PRIMARY KEY AUTOINCREMENT"
		case "postgres":
			return "SERIAL PRIMARY KEY"
		case "mysql", "mariadb":
			return "INT AUTO_INCREMENT PRIMARY KEY"
		case "mssql":
			return "INT IDENTITY(1,1) PRIMARY KEY"
		case "oracle":
			return "NUMBER GENERATED ALWAYS AS IDENTITY PRIMARY KEY"
		default:
			return "INTEGER PRIMARY KEY"
		}
	}

	// Handle pointers (e.g. *time.Time, *string)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	switch t.Kind() {
	case reflect.String:
		switch dialectName {
		case "postgres", "sqlite":
			return "TEXT"
		case "oracle":
			return "VARCHAR2(255)"
		case "mssql":
			return "NVARCHAR(255)"
		default:
			return "VARCHAR(255)"
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if dialectName == "oracle" {
			return "NUMBER(19)"
		}
		return "INTEGER"
	case reflect.Float32, reflect.Float64:
		switch dialectName {
		case "sqlite", "postgres":
			return "REAL"
		case "mysql", "mariadb":
			return "DOUBLE"
		case "oracle", "mssql":
			return "FLOAT"
		default:
			return "REAL"
		}
	case reflect.Bool:
		if dialectName == "oracle" || dialectName == "mssql" {
			return "NUMBER(1)" // Many Oracle/MSSQL implementations use 0/1
		}
		return "BOOLEAN"
	case reflect.Struct:
		if t.String() == "time.Time" {
			switch dialectName {
			case "sqlite", "mysql", "mariadb":
				return "DATETIME"
			case "postgres", "oracle":
				return "TIMESTAMP"
			case "mssql":
				return "DATETIME2"
			default:
				return "TIMESTAMP"
			}
		}
	}

	return "TEXT" // Fallback
}
