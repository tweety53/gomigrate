package gomigrate

import (
	"database/sql"
	"gomigrate/internal/migration"
	"runtime"
)

// AddMigration adds a migration.
func AddMigration(up func(*sql.Tx) error, down func(*sql.Tx) error) {
	_, filename, _, _ := runtime.Caller(1)
	migration.AddNamedMigration(filename, up, down)
}
