package sql_dialect

import (
	"database/sql"
	"fmt"
	"strconv"
)

// PostgresDialect struct.
type PostgresDialect struct{}

func (pd PostgresDialect) CreateVersionTableSQL() string {
	return fmt.Sprintf(`CREATE TABLE %s (
			version TEXT NOT NULL
				CONSTRAINT migration_pkey
					PRIMARY KEY,
			apply_time INTEGER
            );`, "migration")
}

func (pd PostgresDialect) InsertVersionSQL() string {
	return fmt.Sprintf("INSERT INTO %s (version, apply_time) VALUES ($1, $2);", "migration")
}

func (pd PostgresDialect) GetMigrationsHistory(db *sql.DB, limit int) (*sql.Rows, error) {
	query := fmt.Sprintf("SELECT version, apply_time FROM %s ORDER BY apply_time DESC, version DESC", "migration")
	if limit > 0 {
		query += " LIMIT " + strconv.Itoa(limit)
	}

	query += ";"

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	return rows, err
}

func (pd PostgresDialect) DeleteVersionSQL() string {
	return fmt.Sprintf("DELETE FROM %s WHERE version=$1;", "migration")
}
