package migration

import (
	"database/sql"
	"reflect"
	"testing"

	"github.com/gojuno/minimock/v3"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"github.com/tweety53/gomigrate/internal/repo"
)

func TestMigration_run(t *testing.T) {
	type fields struct {
		Source     string
		Registered bool
		SafeUpFn   func(tx *sql.Tx) error
		SafeDownFn func(tx *sql.Tx) error
		UpFn       func(db *sql.DB) error
		DownFn     func(db *sql.DB) error
	}
	type args struct {
		repo      repo.MigrationRepo
		direction Direction
		runner    RunnerInterface
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name:   "success case sql safe up",
			fields: fields{Source: "testdata/runner_test/m000000_000000_safe.sql"},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateUpSafeMock.Return(nil)
				return args{
					repo:      nil,
					direction: migrationDirectionUp,
					runner:    runnerMock,
				}
			}(),
			wantErr: false,
		},
		{
			name:   "runner error sql safe up",
			fields: fields{Source: "testdata/runner_test/m000000_000000_safe.sql"},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateUpSafeMock.Return(errors.New("kek"))
				return args{
					repo:      nil,
					direction: migrationDirectionUp,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name:   "success case sql no tx up",
			fields: fields{Source: "testdata/runner_test/m000000_000000_no_tx.sql"},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateUpMock.Return(nil)
				return args{
					repo:      nil,
					direction: migrationDirectionUp,
					runner:    runnerMock,
				}
			}(),
			wantErr: false,
		},
		{
			name:   "runner error sql no tx up",
			fields: fields{Source: "testdata/runner_test/m000000_000000_no_tx.sql"},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateUpMock.Return(errors.New("kek"))
				return args{
					repo:      nil,
					direction: migrationDirectionUp,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name:   "success case sql safe down",
			fields: fields{Source: "testdata/runner_test/m000000_000000_safe.sql"},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateDownSafeMock.Return(nil)
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: false,
		},
		{
			name:   "runner error sql safe down",
			fields: fields{Source: "testdata/runner_test/m000000_000000_safe.sql"},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateDownSafeMock.Return(errors.New("kek"))
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name:   "success case sql no tx down",
			fields: fields{Source: "testdata/runner_test/m000000_000000_no_tx.sql"},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateDownMock.Return(nil)
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: false,
		},
		{
			name:   "runner error sql no tx down",
			fields: fields{Source: "testdata/runner_test/m000000_000000_no_tx.sql"},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateDownMock.Return(errors.New("kek"))
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name:   "sql file not found error",
			fields: fields{Source: "testdata/runner_test/m000000_000000_some_name.sql"},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc)
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name:   "cannot parse sql error",
			fields: fields{Source: "testdata/runner_test/m000000_000000_bad_content.sql"},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc)
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name: "success case go safe up",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_safe.go",
				Registered: true,
				SafeUpFn: func(tx *sql.Tx) error {
					return nil
				},
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateUpSafeMock.Return(nil)
				return args{
					repo:      nil,
					direction: migrationDirectionUp,
					runner:    runnerMock,
				}
			}(),
			wantErr: false,
		},
		{
			name: "no fn go safe up",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_safe.go",
				Registered: true,
				SafeUpFn:   nil,
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc)
				return args{
					repo:      nil,
					direction: migrationDirectionUp,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name: "runner error go safe up",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_safe.go",
				Registered: true,
				SafeUpFn: func(tx *sql.Tx) error {
					return nil
				},
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateUpSafeMock.Return(errors.New("kek"))
				return args{
					repo:      nil,
					direction: migrationDirectionUp,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name: "success case go no tx up",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_no_tx.go",
				Registered: true,
				UpFn: func(db *sql.DB) error {
					return nil
				},
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateUpMock.Return(nil)
				return args{
					repo:      nil,
					direction: migrationDirectionUp,
					runner:    runnerMock,
				}
			}(),
			wantErr: false,
		},
		{
			name: "no fn go no tx up",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_no_tx.go",
				Registered: true,
				UpFn:       nil,
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc)
				return args{
					repo:      nil,
					direction: migrationDirectionUp,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name: "runner error go no tx up",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_no_tx.go",
				Registered: true,
				UpFn: func(db *sql.DB) error {
					return nil
				},
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateUpMock.Return(errors.New("kek"))
				return args{
					repo:      nil,
					direction: migrationDirectionUp,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name: "success case go safe down",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_safe.go",
				Registered: true,
				SafeDownFn: func(tx *sql.Tx) error {
					return nil
				},
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateDownSafeMock.Return(nil)
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: false,
		},
		{
			name: "no fn go safe down",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_safe.go",
				Registered: true,
				SafeDownFn: nil,
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc)
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name: "runner error go safe down",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_safe.go",
				Registered: true,
				SafeDownFn: func(tx *sql.Tx) error {
					return nil
				},
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateDownSafeMock.Return(errors.New("kek"))
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name: "success case go no tx down",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_no_tx.go",
				Registered: true,
				DownFn: func(db *sql.DB) error {
					return nil
				},
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateDownMock.Return(nil)
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: false,
		},
		{
			name: "no fn go no tx down",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_no_tx.go",
				Registered: true,
				DownFn:     nil,
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc)
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name: "runner error go no tx down",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_no_tx.go",
				Registered: true,
				DownFn: func(db *sql.DB) error {
					return nil
				},
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc).MigrateDownMock.Return(errors.New("kek"))
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name: "not registered go",
			fields: fields{
				Source:     "testdata/runner_test/m000000_000000_no_tx.go",
				Registered: false,
				DownFn: func(db *sql.DB) error {
					return nil
				},
			},
			args: func() args {
				mc := minimock.NewController(t)
				runnerMock := NewRunnerInterfaceMock(mc)
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    runnerMock,
				}
			}(),
			wantErr: true,
		},
		{
			name: "unknown type",
			fields: fields{
				Source: "testdata/runner_test/m000000_000000_txt.txt",
			},
			args: func() args {
				return args{
					repo:      nil,
					direction: migrationDirectionDown,
					runner:    nil,
				}
			}(),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Migration{
				Source:     tt.fields.Source,
				Registered: tt.fields.Registered,
				SafeUpFn:   tt.fields.SafeUpFn,
				SafeDownFn: tt.fields.SafeDownFn,
				UpFn:       tt.fields.UpFn,
				DownFn:     tt.fields.DownFn,
			}
			if err := m.run(tt.args.repo, tt.args.direction, tt.args.runner); (err != nil) != tt.wantErr {
				t.Errorf("run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGetVersionFromFileName(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name:    "wrong file extension",
			args:    args{name: "m000000_000000_kek.php"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "cannot split file by dot",
			args:    args{name: "m000000_000000_kek"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "a lot of dots",
			args:    args{name: "m000000_000000_kek.sql.go.kek"},
			want:    "",
			wantErr: true,
		},
		{
			name:    "success go",
			args:    args{name: "m000000_000000_kek.go"},
			want:    "m000000_000000_kek",
			wantErr: false,
		},
		{
			name:    "success sql",
			args:    args{name: "m000000_000000_kek.sql"},
			want:    "m000000_000000_kek",
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := GetVersionFromFileName(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetVersionFromFileName() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetVersionFromFileName() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestConvert(t *testing.T) {
	type args struct {
		records repo.MigrationRecords
	}
	tests := []struct {
		name string
		args args
		want Migrations
	}{
		{
			name: "OK",
			args: args{records: repo.MigrationRecords{
				&repo.MigrationRecord{
					Version:   "m000000_000000_kek",
					ApplyTime: 12345,
				},
				&repo.MigrationRecord{
					Version:   "m000000_000001_keks",
					ApplyTime: 12345,
				},
				&repo.MigrationRecord{
					Version:   "m000000_000002_kekss",
					ApplyTime: 12345,
				},
			}},
			want: Migrations{
				&Migration{Version: "m000000_000000_kek"},
				&Migration{Version: "m000000_000001_keks"},
				&Migration{Version: "m000000_000002_kekss"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Convert(tt.args.records); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Convert() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMigrations_Reverse(t *testing.T) {
	tests := []struct {
		name string
		ms   Migrations
		want Migrations
	}{
		{
			name: "OK",
			ms: Migrations{
				&Migration{Version: "m000000_000000_kek"},
				&Migration{Version: "m000000_000001_keks"},
				&Migration{Version: "m000000_000002_kekss"},
			},
			want: Migrations{
				&Migration{Version: "m000000_000002_kekss"},
				&Migration{Version: "m000000_000001_keks"},
				&Migration{Version: "m000000_000000_kek"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.ms.Reverse())
		})
	}
}

func TestMigrations_Less(t *testing.T) {
	ms := Migrations{
		&Migration{Version: "m000000_000000_kekss"},
		&Migration{Version: "m000000_000001_keks"},
		&Migration{Version: "m000000_000002_kek"},
	}
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		ms   Migrations
		args args
		want bool
	}{
		{
			name: "less #1",
			ms:   ms,
			args: args{
				i: 0,
				j: 1,
			},
			want: true,
		},
		{
			name: "less #2",
			ms:   ms,
			args: args{
				i: 0,
				j: 2,
			},
			want: true,
		},
		{
			name: "more #1",
			ms:   ms,
			args: args{
				i: 1,
				j: 0,
			},
			want: false,
		},
		{
			name: "more #2",
			ms:   ms,
			args: args{
				i: 2,
				j: 0,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ms.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("Less() = %v, want %v", got, tt.want)
			}
		})
	}
}
