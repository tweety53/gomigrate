package action

import (
	"database/sql"
)

type HistoryAction struct {
	name string
}

func (HistoryAction) Run(_ *sql.DB, _ ...string) error {
	return nil
}
