package gomigrate

import (
	"database/sql"
	"fmt"
	"runtime"

	"github.com/tweety53/gomigrate/internal/action"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/repo"
	"github.com/tweety53/gomigrate/internal/service"
	"github.com/tweety53/gomigrate/internal/sqldialect"
	"github.com/tweety53/gomigrate/pkg/config"
)

func Run(a string, db *sql.DB, config *config.GoMigrateConfig, args []string) error {
	log.SetVerbose(!config.Compact)

	dialect, err := sqldialect.InitDialect(config.SQLDialect, config.MigrationTable)
	if err != nil {
		return err
	}

	migrationsSvc := service.NewMigrationService(
		db,
		repo.NewMigrationsRepository(db, dialect),
		repo.NewDbOperationsRepository(db, dialect),
		&migration.MigrationsCollector{},
		config.MigrationsPath)

	switch a {
	case "create":
		createAction := action.NewCreateAction(config.MigrationsPath)
		params := new(action.CreateActionParams)
		if err := params.ValidateAndFill(args); err != nil {
			return err
		}
		if err := createAction.Run(params); err != nil {
			return err
		}
	case "down":
		downAction := action.NewDownAction(migrationsSvc)
		params := new(action.DownActionParams)
		if err := params.ValidateAndFill(args); err != nil {
			return err
		}
		if err := downAction.Run(params); err != nil {
			return err
		}
	case "fresh":
		freshAction := action.NewFreshAction(migrationsSvc)
		params := new(action.FreshActionParams)
		if err := params.ValidateAndFill(args); err != nil {
			return err
		}
		if err := freshAction.Run(params); err != nil {
			return err
		}
	case "history":
		historyAction := action.NewHistoryAction(migrationsSvc)
		params := new(action.HistoryActionParams)
		if err := params.ValidateAndFill(args); err != nil {
			return err
		}
		if err := historyAction.Run(params); err != nil {
			return err
		}
	case "mark":
		markAction := action.NewMarkAction(migrationsSvc)
		params := new(action.MarkActionParams)
		if err := params.ValidateAndFill(args); err != nil {
			return err
		}
		if err := markAction.Run(params); err != nil {
			return err
		}
	case "new":
		newAction := action.NewNewAction(migrationsSvc)
		params := new(action.NewActionParams)
		if err := params.ValidateAndFill(args); err != nil {
			return err
		}
		if err := newAction.Run(params); err != nil {
			return err
		}
	case "redo":
		redoAction := action.NewRedoAction(migrationsSvc)
		params := new(action.RedoActionParams)
		if err := params.ValidateAndFill(args); err != nil {
			return err
		}
		if err := redoAction.Run(params); err != nil {
			return err
		}
	case "to":
		toAction := action.NewToAction(migrationsSvc)
		params := new(action.ToActionParams)
		if err := params.ValidateAndFill(args); err != nil {
			return err
		}
		if err := toAction.Run(params); err != nil {
			return err
		}
	case "up":
		upAction := action.NewUpAction(migrationsSvc)
		params := new(action.UpActionParams)
		if err := params.ValidateAndFill(args); err != nil {
			return err
		}
		if err := upAction.Run(params); err != nil {
			return err
		}
	default:
		return fmt.Errorf("%q: no such action", a)
	}

	return nil
}

func AddMigration(up func(*sql.Tx) error, down func(*sql.Tx) error) {
	_, filename, _, _ := runtime.Caller(1) //nolint:dogsled
	migration.AddNamedMigration(filename, up, down)
}
