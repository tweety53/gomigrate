package sql_dialect

import (
	"database/sql"
	"fmt"
)

type SQLDialect interface {
	CreateVersionTableSQL() string
	InsertVersionSQL() string
	DeleteVersionSQL() string
	GetMigrationsHistory(db *sql.DB, limit int) (*sql.Rows, error)
}

var dialect SQLDialect = &PostgresDialect{}

// GetDialect gets the SQLDialect
func GetDialect() SQLDialect {
	return dialect
}

// SetDialect sets the SQLDialect
func SetDialect(d string) error {
	switch d {
	case "postgres":
		dialect = &PostgresDialect{}
	default:
		return fmt.Errorf("%q: unknown dialect", d)
	}

	return nil
}
