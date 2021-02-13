package gomigrate

import (
	"database/sql"
	"github.com/tweety53/gomigrate/internal/migration"
	"runtime"
)

func AddMigration(up func(*sql.Tx) error, down func(*sql.Tx) error) {
	_, filename, _, _ := runtime.Caller(1)
	migration.AddNamedMigration(filename, up, down)
}
