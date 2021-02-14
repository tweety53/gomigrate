package errors

import (
	"github.com/pkg/errors"
	"github.com/tweety53/gomigrate/pkg/exit_code"
	"testing"
)

func TestErrorExitCode(t *testing.T) {
	type args struct {
		err error
	}
	tests := []struct {
		name string
		args args
		want exit_code.ExitCode
	}{
		{
			name: "error with undefined exit code",
			args: args{
				err: ErrInvalidActionParamsType,
			},
			want: exit_code.ExitCodeUnspecified,
		},
		{
			name: "no error",
			args: args{
				err: nil,
			},
			want: exit_code.ExitCodeOK,
		},
		{
			name: "GoMigrateError with defined exit code",
			args: args{
				err: &GoMigrateError{
					Err:      errors.New("test err"),
					ExitCode: exit_code.ExitCodeIOErr,
				},
			},
			want: exit_code.ExitCodeIOErr,
		},
		{
			name: "nested GoMigrateError with defined exit code",
			args: args{
				err: &GoMigrateError{
					Err: &GoMigrateError{
						Err:      errors.New("test err"),
						ExitCode: exit_code.ExitCodeIOErr,
					},
					ExitCode: exit_code.ExitCodeOK,
				},
			},
			want: exit_code.ExitCodeIOErr,
		},
		{
			name: "nested GoMigrateError with undefined exit code",
			args: args{
				err: &GoMigrateError{
					Err: &GoMigrateError{
						Err: errors.New("test err"),
					},
					ExitCode: exit_code.ExitCodeOK,
				},
			},
			want: exit_code.ExitCodeUnspecified,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ErrorExitCode(tt.args.err); got != tt.want {
				t.Errorf("ErrorExitCode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestError_Error(t *testing.T) {
	type fields struct {
		Err      error
		ExitCode exit_code.ExitCode
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "simple err case",
			fields: fields{
				Err: errors.New("test error"),
			},
			want: "test error",
		},
		{
			name: "nil case",
			fields: fields{
				Err: nil,
			},
			want: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &GoMigrateError{
				Err:      tt.fields.Err,
				ExitCode: tt.fields.ExitCode,
			}
			if got := e.Error(); got != tt.want {
				t.Errorf("GoMigrateError() = %v, want %v", got, tt.want)
			}
		})
	}
}
