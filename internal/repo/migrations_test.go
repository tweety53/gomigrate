package repo

import (
	"database/sql"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/internal/sqldialect"
	"log"
	"reflect"
	"testing"
)

func TestMigrationsRepository_GetMigrationsHistory(t *testing.T) {
	type fields struct {
		db      *sql.DB
		dialect sqldialect.SQLDialect
		dbMock  sqlmock.Sqlmock
	}
	type args struct {
		limit int
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    MigrationRecords
		wantErr bool
	}{
		{
			name: "Query() error",
			fields: func() fields {
				db, mock, err := sqlmock.New()
				if err != nil {
					log.Fatal(err)
				}

				mock.ExpectQuery("SELECT version, apply_time FROM some_table ORDER BY apply_time DESC, version DESC;").
					WillReturnError(errors.New("some error"))
				dialect, err := sqldialect.InitDialect("postgres", "some_table")
				if err != nil {
					log.Fatal(err)
				}

				return fields{
					db:      db,
					dialect: dialect,
					dbMock:  mock,
				}
			}(),
			args:    args{limit: 0},
			want:    nil,
			wantErr: true,
		},
		{
			name: "rows error",
			fields: func() fields {
				db, mock, err := sqlmock.New()
				if err != nil {
					log.Fatal(err)
				}

				rows := sqlmock.NewRows([]string{"version", "apply_time"}).
					AddRow("m000000_000000_q", "12345").
					AddRow("m000000_000001_w", "12345").
					RowError(0, errors.New("qwe"))
				mock.ExpectQuery("SELECT version, apply_time FROM some_table ORDER BY apply_time DESC, version DESC;").
					WillReturnRows(rows)

				dialect, err := sqldialect.InitDialect("postgres", "some_table")
				if err != nil {
					log.Fatal(err)
				}

				return fields{
					db:      db,
					dialect: dialect,
					dbMock:  mock,
				}
			}(),
			args:    args{limit: 0},
			want:    nil,
			wantErr: true,
		},
		{
			name: "rows scan error",
			fields: func() fields {
				db, mock, err := sqlmock.New()
				if err != nil {
					log.Fatal(err)
				}

				rows := sqlmock.NewRows([]string{"lol", "kek", "cheburek"}).
					AddRow("1", "m000000_000000_q", "12345").
					AddRow("2", "m000000_000001_w", "12345")
				mock.ExpectQuery("SELECT version, apply_time FROM some_table ORDER BY apply_time DESC, version DESC;").
					WillReturnRows(rows)

				dialect, err := sqldialect.InitDialect("postgres", "some_table")
				if err != nil {
					log.Fatal(err)
				}

				return fields{
					db:      db,
					dialect: dialect,
					dbMock:  mock,
				}
			}(),
			args:    args{limit: 0},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success",
			fields: func() fields {
				db, mock, err := sqlmock.New()
				if err != nil {
					log.Fatal(err)
				}

				rows := sqlmock.NewRows([]string{"version", "apply_time"}).
					AddRow("m000000_000000_q", "12345").
					AddRow("m000000_000001_w", "12345")
				mock.ExpectQuery("SELECT version, apply_time FROM some_table ORDER BY apply_time DESC, version DESC;").
					WillReturnRows(rows)

				dialect, err := sqldialect.InitDialect("postgres", "some_table")
				if err != nil {
					log.Fatal(err)
				}

				return fields{
					db:      db,
					dialect: dialect,
					dbMock:  mock,
				}
			}(),
			args: args{limit: 0},
			want: MigrationRecords{
				&MigrationRecord{
					Version:   "m000000_000000_q",
					ApplyTime: 12345,
				},
				&MigrationRecord{
					Version:   "m000000_000001_w",
					ApplyTime: 12345,
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MigrationsRepository{
				db:      tt.fields.db,
				dialect: tt.fields.dialect,
			}
			got, err := r.GetMigrationsHistory(tt.args.limit)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetMigrationsHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetMigrationsHistory() got = %v, want %v", got, tt.want)
			}
			if err := tt.fields.dbMock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func Test_buildMigrationsHistoryQuery(t *testing.T) {
	type args struct {
		limit int
		query string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "zero",
			args: args{
				limit: 0,
				query: "q;",
			},
			want: "q;",
		},
		{
			name: "one",
			args: args{
				limit: 1,
				query: "q;",
			},
			want: "q LIMIT 1;",
		},
		{
			name: "ten",
			args: args{
				limit: 10,
				query: "q;",
			},
			want: "q LIMIT 10;",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := buildMigrationsHistoryQuery(tt.args.limit, tt.args.query); got != tt.want {
				t.Errorf("buildMigrationsHistoryQuery() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMigrationsRepository_GetDB(t *testing.T) {
	type fields struct {
		db      *sql.DB
		dialect sqldialect.SQLDialect
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "success",
			fields: func() fields {
				db, _, err := sqlmock.New()
				if err != nil {
					log.Fatal(err)
				}

				dialect, err := sqldialect.InitDialect("postgres", "some_table")
				if err != nil {
					log.Fatal(err)
				}
				return fields{
					db:      db,
					dialect: dialect,
				}
			}(),
			wantErr: false,
		},
		{
			name: "not initialized",
			fields: func() fields {
				return fields{
					db:      nil,
					dialect: nil,
				}
			}(),
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MigrationsRepository{
				db:      tt.fields.db,
				dialect: tt.fields.dialect,
			}
			got, err := r.GetDB()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetDB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.fields.db) {
				t.Errorf("GetDB() got = %v, want %v", got, tt.fields.db)
			}
		})
	}
}

func TestMigrationsRepository_InsertVersion(t *testing.T) {
	type fields struct {
		db      *sql.DB
		dialect sqldialect.SQLDialect
		dbMock  sqlmock.Sqlmock
	}
	type args struct {
		v string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: func() fields {
				db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
				if err != nil {
					log.Fatal(err)
				}

				dialect, err := sqldialect.InitDialect("postgres", "some_table")
				if err != nil {
					log.Fatal(err)
				}
				mock.ExpectExec("INSERT INTO some_table (version, apply_time) VALUES ($1, $2);").
					WillReturnResult(sqlmock.NewResult(1, 1))
				return fields{
					db:      db,
					dialect: dialect,
					dbMock:  mock,
				}
			}(),
			args:    args{v: "m000000_000000_test"},
			wantErr: false,
		},
		{
			name: "error",
			fields: func() fields {
				db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
				if err != nil {
					log.Fatal(err)
				}

				dialect, err := sqldialect.InitDialect("postgres", "some_table")
				if err != nil {
					log.Fatal(err)
				}
				mock.ExpectExec("INSERT INTO some_table (version, apply_time) VALUES ($1, $2);").
					WillReturnError(errors.New("some err"))
				return fields{
					db:      db,
					dialect: dialect,
					dbMock:  mock,
				}
			}(),
			args:    args{v: "m000000_000000_test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MigrationsRepository{
				db:      tt.fields.db,
				dialect: tt.fields.dialect,
			}
			if err := r.InsertVersion(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("InsertVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.fields.dbMock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}

func TestMigrationsRepository_DeleteVersion(t *testing.T) {
	type fields struct {
		db      *sql.DB
		dialect sqldialect.SQLDialect
		dbMock  sqlmock.Sqlmock
	}
	type args struct {
		v string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "success",
			fields: func() fields {
				db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
				if err != nil {
					log.Fatal(err)
				}

				dialect, err := sqldialect.InitDialect("postgres", "some_table")
				if err != nil {
					log.Fatal(err)
				}
				mock.ExpectExec("DELETE FROM some_table WHERE version=$1;").
					WillReturnResult(sqlmock.NewResult(1, 1))
				return fields{
					db:      db,
					dialect: dialect,
					dbMock:  mock,
				}
			}(),
			args:    args{v: "m000000_000000_test"},
			wantErr: false,
		},
		{
			name: "error",
			fields: func() fields {
				db, mock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherEqual))
				if err != nil {
					log.Fatal(err)
				}

				dialect, err := sqldialect.InitDialect("postgres", "some_table")
				if err != nil {
					log.Fatal(err)
				}
				mock.ExpectExec("DELETE FROM some_table WHERE version=$1;").
					WillReturnError(errors.New("some err"))
				return fields{
					db:      db,
					dialect: dialect,
					dbMock:  mock,
				}
			}(),
			args:    args{v: "m000000_000000_test"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &MigrationsRepository{
				db:      tt.fields.db,
				dialect: tt.fields.dialect,
			}
			if err := r.DeleteVersion(tt.args.v); (err != nil) != tt.wantErr {
				t.Errorf("InsertVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err := tt.fields.dbMock.ExpectationsWereMet(); err != nil {
				t.Errorf("there were unfulfilled expectations: %s", err)
			}
		})
	}
}
