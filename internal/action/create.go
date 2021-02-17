package action

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/version"
	errorsInternal "github.com/tweety53/gomigrate/pkg/errors"
	"github.com/tweety53/gomigrate/pkg/exitcode"
)

var (
	ErrInvalidName           = errors.New("the migration name should contain letters, digits, underscore and/or backslash characters only")
	ErrCannotSelectTmpl      = errors.New("something wrong, cannot select template")
	ErrEmptyName             = errors.New("name cannot be empty")
	ErrUnknownMigrationType  = errors.New("unknown migration type passed")
	ErrUnknownSafeParamValue = errors.New("create action 'safe' param must be true or false")
)

type tmplVars struct {
	CamelName string
}

type CreateAction struct {
	migrationsPath string
}

type CreateActionParams struct {
	name  string
	mType migration.Type
	safe  bool
}

func (p *CreateActionParams) Get() interface{} {
	return &CreateActionParams{
		name:  p.name,
		mType: p.mType,
		safe:  p.safe,
	}
}

// todo: tests for regex
var migrationNameRegex = regexp.MustCompile(`^[\w\\]+$`)

func (p *CreateActionParams) ValidateAndFill(args []string) error {
	if len(args) < 1 {
		return errorsInternal.ErrNotEnoughArgs
	}

	if args[0] == "" {
		return ErrEmptyName
	}
	name := args[0]
	if !migrationNameRegex.MatchString(name) {
		return ErrInvalidName
	}

	var mType migration.Type
	if len(args) == 1 {
		mType = migration.TypeGo
	} else {
		mType = migration.Type(args[1])
		if mType != migration.TypeGo && mType != migration.TypeSQL {
			return ErrUnknownMigrationType
		}
	}

	var safe bool
	if len(args) <= 2 {
		safe = true
	} else {
		switch args[2] {
		case "true":
			safe = true
		case "false":
			safe = false
		default:
			return ErrUnknownSafeParamValue
		}
	}

	p.name = name
	p.mType = mType
	p.safe = safe

	return nil
}

func NewCreateAction(migrationsPath string) *CreateAction {
	return &CreateAction{
		migrationsPath: migrationsPath,
	}
}

func (a *CreateAction) Run(params interface{}) error {
	p, ok := params.(*CreateActionParams)
	if !ok {
		return errorsInternal.ErrInvalidActionParamsType
	}

	var tmpl *template.Template

	if p.mType == migration.TypeGo {
		if p.safe {
			tmpl = MigrationTemplateGoSafe
		} else {
			tmpl = MigrationTemplateGo
		}
	}
	if p.mType == migration.TypeSQL {
		if p.safe {
			tmpl = MigrationTemplateSQLSafe
		} else {
			tmpl = MigrationTemplateSQL
		}
	}
	if tmpl == nil {
		return ErrCannotSelectTmpl
	}

	path := filepath.Join(a.migrationsPath, version.BuildVersion(p.name, p.mType))
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		log.Err("Failed to create new migration.")

		return &errorsInternal.GoMigrateError{
			Err:      err,
			ExitCode: exitcode.IoErr,
		}
	}

	f, err := os.Create(path)
	if err != nil {
		log.Err("Failed to create new migration.")

		return &errorsInternal.GoMigrateError{
			Err:      err,
			ExitCode: exitcode.IoErr,
		}
	}
	defer f.Close()

	vars := tmplVars{
		CamelName: nameToCamelCase(p.name),
	}
	if err := tmpl.Execute(f, vars); err != nil {
		return err
	}

	log.Infof("New migration created successfully: %s\n", f.Name())

	return nil
}

var nameToCamelRegex = regexp.MustCompile("(^[A-Za-z])|_([A-Za-z])")

func nameToCamelCase(name string) string {
	return nameToCamelRegex.ReplaceAllStringFunc(name, func(s string) string {
		return strings.ToUpper(strings.ReplaceAll(s, "_", ""))
	})
}

var MigrationTemplateSQLSafe = template.Must(template.New("gomigrate.sql-migration-safe").Parse(`-- +gomigrate Up
-- +gomigrate StatementBegin
// write down up SQL here
-- +gomigrate StatementEnd

-- +gomigrate Down
-- +gomigrate StatementBegin
// write down down SQL here
-- +gomigrate StatementEnd
`))

var MigrationTemplateSQL = template.Must(template.New("gomigrate.sql-migration").Parse(`-- +gomigrate NO TRANSACTION
-- +gomigrate Up
-- +gomigrate StatementBegin
// write down up SQL here
-- +gomigrate StatementEnd

-- +gomigrate Down
-- +gomigrate StatementBegin
// write down down SQL here
-- +gomigrate StatementEnd
`))

var MigrationTemplateGo = template.Must(template.New("gomigrate.go-migration").Parse(`package migrations

import (
	"database/sql"
	"github.com/tweety53/gomigrate/pkg/gomigrate"
)

func init() {
	gomigrate.AddMigration(up{{.CamelName}}, down{{.CamelName}})
}

func up{{.CamelName}}(db *sql.DB) error {
	// This code is executed when the migration is applied.
	return nil
}

func down{{.CamelName}}(db *sql.DB) error {
	// This code is executed when the migration is rolled back.
	return nil
}
`))

var MigrationTemplateGoSafe = template.Must(template.New("gomigrate.go-migration-safe").Parse(`package migrations

import (
	"database/sql"
	"github.com/tweety53/gomigrate/pkg/gomigrate"
)

func init() {
	gomigrate.AddSafeMigration(safeUp{{.CamelName}}, safeDown{{.CamelName}})
}

func safeUp{{.CamelName}}(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	return nil
}

func safeDown{{.CamelName}}(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
`))
