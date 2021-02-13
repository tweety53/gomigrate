package action

import (
	"database/sql"
	"github.com/tweety53/gomigrate/internal/helpers"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/repo"
)

type FreshAction struct {
	db             *sql.DB
	migrationsPath string
}

func NewFreshAction(db *sql.DB, migrationsPath string) *FreshAction {
	return &FreshAction{db: db, migrationsPath: migrationsPath}
}

type FreshActionParams struct{}

func (a *FreshAction) Run(_ interface{}) error {
	//todo: restrict action also for local env only
	res := helpers.AskForConfirmation("Are you sure you want to drop all tables and related constraints and start the migration from the beginning?\nAll data will be lost irreversibly!")
	if !res {
		log.Info("Action was cancelled by user. Nothing has been performed.")
		return nil
	}

	// truncate repo
	if err := repo.TruncateDatabase(a.db); err != nil {
		return err
	}

	// exec up action
	upAction := NewUpAction(a.db, a.migrationsPath)
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
