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

func (pd PostgresDialect) TableForeignKeysSQL() string {
	return fmt.Sprintf(`
SELECT
    tc.constraint_name as fk_name
FROM 
    information_schema.table_constraints AS tc 
    JOIN information_schema.key_column_usage AS kcu
      ON tc.constraint_name = kcu.constraint_name
      AND tc.table_schema = kcu.table_schema
    JOIN information_schema.constraint_column_usage AS ccu
      ON ccu.constraint_name = tc.constraint_name
      AND ccu.table_schema = tc.table_schema
WHERE tc.constraint_type = 'FOREIGN KEY' AND tc.table_name=$1;
`)
}

func (pd PostgresDialect) AllTableNamesSQL() string {
	return fmt.Sprintf(`
SELECT
    table_schema || '.' || table_name as table_name
FROM
    information_schema.tables
WHERE
    table_type = 'BASE TABLE'
AND
    table_schema NOT IN ('pg_catalog', 'information_schema');
`)
}

func (pd PostgresDialect) DropFkSQL(tableName string, fkName string) string {
	return fmt.Sprintf(`
ALTER TABLE %s DROP CONSTRAINT %s;
`, tableName, fkName)
}

func (pd PostgresDialect) DropTableSQL(tableName string) string {
	return fmt.Sprintf(`
DROP TABLE IF EXISTS %s;
`, tableName)
}

func (pd PostgresDialect) DeleteVersionSQL() string {
	return fmt.Sprintf("DELETE FROM %s WHERE version=$1;", "migration")
}
