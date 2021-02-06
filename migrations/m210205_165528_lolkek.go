package migrations

import (
	"database/sql"
	"github.com/tweety53/gomigrate/pkg/gomigrate"
)

func init() {
	gomigrate.AddMigration(upLolkek, downLolkek)
}

func upLolkek(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	return nil
}

func downLolkek(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
