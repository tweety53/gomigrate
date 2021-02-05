package action

import (
	"database/sql"
)

type ToAction struct {
	name string
}

func (ToAction) Run(_ *sql.DB, _ ...string) error {
	return nil
}
