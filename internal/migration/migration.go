package migration

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/repo"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	// The name of the dummy migration that marks the beginning of the whole migration history.
	BaseMigrationVersion = "m000000_000000_base"
)

var registeredMigrations = map[string]*Migration{}

type Migration struct {
	Version    string
	Next       string
	Previous   string
	Source     string // path to .sql\.go file
	Registered bool
	UpFn       func(*sql.Tx) error
	DownFn     func(*sql.Tx) error
}

type MigrationDirection string

var (
	migrationDirectionUp   MigrationDirection = "up"
	migrationDirectionDown MigrationDirection = "down"
)

func (m *Migration) String() string {
	return fmt.Sprintf(m.Version)
}

func (m *Migration) Up(repo *repo.MigrationsRepository) error {
	if err := m.run(repo, migrationDirectionUp); err != nil {
		return err
	}
	return nil
}

func (m *Migration) Down(repo *repo.MigrationsRepository) error {
	if err := m.run(repo, migrationDirectionDown); err != nil {
		return err
	}
	return nil
}

func (m *Migration) run(repo *repo.MigrationsRepository, direction MigrationDirection) error {
	switch filepath.Ext(m.Source) {
	case ".sql":
		f, err := os.Open(m.Source)
		if err != nil {
			return errors.Wrapf(err, "failed to open SQL migration file, err: %v", filepath.Base(m.Source))
		}
		defer f.Close()

		statements, useTx, err := parseSQLMigration(f, direction)
		if err != nil {
			return errors.Wrapf(err, "failed to parse SQL migration file: %v", filepath.Base(m.Source))
		}

		if direction == migrationDirectionUp {
			assembleFnFromStatements(statements, useTx, m, direction)
			return migrateUpGo(repo, m)
		} else {
			assembleFnFromStatements(statements, useTx, m, direction)
			return migrateDownGo(repo, m)
		}

	case ".go":
		if !m.Registered {
			return errors.Errorf("not registered %v", m.Source)
		}

		if direction == migrationDirectionUp {
			return migrateUpGo(repo, m)
		} else {
			return migrateDownGo(repo, m)
		}
	}

	return nil
}

func GetVersionFromFileName(name string) (string, error) {
	base := filepath.Base(name)

	if ext := filepath.Ext(base); ext != ".go" && ext != ".sql" {
		return "", errors.New("only .go and .sql migrations supported")
	}

	version := strings.Split(base, ".")
	if len(version) != 2 {
		return "", errors.New("cannot extract migration version from filename")
	}

	return version[0], nil
}

func GetComparableVersion(version string) int {
	parts := strings.Split(version, "_")
	prefix := strings.TrimLeft(parts[0], "m") + parts[1]
	val, err := strconv.Atoi(prefix)
	if err != nil {
		panic("cannot get comparable version from string:" + version)
	}

	return val
}

type Migrations []*Migration

// helpers so we can use pkg sort
func (ms Migrations) Len() int      { return len(ms) }
func (ms Migrations) Swap(i, j int) { ms[i], ms[j] = ms[j], ms[i] }
func (ms Migrations) Less(i, j int) bool {
	if ms[i].Version == ms[j].Version {
		panic(fmt.Sprintf("gomigrate: duplicate version %v detected:\n%v\n%v", ms[i].Version, ms[i].Source, ms[j].Source))
	}

	// extract prefixes
	iParts := strings.Split(ms[i].Version, "_")
	iPrefix := strings.TrimLeft(iParts[0], "m") + iParts[1]
	iVal, err := strconv.Atoi(iPrefix)
	if err != nil {
		panic(fmt.Sprintf("gomigrate: Less() %v:\n%v\n%v,err: %v", ms[i].Version, ms[i].Source, ms[j].Source, err))
	}

	jParts := strings.Split(ms[j].Version, "_")
	jPrefix := strings.TrimLeft(jParts[0], "m") + jParts[1]
	jVal, err := strconv.Atoi(jPrefix)
	if err != nil {
		panic(fmt.Sprintf("gomigrate: Less() %v:\n%v\n%v,err: %v", ms[i].Version, ms[i].Source, ms[j].Source, err))
	}

	return iVal < jVal
}

func (ms Migrations) Reverse() {
	for i, j := 0, len(ms)-1; i < j; i, j = i+1, j-1 {
		ms[i], ms[j] = ms[j], ms[i]
	}
}

func (ms Migrations) Current(current string) (*Migration, error) {
	for i, migration := range ms {
		if migration.Version == current {
			return ms[i], nil
		}
	}

	return nil, errors.New("gomigrate: cannot get current version from migrations slice")
}

func (ms Migrations) Next(current string) (*Migration, error) {
	for i, migration := range ms {
		if migration.Version > current {
			return ms[i], nil
		}
	}

	return nil, errors.New("gomigrate: cannot get next version from migrations slice")
}

func (ms Migrations) Previous(current string) (*Migration, error) {
	for i := len(ms) - 1; i >= 0; i-- {
		if ms[i].Version < current {
			return ms[i], nil
		}
	}

	return nil, errors.New("gomigrate: cannot get previous version from migrations slice")
}

func (ms Migrations) Last() (*Migration, error) {
	if len(ms) == 0 {
		return nil, errors.New("gomigrate: cannot get last version from migrations slice")
	}

	return ms[len(ms)-1], nil
}

func (ms Migrations) String() string {
	str := "\n"
	for _, m := range ms {
		str += fmt.Sprintln(m)
	}
	return str
}

func AddNamedMigration(filename string, up func(*sql.Tx) error, down func(*sql.Tx) error) {
	v, _ := GetVersionFromFileName(filename)
	migration := &Migration{Version: v, Next: "", Previous: "", Registered: true, UpFn: up, DownFn: down, Source: filename}

	if existing, ok := registeredMigrations[v]; ok {
		panic(fmt.Sprintf("failed to add migration %q: version conflicts with %q", filename, existing.Source))
	}

	registeredMigrations[v] = migration
}

func Convert(records repo.MigrationRecords) Migrations {
	var migrations Migrations

	for i := range records {
		// skip base migration
		if records[i].Version == BaseMigrationVersion {
			continue
		}

		migrations = append(migrations, &Migration{
			Version:    records[i].Version,
			Next:       "",
			Previous:   "",
			Source:     "",
			Registered: false,
			UpFn:       nil,
			DownFn:     nil,
		})
	}

	return migrations
}
