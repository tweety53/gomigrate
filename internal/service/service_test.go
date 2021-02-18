package service

import (
	"database/sql"
	"github.com/pkg/errors"
	"reflect"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/repo"
)

func TestMigrationService_GetNewMigrations(t *testing.T) {
	type fields struct {
		Db                  *sql.DB
		MigrationsRepo      *repo.MigrationRepoMock
		DbOperationRepo     *repo.DBOperationRepoMock
		MigrationsPath      string
		MigrationsCollector *migration.MigrationsCollectorInterfaceMock
	}
	tests := []struct {
		name    string
		fields  fields
		want    migration.Migrations
		wantErr bool
	}{
		{
			name: "create db error",
			fields: fields{
				Db: nil,
				MigrationsRepo: func() *repo.MigrationRepoMock {
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBVersionMock.Return("", errors.New("some error"))

					return mRepoMock
				}(),
				DbOperationRepo:     nil,
				MigrationsPath:      "",
				MigrationsCollector: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "migrations history fetch error",
			fields: fields{
				Db: nil,
				MigrationsRepo: func() *repo.MigrationRepoMock {
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBVersionMock.Return("", nil).
						GetMigrationsHistoryMock.Return(nil, errors.New("some error"))

					return mRepoMock
				}(),
				DbOperationRepo:     nil,
				MigrationsPath:      "",
				MigrationsCollector: nil,
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "base migration skipped",
			fields: fields{
				Db: nil,
				MigrationsRepo: func() *repo.MigrationRepoMock {
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBVersionMock.Return("", nil).
						GetMigrationsHistoryMock.Return(repo.MigrationRecords{
						&repo.MigrationRecord{
							Version:   "m000000_000000_base",
							ApplyTime: 1,
						},
						&repo.MigrationRecord{
							Version:   "m200101_000000_test",
							ApplyTime: 1,
						},
					}, nil)
					return mRepoMock
				}(),
				DbOperationRepo: nil,
				MigrationsPath:  "",
				MigrationsCollector: func() *migration.MigrationsCollectorInterfaceMock {
					mc := minimock.NewController(t)
					cMock := migration.NewMigrationsCollectorInterfaceMock(mc).
						CollectMigrationsMock.Return(migration.Migrations{
						&migration.Migration{
							Version: "m200101_000000_test",
						},
					}, nil)

					return cMock
				}(),
			},
			want:    migration.Migrations{},
			wantErr: false,
		},
		{
			name: "collector error",
			fields: fields{
				Db: nil,
				MigrationsRepo: func() *repo.MigrationRepoMock {
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBVersionMock.Return("", nil).
						GetMigrationsHistoryMock.Return(repo.MigrationRecords{
						&repo.MigrationRecord{
							Version:   "m200101_000000_test",
							ApplyTime: 1,
						},
					}, nil)
					return mRepoMock
				}(),
				DbOperationRepo: nil,
				MigrationsPath:  "",
				MigrationsCollector: func() *migration.MigrationsCollectorInterfaceMock {
					mc := minimock.NewController(t)
					cMock := migration.NewMigrationsCollectorInterfaceMock(mc).
						CollectMigrationsMock.Return(nil, errors.New("some error"))

					return cMock
				}(),
			},
			want:    nil,
			wantErr: true,
		},
		{
			name: "success case",
			fields: fields{
				Db: nil,
				MigrationsRepo: func() *repo.MigrationRepoMock {
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetDBVersionMock.Return("", nil).
						GetMigrationsHistoryMock.Return(repo.MigrationRecords{
						&repo.MigrationRecord{
							Version:   "m200101_000000_test",
							ApplyTime: 1,
						},
					}, nil)
					return mRepoMock
				}(),
				DbOperationRepo: nil,
				MigrationsPath:  "",
				MigrationsCollector: func() *migration.MigrationsCollectorInterfaceMock {
					mc := minimock.NewController(t)
					cMock := migration.NewMigrationsCollectorInterfaceMock(mc).
						CollectMigrationsMock.Return(migration.Migrations{
						&migration.Migration{
							Version: "m200101_000000_test",
						},
						&migration.Migration{
							Version: "m200101_000001_test1",
						},
					}, nil)

					return cMock
				}(),
			},
			want: migration.Migrations{
				&migration.Migration{
					Version: "m200101_000001_test1",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MigrationService{
				DB:                  tt.fields.Db,
				MigrationsRepo:      tt.fields.MigrationsRepo,
				DBOperationRepo:     tt.fields.DbOperationRepo,
				MigrationsPath:      tt.fields.MigrationsPath,
				MigrationsCollector: tt.fields.MigrationsCollector,
			}
			got, err := s.GetNewMigrations()
			if (err != nil) != tt.wantErr {
				t.Errorf("GetNewMigrations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetNewMigrations() got = %v, want %v", got, tt.want)
			}
		})
	}
}
