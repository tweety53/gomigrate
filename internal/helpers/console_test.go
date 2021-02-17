package helpers

import "testing"

func Test_processResponse(t *testing.T) {
	type args struct {
		response string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "yes #1",
			args: args{response: "y"},
			want: true,
		},
		{
			name: "yes #2",
			args: args{response: "yes"},
			want: true,
		},
		{
			name: "yes #3",
			args: args{response: "Y"},
			want: true,
		},
		{
			name: "yes #4",
			args: args{response: "YES"},
			want: true,
		},
		{
			name: "no #1",
			args: args{response: "n"},
			want: false,
		},
		{
			name: "no #2",
			args: args{response: "no"},
			want: false,
		},
		{
			name: "no #3",
			args: args{response: "N"},
			want: false,
		},
		{
			name: "no #4",
			args: args{response: "NO"},
			want: false,
		},
		{
			name: "random word lowercase",
			args: args{response: "kek"},
			want: false,
		},
		{
			name: "random word uppercase",
			args: args{response: "KEK"},
			want: false,
		},
		{
			name: "empty",
			args: args{response: ""},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := processResponse(tt.args.response); got != tt.want {
				t.Errorf("processResponse() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChooseLogText(t *testing.T) {
	type args struct {
		n         int
		beforeRun bool
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "one before migrations apply/revert",
			args: args{
				n:         1,
				beforeRun: true,
			},
			want: migrationText,
		},
		{
			name: "few before migrations apply/revert",
			args: args{
				n:         3,
				beforeRun: true,
			},
			want: migrationsText,
		},
		{
			name: "one after migrations apply/revert",
			args: args{
				n:         1,
				beforeRun: false,
			},
			want: migrationWasText,
		},
		{
			name: "few after migrations apply/revert",
			args: args{
				n:         3,
				beforeRun: false,
			},
			want: migrationsWereText,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ChooseLogText(tt.args.n, tt.args.beforeRun); got != tt.want {
				t.Errorf("ChooseLogText() = %v, want %v", got, tt.want)
			}
		})
	}
}
