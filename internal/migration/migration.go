package migration

import (
	"database/sql"
	"fmt"
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/sql_dialect"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

var registeredMigrations = map[string]*Migration{}

// MigrationRecord struct.
type MigrationRecord struct {
	Version   string
	ApplyTime int
}

// Migration struct.
type Migration struct {
	Version    string
	Next       string // next version, or -1 if none
	Previous   string // previous version, -1 if none
	Source     string // path to .sql script or go file
	Registered bool
	UpFn       func(*sql.Tx) error // Up go migration function
	DownFn     func(*sql.Tx) error // Down go migration function
}

func (m *Migration) String() string {
	return fmt.Sprintf(m.Version)
}

// Up runs an up migration.
func (m *Migration) Up(db *sql.DB) error {
	if err := m.run(db, true); err != nil {
		return err
	}
	return nil
}

// Down runs a down migration.
func (m *Migration) Down(db *sql.DB) error {
	if err := m.run(db, false); err != nil {
		return err
	}
	return nil
}

func (m *Migration) run(db *sql.DB, direction bool) error {
	switch filepath.Ext(m.Source) {
	case ".sql":
		f, err := os.Open(m.Source)
		if err != nil {
			return errors.Wrapf(err, "ERROR %v: failed to open SQL migration file", filepath.Base(m.Source))
		}
		defer f.Close()

		statements, useTx, err := parseSQLMigration(f, direction)
		if err != nil {
			return errors.Wrapf(err, "ERROR %v: failed to parse SQL migration file", filepath.Base(m.Source))
		}

		if err := runSQLMigration(db, statements, useTx, m.Version, direction); err != nil {
			return errors.Wrapf(err, "ERROR %v: failed to run SQL migration", filepath.Base(m.Source))
		}

		if len(statements) > 0 {
			log.Infoln("OK   ", filepath.Base(m.Source))
		} else {
			log.Warnln("EMPTY", filepath.Base(m.Source))
		}

	case ".go":
		if !m.Registered {
			return errors.Errorf("ERROR %v", m.Source)
		}
		tx, err := db.Begin()
		if err != nil {
			return errors.Wrap(err, "ERROR failed to begin transaction")
		}

		fn := m.UpFn
		if !direction {
			fn = m.DownFn
		}

		if fn != nil {
			// Run Go migration function.
			if err := fn(tx); err != nil {
				tx.Rollback()
				return errors.Wrapf(err, "ERROR %v: failed to run Go migration function %T", filepath.Base(m.Source), fn)
			}
		}

		if direction {
			if _, err := tx.Exec(sql_dialect.GetDialect().InsertVersionSQL(), m.Version, int(time.Now().Unix())); err != nil {
				tx.Rollback()
				return errors.Wrap(err, "ERROR failed to execute transaction")
			}
		} else {
			if _, err := tx.Exec(sql_dialect.GetDialect().DeleteVersionSQL(), m.Version); err != nil {
				tx.Rollback()
				return errors.Wrap(err, "ERROR failed to execute transaction")
			}
		}

		if err := tx.Commit(); err != nil {
			return errors.Wrap(err, "ERROR failed to commit transaction")
		}

		if fn != nil {
			log.Infoln("OK   ", filepath.Base(m.Source))
		} else {
			log.Warnln("EMPTY", filepath.Base(m.Source))
		}

		return nil
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

// Current gets the current migration.
func (ms Migrations) Current(current string) (*Migration, error) {
	for i, migration := range ms {
		if migration.Version == current {
			return ms[i], nil
		}
	}

	return nil, errors.New("CURR")
}

// Next gets the next migration.
func (ms Migrations) Next(current string) (*Migration, error) {
	for i, migration := range ms {
		if migration.Version > current {
			return ms[i], nil
		}
	}

	return nil, errors.New("NEXT")
}

// Previous : Get the previous migration.
func (ms Migrations) Previous(current string) (*Migration, error) {
	for i := len(ms) - 1; i >= 0; i-- {
		if ms[i].Version < current {
			return ms[i], nil
		}
	}

	return nil, errors.New("PREV")
}

// Last gets the last migration.
func (ms Migrations) Last() (*Migration, error) {
	if len(ms) == 0 {
		return nil, errors.New("LAST")
	}

	return ms[len(ms)-1], nil
}

func (ms Migrations) String() string {
	str := ""
	for _, m := range ms {
		str += fmt.Sprintln(m)
	}
	return str
}

// AddNamedMigration : Add a named migration.
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
		if row.Version == "m000000_000000_base" {
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
