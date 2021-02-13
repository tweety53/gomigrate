package sql_dialect

import (
	"fmt"
)

type SQLDialect interface {
	CreateVersionTableSQL() string
	InsertVersionSQL() string
	InsertUnAppliedVersionSQL() string
	UpdateApplyTimeSQL() string
	LockVersionSQL() string
	DeleteVersionSQL() string
	AllTableNamesSQL() string
	TableForeignKeysSQL() string
	DropFkSQL(tableName string, fkName string) string
	DropTableSQL(tableName string) string
	MigrationsHistorySQL() string
}

func InitDialect(v, migrationTable string) (SQLDialect, error) {
	var dialect SQLDialect
	switch v {
	case "postgres":
		dialect = &PostgresDialect{
			migrationTable: migrationTable,
		}
	default:
		return nil, fmt.Errorf("%q: unknown dialect", v)
	}

	return dialect, nil
}
