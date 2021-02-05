package action

import (
	"database/sql"
)

type NewAction struct {
	name string
}

func (NewAction) Run(_ *sql.DB, _ ...string) error {
	return nil
}
