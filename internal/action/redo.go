package action

import (
	"database/sql"
)

type RedoAction struct {
	name string
}

func (RedoAction) Run(_ *sql.DB, _ ...string) error {
	return nil
}
