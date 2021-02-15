package action

import (
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/stretchr/testify/require"
	"github.com/tweety53/gomigrate/internal/helpers"
	"github.com/tweety53/gomigrate/internal/migration"
	"github.com/tweety53/gomigrate/internal/repo"
	"github.com/tweety53/gomigrate/internal/service"
)

func TestDownActionParams_ValidateAndFill(t *testing.T) {
	type args struct {
		args []string
	}
	tests := []struct {
		name           string
		args           args
		expectedParams *DownActionParams
		wantErr        bool
	}{
		{
			name:           "no args",
			args:           args{args: []string{}},
			expectedParams: &DownActionParams{limit: 1},
			wantErr:        false,
		},
		{
			name:           "handle all limit",
			args:           args{args: []string{helpers.LimitAll}},
			expectedParams: &DownActionParams{limit: 0},
			wantErr:        false,
		},
		{
			name:           "handle some limit",
			args:           args{args: []string{"3"}},
			expectedParams: &DownActionParams{limit: 3},
			wantErr:        false,
		},
		{
			name:           "non-numeric limit",
			args:           args{args: []string{"kek"}},
			expectedParams: &DownActionParams{},
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &DownActionParams{}
			err := p.ValidateAndFill(tt.args.args)
			require.Equal(t, tt.wantErr, err != nil)
			require.Equal(t, tt.expectedParams, p)
		})
	}
}

func TestDownAction_Run(t *testing.T) {
	type fields struct {
		svc *service.MigrationService
	}
	type args struct {
		params interface{}
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:    "invalid action params type passed",
			fields:  fields{},
			args:    args{params: struct{}{}},
			wantErr: true,
		},
		{
			name: "no migrations done before",
			fields: fields{
				svc: func() *service.MigrationService {
					mc := minimock.NewController(t)
					mRepoMock := repo.NewMigrationRepoMock(mc).
						GetMigrationsHistoryMock.Return(repo.MigrationRecords{}, nil)
					return service.NewMigrationService(
						nil,
						mRepoMock,
						nil,
						&migration.MigrationsCollector{},
						"")
				}(),
			},
			args:    args{params: &DownActionParams{limit: 0}},
			wantErr: false,
		},
		//{
		//	name: "success case",
		//	fields: fields{
		//		svc: func() *service.MigrationService {
		//			mc := minimock.NewController(t)
		//			mRepoMock := repo.NewMigrationRepoMock(mc).
		//				GetMigrationsHistoryMock.Return(repo.MigrationRecords{
		//				&repo.MigrationRecord{
		//					Version:   "m200101_000000_test",
		//					ApplyTime: 1,
		//				},
		//				&repo.MigrationRecord{
		//					Version:   "m200102_000000_test",
		//					ApplyTime: 1,
		//				},
		//			}, nil)
		//			cMock := migration.NewMigrationsCollectorInterfaceMock(mc).
		//				CollectMigrationsMock.Return(migration.Migrations{
		//				&migration.Migration{
		//					Version: "m200101_000000_test",
		//					DownFn:  nil,
		//				},
		//				&migration.Migration{
		//					Version: "m200102_000000_test",
		//					DownFn:  nil,
		//				},
		//			}, nil)
		//
		//			return service.NewMigrationService(
		//				nil,
		//				mRepoMock,
		//				repo.NewDbOperationRepoMock(mc),
		//				cMock,
		//				"")
		//		}(),
		//	},
		//	args:    args{params: &DownActionParams{limit: 0}},
		//	wantErr: false,
		//},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &DownAction{
				svc: tt.fields.svc,
			}
			if err := a.Run(tt.args.params); (err != nil) != tt.wantErr {
				t.Errorf("Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
