package action

import (
	"database/sql"
)

type MarkAction struct {
	name string
}

func (MarkAction) Run(_ *sql.DB, _ ...string) error {
	return nil
}
