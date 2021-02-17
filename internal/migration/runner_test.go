package migration

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gojuno/minimock/v3"
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/log"
	"github.com/tweety53/gomigrate/internal/repo"
)

func Test_migrateUpGo(t *testing.T) {
	type repoMock struct {
		mRepo  repo.MigrationRepo
		dbMock sqlmock.Sqlmock
	}
	type args struct {
		repo repoMock
		m    *Migration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "repo db not initialized",
			args: args{
				repo: func() repoMock {
					_, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(nil, errors.New("some error"))

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeUpFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "cannot init tx",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil)
					mock.ExpectBegin().WillReturnError(errors.New("failed to begin transaction"))

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeUpFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "insert unapplied version failed",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertUnAppliedVersionMock.Return(errors.New("some error"))
					mock.ExpectBegin()
					mock.ExpectRollback()

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeUpFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "insert unapplied version and rollback failed",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertUnAppliedVersionMock.Return(errors.New("some error"))
					mock.ExpectBegin()
					mock.ExpectRollback().WillReturnError(errors.New("tx rollback err"))

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeUpFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "go up func error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertUnAppliedVersionMock.Return(nil).
						DeleteVersionMock.Return(nil)
					mock.ExpectBegin()
					mock.ExpectRollback()
					mock.ExpectBegin()
					mock.ExpectCommit()

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeUpFn: func(tx *sql.Tx) error {
						return errors.New("some go fn error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "go up func error with delete version error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertUnAppliedVersionMock.Return(nil).
						DeleteVersionMock.Return(errors.New("some delete version error"))
					mock.ExpectBegin()
					mock.ExpectRollback()
					mock.ExpectBegin()
					mock.ExpectRollback()

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeUpFn: func(tx *sql.Tx) error {
						return errors.New("some go fn error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update apply time error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertUnAppliedVersionMock.Return(nil).
						UpdateApplyTimeMock.Return(errors.New("some update apply time error")).
						DeleteVersionMock.Return(nil)
					mock.ExpectBegin()
					mock.ExpectRollback()
					mock.ExpectBegin()
					mock.ExpectCommit()

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeUpFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "update apply time error with delete version error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertUnAppliedVersionMock.Return(nil).
						UpdateApplyTimeMock.Return(errors.New("some update apply time error")).
						DeleteVersionMock.Return(errors.New("some delete version error"))
					mock.ExpectBegin()
					mock.ExpectRollback()
					mock.ExpectBegin()
					mock.ExpectRollback()

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeUpFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "final commit error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertUnAppliedVersionMock.Return(nil).
						UpdateApplyTimeMock.Return(nil)
					mock.ExpectBegin()
					mock.ExpectCommit().WillReturnError(errors.New("tx commit err"))

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeUpFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "err case with some error inside go fn()",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mock.ExpectBegin()
					mock.ExpectExec(`CREATE TABLE accounts (user_id serial PRIMARY KEY);`).WillReturnError(errors.New("some error inside go fn()"))
					mock.ExpectRollback()
					mock.ExpectBegin()
					mock.ExpectCommit()

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertUnAppliedVersionMock.Return(nil).
						UpdateApplyTimeMock.Return(nil).
						DeleteVersionMock.Return(nil)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeUpFn: func(tx *sql.Tx) error {
						_, err := tx.Exec(`CREATE TABLE accounts (user_id serial PRIMARY KEY);`)
						if err != nil {
							return err
						}
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "success case simple",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mock.ExpectBegin()
					mock.ExpectCommit()

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertUnAppliedVersionMock.Return(nil).
						UpdateApplyTimeMock.Return(nil)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeUpFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success case no fn",
			args: args{
				repo: func() repoMock {
					_, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version:  "m000000_000000_test",
					Source:   "m000000_000000_test",
					SafeUpFn: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "success case with some go fn()",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mock.ExpectBegin()
					mock.ExpectExec(`CREATE TABLE accounts (user_id serial PRIMARY KEY);`).WillReturnResult(sqlmock.NewResult(0, 0))
					mock.ExpectCommit()

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertUnAppliedVersionMock.Return(nil).
						UpdateApplyTimeMock.Return(nil)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeUpFn: func(tx *sql.Tx) error {
						_, err := tx.Exec(`CREATE TABLE accounts (user_id serial PRIMARY KEY);`)
						if err != nil {
							log.Errf("err %v", err)
							return err
						}
						return nil
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &Runner{}
			if err := runner.MigrateUpSafe(tt.args.repo.mRepo, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("MigrateUpSafe() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.args.repo.dbMock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func Test_migrateUp(t *testing.T) {
	type repoMock struct {
		mRepo  repo.MigrationRepo
		dbMock sqlmock.Sqlmock
	}
	type args struct {
		repo repoMock
		m    *Migration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "db not initialized",
			args: args{
				repo: func() repoMock {
					_, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(nil, errors.New("db not initialized"))

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					UpFn: func(db *sql.DB) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "some go fn() error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					UpFn: func(db *sql.DB) error {
						return errors.New("some go fn() error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "insert version error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertVersionMock.Return(errors.New("some insert version error"))

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					UpFn: func(db *sql.DB) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "success case simple",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertVersionMock.Return(nil)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					UpFn: func(db *sql.DB) error {
						return nil
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success case no fn",
			args: args{
				repo: func() repoMock {
					_, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					UpFn:    nil,
				},
			},
			wantErr: false,
		},
		{
			name: "success case with some go fn()",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mock.ExpectExec(`CREATE TABLE accounts (user_id serial PRIMARY KEY);`).WillReturnResult(sqlmock.NewResult(0, 0))

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						InsertVersionMock.Return(nil)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					UpFn: func(db *sql.DB) error {
						_, err := db.Exec(`CREATE TABLE accounts (user_id serial PRIMARY KEY);`)
						if err != nil {
							return err
						}
						return nil
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &Runner{}
			if err := runner.MigrateUp(tt.args.repo.mRepo, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("MigrateUp() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.args.repo.dbMock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func Test_migrateDownSafe(t *testing.T) {
	type repoMock struct {
		mRepo  repo.MigrationRepo
		dbMock sqlmock.Sqlmock
	}
	type args struct {
		repo repoMock
		m    *Migration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "db not initialized",
			args: args{
				repo: func() repoMock {
					_, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(nil, errors.New("db not initialized"))

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeDownFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "db not initialized",
			args: args{
				repo: func() repoMock {
					_, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(nil, errors.New("db not initialized"))

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeDownFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "transaction begin error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil)
					mock.ExpectBegin().WillReturnError(errors.New("some tx begin error"))

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeDownFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "lock version error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						LockVersionMock.Return(errors.New("some lock version error"))
					mock.ExpectBegin()
					mock.ExpectRollback()

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeDownFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "lock version error + rollback error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						LockVersionMock.Return(errors.New("some lock version error"))
					mock.ExpectBegin()
					mock.ExpectRollback().WillReturnError(errors.New("some rollback error"))

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeDownFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "go fn() error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						LockVersionMock.Return(nil).
						DeleteVersionMock.Return(nil)
					mock.ExpectBegin()
					mock.ExpectRollback()
					mock.ExpectBegin()
					mock.ExpectCommit()

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeDownFn: func(tx *sql.Tx) error {
						return errors.New("some go fn() error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "go fn() error + delete version error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						LockVersionMock.Return(nil).
						DeleteVersionMock.Return(errors.New("some delete version error"))
					mock.ExpectBegin()
					mock.ExpectRollback()
					mock.ExpectBegin()
					mock.ExpectRollback()

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeDownFn: func(tx *sql.Tx) error {
						return errors.New("some go fn() error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "delete version error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						LockVersionMock.Return(nil).
						DeleteVersionMock.Return(errors.New("some delete version error"))
					mock.ExpectBegin()
					mock.ExpectRollback()

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeDownFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "success case simple",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mock.ExpectBegin()
					mock.ExpectCommit()

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						LockVersionMock.Return(nil).
						DeleteVersionMock.Return(nil)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeDownFn: func(tx *sql.Tx) error {
						return nil
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success case no fn()",
			args: args{
				repo: func() repoMock {
					_, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version:    "m000000_000000_test",
					Source:     "m000000_000000_test",
					SafeDownFn: nil,
				},
			},
			wantErr: false,
		},
		{
			name: "success case with some go fn()",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mock.ExpectBegin()
					mock.ExpectExec(`DROP TABLE accounts;`).WillReturnResult(sqlmock.NewResult(0, 0))
					mock.ExpectCommit()

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						LockVersionMock.Return(nil).
						DeleteVersionMock.Return(nil)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					SafeDownFn: func(tx *sql.Tx) error {
						_, err := tx.Exec(`DROP TABLE accounts;`)
						if err != nil {
							return err
						}
						return nil
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &Runner{}
			if err := runner.MigrateDownSafe(tt.args.repo.mRepo, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("MigrateDownSafe() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.args.repo.dbMock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func Test_migrateDown(t *testing.T) {
	type repoMock struct {
		mRepo  repo.MigrationRepo
		dbMock sqlmock.Sqlmock
	}
	type args struct {
		repo repoMock
		m    *Migration
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "db initialization error",
			args: args{
				repo: func() repoMock {
					_, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(nil, errors.New("db not initialized"))

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					DownFn: func(db *sql.DB) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "fn() error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					DownFn: func(db *sql.DB) error {
						return errors.New("some fn() error")
					},
				},
			},
			wantErr: true,
		},
		{
			name: "delete version error",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						DeleteVersionMock.Return(errors.New("some delete version error"))

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					DownFn: func(db *sql.DB) error {
						return nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "success case simple",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						DeleteVersionMock.Return(nil)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					DownFn: func(db *sql.DB) error {
						return nil
					},
				},
			},
			wantErr: false,
		},
		{
			name: "success case no fn()",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New()
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						DeleteVersionMock.Return(nil)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					DownFn:  nil,
				},
			},
			wantErr: false,
		},
		{
			name: "success case with some go fn()",
			args: args{
				repo: func() repoMock {
					db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
					if err != nil {
						t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
					}
					mock.ExpectExec(`DROP TABLE accounts;`).WillReturnResult(sqlmock.NewResult(0, 0))

					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBMock.Return(db, nil).
						DeleteVersionMock.Return(nil)

					return repoMock{mRepo: mRepoMock, dbMock: mock}
				}(),
				m: &Migration{
					Version: "m000000_000000_test",
					Source:  "m000000_000000_test",
					DownFn: func(db *sql.DB) error {
						_, err := db.Exec(`DROP TABLE accounts;`)
						if err != nil {
							return err
						}
						return nil
					},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			runner := &Runner{}
			if err := runner.MigrateDown(tt.args.repo.mRepo, tt.args.m); (err != nil) != tt.wantErr {
				t.Errorf("MigrateDown() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.args.repo.dbMock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
