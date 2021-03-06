package repo

import (
	"database/sql"

	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/sqldialect"
)

type DBOperationsRepository struct {
	db      *sql.DB
	dialect sqldialect.SQLDialect
}

func NewDBOperationsRepository(db *sql.DB, dialect sqldialect.SQLDialect) *DBOperationsRepository {
	return &DBOperationsRepository{db: db, dialect: dialect}
}

type ForeignKey struct {
	name string
}

type ForeignKeys []*ForeignKey

func (r *DBOperationsRepository) TruncateDatabase() error {
	tableNames, err := r.AllTableNames()
	if err != nil {
		return err
	}

	// first drop all foreign keys
	for i := range tableNames {
		fKeys, err := r.GetForeignKeys(tableNames[i])
		if err != nil {
			return err
		}

		for i := range fKeys {
			_, err := r.db.Exec(r.dialect.DropFkSQL(tableNames[i], fKeys[i].name))
			if err != nil {
				log.Errf("Foreign key drop err: %v\n", err)

				return err
			}

			log.Infof("Foreign key %s dropped.", fKeys[i].name)
		}
	}

	// Then drop the tables
	for _, name := range tableNames {
		// todo: handle db view errors
		err := r.DropTable(name)
		if err != nil {
			log.Errf("Cannot drop %s table, err: %v\n", err)

			return err
		}

		log.Infof("Table %s dropped.", name)
	}

	return nil
}

func (r *DBOperationsRepository) GetForeignKeys(tableName string) (ForeignKeys, error) {
	fkRows, err := r.db.Query(r.dialect.TableForeignKeysSQL(), tableName)
	if err != nil {
		return nil, err
	}
	defer fkRows.Close()

	var fKeys ForeignKeys
	for fkRows.Next() {
		var fk ForeignKey
		if err = fkRows.Scan(&fk.name); err != nil {
			return nil, err
		}

		fKeys = append(fKeys, &fk)
	}
	if fkRows.Err() != nil {
		return nil, fkRows.Err()
	}

	return fKeys, nil
}

func (r *DBOperationsRepository) DropForeignKey(tableName string, fkName string) error {
	if _, err := r.db.Exec(r.dialect.DropFkSQL(tableName, fkName)); err != nil {
		return err
	}

	return nil
}

func (r *DBOperationsRepository) DropTable(tableName string) error {
	if _, err := r.db.Exec(r.dialect.DropTableSQL(tableName)); err != nil {
		return err
	}

	return nil
}

func (r *DBOperationsRepository) AllTableNames() ([]string, error) {
	tableNamesRows, err := r.db.Query(r.dialect.AllTableNamesSQL())
	if err != nil {
		return nil, err
	}
	defer tableNamesRows.Close()

	tableNames := make([]string, 0, 1024*1024*4)
	// First drop all foreign keys
	for tableNamesRows.Next() {
		var tableName string
		if err = tableNamesRows.Scan(&tableName); err != nil {
			return nil, err
		}

		tableNames = append(tableNames, tableName)
	}
	if tableNamesRows.Err() != nil {
		return nil, tableNamesRows.Err()
	}

	return tableNames, err
}
