package service

import (
	"database/sql"
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
		DbOperationRepo     *repo.DbOperationRepoMock
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
