package sqldialect

import (
	"reflect"
	"testing"
)

func TestInitDialect(t *testing.T) {
	type args struct {
		v              string
		migrationTable string
	}
	tests := []struct {
		name    string
		args    args
		want    SQLDialect
		wantErr bool
	}{
		{
			name: "success init",
			args: args{
				v:              "postgres",
				migrationTable: "some_table_name",
			},
			want:    &PostgresDialect{migrationTable: "some_table_name"},
			wantErr: false,
		},
		{
			name: "unknown dialect",
			args: args{
				v:              "mysql",
				migrationTable: "some_table_name",
			},
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := InitDialect(tt.args.v, tt.args.migrationTable)
			if (err != nil) != tt.wantErr {
				t.Errorf("InitDialect() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("InitDialect() got = %v, want %v", got, tt.want)
			}
		})
	}
}
