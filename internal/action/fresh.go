package action

import (
	"database/sql"
)

type FreshAction struct {
	name string
}

func (FreshAction) Run(_ *sql.DB, _ ...string) error {
	return nil
}
