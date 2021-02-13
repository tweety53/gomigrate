package sql_dialect

import (
	"fmt"
	"github.com/tweety53/gomigrate/internal/config"
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

var dialect SQLDialect

func GetDialect() SQLDialect {
	return dialect
}

func InitDialect(config *config.AppConfig) error {
	switch config.SQLDialect {
	case "postgres":
		dialect = &PostgresDialect{
			config: config,
		}
	default:
		return fmt.Errorf("%q: unknown dialect", config.SQLDialect)
	}

	return nil
}
