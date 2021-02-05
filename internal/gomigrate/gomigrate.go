package gomigrate

import (
	"database/sql"
	"fmt"
	"github.com/tweety53/gomigrate/internal/action"
)

const (
	// The name of the dummy migration that marks the beginning of the whole migration history.
	BaseMigrationVersion = "m000000_000000_base"
)

var (
	// The default action
	defaultAction = "up"

	// The directory containing the migration classes
	MigrationsPath string

	// Table name which contains migrations data
	MigrationTable string

	// "create" action template file path
	goTemplateFilePath string

	// If this is set to true, the individual commands ran within the migration will not be output to the console.
	// Default is false, in other words the output is fully verbose by default.
	Compact bool
)

func SetMigrationsPath(p string) {
	MigrationsPath = p
}

func SetMigrationTable(n string) {
	MigrationTable = n
}

func SetGoTemplateFilePath(p string) {
	goTemplateFilePath = p
}

func SetCompact(c bool) {
	Compact = c
}

// Run runs an action.
func Run(a string, db *sql.DB, args []string) error {
	switch a {
	case "create":
		createAction := action.NewCreateAction(db, MigrationsPath)
		params := new(action.CreateActionParams)
		if err := params.ValidateAndFill(args); err != nil {
			return err
		}
		if err := createAction.Run(params); err != nil {
			return err
		}
	case "down":
		downAction := action.NewDownAction(db)
		params := new(action.DownActionParams)
		if err := params.ValidateAndFill(args); err != nil {
			return err
		}
		if err := downAction.Run(params); err != nil {
			return err
		}
	case "fresh":
		return nil
	case "history":
		return nil
	case "mark":
		return nil
	case "new":
		return nil
	case "redo":
		return nil
	case "to":
		return nil
	case "up":
		upAction := action.NewUpAction(db)
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
