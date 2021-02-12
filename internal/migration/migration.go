package migration

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

const (
	// The name of the dummy migration that marks the beginning of the whole migration history.
	BaseMigrationVersion = "m000000_000000_base"
)

var registeredMigrations = map[string]*Migration{}

type MigrationRecord struct {
	Version   string
	ApplyTime int
}

type Migration struct {
	Version    string
	Next       string // next version, or -1 if none
	Previous   string // previous version, -1 if none
	Source     string // path to .sql script or go file
	Registered bool
	UpFn       func(*sql.Tx) error // Up go migration function
	DownFn     func(*sql.Tx) error // Down go migration function
}

type MigrationDirection string

var (
	migrationDirectionUp   MigrationDirection = "up"
	migrationDirectionDown MigrationDirection = "down"
)

func (m *Migration) String() string {
	return fmt.Sprintf(m.Version)
}

func (m *Migration) Up(db *sql.DB) error {
	if err := m.run(db, migrationDirectionUp); err != nil {
		return err
	}
	return nil
}

// Down runs a down migration.
func (m *Migration) Down(db *sql.DB) error {
	if err := m.run(db, migrationDirectionDown); err != nil {
		return err
	}
	return nil
}

func (m *Migration) run(db *sql.DB, direction MigrationDirection) error {
	switch filepath.Ext(m.Source) {
	case ".sql":
		f, err := os.Open(m.Source)
		if err != nil {
			return errors.Wrapf(err, "ERROR %v: failed to open SQL migration file", filepath.Base(m.Source))
		}
		defer f.Close()

		statements, useTx, err := parseSQLMigration(f, direction)
		if err != nil {
			return errors.Wrapf(err, "failed to parse SQL migration file: %v", filepath.Base(m.Source))
		}

		if direction == migrationDirectionUp {
			assembleUpFnFromStatements(statements, useTx, m)
			return migrateUpGo(db, m)
		} else {
			assembleDownFnFromStatements(statements, useTx, m)
			return migrateDownGo(db, m)
		}

	case ".go":
		if !m.Registered {
			return errors.Errorf("ERROR %v", m.Source)
		}

		if direction == migrationDirectionUp {
			return migrateUpGo(db, m)
		} else {
			return migrateDownGo(db, m)
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
		panic("GetComparableVersion")
	}

	return val
}

// Migrations slice.
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
		panic(fmt.Sprintf("gomigrate: LESS %v:\n%v\n%v", ms[i].Version, ms[i].Source, ms[j].Source))
	}

	jParts := strings.Split(ms[j].Version, "_")
	jPrefix := strings.TrimLeft(jParts[0], "m") + jParts[1]
	jVal, err := strconv.Atoi(jPrefix)
	if err != nil {
		panic(fmt.Sprintf("gomigrate: LESS %v:\n%v\n%v", ms[i].Version, ms[i].Source, ms[j].Source))
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

	return nil, errors.New("CURR")
}

func (ms Migrations) Next(current string) (*Migration, error) {
	for i, migration := range ms {
		if migration.Version > current {
			return ms[i], nil
		}
	}

	return nil, errors.New("NEXT")
}

func (ms Migrations) Previous(current string) (*Migration, error) {
	for i := len(ms) - 1; i >= 0; i-- {
		if ms[i].Version < current {
			return ms[i], nil
		}
	}

	return nil, errors.New("PREV")
}

func (ms Migrations) Last() (*Migration, error) {
	if len(ms) == 0 {
		return nil, errors.New("LAST")
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

func sortAndConnectMigrations(migrations Migrations) Migrations {
	sort.Sort(migrations)

	for i, m := range migrations {
		prev := ""
		if i > 0 {
			prev = migrations[i-1].Version
			migrations[i-1].Next = m.Version
		}
		migrations[i].Previous = prev
	}

	return migrations
}

func ConvertDbRecordsToMigrationObjects(rows *sql.Rows) (Migrations, error) {
	var migrations Migrations

	for rows.Next() {

		var (
			err error
			row MigrationRecord
		)

		if err = rows.Scan(&row.Version, &row.ApplyTime); err != nil {
			return Migrations{}, errors.Wrap(err, "failed to scan row")
		}
		// skip base migration
		if row.Version == BaseMigrationVersion {
			continue
		}

		migrations = append(migrations, &Migration{
			Version:    row.Version,
			Next:       "",
			Previous:   "",
			Source:     "",
			Registered: false,
			UpFn:       nil,
			DownFn:     nil,
		})
	}

	return migrations, nil
}
