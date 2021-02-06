package action

import (
	"database/sql"
	"github.com/tweety53/gomigrate/internal/db"
	"github.com/tweety53/gomigrate/internal/helpers"
	"github.com/tweety53/gomigrate/internal/log"
)

type FreshAction struct {
	db *sql.DB
}

func NewFreshAction(db *sql.DB) *FreshAction {
	return &FreshAction{db: db}
}

type FreshActionParams struct{}

func (a *FreshAction) Run(_ interface{}) error {
	//todo: restrict action also for local dev only
	res := helpers.AskForConfirmation("Are you sure you want to drop all tables and related constraints and start the migration from the beginning?\\nAll data will be lost irreversibly!", false)
	if !res {
		log.Info("Action was cancelled by user. Nothing has been performed.")
		return nil
	}

	// truncate db
	if err := db.TruncateDatabase(a.db); err != nil {
		return err
	}

	// exec up action
	upAction := NewUpAction(a.db)
	params := new(UpActionParams)
	if err := params.ValidateAndFill([]string{}); err != nil {
		return err
	}
	if err := upAction.Run(params); err != nil {
		return err
	}

	return nil
}

func (p *FreshActionParams) Get() interface{} {
	return &FreshActionParams{}
}

func (p *FreshActionParams) ValidateAndFill(_ []string) error {
	return nil
}
