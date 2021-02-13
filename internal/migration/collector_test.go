package migration

import (
	"reflect"
	"testing"
)

func TestCollectMigrations(t *testing.T) {
	type args struct {
		current int
		target  int
		dirpath string
	}
	tests := []struct {
		name    string
		args    args
		want    Migrations
		wantErr bool
	}{
		{
			name: "get all migrations",
			args: args{
				dirpath: "testdata/migrations_test/",
			},
			want: Migrations{
				&Migration{
					Version:  "m200101_000000_add_accounts_table",
					Next:     "m200101_000001_add_zulul_table",
					Previous: "",
					Source:   "testdata/migrations_test/m200101_000000_add_accounts_table.go",
				},
				&Migration{
					Version:  "m200101_000001_add_zulul_table",
					Next:     "m200101_000002_add_another_accounts_table",
					Previous: "m200101_000000_add_accounts_table",
					Source:   "testdata/migrations_test/m200101_000001_add_zulul_table.sql",
				},
				&Migration{
					Version:  "m200101_000002_add_another_accounts_table",
					Next:     "",
					Previous: "m200101_000001_add_zulul_table",
					Source:   "testdata/migrations_test/m200101_000002_add_another_accounts_table.go",
				},
			},
			wantErr: false,
		},
		{
			name: "migrations dir not exists",
			args: args{
				dirpath: "testdata/iamnotexists/",
			},
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CollectMigrations(tt.args.dirpath, tt.args.current, tt.args.target)
			if (err != nil) != tt.wantErr {
				t.Errorf("CollectMigrations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("CollectMigrations() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_versionInRange(t *testing.T) {
	type args struct {
		v       int
		current int
		target  int
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "get all migrations case",
			args: args{},
			want: true,
		},
		{
			name: "action up not in range",
			args: args{
				v:       1,
				current: 2,
				target:  3,
			},
			want: false,
		},
		{
			name: "action up not in range (target == current)",
			args: args{
				v:       1,
				current: 2,
				target:  2,
			},
			want: false,
		},
		{
			name: "action up in range (v != current)",
			args: args{
				v:       3,
				current: 2,
				target:  4,
			},
			want: true,
		},
		{
			name: "action up in range (v == current)",
			args: args{
				v:       2,
				current: 2,
				target:  3,
			},
			want: true,
		},
		{
			name: "action down not in range",
			args: args{
				v:       3,
				current: 2,
				target:  1,
			},
			want: false,
		},
		{
			name: "action down in range (v != current)",
			args: args{
				v:       1,
				current: 2,
				target:  1,
			},
			want: true,
		},
		{
			name: "action down in range (v == current)",
			args: args{
				v:       2,
				current: 2,
				target:  1,
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := versionInRange(tt.args.v, tt.args.current, tt.args.target); got != tt.want {
				t.Errorf("versionInRange() = %v, want %v", got, tt.want)
			}
		})
	}
}
