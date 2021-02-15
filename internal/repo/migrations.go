package repo

import (
	"database/sql"
	"strconv"
	"time"

	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/sqldialect"
)

type MigrationsRepository struct {
	DB      *sql.DB
	dialect sqldialect.SQLDialect
}

func NewMigrationsRepository(db *sql.DB, dialect sqldialect.SQLDialect) *MigrationsRepository {
	return &MigrationsRepository{DB: db, dialect: dialect}
}

func (r *MigrationsRepository) GetMigrationsHistory(limit int) (MigrationRecords, error) {
	query := r.dialect.MigrationsHistorySQL()
	if limit > 0 {
		query += " LIMIT " + strconv.Itoa(limit)
	}

	query += ";"

	rows, err := r.DB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	var records MigrationRecords

	for rows.Next() {
		var row MigrationRecord
		if err = rows.Scan(&row.Version, &row.ApplyTime); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		records = append(records, &row)
	}

	return records, err
}

func (r *MigrationsRepository) InsertVersion(v string) error {
	if _, err := r.DB.Exec(r.dialect.InsertVersionSQL(), v, int(time.Now().Unix())); err != nil {
		return err
	}

	return nil
}

func (r *MigrationsRepository) DeleteVersion(v string) error {
	if _, err := r.DB.Exec(r.dialect.DeleteVersionSQL(), v); err != nil {
		return err
	}

	return nil
}

func (r *MigrationsRepository) EnsureDBVersion() (string, error) {
	_, err := r.GetMigrationsHistory(1)
	if err != nil {
		return "", r.CreateVersionTable()
	}

	return "", nil
}

func (r *MigrationsRepository) GetDBVersion() (string, error) {
	version, err := r.EnsureDBVersion()
	if err != nil {
		return "", err
	}

	return version, nil
}

func (r *MigrationsRepository) CreateVersionTable() error {
	if _, err := r.DB.Exec(r.dialect.CreateVersionTableSQL()); err != nil {
		return err
	}

	return nil
}

func (r *MigrationsRepository) InsertUnAppliedVersion(v string) error {
	if _, err := r.DB.Exec(r.dialect.InsertUnAppliedVersionSQL(), v); err != nil {
		return err
	}

	return nil
}

func (r *MigrationsRepository) UpdateApplyTime(v string) error {
	if _, err := r.DB.Exec(r.dialect.UpdateApplyTimeSQL(), int(time.Now().Unix()), v); err != nil {
		return err
	}

	return nil
}

func (r *MigrationsRepository) LockVersion(v string) error {
	if _, err := r.DB.Exec(r.dialect.LockVersionSQL(), v); err != nil {
		return err
	}

	return nil
}
