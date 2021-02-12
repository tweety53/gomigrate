package sql_dialect

import (
	"fmt"
	"github.com/tweety53/gomigrate/internal/config"
)

type PostgresDialect struct {
	config *config.AppConfig
}

func (pd PostgresDialect) CreateVersionTableSQL() string {
	return fmt.Sprintf(`CREATE TABLE %s (
			version TEXT NOT NULL
				CONSTRAINT migration_pkey
					PRIMARY KEY,
			apply_time INTEGER
            );`, pd.config.MigrationTable)
}

func (pd PostgresDialect) InsertVersionSQL() string {
	return fmt.Sprintf("INSERT INTO %s (version, apply_time) VALUES ($1, $2);", pd.config.MigrationTable)
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
	return fmt.Sprintf("DELETE FROM %s WHERE version=$1;", pd.config.MigrationTable)
}

func (pd PostgresDialect) InsertUnAppliedVersionSQL() string {
	return fmt.Sprintf("INSERT INTO %s (version) VALUES ($1);", pd.config.MigrationTable)
}

func (pd PostgresDialect) UpdateApplyTimeSQL() string {
	return fmt.Sprintf("UPDATE %s SET apply_time=$1 WHERE version=$2;", pd.config.MigrationTable)
}

func (pd PostgresDialect) LockVersionSQL() string {
	return fmt.Sprintf("SELECT * FROM %s WHERE version=$1 FOR UPDATE NOWAIT;", pd.config.MigrationTable)
}