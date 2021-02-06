package migrations

import (
	"database/sql"
	"github.com/tweety53/gomigrate/pkg/gomigrate"
)

func init() {
	gomigrate.AddMigration(upZululKek, downZululKek)
}

func upZululKek(tx *sql.Tx) error {
	_, err := tx.Exec(`
ALTER TABLE accounts 
RENAME COLUMN username TO user_name;`)
	if err != nil {
		return err
	}
	return nil
}

func downZululKek(tx *sql.Tx) error {
	_, err := tx.Exec(`
ALTER TABLE accounts 
RENAME COLUMN user_name TO username;`)
	if err != nil {
		return err
	}
	return nil
}
