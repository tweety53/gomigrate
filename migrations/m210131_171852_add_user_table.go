package migrations

import (
	"database/sql"
	"github.com/tweety53/gomigrate/pkg/gomigrate"
)

func init() {
	gomigrate.AddMigration(upAddUserTable, downAddUserTable)
}

func upAddUserTable(tx *sql.Tx) error {
	_, err := tx.Exec(`CREATE TABLE accounts (
	user_id serial PRIMARY KEY,
	username VARCHAR ( 50 ) UNIQUE NOT NULL,
	password VARCHAR ( 50 ) NOT NULL,
	email VARCHAR ( 255 ) UNIQUE NOT NULL,
	created_on TIMESTAMP NOT NULL,
        last_login TIMESTAMP 
);`)
	if err != nil {
		return err
	}
	return nil
}

func downAddUserTable(tx *sql.Tx) error {
	_, err := tx.Exec(`DROP TABLE accounts;`)
	if err != nil {
		return err
	}
	return nil
}