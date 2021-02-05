package action

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	errors2 "gomigrate/internal/errors"
	"gomigrate/internal/exit_code"
	"gomigrate/internal/log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"
)

// Migrations version prefix format
const versionFormat = "m060102_150405"

var (
	ErrInvalidName          = errors.New("the migration name should contain letters, digits, underscore and/or backslash characters only")
	ErrCannotSelectTmpl     = errors.New("something wrong, cannot select template")
	ErrEmptyName            = errors.New("name cannot be empty")
	ErrUnknownMigrationType = errors.New("unknown migration type passed")
)

type MigrationType string

var (
	migrationTypeGo  MigrationType = "go"
	migrationTypeSQL MigrationType = "sql"
)

type tmplVars struct {
	Version   string
	CamelName string
}

type CreateAction struct {
	db             *sql.DB
	migrationsPath string
}

type CreateActionParams struct {
	name  string
	mType MigrationType
}

func (c *CreateActionParams) Get() interface{} {
	return &CreateActionParams{
		name:  c.name,
		mType: c.mType,
	}
}

var migrationNameRegex = regexp.MustCompile("^[\\w\\\\]+$")

func (p *CreateActionParams) ValidateAndFill(args []string) error {
	if len(args) < 1 {
		return errors2.ErrNotEnoughArgs
	}

	if args[0] == "" {
		return ErrEmptyName
	}
	name := args[0]
	if !migrationNameRegex.MatchString(name) {
		return ErrInvalidName
	}
	p.name = name

	var mType MigrationType
	if len(args) != 2 {
		p.mType = migrationTypeGo
		return nil
	}

	p.mType = MigrationType(args[1])
	if mType != migrationTypeGo && mType != migrationTypeSQL {
		return ErrUnknownMigrationType
	}

	return nil
}

func NewCreateAction(db *sql.DB, migrationsPath string) *CreateAction {
	return &CreateAction{
		migrationsPath: migrationsPath,
		db:             db,
	}
}

func (a *CreateAction) Run(params interface{}) error {
	p, ok := params.(*CreateActionParams)
	if !ok {
		return errors2.ErrInvalidActionParamsType
	}

	var (
		tmpl *template.Template
	)

	versionPrefix := time.Now().Format(versionFormat)

	if p.mType == migrationTypeGo {
		tmpl = MigrationTemplateGo
	}
	if p.mType == migrationTypeSQL {
		tmpl = MigrationTemplateSQL
	}
	if tmpl == nil {
		return ErrCannotSelectTmpl
	}

	version := fmt.Sprintf("%s_%s", versionPrefix, p.name)
	fileName := fmt.Sprintf("%s.%s", version, p.mType)

	path := filepath.Join(a.migrationsPath, fileName)
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		log.Err("Failed to create new migration.")
		return &errors2.Error{
			Err:      err,
			ExitCode: exit_code.ExitCodeIOErr,
		}
	}

	f, err := os.Create(path)
	if err != nil {
		log.Err("Failed to create new migration.")
		return &errors2.Error{
			Err:      err,
			ExitCode: exit_code.ExitCodeIOErr,
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
		return strings.ToUpper(strings.Replace(s, "_", "", -1))
	})
}

var MigrationTemplateSQL = template.Must(template.New("gomigrate.sql-migration").Parse(`-- +gomigrate Up
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
	"gomigrate/pkg/gomigrate"
)

func init() {
	gomigrate.AddMigration(up{{.CamelName}}, down{{.CamelName}})
}

func up{{.CamelName}}(tx *sql.Tx) error {
	// This code is executed when the migration is applied.
	return nil
}

func down{{.CamelName}}(tx *sql.Tx) error {
	// This code is executed when the migration is rolled back.
	return nil
}
`))
